package sdk

import (
	"context"
	"errors"
	"fmt"
	"github.com/peakedshout/anchorage-core/pkg/client"
	"github.com/peakedshout/anchorage-core/pkg/sdk/plugin"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"github.com/peakedshout/go-pandorasbox/tool/dcopy"
	"github.com/peakedshout/go-pandorasbox/xnet"
	"io"
	"net"
	"sync"
)

func (sm *sdkManager) AddProxy(id string, cfg *ProxyConfig) (string, error) {
	var sid string
	return sid, sm.getClient(id, func(cs *clientSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-listenSdk:", "add proxy", id, sid)
			} else {
				sm.logger.Warn("clientSdk-listenSdk:", "add proxy", id, sid, "err:", err.Error())
			}
		}()
		cs.mux.Lock()
		defer cs.mux.Unlock()
		sdk, err := cs.newProxySdk(cfg)
		if err != nil {
			return err
		}
		cs.config.Proxy = append(cs.config.Proxy, cfg)
		err = sm.save()
		if err != nil {
			cs.config.Proxy = cs.config.Proxy[:len(cs.config.Proxy)-1]
			_ = sdk.stop()
			return err
		}
		cs.pl = append(cs.pl, sdk)
		sid = sdk.GetId()
		return nil
	})
}

func (sm *sdkManager) DelProxy(id string, sid string) error {
	return sm.getClient(id, func(cs *clientSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-listenSdk:", "del proxy", id, sid)
			} else {
				sm.logger.Warn("clientSdk-listenSdk:", "del proxy", id, sid, "err:", err.Error())
			}
		}()
		cs.mux.Lock()
		defer cs.mux.Unlock()
		sub := findIndex(cs.pl, sid)
		if sub < 0 || sub >= len(cs.pl) {
			return errors.New("invalid sid")
		}
		sdk := cs.pl[sub]
		cs.pl = append(cs.pl[:sub], cs.pl[sub+1:]...)
		cs.config.Proxy = append(cs.config.Proxy[:sub], cs.config.Proxy[sub+1:]...)
		_ = sdk.stop()
		return sm.save()
	})
}

func (sm *sdkManager) StartProxy(id string, sid string) error {
	return sm.getProxy(id, sid, func(sdk *proxySdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-listenSdk:", "start proxy", id, sid)
			} else {
				sm.logger.Warn("clientSdk-listenSdk:", "start proxy", id, sid, "err:", err.Error())
				_ = sdk.stop()
			}
		}()
		return sdk.start()
	})
}

func (sm *sdkManager) ReloadProxy(id string, sid string) error {
	return sm.getProxy(id, sid, func(sdk *proxySdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-listenSdk:", "reload proxy", id, sid)
			} else {
				sm.logger.Warn("clientSdk-listenSdk:", "reload proxy", id, sid, "err:", err.Error())
				_ = sdk.stop()
			}
		}()
		return sdk.reload()
	})
}

func (sm *sdkManager) StopProxy(id string, sid string) error {
	return sm.getProxy(id, sid, func(sdk *proxySdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-listenSdk:", "stop proxy", id, sid)
			} else {
				sm.logger.Warn("clientSdk-listenSdk:", "stop proxy", id, sid, "err:", err.Error())
			}
		}()
		return sdk.stop()
	})
}

func (sm *sdkManager) UpdateProxy(id string, sid string, fn func(cfg *ProxyConfig)) error {
	return sm.getProxy(id, sid, func(sdk *proxySdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-dialSdk:", "update proxy", id, sid)
			} else {
				sm.logger.Warn("clientSdk-dialSdk:", "update proxy", id, sid, "err:", err.Error())
				_ = sdk.stop()
			}
		}()
		return sdk.update(fn)
	})
}

