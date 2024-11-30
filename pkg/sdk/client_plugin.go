package sdk

import (
	"errors"
	"fmt"
	"github.com/peakedshout/anchorage-core/pkg/sdk/plugin"
	"github.com/peakedshout/go-pandorasbox/tool/dcopy"
	"github.com/peakedshout/go-pandorasbox/tool/hjson"
	"sync"
)

func (sm *sdkManager) GetPluginUnitTempCfg(t, n string) (any, error) {
	switch t {
	case PluginTypeDial:
		dialPlugin, ok := plugin.GetDialPlugin(n)
		if !ok {
			return nil, fmt.Errorf("no found %s plugin: %s", t, n)
		}
		return dialPlugin.TempDialCfg(), nil
	case PluginTypeProxy:
		proxyPlugin, ok := plugin.GetProxyPlugin(n)
		if !ok {
			return nil, fmt.Errorf("no found %s plugin: %s", t, n)
		}
		return proxyPlugin.TempProxyCfg(), nil
	case PluginTypeListen:
		return nil, errors.New("todo")
	default:
		return nil, errors.New("invalid plugin type: " + t)
	}
}

func (sm *sdkManager) ListPluginUnit(t string) ([]string, error) {
	switch t {
	case PluginTypeDial:
		return plugin.ListDialPlugin(), nil
	case PluginTypeProxy:
		return plugin.ListProxyPlugin(), nil
	case PluginTypeListen:
		return nil, errors.New("todo")
	default:
		return nil, errors.New("invalid plugin type: " + t)
	}
}

func (sm *sdkManager) AddPlugin(id string, cfg *PluginConfig) error {
	return sm.getClient(id, func(sdk *clientSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-listenSdk:", "add plugin", id, cfg.Name)
			} else {
				sm.logger.Warn("clientSdk-listenSdk:", "add plugin", id, cfg.Name, "err:", err.Error())
			}
		}()
		sdk.mux.Lock()
		defer sdk.mux.Unlock()
		if _, ok := sdk.pm[cfg.Name]; ok {
			return errors.New("plugin name duplicate")
		}
		subSdk, err := newPluginSdk(cfg)
		if err != nil {
			return err
		}
		sdk.config.Plugin = append(sdk.config.Plugin, cfg)
		sdk.pm[cfg.Name] = subSdk
		err = sm.save()
		if err != nil {
			sdk.config.Plugin = sdk.config.Plugin[:len(sdk.config.Plugin)-1]
			delete(sdk.pm, cfg.Name)
			return err
		}
		return nil
	})
}

func (sm *sdkManager) DelPlugin(id string, name string) error {
	return sm.getClient(id, func(sdk *clientSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-listenSdk:", "del plugin", id, name)
			} else {
				sm.logger.Warn("clientSdk-listenSdk:", "del plugin", id, name, "err:", err.Error())
			}
		}()
		sdk.mux.Lock()
		defer sdk.mux.Unlock()
		if _, ok := sdk.pm[name]; !ok {
			return fmt.Errorf("not found plugin: %s", name)
		}
		for i, config := range sdk.config.Plugin {
			if config.Name == name {
				sdk.config.Plugin = append(sdk.config.Plugin[:i], sdk.config.Plugin[i+1:]...)
				break
			}
		}
		delete(sdk.pm, name)
		return sm.save()
	})
}

func (sm *sdkManager) UpdatePlugin(id string, name string, fn func(cfg *PluginConfig)) error {
	return sm.getClient(id, func(sdk *clientSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("clientSdk-listenSdk:", "update plugin", id, name)
			} else {
				sm.logger.Warn("clientSdk-listenSdk:", "update plugin", id, name, "err:", err.Error())
			}
		}()
		sdk.mux.Lock()
		defer sdk.mux.Unlock()
		sdk.pmMux.Lock()
		defer sdk.pmMux.Unlock()
		subSdk, ok := sdk.pm[name]
		if !ok {
			return fmt.Errorf("not found plugin: %s", name)
		}
		subSdk.mux.Lock()
		defer subSdk.mux.Unlock()
		tmp := dcopy.CopyT(subSdk.config)
		fn(subSdk.config)
		if tmp.Name != subSdk.config.Name {
			_, ok = sdk.pm[subSdk.config.Name]
			if ok {
				subSdk.config = tmp
				return errors.New("plugin name duplicate")
			}
		}
		err = subSdk.build()
		if err != nil {
			subSdk.config = tmp
			return err
		}
		if tmp.Name != subSdk.config.Name {
			delete(sdk.pm, tmp.Name)
			sdk.pm[subSdk.config.Name] = subSdk
		}
		return sm.save()
	})
}

