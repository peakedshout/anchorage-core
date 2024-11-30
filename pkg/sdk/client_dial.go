package sdk

import (
	"context"
	"errors"
	"fmt"
	"github.com/peakedshout/anchorage-core/pkg/comm"
	"github.com/peakedshout/anchorage-core/pkg/sdk/plugin"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"github.com/peakedshout/go-pandorasbox/protocol/cfcprotocol"
	"github.com/peakedshout/go-pandorasbox/tool/dcopy"
	"github.com/peakedshout/go-pandorasbox/xnet"
	"github.com/peakedshout/go-pandorasbox/xnet/fasttool"
	"github.com/peakedshout/go-pandorasbox/xnet/multiplex"
	"io"
	"net"
	"sync"
)

func (sm *sdkManager) AddDial(id string, cfg *DialConfig) (string, error) {
	var sid string
	return sid, sm.getClient(id, func(cs *clientSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-dialSdk:", "add dial", id, sid)
			} else {
				sm.logger.Warn("clientSdk-dialSdk:", "add dial", id, sid, "err:", err.Error())
			}
		}()
		cs.mux.Lock()
		defer cs.mux.Unlock()
		sdk, err := cs.newDialSdk(cfg)
		if err != nil {
			return err
		}
		cs.config.Dial = append(cs.config.Dial, cfg)
		err = sm.save()
		if err != nil {
			cs.config.Dial = cs.config.Dial[:len(cs.config.Dial)-1]
			_ = sdk.stop()
			return err
		}
		cs.dl = append(cs.dl, sdk)
		sid = sdk.GetId()
		return nil
	})
}

func (sm *sdkManager) DelDial(id string, sid string) error {
	return sm.getClient(id, func(cs *clientSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-dialSdk:", "del dial", id, sid)
			} else {
				sm.logger.Warn("clientSdk-dialSdk:", "del dial", id, sid, "err:", err.Error())
			}
		}()
		cs.mux.Lock()
		defer cs.mux.Unlock()
		sub := findIndex(cs.dl, sid)
		if sub < 0 || sub >= len(cs.dl) {
			return errors.New("invalid sid")
		}
		sdk := cs.dl[sub]
		cs.dl = append(cs.dl[:sub], cs.dl[sub+1:]...)
		cs.config.Dial = append(cs.config.Dial[:sub], cs.config.Dial[sub+1:]...)
		_ = sdk.stop()
		return sm.save()
	})
}

func (sm *sdkManager) StartDial(id string, sid string) error {
	return sm.getDial(id, sid, func(sdk *dialSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-dialSdk:", "start dial", id, sid)
			} else {
				sm.logger.Warn("clientSdk-dialSdk:", "start dial", id, sid, "err:", err.Error())
				_ = sdk.stop()
			}
		}()
		return sdk.start()
	})
}

func (sm *sdkManager) ReloadDial(id string, sid string) error {
	return sm.getDial(id, sid, func(sdk *dialSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-dialSdk:", "reload dial", id, sid)
			} else {
				sm.logger.Warn("clientSdk-dialSdk:", "reload dial", id, sid, "err:", err.Error())
				_ = sdk.stop()
			}
		}()
		return sdk.reload()
	})
}

func (sm *sdkManager) StopDial(id string, sid string) error {
	return sm.getDial(id, sid, func(sdk *dialSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-dialSdk:", "stop dial", id, sid)
			} else {
				sm.logger.Warn("clientSdk-dialSdk:", "stop dial", id, sid, "err:", err.Error())
			}
		}()
		return sdk.stop()
	})
}

func (sm *sdkManager) UpdateDial(id string, sid string, fn func(cfg *DialConfig)) error {
	return sm.getDial(id, sid, func(sdk *dialSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-dialSdk:", "update dial", id, sid)
			} else {
				sm.logger.Warn("clientSdk-dialSdk:", "update dial", id, sid, "err:", err.Error())
				_ = sdk.stop()
			}
		}()
		return sdk.update(fn)
	})
}

