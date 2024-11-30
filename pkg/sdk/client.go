package sdk

import (
	"errors"
	"fmt"
	"github.com/peakedshout/anchorage-core/pkg/client"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"github.com/peakedshout/go-pandorasbox/tool/dcopy"
	"github.com/peakedshout/go-pandorasbox/xrpc"
	"sync"
)

func (sm *sdkManager) handleClient() error {
	for _, cc := range sm.cfg.Config.Client {
		sdk, err := sm.newClientSdk(cc)
		if err != nil {
			return err
		}
		sm.cList = append(sm.cList, sdk)
	}
	return nil
}

func (sm *sdkManager) newClientSdk(cfg *ClientConfig) (*clientSdk, error) {
	cs := &clientSdk{
		id:     newSdkId(IdPrefixClient),
		sm:     sm,
		config: cfg,
	}
	if cs.config.Enable {
		err := cs.reload()
		if err != nil {
			return nil, err
		}
	}

	err := cs.newPlugin(cs.config.Plugin)
	if err != nil {
		return nil, err
	}
	for _, scfg := range cs.config.Listen {
		sdk, err := cs.newListenSdk(scfg)
		if err != nil {
			_ = cs.stop()
			return nil, err
		}
		cs.ll = append(cs.ll, sdk)
	}
	for _, scfg := range cs.config.Dial {
		sdk, err := cs.newDialSdk(scfg)
		if err != nil {
			_ = cs.stop()
			return nil, err
		}
		cs.dl = append(cs.dl, sdk)
	}
	for _, scfg := range cs.config.Proxy {
		sdk, err := cs.newProxySdk(scfg)
		if err != nil {
			_ = cs.stop()
			return nil, err
		}
		cs.pl = append(cs.pl, sdk)
	}
	return cs, nil
}

type clientSdk struct {
	id     string
	sm     *sdkManager
	mux    sync.Mutex
	client *client.Client
	config *ClientConfig
	status bool

	ll    []*listenSdk
	dl    []*dialSdk
	pl    []*proxySdk
	pmMux sync.Mutex
	pm    map[string]*pluginSdk
}

func (cs *clientSdk) run(lock, reload bool) error {
	if lock {
		cs.mux.Lock()
		defer cs.mux.Unlock()
	}
	if cs.client != nil {
		if reload {
			_ = cs.client.Close()
		} else {
			return nil
		}
	}
	c, err := client.NewClientContext(cs.sm.context(), cs.config.ClientConfig)
	if err != nil {
		return err
	}
	cs.sm.wg.Add(1)
	ctxtool.GWaitFunc(c.Context(), func() {
		defer cs.sm.wg.Done()
		cs.mux.Lock()
		defer cs.mux.Unlock()
		if c == cs.client {
			cs.client = nil
			cs.status = false
		}
		_ = c.Close()
	})
	cs.client = c
	cs.status = true

	return nil
}

func (sm *sdkManager) AddClient(cc *ClientConfig) (id string, err error) {
	defer func() {
		if err == nil {
			sm.logger.Info("clientSdk:", "add client", id)
		} else {
			sm.logger.Warn("clientSdk:", "add client", id, "err:", err.Error())
		}
	}()
	defer sm.Lock().Unlock()
	sdk, err := sm.newClientSdk(cc)
	if err != nil {
		return "", err
	}
	sm.cfg.Config.Client = append(sm.cfg.Config.Client, cc)
	err = sm.save()
	if err != nil {
		sm.cfg.Config.Client = sm.cfg.Config.Client[:len(sm.cfg.Config.Client)-1]
		_ = sdk.stop()
		return "", err
	}
	sm.cList = append(sm.cList, sdk)
	return sdk.GetId(), nil
}

func (sm *sdkManager) AddClient2(cc *ClientConfigUnit) (id string, err error) {
	defer func() {
		if err == nil {
			sm.logger.Info("clientSdk:", "add client", id)
		} else {
			sm.logger.Warn("clientSdk:", "add client", id, "err:", err.Error())
		}
	}()
	defer sm.Lock().Unlock()
	cfg := &ClientConfig{
		ClientConfigUnit: cc,
	}
	sdk, err := sm.newClientSdk(cfg)
	if err != nil {
		return "", err
	}
	sm.cfg.Config.Client = append(sm.cfg.Config.Client, cfg)
	err = sm.save()
	if err != nil {
		sm.cfg.Config.Client = sm.cfg.Config.Client[:len(sm.cfg.Config.Client)-1]
		_ = sdk.stop()
		return "", err
	}
	sm.cList = append(sm.cList, sdk)
	return sdk.GetId(), nil
}

func (sm *sdkManager) DelClient(id string) (err error) {
	defer func() {
		if err == nil {
			sm.logger.Info("clientSdk:", "del client", id)
		} else {
			sm.logger.Warn("clientSdk:", "del client", id, "err:", err.Error())
		}
	}()
	defer sm.Lock().Unlock()
	index := findIndex(sm.cList, id)
	if index < 0 || index >= len(sm.cList) {
		return errors.New("invalid id")
	}
	sdk := sm.cList[index]
	sm.cList = append(sm.cList[:index], sm.cList[index+1:]...)
	sm.cfg.Config.Client = append(sm.cfg.Config.Client[:index], sm.cfg.Config.Client[index+1:]...)
	_ = sdk.stop()
	return sm.save()
}

