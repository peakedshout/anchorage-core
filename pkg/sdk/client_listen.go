package sdk

import (
	"context"
	"errors"
	"github.com/peakedshout/anchorage-core/pkg/comm"
	"github.com/peakedshout/anchorage-core/pkg/config"
	"github.com/peakedshout/go-pandorasbox/protocol/cfcprotocol"
	"github.com/peakedshout/go-pandorasbox/tool/dcopy"
	"github.com/peakedshout/go-pandorasbox/xnet/multiplex"
	"io"
	"net"
	"regexp"
	"sync"
)

func (sm *sdkManager) AddListen(id string, cfg *ListenConfig) (string, error) {
	var sid string
	return sid, sm.getClient(id, func(cs *clientSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-listenSdk:", "add listen", id, sid)
			} else {
				sm.logger.Warn("clientSdk-listenSdk:", "add listen", id, sid, "err:", err.Error())
			}
		}()
		cs.mux.Lock()
		defer cs.mux.Unlock()
		sdk, err := cs.newListenSdk(cfg)
		if err != nil {
			return err
		}
		cs.config.Listen = append(cs.config.Listen, cfg)
		err = sm.save()
		if err != nil {
			cs.config.Listen = cs.config.Listen[:len(cs.config.Listen)-1]
			_ = sdk.stop()
			return err
		}
		cs.ll = append(cs.ll, sdk)
		sid = sdk.GetId()
		return nil
	})
}

func (sm *sdkManager) DelListen(id string, sid string) error {
	return sm.getClient(id, func(cs *clientSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-listenSdk:", "del listen", id, sid)
			} else {
				sm.logger.Warn("clientSdk-listenSdk:", "del listen", id, sid, "err:", err.Error())
			}
		}()
		cs.mux.Lock()
		defer cs.mux.Unlock()
		sub := findIndex(cs.ll, sid)
		if sub < 0 || sub >= len(cs.ll) {
			return errors.New("invalid sid")
		}
		sdk := cs.ll[sub]
		cs.ll = append(cs.ll[:sub], cs.ll[sub+1:]...)
		cs.config.Listen = append(cs.config.Listen[:sub], cs.config.Listen[sub+1:]...)
		_ = sdk.stop()
		return sm.save()
	})
}

func (sm *sdkManager) StartListen(id string, sid string) error {
	return sm.getListen(id, sid, func(sdk *listenSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-listenSdk:", "start listen", id, sid)
			} else {
				sm.logger.Warn("clientSdk-listenSdk:", "start listen", id, sid, "err:", err.Error())
				_ = sdk.stop()
			}
		}()
		return sdk.start()
	})
}

func (sm *sdkManager) ReloadListen(id string, sid string) error {
	return sm.getListen(id, sid, func(sdk *listenSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-dialSdk:", "reload listen", id, sid)
			} else {
				sm.logger.Warn("clientSdk-dialSdk:", "reload listen", id, sid, "err:", err.Error())
				_ = sdk.stop()
			}
		}()
		return sdk.reload()
	})
}

func (sm *sdkManager) StopListen(id string, sid string) error {
	return sm.getListen(id, sid, func(sdk *listenSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-dialSdk:", "stop listen", id, sid)
			} else {
				sm.logger.Warn("clientSdk-dialSdk:", "stop listen", id, sid, "err:", err.Error())
			}
		}()
		return sdk.stop()
	})
}

func (sm *sdkManager) UpdateListen(id string, sid string, fn func(cfg *ListenConfig)) error {
	return sm.getListen(id, sid, func(sdk *listenSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-dialSdk:", "update listen", id, sid)
			} else {
				sm.logger.Warn("clientSdk-dialSdk:", "update listen", id, sid, "err:", err.Error())
				_ = sdk.stop()
			}
		}()
		return sdk.update(fn)
	})
}

