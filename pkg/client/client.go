package client

import (
	"context"
	"github.com/peakedshout/anchorage-core/pkg/comm"
	"github.com/peakedshout/anchorage-core/pkg/config"
	"github.com/peakedshout/go-pandorasbox/logger"
	"github.com/peakedshout/go-pandorasbox/tool/tmap"
	"github.com/peakedshout/go-pandorasbox/xrpc"
	"net"
	"time"
)

type Client struct {
	ctx      context.Context
	cl       context.CancelFunc
	launcher *launcher
	proxy    tmap.SyncMap[any, *ProxyDialer]

	logger logger.Logger
}

func NewClient(config *config.ClientConfig) (*Client, error) {
	return NewClientContext(context.Background(), config)
}

func NewClientContext(ctx context.Context, config *config.ClientConfig) (*Client, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(config.Nodes) == 0 {
		return nil, ErrNilNodes
	}
	nCtx, cl := context.WithCancel(ctx)
	nm, err := comm.MakeNodeUnit(nCtx, config.Nodes)
	if err != nil {
		cl()
		return nil, err
	}
	c := &Client{
		ctx:    nCtx,
		cl:     cl,
		logger: logger.MustLogger(ctx),
	}
	c.newLauncher(nm)
	return c, nil
}

func (c *Client) Close() error {
	c.cl()
	return nil
}

func (c *Client) Serve(ctx context.Context, cfg comm.RegisterListenerInfo, fn func(conn net.Conn) error) error {
	return c.serve(ctx, cfg, fn)
}

func (c *Client) serve(ctx context.Context, cfg comm.RegisterListenerInfo, fn func(conn net.Conn) error) error {
	sctx := xrpc.SetClientShareStreamClass(ctx, xrpc.ClientNotShareStream)
	return c.launcher.nodeCallBack(ctx, 1*time.Second, cfg.Node, func(ctx context.Context, nu *comm.NodeUnit) error {
		c.logger.Info("client:", "start listener:", cfg.Name)
		defer c.logger.Info("client:", "failed listener:", cfg.Name)
		_ = nu.ReverseRpc(sctx, comm.CallRegister, cfg, map[string]xrpc.ClientReverseRpcHandler{
			comm.CallLinkReq: func(rpcContext xrpc.ClientReverseRpcContext) (any, error) {
				var info comm.LinkRequest
				err := rpcContext.Bind(&info)
				if err != nil {
					return nil, err
				}
				if !cfg.Equal(&info) {
					return nil, ErrLinkRefuse
				}
				pinfo := &p2pInfo{isListener: true}
				network := selectP2PNetwork(nu, &cfg, &info)
				if info.ForceP2P && network == "" {
					return nil, ErrLinkRefuse
				}
				pinfo.network = network

				sCtx, err := xrpc.SetStreamAuthInfoT[uint64](xrpc.CloneSessionAuthInfo(rpcContext.Context(), ctx), comm.KeyLinkId, info.BoxId)
				if err != nil {
					return nil, err
				}
				sCtx, err = xrpc.SetStreamAuthInfoT[string](sCtx, comm.KeyLinkLId, info.BoxLId)
				if err != nil {
					return nil, err
				}
				stream, err := c.launcher.handleLink(sCtx, nu, pinfo)
				if err != nil {
					return nil, err
				}
				c.logger.Info("client:", "listener link req:", info.Link, "id", info.BoxLId)
				go func() {
					conn, err := c.launcher.handleConn(stream, pinfo)
					if err != nil {
						_ = stream.Close()
						return
					}
					err = fn(conn)
					if err != nil {
						_ = conn.Close()
						return
					}
				}()
				return comm.LinkResponse{P2PNetwork: network}, nil
			},
		})
		return nil
	})
}

func (c *Client) Dial(ctx context.Context, cfg comm.LinkRequest) (x net.Conn, err error) {
	units := c.launcher.selectNodeUnit(&cfg)
	var recv comm.LinkResponse
	nu, err := c.launcher.rpc(ctx, units, comm.CallLinkReq, cfg, &recv)
	if err != nil {
		return nil, err
	}
	if cfg.ForceP2P && recv.P2PNetwork == "" {
		return nil, ErrLinkRefuse
	}

	pinfo := &p2pInfo{isListener: false, network: recv.P2PNetwork}
	sCtx, err := xrpc.SetStreamAuthInfoT[uint64](xrpc.CloneSessionAuthInfo(nu.Context(), ctx), comm.KeyLinkId, recv.BoxId)
	if err != nil {
		return nil, err
	}
	sCtx, err = xrpc.SetStreamAuthInfoT[string](sCtx, comm.KeyLinkLId, recv.BoxLId)
	if err != nil {
		return nil, err
	}
	stream, err := c.launcher.handleLink(sCtx, nu, pinfo)
	if err != nil {
		return nil, err
	}
	conn, err := c.launcher.handleConn(stream, pinfo)
	if err != nil {
		_ = stream.Close()
		return nil, err
	}
	c.logger.Info("client:", "dialer link req:", cfg.Link, "id", recv.BoxLId)
	return conn, nil
}

func (c *Client) Listen(ctx context.Context, cfg comm.RegisterListenerInfo) net.Listener {
	ln := c.newListener(ctx, nil)
	go c.serve(ln.ctx, cfg, func(conn net.Conn) error {
		return ln.addConn(conn)
	})
	return ln
}

func (c *Client) GetServiceRoute(ctx context.Context) (*comm.ServiceRouteView, error) {
	var info comm.ServiceRouteView
	_, err := c.launcher.rpcByNode(ctx, "", comm.CallRouteView, nil, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (c *Client) Context() context.Context {
	return c.ctx
}