func (sm *sdkManager) StartClient(id string) error {
	return sm.getClient(id, func(sdk *clientSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk:", "start client", id)
			} else {
				sm.logger.Warn("clientSdk:", "start client", id, "err:", err.Error())
				_ = sdk.stop()
			}
		}()
		sdk.mux.Lock()
		defer sdk.mux.Unlock()
		err = sdk.notLock(false)
		if err != nil {
			return err
		}
		var errs []error
		for _, l := range sdk.ll {
			l.mux.Lock()
			if l.config.Enable {
				suberr := l.notLock(false)
				if suberr != nil {
					errs = append(errs, fmt.Errorf("[listen] %s err: %w", l.config.Name, suberr))
				}
			}
			l.mux.Unlock()
		}
		for _, l := range sdk.dl {
			l.mux.Lock()
			if l.config.Enable {
				suberr := l.notLock(false)
				if suberr != nil {
					errs = append(errs, fmt.Errorf("[dial] %s err: %w", l.config.Link, suberr))
				}
			}
			l.mux.Unlock()
		}
		for _, l := range sdk.pl {
			l.mux.Lock()
			if l.config.Enable {
				suberr := l.notLock(false)
				if suberr != nil {
					errs = append(errs, fmt.Errorf("[proxy] %s err: %w", l.config.InNetwork.Address, suberr))
				}
			}
			l.mux.Unlock()
		}
		return errors.Join(errs...)
	})
}

func (sm *sdkManager) StartClient2(id string) error {
	return sm.getClient(id, func(sdk *clientSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk:", "start client", id)
			} else {
				sm.logger.Warn("clientSdk:", "start client", id, "err:", err.Error())
				_ = sdk.stop()
			}
		}()
		return sdk.start()
	})
}

func (sm *sdkManager) ReloadClient(id string) error {
	return sm.getClient(id, func(sdk *clientSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk:", "reload client", id)
			} else {
				sm.logger.Warn("clientSdk:", "reload client", id, "err:", err.Error())
				_ = sdk.stop()
			}
		}()
		sdk.mux.Lock()
		defer sdk.mux.Unlock()
		err = sdk.notLock(true)
		if err != nil {
			return err
		}
		var errs []error
		for _, l := range sdk.ll {
			l.mux.Lock()
			if l.config.Enable {
				suberr := l.notLock(true)
				if suberr != nil {
					errs = append(errs, fmt.Errorf("[listen] %s err: %w", l.config.Name, suberr))
				}
			}
			l.mux.Unlock()
		}
		for _, l := range sdk.dl {
			l.mux.Lock()
			if l.config.Enable {
				suberr := l.notLock(true)
				if suberr != nil {
					errs = append(errs, fmt.Errorf("[dial] %s err: %w", l.config.Link, suberr))
				}
			}
			l.mux.Unlock()
		}
		for _, l := range sdk.pl {
			l.mux.Lock()
			if l.config.Enable {
				suberr := l.notLock(true)
				if suberr != nil {
					errs = append(errs, fmt.Errorf("[proxy] %s err: %w", l.config.InNetwork.Address, suberr))
				}
			}
			l.mux.Unlock()
		}
		return errors.Join(errs...)
	})
}

func (sm *sdkManager) ReloadClient2(id string) error {
	return sm.getClient(id, func(sdk *clientSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk:", "reload client", id)
			} else {
				sm.logger.Warn("clientSdk:", "reload client", id, "err:", err.Error())
				_ = sdk.stop()
			}
		}()
		return sdk.reload()
	})
}

func (sm *sdkManager) StopClient(id string) error {
	return sm.getClient(id, func(sdk *clientSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk:", "stop client", id)
			} else {
				sm.logger.Warn("clientSdk:", "stop client", id, "err:", err.Error())
			}
		}()
		return sdk.stop()
	})
}

func (sm *sdkManager) UpdateClient(id string, fn func(cfg *ClientConfig)) error {
	return sm.getClient(id, func(sdk *clientSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk:", "update client", id)
			} else {
				sm.logger.Warn("clientSdk:", "update client", id, "err:", err.Error())
				_ = sdk.stop()
			}
		}()
		return sdk.update(fn)
	})
}

func (sm *sdkManager) UpdateClient2(id string, fn func(cfg *ClientConfigUnit)) error {
	return sm.getClient(id, func(sdk *clientSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk:", "update client", id)
			} else {
				sm.logger.Warn("clientSdk:", "update client", id, "err:", err.Error())
				_ = sdk.stop()
			}
		}()
		return sdk.update2(fn)
	})
}