func (sm *sdkManager) GetConfigListen(id string, sid string) (*ListenConfig, error) {
	var cfg *ListenConfig
	err := sm.getListen(id, sid, func(sdk *listenSdk) error {
		cfg = sdk.getConfig()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (sm *sdkManager) getListen(id string, sid string, fn func(sdk *listenSdk) error) error {
	return sm.getClient(id, func(cs *clientSdk) error {
		cs.mux.Lock()
		defer cs.mux.Unlock()
		sub := findIndex(cs.ll, sid)
		if sub < 0 || sub >= len(cs.ll) {
			return errors.New("invalid sid")
		}
		sdk := cs.ll[sub]
		return fn(sdk)
	})
}

func (cs *clientSdk) newListenSdk(config *ListenConfig) (*listenSdk, error) {
	ls := &listenSdk{
		id:     newSdkId(IdPrefixListen),
		cs:     cs,
		mux:    sync.Mutex{},
		config: config,
		status: false,
	}
	if cs.getStatus() && ls.config.Enable {
		err := ls.reload()
		if err != nil {
			return nil, err
		}
	}
	return ls, nil
}

type listenSdk struct {
	id     string
	cs     *clientSdk
	mux    sync.Mutex
	ctx    context.Context
	cl     context.CancelFunc
	config *ListenConfig
	status bool
}

func (ls *listenSdk) GetId() string {
	return ls.id
}

func (ls *listenSdk) run(lock, reload bool) error {
	if lock {
		ls.mux.Lock()
		defer ls.mux.Unlock()
	}
	if ls.ctx != nil {
		if reload {
			ls.cl()
		} else {
			return nil
		}
	}
	if ls.cs.client == nil {
		return errors.New("client not running")
	}
	ctx, cl := context.WithCancel(ls.cs.client.Context())
	rln := ls.cs.client.Listen(ctx, comm.RegisterListenerInfo{
		Name:  ls.config.Name,
		Notes: ls.config.Notes,
		Auth:  dcopy.CopyT[*config.AuthInfo](ls.config.Auth),
		Settings: comm.Settings{
			SwitchHide: ls.config.SwitchHide,
			SwitchLink: ls.config.SwitchLink,
			SwitchUP2P: ls.config.SwitchUP2P,
			SwitchTP2P: ls.config.SwitchTP2P,
		},
	})
	var afn func() (io.ReadWriteCloser, error)
	var fn func()
	if ls.config.Multi {
		multi := multiplex.NewMultiplex(ctx, 0, 0, 0)
		go func() {
			defer rln.Close()
			defer multi.Stop()
			_ = multi.Listen(func(ctx context.Context) (io.ReadWriteCloser, error) {
				return rln.Accept()
			})
		}()
		afn = multi.Accept
		fn = multi.Stop
	} else {
		afn = func() (io.ReadWriteCloser, error) {
			return rln.Accept()
		}
		fn = func() {}
	}
	out := dcopy.CopyT(ls.config.OutNetwork)

	go func() {
		defer func() {
			cl()
			fn()
			_ = rln.Close()
			ls.mux.Lock()
			defer ls.mux.Unlock()
			if ls.ctx == ctx {
				ls.ctx = nil
				ls.cl = nil
				ls.status = false
			}
		}()
		for {
			conn, err := afn()
			if err != nil {
				return
			}
			go ls.handle(ctx, conn, out)
		}
	}()
	ls.ctx, ls.cl = ctx, cl
	ls.status = true
	return nil
}

func (ls *listenSdk) handle(ctx context.Context, rwc io.ReadWriteCloser, out *NetworkConfig) {
	defer rwc.Close()
	var sl [2]string // nk na
	err := cfcprotocol.CFCPlaintext.Decode(rwc, &sl)
	if err != nil {
		return
	}
	if out != nil {
		okn, _ := regexp.MatchString(out.Network, sl[0])
		oka, _ := regexp.MatchString(out.Address, sl[1])
		if !okn || !oka {
			return
		}
	}
	dr := new(net.Dialer)
	conn, err := dr.DialContext(ctx, sl[0], sl[1])
	if err != nil {
		return
	}
	defer conn.Close()
	go io.Copy(conn, rwc)
	_, _ = io.Copy(rwc, conn)
}

func (ls *listenSdk) update(fn func(cfg *ListenConfig)) error {
	ls.mux.Lock()
	defer ls.mux.Unlock()
	fn(ls.config)
	err := ls.cs.sm.save()
	if err != nil {
		return err
	}
	if !ls.status {
		return nil
	}
	return ls.run(false, true)
}

func (ls *listenSdk) notLock(r bool) error {
	return ls.run(false, r)
}

func (ls *listenSdk) reload() error {
	return ls.run(true, true)
}

func (ls *listenSdk) start() error {
	return ls.run(true, false)
}

func (ls *listenSdk) stop() error {
	ls.mux.Lock()
	defer ls.mux.Unlock()
	if ls.ctx == nil {
		return errors.New("no running")
	}
	ls.cl()
	ls.ctx = nil
	ls.cl = nil
	ls.status = false
	return nil
}

func (ls *listenSdk) getConfig() *ListenConfig {
	ls.mux.Lock()
	defer ls.mux.Unlock()
	return dcopy.CopyT(ls.config)
}