func (sm *sdkManager) GetConfigPlugin(id string, name string) (*PluginConfig, error) {
	var cfg *PluginConfig
	err := sm.getClient(id, func(cs *clientSdk) error {
		cs.mux.Lock()
		defer cs.mux.Unlock()
		cs.pmMux.Lock()
		defer cs.pmMux.Unlock()
		sdk, ok := cs.pm[name]
		if !ok {
			return fmt.Errorf("not found plugin: %s", name)
		}
		cfg = sdk.getConfig()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (cs *clientSdk) newPlugin(cfgs []*PluginConfig) error {
	cs.pm = make(map[string]*pluginSdk)
	for _, cfg := range cfgs {
		if cfg.Name == "" {
			return errors.New("plugin nil name")
		}
		if _, ok := cs.pm[cfg.Name]; ok {
			return errors.New("plugin name duplicate")
		}
		sdk, err := newPluginSdk(cfg)
		if err != nil {
			return err
		}
		cs.pm[cfg.Name] = sdk
	}
	return nil
}

func newPluginSdk(cfg *PluginConfig) (*pluginSdk, error) {
	sdk := &pluginSdk{
		config: cfg,
	}
	err := sdk.build()
	if err != nil {
		return nil, err
	}
	return sdk, nil
}

type pluginSdk struct {
	a      map[string]any
	mux    sync.Mutex
	config *PluginConfig
}

func (ps *pluginSdk) build() error {
	ps.a = make(map[string]any)
	for _, t := range ps.config.Type {
		switch t {
		case PluginTypeDial:
			var dl []plugin.DialPlugin
			for _, m := range ps.config.List {
				str := m[plugin.Name]
				if str == nil {
					return fmt.Errorf("no found dial plugin name")
				}
				name, ok := str.(string)
				if !ok || name == "" {
					return fmt.Errorf("no found dial plugin name")
				}
				dialPlugin, ok := plugin.GetDialPlugin(name)
				if !ok {
					return fmt.Errorf("no found %s dial plugin", name)
				}
				cfg, err := dialPlugin.LoadDialCfg(hjson.MustMarshalStr(m))
				if err != nil {
					return err
				}
				dl = append(dl, cfg)
			}
			ps.a[t] = dl
		case PluginTypeListen:
			return errors.New("todo")
		case PluginTypeProxy:
			var pl []plugin.ProxyPlugin
			for _, m := range ps.config.List {
				str := m[plugin.Name]
				if str == nil {
					return fmt.Errorf("no found proxy plugin name")
				}
				name, ok := str.(string)
				if !ok || name == "" {
					return fmt.Errorf("no found proxy plugin name")
				}
				proxylugin, ok := plugin.GetProxyPlugin(name)
				if !ok {
					return fmt.Errorf("no found %s proxy plugin", name)
				}
				cfg, err := proxylugin.LoadProxyCfg(hjson.MustMarshalStr(m))
				if err != nil {
					return err
				}
				pl = append(pl, cfg)
			}
			ps.a[t] = pl
		default:
			return errors.New("invalid plugin type")
		}
	}
	return nil
}

//func (ps *pluginSdk) newPluginHttp() {
//	scfg := &httpproxy.ServerConfig{}
//	u := ps.config.Args[PluginArgsAuthUserName]
//	p := ps.config.Args[PluginArgsAuthPassword]
//	if u != "" && p != "" {
//		scfg.ReqAuthCb = httpproxy.UserInfoAuth(u, p)
//	}
//	ps.a = scfg
//}
//
//func (ps *pluginSdk) newPluginSocks() {
//	s4, s5 := true, true
//	if ps.config.Type == PluginTypeSocks4 {
//		s5 = false
//	} else if ps.config.Type == PluginTypeSocks5 {
//		s4 = false
//	}
//	u := ps.config.Args[PluginArgsAuthUserName]
//	p := ps.config.Args[PluginArgsAuthPassword]
//	scfg := &socks.ServerSimplifyConfig{
//		SwitchSocksVersion4:   s4,
//		SwitchSocksVersion5:   s5,
//		SwitchCMDCONNECT:      true,
//		SwitchCMDBIND:         false,
//		SwitchCMDUDPASSOCIATE: true,
//		Socks5Auth:            nil,
//		Socks4Auth:            nil,
//	}
//	if u != "" && p != "" {
//		scfg.Socks5Auth = &socks.SimplifySocks5Auth{
//			User:     ps.config.Args[PluginArgsAuthUserName],
//			Password: ps.config.Args[PluginArgsAuthPassword],
//		}
//
//		scfg.Socks4Auth = &socks.SimplifySocks4Auth{UserId: ps.config.Args[PluginArgsAuthUserName]}
//	}
//	xcfg := scfg.Build()
//	ps.a = xcfg
//}

func (ps *pluginSdk) getConfig() *PluginConfig {
	ps.mux.Lock()
	defer ps.mux.Unlock()
	return dcopy.CopyT(ps.config)
}