func (sm *sdkManager) GetConfigClient(id string) (*ClientConfig, error) {
	var cfg *ClientConfig
	err := sm.getClient(id, func(sdk *clientSdk) error {
		cfg = sdk.getConfig()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (sm *sdkManager) GetConfigClient2(id string) (*ClientConfigUnit, error) {
	var cfg *ClientConfigUnit
	err := sm.getClient(id, func(sdk *clientSdk) error {
		cfg = sdk.getConfig2()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (sm *sdkManager) GetClientSessionView(id string) (view map[string][]xrpc.SessionView, err error) {
	err = sm.getClient(id, func(sdk *clientSdk) error {
		view, err = sdk.getSessionView()
		return err
	})
	if err != nil {
		return nil, err
	}
	return view, nil
}

func (sm *sdkManager) GetClientProxyView(id string) (view map[string][]client.ProxyUnitView, err error) {
	err = sm.getClient(id, func(sdk *clientSdk) error {
		view, err = sdk.getProxyView()
		return err
	})
	if err != nil {
		return nil, err
	}
	return view, nil
}

func (sm *sdkManager) getClient(id string, fn func(sdk *clientSdk) error) error {
	defer sm.Lock().Unlock()
	index := findIndex(sm.cList, id)
	if index < 0 || index >= len(sm.cList) {
		return errors.New("invalid id")
	}
	sdk := sm.cList[index]
	return fn(sdk)
}

func (cs *clientSdk) GetId() string {
	return cs.id
}

func (cs *clientSdk) update(fn func(cfg *ClientConfig)) (err error) {
	cs.mux.Lock()
	defer cs.mux.Unlock()
	fn(cs.config)
	err = cs.sm.save()
	if err != nil {
		return err
	}
	if !cs.status {
		return nil
	}
	sub := cs.updateSub()
	defer func() {
		if err == nil {
			_ = sub()
		}
	}()
	return cs.run(false, true)
}

func (cs *clientSdk) update2(fn func(cfg *ClientConfigUnit)) (err error) {
	cs.mux.Lock()
	defer cs.mux.Unlock()
	fn(cs.config.ClientConfigUnit)
	err = cs.sm.save()
	if err != nil {
		return err
	}
	if !cs.status {
		return nil
	}
	sub := cs.updateSub()
	defer func() {
		if err == nil {
			_ = sub()
		}
	}()
	return cs.run(false, true)
}

func (cs *clientSdk) updateSub() func() error {
	var fns []func()
	var errs []error
	for _, l := range cs.ll {
		l.mux.Lock()
		if l.status {
			sdk := l
			fns = append(fns, func() {
				suberr := sdk.run(true, true)
				if suberr != nil {
					errs = append(errs, fmt.Errorf("[listen] %s err: %w", sdk.config.Name, suberr))
				}
			})
		}
		l.mux.Unlock()
	}
	for _, l := range cs.dl {
		l.mux.Lock()
		if l.status {
			sdk := l
			fns = append(fns, func() {
				suberr := sdk.run(true, true)
				if suberr != nil {
					errs = append(errs, fmt.Errorf("[dial] %s err: %w", sdk.config.Link, suberr))
				}
			})
		}
		l.mux.Unlock()
	}
	for _, l := range cs.pl {
		l.mux.Lock()
		if l.status {
			sdk := l
			fns = append(fns, func() {
				suberr := sdk.run(true, true)
				if suberr != nil {
					errs = append(errs, fmt.Errorf("[proxy] %s err: %w", sdk.config.InNetwork.Address, suberr))
				}
			})
		}
		l.mux.Unlock()
	}
	return func() error {
		for _, fn := range fns {
			fn()
		}
		return errors.Join(errs...)
	}
}

func (cs *clientSdk) notLock(r bool) error {
	return cs.run(false, r)
}

func (cs *clientSdk) reload() error {
	return cs.run(true, true)
}

func (cs *clientSdk) start() error {
	return cs.run(true, false)
}

func (cs *clientSdk) stop() error {
	cs.mux.Lock()
	defer cs.mux.Unlock()
	if cs.client == nil {
		return errors.New("no running")
	}
	_ = cs.client.Close()
	cs.client = nil
	cs.status = false
	return nil
}

func (cs *clientSdk) getStatus() bool {
	return cs.status
}

func (cs *clientSdk) getConfig() *ClientConfig {
	cs.mux.Lock()
	defer cs.mux.Unlock()
	return dcopy.CopyT(cs.config)
}

func (cs *clientSdk) getConfig2() *ClientConfigUnit {
	cs.mux.Lock()
	defer cs.mux.Unlock()
	return dcopy.CopyT(cs.config.ClientConfigUnit)
}

func (cs *clientSdk) getSessionView() (map[string][]xrpc.SessionView, error) {
	cs.mux.Lock()
	defer cs.mux.Unlock()
	if !cs.status {
		return nil, errors.New("no running")
	}
	return cs.client.GetClientSessionView(), nil
}

func (cs *clientSdk) getProxyView() (map[string][]client.ProxyUnitView, error) {
	cs.mux.Lock()
	defer cs.mux.Unlock()
	if !cs.status {
		return nil, errors.New("no running")
	}
	return cs.client.GetProxyView(), nil
}
