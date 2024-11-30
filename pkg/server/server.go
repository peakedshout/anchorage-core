package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/peakedshout/anchorage-core/pkg/comm"
	"github.com/peakedshout/anchorage-core/pkg/config"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"github.com/peakedshout/go-pandorasbox/logger"
	"github.com/peakedshout/go-pandorasbox/tool/expired"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"github.com/peakedshout/go-pandorasbox/xnet/xmulti"
	"github.com/peakedshout/go-pandorasbox/xrpc"
	"net"
	"time"
)

type Server struct {
	nodeName string
	server   *xrpc.Server
	addrs    []net.Addr

	lm    *linkManager
	route *routeManager
	sm    *syncManager
	proxy *proxyManager

	logger logger.Logger
}

func NewServer(config *config.ServerConfig) (*Server, error) {
	return NewServerContext(context.Background(), config)
}

func NewServerContext(ctx context.Context, config *config.ServerConfig) (*Server, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cryptoList, err := comm.MakeCrypto(config.NodeInfo.Crypto)
	if err != nil {
		return nil, err
	}
	upgrader, err := comm.MakeUpgrader(config.NodeInfo.ExNetworks)
	if err != nil {
		return nil, err
	}
	nodeName := config.NodeInfo.NodeName
	if nodeName == "" {
		return nil, errors.New("nil node name")
	}
	sc := &xrpc.ServerConfig{
		Ctx:                      ctx,
		HandshakeTimeout:         time.Duration(config.NodeInfo.HandshakeTimeout) * time.Millisecond,
		SwitchNetworkSpeedTicker: true,
		SessionAuthCallback: func(info xrpc.AuthInfo) (xrpc.AuthInfo, error) {
			n, err := xrpc.GetAuthInfo[string](info, comm.NodeName)
			if err != nil || n != nodeName {
				return nil, ErrNodeAuthFailed
			}
			return nil, nil
		},
		CryptoList: cryptoList,
		Upgrader:   upgrader,
		CacheTime:  10 * time.Second,
	}
	if config.NodeInfo.Auth != nil {
		sc.SessionAuthCallback = xrpc.UPAuthCallback(config.NodeInfo.Auth.UserName, config.NodeInfo.Auth.Password)
	}
	server := xrpc.NewServer(sc)
	addrs := comm.MakeBaseAddress(config.NodeInfo.BaseNetwork)
	s := &Server{
		nodeName: config.NodeInfo.NodeName,
		server:   server,
		addrs:    addrs,
		logger:   logger.MustLogger(ctx),
	}
	s.lm = s.newLinkManager(config.LinkTimeout)
	s.route = s.newRoute(s.nodeName)
	err = s.newNodeManager(time.Duration(config.SyncTimeInterval)*time.Millisecond, config.SyncNodes)
	if err != nil {
		_ = server.Close()
		return nil, err
	}
	s.proxy = newProxyManager(s.sm.nm, expired.NewTODO(expired.Init(s.server.Context(), 1)), config.ProxyMulti)
	s.handle()
	return s, nil
}

func (s *Server) Serve() error {
	defer s.Close()
	ln, err := xmulti.NewMultiListenerFromAddr(s.server.Context(), true, s.addrs...)
	if err != nil {
		s.logger.Warn("server:", s.nodeName, " serve failed", err)
		return err
	}
	defer ln.Close()
	cl := s.sm.run()
	defer cl()
	s.logger.Info("server:", s.nodeName, "start serve ...")
	return s.server.Serve(ln)
}

func (s *Server) SyncServe(ch chan<- error) error {
	defer s.Close()
	ln, err := xmulti.NewMultiListenerFromAddr(s.server.Context(), true, s.addrs...)
	if err != nil {
		s.logger.Warn("server:", s.nodeName, "serve failed", err)
		ch <- err
		return err
	}
	defer ln.Close()
	cl := s.sm.run()
	defer cl()
	ch <- nil
	s.logger.Info("server:", s.nodeName, "start serve ...")
	return s.server.Serve(ln)
}

func (s *Server) Close() error {
	return s.server.Close()
}

func (s *Server) Context() context.Context {
	return s.server.Context()
}

func (s *Server) Logger() logger.Logger {
	return s.logger
}

func (s *Server) handle() {
	s.server.MustAddHandler(comm.CallRegister, s.handleListener)
	s.server.MustAddHandler(comm.CallLink, s.handleLink)
	s.server.MustAddHandler(comm.CallLinkReq, s.handleLinkReq)
	s.server.MustAddHandler(comm.CallSync, s.handleSync)
	s.server.MustAddHandler(comm.CallRouteView, s.handleRouteViewReq)
	s.server.MustAddHandler(comm.CallProxy, s.handleProxy)
}

func (s *Server) handleListener(ctx xrpc.ReverseRpc) error {
	var info comm.RegisterListenerInfo
	err := ctx.Bind(&info)
	if err != nil {
		return ErrRegisterListenerInfo.Errorf(err)
	}
	if info.Name == "" {
		return ErrRegisterListenerInfo.Errorf("nil service name")
	}
	unit := &localRouteUnit{
		ReverseRpc: ctx,
		info:       info,
	}
	s.route.setLocal(info.Name, unit)
	s.logger.Info("server:", s.nodeName, "add listener:", info.Name)
	defer func() {
		s.route.delLocal(info.Name, unit)
		s.logger.Info("server:", s.nodeName, "del listener:", info.Name)
	}()
	<-ctx.Context().Done()
	return nil
}

