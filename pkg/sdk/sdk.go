package sdk

import (
	"context"
	"errors"
	"fmt"
	"github.com/peakedshout/anchorage-core/pkg/comm"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"github.com/peakedshout/go-pandorasbox/logger"
	"github.com/peakedshout/go-pandorasbox/tool/dcopy"
	"github.com/peakedshout/go-pandorasbox/tool/hyaml"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"io"
	"slices"
	"sync"
)

type Sdk struct {
	*sdkManager
}

func NewSdkFromFile(ctx context.Context, fp string) (*Sdk, error) {
	cfg := &hyaml.Config[Config]{}
	cfg.SetPath(fp)
	return NewSdk(ctx, cfg)
}

func NewSdk(ctx context.Context, cfg *hyaml.Config[Config]) (*Sdk, error) {
	manager, err := newSdkManager(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Sdk{manager}, nil
}

type sdkManager struct {
	mux sync.Mutex
	ctx context.Context
	cl  context.CancelFunc

	logger logger.Logger

	cfg *hyaml.Config[Config]

	//smux  sync.Mutex
	sList []*serverSdk
	//cmux  sync.Mutex
	cList []*clientSdk

	wg sync.WaitGroup
}

func newSdkManager(ctx context.Context, cfg *hyaml.Config[Config]) (*sdkManager, error) {
	err := cfg.Update()
	if err != nil {
		return nil, err
	}
	if cfg.Config == nil {
		return nil, errors.New("nil config")
	}
	sm := &sdkManager{
		cfg: cfg,
	}
	sm.ctx, sm.cl = context.WithCancel(ctx)
	sm.logger = comm.MakeLogger(sm.ctx, logger.Init("anchorage"), cfg.Config.Logger)
	sm.ctx = logger.SetLogger(sm.ctx, sm.logger)
	defer func() {
		if err != nil {
			sm.Stop()
		}
	}()
	err = sm.handleServer()
	if err != nil {
		return nil, err
	}
	err = sm.handleClient()
	if err != nil {
		return nil, err
	}
	return sm, nil
}

func (sm *sdkManager) Stop() {
	defer sm.Lock().Unlock()
	sm.cl()
	sm.wg.Wait()
}

func (sm *sdkManager) Context() context.Context {
	sm.mux.Lock()
	defer sm.mux.Unlock()
	return sm.context()
}

func (sm *sdkManager) context() context.Context {
	return sm.ctx
}

func (sm *sdkManager) GetServerView() []*ServerView {
	defer sm.Lock().Unlock()
	list := make([]*ServerView, 0, len(sm.sList))
	for _, sdk := range sm.sList {
		sdk.mux.Lock()
		v := &ServerView{
			Id:     sdk.GetId(),
			Status: sdk.status,
			Enable: sdk.config.Enable,
			Config: dcopy.CopyT(sdk.config.ServerConfig),
		}
		sdk.mux.Unlock()
		list = append(list, v)
	}
	return list
}

func (sm *sdkManager) GetClientView() []*ClientView {
	defer sm.Lock().Unlock()
	list := make([]*ClientView, 0, len(sm.cList))
	for _, sdk := range sm.cList {
		sdk.mux.Lock()
		v := &ClientView{
			Id:     sdk.GetId(),
			Status: sdk.status,
			Enable: sdk.config.Enable,
			Config: dcopy.CopyT(sdk.config.ClientConfig),
			Listen: nil,
			Dial:   nil,
			Proxy:  nil,
			Plugin: dcopy.CopyT(sdk.config.Plugin),
		}
		v.Listen = make([]*ListenView, 0, len(sdk.ll))
		for _, l := range sdk.ll {
			l.mux.Lock()
			v.Listen = append(v.Listen, &ListenView{
				Id:           l.GetId(),
				Status:       l.status,
				ListenConfig: dcopy.CopyT(l.config),
			})
			l.mux.Unlock()
		}
		v.Dial = make([]*DialView, 0, len(sdk.dl))
		for _, d := range sdk.dl {
			d.mux.Lock()
			v.Dial = append(v.Dial, &DialView{
				Id:         d.GetId(),
				Status:     d.status,
				DialConfig: dcopy.CopyT(d.config),
			})
			d.mux.Unlock()
		}
		v.Proxy = make([]*ProxyView, 0, len(sdk.pl))
		for _, p := range sdk.pl {
			p.mux.Lock()
			v.Proxy = append(v.Proxy, &ProxyView{
				Id:          p.GetId(),
				Status:      p.status,
				ProxyConfig: dcopy.CopyT(p.config),
			})
			p.mux.Unlock()
		}
		list = append(list, v)
		sdk.mux.Unlock()
	}
	return list
}

func (sm *sdkManager) GetClientView2() []*ClientView {
	defer sm.Lock().Unlock()
	list := make([]*ClientView, 0, len(sm.cList))
	for _, sdk := range sm.cList {
		sdk.mux.Lock()
		v := &ClientView{
			Id:     sdk.GetId(),
			Status: sdk.status,
			Enable: sdk.config.Enable,
			Config: dcopy.CopyT(sdk.config.ClientConfig),
			Listen: nil,
			Dial:   nil,
			Proxy:  nil,
			Plugin: nil,
		}
		list = append(list, v)
		sdk.mux.Unlock()
	}
	return list
}

func (sm *sdkManager) GetServerViewById(id string) (*ServerView, error) {
	var v *ServerView
	return v, sm.getServer(id, func(sdk *serverSdk) error {
		sdk.mux.Lock()
		v = &ServerView{
			Id:     sdk.GetId(),
			Status: sdk.status,
			Enable: sdk.config.Enable,
			Config: dcopy.CopyT(sdk.config.ServerConfig),
		}
		sdk.mux.Unlock()
		return nil
	})
}

func (sm *sdkManager) GetClientViewById(id string) (*ClientView, error) {
	var v *ClientView
	return v, sm.getClient(id, func(sdk *clientSdk) error {
		sdk.mux.Lock()
		v = &ClientView{
			Id:     sdk.GetId(),
			Status: sdk.status,
			Enable: sdk.config.Enable,
			Config: dcopy.CopyT(sdk.config.ClientConfig),
			Listen: nil,
			Dial:   nil,
			Proxy:  nil,
			Plugin: dcopy.CopyT(sdk.config.Plugin),
		}
		v.Listen = make([]*ListenView, 0, len(sdk.ll))
		for _, l := range sdk.ll {
			l.mux.Lock()
			v.Listen = append(v.Listen, &ListenView{
				Id:           l.GetId(),
				Status:       l.status,
				ListenConfig: dcopy.CopyT(l.config),
			})
			l.mux.Unlock()
		}
		v.Dial = make([]*DialView, 0, len(sdk.dl))
		for _, d := range sdk.dl {
			d.mux.Lock()
			v.Dial = append(v.Dial, &DialView{
				Id:         d.GetId(),
				Status:     d.status,
				DialConfig: dcopy.CopyT(d.config),
			})
			d.mux.Unlock()
		}
		v.Proxy = make([]*ProxyView, 0, len(sdk.pl))
		for _, p := range sdk.pl {
			p.mux.Lock()
			v.Proxy = append(v.Proxy, &ProxyView{
				Id:          p.GetId(),
				Status:      p.status,
				ProxyConfig: dcopy.CopyT(p.config),
			})
			p.mux.Unlock()
		}
		sdk.mux.Unlock()
		return nil
	})
}

func (sm *sdkManager) GetClientViewById2(id string) (*ClientView, error) {
	var v *ClientView
	return v, sm.getClient(id, func(sdk *clientSdk) error {
		sdk.mux.Lock()
		v = &ClientView{
			Id:     sdk.GetId(),
			Status: sdk.status,
			Enable: sdk.config.Enable,
			Config: dcopy.CopyT(sdk.config.ClientConfig),
			Listen: nil,
			Dial:   nil,
			Proxy:  nil,
			Plugin: dcopy.CopyT(sdk.config.Plugin),
		}
		sdk.mux.Unlock()
		return nil
	})
}

func (sm *sdkManager) GetListenViewById(id string, sid string) (*ListenView, error) {
	var v *ListenView
	return v, sm.getListen(id, sid, func(sdk *listenSdk) error {
		sdk.mux.Lock()
		v = &ListenView{
			Id:           sdk.GetId(),
			Status:       sdk.status,
			ListenConfig: dcopy.CopyT(sdk.config),
		}
		sdk.mux.Unlock()
		return nil
	})
}

func (sm *sdkManager) GetDialViewById(id string, sid string) (*DialView, error) {
	var v *DialView
	return v, sm.getDial(id, sid, func(sdk *dialSdk) error {
		sdk.mux.Lock()
		v = &DialView{
			Id:         sdk.GetId(),
			Status:     sdk.status,
			DialConfig: dcopy.CopyT(sdk.config),
		}
		sdk.mux.Unlock()
		return nil
	})
}

func (sm *sdkManager) GetProxyViewById(id string, sid string) (*ProxyView, error) {
	var v *ProxyView
	return v, sm.getProxy(id, sid, func(sdk *proxySdk) error {
		sdk.mux.Lock()
		v = &ProxyView{
			Id:          sdk.GetId(),
			Status:      sdk.status,
			ProxyConfig: dcopy.CopyT(sdk.config),
		}
		sdk.mux.Unlock()
		return nil
	})
}

func (sm *sdkManager) Lock() *sync.Mutex {
	sm.mux.Lock()
	return &sm.mux
}

func (sm *sdkManager) save() error {
	return sm.cfg.Save()
}

func (sm *sdkManager) Save() error {
	defer sm.Lock().Unlock()
	return sm.save()
}

func (sm *sdkManager) GetConfig() *Config {
	defer sm.Lock().Unlock()
	return dcopy.CopyT(sm.cfg.Config)
}

func (sm *sdkManager) GetLogger(ctx context.Context) (io.Reader, error) {
	reader, writer := io.Pipe()
	tCtx, tCl := ctxtool.ContextsWithCancel(sm.Context(), ctx)
	go func() {
		defer sm.logger.Info("del sync logger copy...")
		_, _ = writer.Write([]byte{})
		_ = sm.logger.SyncLoggerCopy(tCtx, writer)
		tCl()
		_ = writer.Close()
	}()
	sm.logger.Info("add sync logger copy...")
	return reader, nil
}

func findIndex[S interface{ ~[]E }, E interface{ GetId() string }](s S, id string) int {
	return slices.IndexFunc(s, func(e E) bool {
		return id == e.GetId()
	})
}

func newSdkId(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, uuid.NewUrlBase64())
}
