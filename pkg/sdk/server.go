package sdk

import (
	"errors"
	"github.com/peakedshout/anchorage-core/pkg/comm"
	"github.com/peakedshout/anchorage-core/pkg/server"
	"github.com/peakedshout/go-pandorasbox/tool/dcopy"
	"github.com/peakedshout/go-pandorasbox/xrpc"
	"sync"
)

func (sm *sdkManager) handleServer() error {
	for _, sc := range sm.cfg.Config.Server {
		sdk, err := sm.newServerSdk(sc)
		if err != nil {
			return err
		}
		sm.sList = append(sm.sList, sdk)
	}
	return nil
}

func (sm *sdkManager) newServerSdk(cfg *ServerConfig) (*serverSdk, error) {
	ss := &serverSdk{
		id:     newSdkId(IdPrefixServer),
		sm:     sm,
		config: cfg,
	}
	if ss.config.Enable {
		err := ss.reload()
		if err != nil {
			return nil, err
		}
	}
	return ss, nil
}

type serverSdk struct {
	id     string
	sm     *sdkManager
	mux    sync.Mutex
	server *server.Server
	config *ServerConfig
	status bool
}

func (ss *serverSdk) run(lock, reload bool) error {
	if lock {
		ss.mux.Lock()
		defer ss.mux.Unlock()
	}
	if ss.server != nil {
		if reload {
			_ = ss.server.Close()
		} else {
			return nil
		}
	}
	s, err := server.NewServerContext(ss.sm.context(), ss.config.ServerConfig)
	if err != nil {
		return err
	}
	ss.server = s
	ss.sm.wg.Add(1)
	ch := make(chan error, 1)
	go func() {
		defer ss.sm.wg.Done()
		defer s.Close()
		_ = s.SyncServe(ch)
		ss.mux.Lock()
		defer ss.mux.Unlock()
		if ss.server == s {
			ss.server = nil
			ss.status = false
		}
	}()
	ss.status = true
	return <-ch
}

func (sm *sdkManager) AddServer(sc *ServerConfig) (id string, err error) {
	defer func() {
		if err == nil {
			sm.logger.Info("serverSdk:", "add server", sc.NodeInfo.NodeName, "->", id)
		} else {
			sm.logger.Warn("serverSdk:", "add server", sc.NodeInfo.NodeName, "->", id, "err:", err.Error())
		}
	}()
	defer sm.Lock().Unlock()
	sdk, err := sm.newServerSdk(sc)
	if err != nil {
		return "", err
	}
	sm.cfg.Config.Server = append(sm.cfg.Config.Server, sc)
	err = sm.save()
	if err != nil {
		sm.cfg.Config.Server = sm.cfg.Config.Server[:len(sm.cfg.Config.Server)-1]
		_ = sdk.stop()
		return "", err
	}
	sm.sList = append(sm.sList, sdk)
	return sdk.GetId(), nil
}

func (sm *sdkManager) DelServer(id string) (err error) {
	defer func() {
		if err == nil {
			sm.logger.Info("serverSdk:", "del server", id)
		} else {
			sm.logger.Warn("serverSdk:", "del server", id, "err:", err.Error())
		}
	}()
	defer sm.Lock().Unlock()
	index := findIndex(sm.sList, id)
	if index < 0 || index >= len(sm.sList) {
		return errors.New("invalid id")
	}
	sdk := sm.sList[index]
	sm.sList = append(sm.sList[:index], sm.sList[index+1:]...)
	sm.cfg.Config.Server = append(sm.cfg.Config.Server[:index], sm.cfg.Config.Server[index+1:]...)
	_ = sdk.stop()
	sm.logger.Info("serverSdk:", "del server", id)
	return sm.save()
}

func (sm *sdkManager) StartServer(id string) error {
	return sm.getServer(id, func(sdk *serverSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("serverSdk:", "start server", id)
			} else {
				sm.logger.Warn("serverSdk:", "start server", id, "err:", err.Error())
				_ = sdk.stop()
			}
		}()
		return sdk.start()
	})
}

func (sm *sdkManager) ReloadServer(id string) error {
	return sm.getServer(id, func(sdk *serverSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("serverSdk:", "reload server", id)
			} else {
				sm.logger.Warn("serverSdk:", "reload server", id, "err:", err.Error())
				_ = sdk.stop()
			}
		}()
		return sdk.reload()
	})
}

func (sm *sdkManager) StopServer(id string) error {
	return sm.getServer(id, func(sdk *serverSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("serverSdk:", "stop server", id)
			} else {
				sm.logger.Warn("serverSdk:", "stop server", id, "err:", err.Error())
			}
		}()
		return sdk.stop()
	})
}