func (sm *sdkManager) GetConfigProxy(id string, sid string) (*ProxyConfig, error) {
	var cfg *ProxyConfig
	err := sm.getProxy(id, sid, func(sdk *proxySdk) error {
		cfg = sdk.getConfig()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (sm *sdkManager) GetProxyUnitView(id string, sid string) ([]client.ProxyUnitView, error) {
	var view []client.ProxyUnitView
	err := sm.getProxy(id, sid, func(sdk *proxySdk) error {
		view = sdk.getView()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return view, nil
}

func (sm *sdkManager) getProxy(id string, sid string, fn func(sdk *proxySdk) error) error {
	return sm.getClient(id, func(cs *clientSdk) error {
		cs.mux.Lock()
		defer cs.mux.Unlock()
		sub := findIndex(cs.pl, sid)
		if sub < 0 || sub >= len(cs.pl) {
			return errors.New("invalid sid")
		}
		sdk := cs.pl[sub]
		return fn(sdk)
	})
}

func (cs *clientSdk) newProxySdk(cfg *ProxyConfig) (*proxySdk, error) {
	ps := &proxySdk{
		id:     newSdkId(IdPrefixProxy),
		cs:     cs,
		config: cfg,
	}
	if cs.getStatus() && cfg.Enable {
		err := ps.reload()
		if err != nil {
			return nil, err
		}
	}
	return ps, nil
}

type proxySdk struct {
	id     string
	cs     *clientSdk
	ln     net.Listener
	proxy  *client.ProxyDialer
	mux    sync.Mutex
	config *ProxyConfig
	status bool
}

func (ps *proxySdk) GetId() string {
	return ps.id
}

func (ps *proxySdk) run(lock, reload bool) (err error) {
	if lock {
		ps.mux.Lock()
		defer ps.mux.Unlock()
	}
	if ps.ln != nil {
		if reload {
			_ = ps.ln.Close()
			if ps.proxy != nil {
				_ = ps.proxy.Close()
			}
		} else {
			return nil
		}
	}
	if ps.cs.client == nil {
		return errors.New("client not running")
	}
	ctx, cl := context.WithCancel(ps.cs.client.Context())
	defer func() {
		if err != nil {
			cl()
		}
	}()
	if ps.config.InNetwork == nil {
		return errors.New("nil in network")
	}
	lc, err := xnet.GetBaseStreamListenerConfig(ps.config.InNetwork.Network)
	if err != nil {
		return err
	}
	ln, err := lc.ListenContext(ctx, ps.config.InNetwork.Network, ps.config.InNetwork.Address)
	if err != nil {
		return err
	}
	ctxtool.GWaitFunc(ctx, func() {
		_ = ln.Close()
	})

	// plugin
	var plugins []plugin.ProxyPlugin
	if ps.config.Plugin != "" {
		ps.cs.pmMux.Lock()
		pl, ok := ps.cs.pm[ps.config.Plugin]
		if !ok || pl.a[PluginTypeProxy] == nil {
			ps.cs.pmMux.Unlock()
			err = fmt.Errorf("not found proxy plugin: %s", ps.config.Plugin)
			return err
		}
		pl.mux.Lock()
		plugins = pl.a[PluginTypeProxy].([]plugin.ProxyPlugin)
		pl.mux.Unlock()
		ps.cs.pmMux.Unlock()
	} else if ps.config.OutNetwork == nil {
		return errors.New("nil out network")
	}

	//out
	out := dcopy.CopyT(ps.config.OutNetwork)

	if len(plugins) == 0 && out == nil {
		return errors.New("nil dst network and address")
	}

	proxy := ps.cs.client.Proxy(ps.config.Multi)
	ps.proxy = proxy
	if len(ps.config.Node) != 0 {
		ctx = context.WithValue(ctx, client.ProxyNodes, ps.config.Node)
	}

	go func() {
		defer func() {
			cl()
			_ = ln.Close()
			_ = proxy.Close()
			ps.mux.Lock()
			defer ps.mux.Unlock()
			if ps.ln == ln {
				ps.ln = nil
				ps.status = false
			}
			if ps.proxy == proxy {
				ps.proxy = nil
			}
		}()
		ps.handle(ctx, ln, out, proxy, plugins)
	}()
	ps.ln = ln
	ps.status = true
	return nil
}

func (ps *proxySdk) handle(ctx context.Context, ln net.Listener, out *NetworkConfig, proxy *client.ProxyDialer, plugins []plugin.ProxyPlugin) {
	defer ln.Close()
	tdFunc := proxy.DialContext
	var p plugin.ProxyPlugin
	if len(plugins) != 0 {
		var err error
		for i := len(plugins) - 1; i >= 0; i-- {
			ctx, ln, tdFunc, err = plugins[i].ProxyUpgrade(ctx, ln, tdFunc)
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
			_ = p.ProxyServe(ctx, ln, tdFunc)
		}
	}
}

func (ps *proxySdk) update(fn func(cfg *ProxyConfig)) error {
	ps.mux.Lock()
	defer ps.mux.Unlock()
	fn(ps.config)
	err := ps.cs.sm.save()
	if err != nil {
		return err
	}
	if !ps.status {
		return nil
	}
	return ps.run(false, true)
}

func (ps *proxySdk) notLock(r bool) error {
	return ps.run(false, r)
}

func (ps *proxySdk) reload() error {
	return ps.run(true, true)
}

func (ps *proxySdk) start() error {
	return ps.run(true, false)
}

func (ps *proxySdk) stop() error {
	ps.mux.Lock()
	defer ps.mux.Unlock()
	if ps.ln == nil {
		return errors.New("no running")
	}
	_ = ps.ln.Close()
	ps.ln = nil
	ps.status = false
	return nil
}

func (ps *proxySdk) getConfig() *ProxyConfig {
	ps.mux.Lock()
	defer ps.mux.Unlock()
	return dcopy.CopyT(ps.config)
}

func (ps *proxySdk) getView() []client.ProxyUnitView {
	ps.mux.Lock()
	defer ps.mux.Unlock()
	return ps.proxy.View()
}