func (sm *sdkManager) GetConfigDial(id string, sid string) (*DialConfig, error) {
	var cfg *DialConfig
	err := sm.getDial(id, sid, func(sdk *dialSdk) error {
		cfg = sdk.getConfig()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (sm *sdkManager) getDial(id string, sid string, fn func(sdk *dialSdk) error) error {
	return sm.getClient(id, func(cs *clientSdk) error {
		cs.mux.Lock()
		defer cs.mux.Unlock()
		sub := findIndex(cs.dl, sid)
		if sub < 0 || sub >= len(cs.dl) {
			return errors.New("invalid sid")
		}
		sdk := cs.dl[sub]
		return fn(sdk)
	})
}

func (cs *clientSdk) newDialSdk(config *DialConfig) (*dialSdk, error) {
	ds := &dialSdk{
		id:     newSdkId(IdPrefixDial),
		cs:     cs,
		mux:    sync.Mutex{},
		ln:     nil,
		config: config,
		status: false,
	}
	if cs.getStatus() && ds.config.Enable {
		err := ds.reload()
		if err != nil {
			return nil, err
		}
	}
	return ds, nil
}

type dialSdk struct {
	id     string
	cs     *clientSdk
	mux    sync.Mutex
	ln     net.Listener
	config *DialConfig
	status bool
}

func (ds *dialSdk) GetId() string {
	return ds.id
}

func (ds *dialSdk) run(lock, reload bool) (err error) {
	if lock {
		ds.mux.Lock()
		defer ds.mux.Unlock()
	}
	if ds.ln != nil {
		if reload {
			_ = ds.ln.Close()
		} else {
			return nil
		}
	}
	if ds.cs.client == nil {
		return errors.New("client not running")
	}
	ctx, cl := context.WithCancel(ds.cs.client.Context())
	defer func() {
		if err != nil {
			cl()
		}
	}()
	if ds.config.InNetwork == nil {
		return errors.New("nil in network")
	}
	lc, err := xnet.GetBaseStreamListenerConfig(ds.config.InNetwork.Network)
	if err != nil {
		return err
	}
	ln, err := lc.ListenContext(ctx, ds.config.InNetwork.Network, ds.config.InNetwork.Address)
	if err != nil {
		return err
	}
	ctxtool.GWaitFunc(ctx, func() {
		_ = ln.Close()
	})

	// plugin
	var plugins []plugin.DialPlugin
	if ds.config.Plugin != "" {
		ds.cs.pmMux.Lock()
		ps, ok := ds.cs.pm[ds.config.Plugin]
		if !ok || ps.a[PluginTypeDial] == nil {
			ds.cs.pmMux.Unlock()
			err = fmt.Errorf("not found dial plugin: %s", ds.config.Plugin)
			return err
		}
		ps.mux.Lock()
		plugins = ps.a[PluginTypeDial].([]plugin.DialPlugin)
		ps.mux.Unlock()
		ds.cs.pmMux.Unlock()
	} else if ds.config.OutNetwork == nil {
		return errors.New("nil out network")
	}

	//out
	out := dcopy.CopyT(ds.config.OutNetwork)

	if len(plugins) == 0 && out == nil {
		return errors.New("nil dst network and address")
	}

	linkReq := comm.LinkRequest{
		Node:        dcopy.CopyT(ds.config.Node),
		Link:        ds.config.Link,
		PP2PNetwork: ds.config.PP2PNetwork,
		SwitchUP2P:  ds.config.SwitchUP2P,
		SwitchTP2P:  ds.config.SwitchTP2P,
		ForceP2P:    ds.config.ForceP2P,
		Auth:        dcopy.CopyT(ds.config.Auth),
	}

	//multi
	var toDialer func(ctx context.Context) (net.Conn, error)
	var fn func()
	if ds.config.Multi < 0 {
		toDialer = func(ctx context.Context) (net.Conn, error) {
			return ds.cs.client.Dial(ctx, linkReq)
		}
		fn = func() {}
	} else {
		dFunc := func(nctx context.Context) (io.ReadWriteCloser, error) {
			return ds.cs.client.Dial(ctx, linkReq)
		}
		multi := multiplex.NewMultiplex(ctx, 0, 0, 0)
		m := ds.config.Multi
		toDialer = func(ctx context.Context) (net.Conn, error) {
			sess, err := multi.Dial(ctx, dFunc, uint32(m))
			if err != nil {
				return nil, err
			}
			return fasttool.NewFakeConn(sess), nil
		}
		if ds.config.MultiIdle > 0 {
			go multi.DialIDle(dFunc, ds.config.MultiIdle, 0)
		}
		fn = multi.Stop
	}

	go func() {
		defer func() {
			fn()
			cl()
			_ = ln.Close()
			ds.mux.Lock()
			defer ds.mux.Unlock()
			if ds.ln == ln {
				ds.ln = nil
				ds.status = false
			}
		}()
		ds.handle(ctx, ln, out, toDialer, plugins)
	}()
	ds.ln = ln
	ds.status = true
	return nil
}

func (ds *dialSdk) handle(ctx context.Context, ln net.Listener, out *NetworkConfig, toDialer func(ctx context.Context) (net.Conn, error), plugins []plugin.DialPlugin) {
	defer ln.Close()
	tdFunc := func(ctx context.Context, network string, address string) (net.Conn, error) {
		conn, err := toDialer(ctx)
		if err != nil {
			return nil, err
		}
		err = cfcprotocol.CFCPlaintext.Encode(conn, [2]string{network, address})
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		return conn, nil
	}
	var p plugin.DialPlugin
	if len(plugins) != 0 {
		var err error
		for i := len(plugins) - 1; i >= 0; i-- {
			ctx, ln, tdFunc, err = plugins[i].DialUpgrade(ctx, ln, tdFunc)
			if err != nil {
				return
			}
		}
		p = plugins[len(plugins)-1]
	}
	if out != nil {
		for {
			src, err := ln.Accept()
			if err != nil {
				return
			}
			go func(src net.Conn) {
				defer src.Close()
				conn, err := tdFunc(ctx, out.Network, out.Address)
				if err != nil {
					return
				}
				defer conn.Close()
				go io.Copy(conn, src)
				_, _ = io.Copy(src, conn)
			}(src)
		}
	} else {
		if p != nil {
			_ = p.DialServe(ctx, ln, tdFunc)
		}
	}
}

func (ds *dialSdk) update(fn func(cfg *DialConfig)) error {
	ds.mux.Lock()
	defer ds.mux.Unlock()
	fn(ds.config)
	err := ds.cs.sm.save()
	if err != nil {
		return err
	}
	if !ds.status {
		return nil
	}
	return ds.run(false, true)
}

func (ds *dialSdk) notLock(r bool) error {
	return ds.run(false, r)
}

func (ds *dialSdk) reload() error {
	return ds.run(true, true)
}

func (ds *dialSdk) start() error {
	return ds.run(true, false)
}

func (ds *dialSdk) stop() error {
	ds.mux.Lock()
	defer ds.mux.Unlock()
	if ds.ln == nil {
		return errors.New("no running")
	}
	_ = ds.ln.Close()
	ds.ln = nil
	ds.status = false
	return nil
}

func (ds *dialSdk) getConfig() *DialConfig {
	ds.mux.Lock()
	defer ds.mux.Unlock()
	return dcopy.CopyT(ds.config)
}