func (s *Server) handleLink(ctx xrpc.Stream) error {
	id, err := xrpc.GetStreamAuthInfoT[uint64](ctx.Context(), comm.KeyLinkId)
	if err != nil {
		return err
	}
	lid, err := xrpc.GetStreamAuthInfoT[string](ctx.Context(), comm.KeyLinkLId)
	if err != nil {
		return err
	}
	box := s.lm.getBox(id)
	if box == nil {
		return ErrInvalidLinkId
	}
	s.logger.Info("server:", s.nodeName, "link to id:", lid)
	return box.join(id, lid, ctx)
}

func (s *Server) handleLinkReq(ctx xrpc.Rpc) (any, error) {
	var info comm.LinkRequest
	err := ctx.Bind(&info)
	if err != nil {
		return nil, err
	}
	rrpc := s.route.get(&info)
	if rrpc == nil {
		return nil, ErrNotFoundService.Errorf(info.Link)
	}

	sid, _ := xrpc.GetSessionAuthInfoT[string](ctx.Context(), xrpc.SessionId)
	tid, _ := xrpc.GetSessionAuthInfoT[string](rrpc.Context(), xrpc.SessionId)
	binfo := &linkInfo{link: info.Link, stSessId: [2]string{sid, tid}, lid: uuid.NewId(1)}
	binfo.sit[1] = s.nodeName
	if unit, ok := rrpc.(*remoteRouteUnit); ok {
		binfo.sit[2] = unit.node
	}

	box := s.lm.newBox(rrpc.Context(), binfo)
	defer func() {
		if err != nil {
			box.doExpired()
		}
	}()
	info.BoxId = box.getId(true)
	info.BoxLId = binfo.lid
	var recv comm.LinkResponse
	err = rrpc.Rpc(ctx.Context(), comm.CallLinkReq, info, &recv)
	if err != nil {
		return nil, err
	}
	recv.BoxId = box.getId(false)
	recv.BoxLId = binfo.lid
	s.logger.Info("server:", s.nodeName, "link req:", info.Link, "id", binfo.lid)
	return recv, nil
}

func (s *Server) handleSync(ctx xrpc.ReverseRpc) error {
	var info syncInfo
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	if info.SourceNode == "" || info.TargetNode == "" {
		return ErrInvalidNode.Errorf("nil node name")
	}
	if info.TargetNode != s.nodeName {
		return ErrInvalidNode.Errorf("node name inconsistent")
	}
	ru := &remoteRouteUnit{
		node:       info.SourceNode,
		ReverseRpc: ctx,
		m:          make(map[string][]remoteRouteInfo),
	}
	s.route.setRemote(info.SourceNode, ru)
	defer s.route.delRemote(info.SourceNode, ru)
	s.logger.Info("server:", s.nodeName, "add sync node:", info.SourceNode)
	defer s.logger.Info("server:", s.nodeName, "del sync node:", info.SourceNode)
	return ctxtool.RunTimerFunc(ctx.Context(), s.sm.interval, func(ctx context.Context) error {
		var m map[string][]remoteRouteInfo
		err := ru.Rpc(ctx, comm.CallSyncMap, nil, &m)
		if err != nil {
			return err
		}
		ru.update(m)
		return nil
	})
}

func (s *Server) handleRouteViewReq(ctx xrpc.Rpc) (any, error) {
	view := s.route.getView(true)
	return view, nil
}

func (s *Server) handleProxy(ctx xrpc.Stream) error {
	var req comm.ProxyRequest
	err := ctx.Recv(&req)
	if err != nil {
		return err
	}
	tmpCtx, cl := context.WithCancel(ctx.Context())
	defer cl()
	stop := make(chan struct{})
	timer := time.NewTimer(s.proxy.timeout)
	go func() {
		defer timer.Stop()
		select {
		case <-timer.C:
			cl()
		case <-tmpCtx.Done():
		case <-stop:
		}
	}()

	node, nodes := req.GetNode(s.nodeName)
	if node == "" {
		conn, err := s.dialProxy(tmpCtx, &req)
		if err != nil {
			return err
		}
		defer conn.Close()
		close(stop)
		defer s.proxy.Record(ctx, s.nodeName, nodes, &req, nil, conn)()
		s.logger.Info("server:", s.nodeName, "proxy direct ->", fmt.Sprintf("%s_%s", req.Network, req.Address))
		go func() {
			defer conn.Close()
			var buf []byte
			for {
				err := ctx.Recv(&buf)
				if err != nil {
					return
				}
				_, err = conn.Write(buf)
				if err != nil {
					return
				}
			}
		}()
		buf := make([]byte, 32*1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return err
			}
			err = ctx.Send(buf[:n])
			if err != nil {
				return err
			}
		}
	} else {
		req.Node = nodes
		pctx := xrpc.SetClientShareStreamTmpClass(tmpCtx, s.proxy.cfg)
		_, stream, err := s.proxy.GetStream(pctx, node, comm.CallProxy, req)
		if err != nil {
			return err
		}
		defer stream.Close()
		close(stop)
		defer s.proxy.Record(ctx, s.nodeName, nodes, &req, stream, nil)()
		s.logger.Info("server:", s.nodeName, "proxy indirect ->", fmt.Sprintf("%s_%s", req.Network, req.Address))
		go func() {
			defer stream.Close()
			var buf []byte
			for {
				err := ctx.Recv(&buf)
				if err != nil {
					return
				}
				err = stream.Send(buf)
				if err != nil {
					return
				}
			}
		}()
		var buf []byte
		for {
			err = stream.Recv(&buf)
			if err != nil {
				return err
			}
			err = ctx.Send(buf)
			if err != nil {
				return err
			}
		}
	}
}