func (sm *sdkManager) UpdateServer(id string, fn func(cfg *ServerConfig)) error {
	return sm.getServer(id, func(sdk *serverSdk) (err error) {
		defer func() {
			if err == nil {
				sm.logger.Info("serverSdk:", "update server", id)
			} else {
				sm.logger.Warn("serverSdk:", "update server", id, "err:", err.Error())
				_ = sdk.stop()
			}
		}()
		return sdk.update(fn)
	})
}

func (sm *sdkManager) GetConfigServer(id string) (*ServerConfig, error) {
	var cfg *ServerConfig
	err := sm.getServer(id, func(sdk *serverSdk) error {
		cfg = sdk.getConfig()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (sm *sdkManager) GetServerSessionView(id string) (view []xrpc.SessionView, err error) {
	err = sm.getServer(id, func(sdk *serverSdk) error {
		view, err = sdk.getSessionView()
		return err
	})
	if err != nil {
		return nil, err
	}
	return view, nil
}

func (sm *sdkManager) GetServerRouteView(id string) (view *comm.ServiceRouteView, err error) {
	err = sm.getServer(id, func(sdk *serverSdk) error {
		view, err = sdk.getRouteView()
		return err
	})
	if err != nil {
		return nil, err
	}
	return view, nil
}

func (sm *sdkManager) GetServerLinkView(id string) (view []comm.LinkView, err error) {
	err = sm.getServer(id, func(sdk *serverSdk) error {
		view, err = sdk.getLinkView()
		return err
	})
	if err != nil {
		return nil, err
	}
	return view, nil
}

func (sm *sdkManager) GetServerSyncView(id string) (view map[string][]xrpc.SessionView, err error) {
	err = sm.getServer(id, func(sdk *serverSdk) error {
		view, err = sdk.getSyncView()
		return err
	})
	if err != nil {
		return nil, err
	}
	return view, nil
}

func (sm *sdkManager) GetServerProxyView(id string) (view map[string][]server.ProxyUnitView, err error) {
	err = sm.getServer(id, func(sdk *serverSdk) error {
		view, err = sdk.getProxyView()
		return err
	})
	if err != nil {
		return nil, err
	}
	return view, nil
}

func (sm *sdkManager) getServer(id string, fn func(sdk *serverSdk) error) error {
	defer sm.Lock().Unlock()
	index := findIndex(sm.sList, id)
	if index < 0 || index >= len(sm.sList) {
		return errors.New("invalid id")
	}
	sdk := sm.sList[index]
	return fn(sdk)
}

func (ss *serverSdk) GetId() string {
	return ss.id
}

func (ss *serverSdk) update(fn func(cfg *ServerConfig)) error {
	ss.mux.Lock()
	defer ss.mux.Unlock()
	fn(ss.config)
	err := ss.sm.save()
	if err != nil {
		return err
	}
	if !ss.status {
		return nil
	}
	return ss.run(false, true)
}

func (ss *serverSdk) reload() error {
	return ss.run(true, true)
}

func (ss *serverSdk) start() error {
	return ss.run(true, false)
}

func (ss *serverSdk) stop() error {
	ss.mux.Lock()
	defer ss.mux.Unlock()
	if ss.server == nil {
		return errors.New("no running")
	}
	_ = ss.server.Close()
	ss.server = nil
	ss.status = false
	return nil
}

func (ss *serverSdk) getConfig() *ServerConfig {
	ss.mux.Lock()
	defer ss.mux.Unlock()
	return dcopy.CopyT(ss.config)
}

func (ss *serverSdk) getSessionView() ([]xrpc.SessionView, error) {
	ss.mux.Lock()
	defer ss.mux.Unlock()
	if !ss.status {
		return nil, errors.New("no running")
	}
	return ss.server.GetServerSessionView(), nil
}

func (ss *serverSdk) getRouteView() (*comm.ServiceRouteView, error) {
	ss.mux.Lock()
	defer ss.mux.Unlock()
	if !ss.status {
		return nil, errors.New("no running")
	}
	return ss.server.GetRouteView(), nil
}

func (ss *serverSdk) getLinkView() ([]comm.LinkView, error) {
	ss.mux.Lock()
	defer ss.mux.Unlock()
	if !ss.status {
		return nil, errors.New("no running")
	}
	return ss.server.GetLinkView(), nil
}

func (ss *serverSdk) getSyncView() (map[string][]xrpc.SessionView, error) {
	ss.mux.Lock()
	defer ss.mux.Unlock()
	if !ss.status {
		return nil, errors.New("no running")
	}
	return ss.server.GetSyncView(), nil
}

func (ss *serverSdk) getProxyView() (map[string][]server.ProxyUnitView, error) {
	ss.mux.Lock()
	defer ss.mux.Unlock()
	if !ss.status {
		return nil, errors.New("no running")
	}
	return ss.server.GetProxyView(), nil
}
