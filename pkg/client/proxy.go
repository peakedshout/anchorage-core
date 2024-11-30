package client

import (
	"context"
	"fmt"
	"github.com/peakedshout/anchorage-core/pkg/comm"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"github.com/peakedshout/go-pandorasbox/logger"
	"github.com/peakedshout/go-pandorasbox/tool/expired"
	"github.com/peakedshout/go-pandorasbox/tool/tmap"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"github.com/peakedshout/go-pandorasbox/xrpc"
	"net"
	"sort"
	"time"
)

const ProxyNodes = "proxyNodes"

type ProxyDialer struct {
	ctx     context.Context
	cl      context.CancelFunc
	header  string
	proxy   *comm.StreamManager
	m       tmap.SyncMap[xrpc.Stream, *proxyInfo]
	cache   *expired.TODO
	timeout time.Duration

	cfg *xrpc.ShareStreamConfig

	logger logger.Logger
}

type proxyInfo struct {
	nodes        []string
	node         string
	laddr, raddr net.Addr
}

func (p *ProxyDialer) Dial(network string, addr string) (net.Conn, error) {
	return p.DialContext(context.Background(), network, addr)
}

func (p *ProxyDialer) DialContext(ctx context.Context, network string, addr string) (net.Conn, error) {
	var nodes []string
	value := ctx.Value(ProxyNodes)
	if value != nil {
		nodes, _ = value.([]string)
	}
	stop := make(chan struct{})
	timer := time.NewTimer(p.timeout)
	defer timer.Stop()
	tmpCtx, cl := context.WithCancel(context.Background())
	go func() {
		select {
		case <-timer.C:
			cl()
		case <-ctx.Done():
			cl()
		case <-stop:
		}
	}()

	n := ""
	if len(nodes) > 0 {
		n = nodes[0]
	}
	pctx := xrpc.SetClientShareStreamTmpClass(tmpCtx, p.cfg)
	nu, stream, err := p.proxy.GetStream(pctx, n, p.header, comm.ProxyRequest{
		Node:    nodes,
		Network: network,
		Address: addr,
	})
	if err != nil {
		cl()
		return nil, err
	}
	close(stop)
	prin, _ := xrpc.GetSessionAuthInfoT[string](stream.Context(), xrpc.LocalPriNetwork)
	pria, _ := xrpc.GetSessionAuthInfoT[string](stream.Context(), xrpc.LocalPriAddress)
	laddr := xnetutil.NewNetAddr(prin, pria)
	raddr := xnetutil.NewNetAddr(network, addr)
	p.m.Store(stream, &proxyInfo{
		nodes: nodes,
		node:  nu.Node,
		laddr: laddr,
		raddr: raddr,
	})
	nConn := newConn(stream, laddr, raddr, func() {
		cl()
		fn := func() {
			p.m.Delete(stream)
		}
		if p.cache != nil {
			p.cache.Duration(10*time.Second, func() {
				fn()
			})
		} else {
			fn()
		}
	})
	ctxtool.GWaitFunc(stream.Context(), func() {
		_ = nConn.Close()
	})
	p.logger.Info("client:", "proxy direct ->", fmt.Sprintf("%s_%s", network, addr))
	return nConn, nil
}

func (p *ProxyDialer) Close() error {
	err := p.ctx.Err()
	p.cl()
	return err
}

func (p *ProxyDialer) View() []ProxyUnitView {
	var list []ProxyUnitView
	p.m.Range(func(stream xrpc.Stream, info *proxyInfo) bool {
		list = append(list, ProxyUnitView{
			StreamView:    xrpc.GetStreamView(stream),
			Nodes:         info.nodes,
			Node:          info.node,
			LocalNetwork:  info.laddr.Network(),
			LocalAddress:  info.laddr.String(),
			RemoteNetwork: info.raddr.Network(),
			RemoteAddress: info.raddr.String(),
		})
		return true
	})
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].MonitorInfo.CreateTime.Before(list[j].MonitorInfo.CreateTime)
	})
	return list
}

func (c *Client) Proxy(multi int) *ProxyDialer {
	ctx, cl := context.WithCancel(c.ctx)
	manager := comm.NewStreamManager(c.launcher.nm)
	pd := &ProxyDialer{
		ctx:     ctx,
		header:  comm.CallProxy,
		proxy:   manager,
		cache:   expired.NewTODO(expired.Init(ctx, 1)),
		timeout: 10 * time.Second,
		cfg:     xrpc.NewTmpShareStreamConfig(false, multi),
		logger:  c.logger,
	}
	c.proxy.Store(pd, pd)
	pd.cl = func() {
		cl()
		c.proxy.Delete(pd)
	}
	return pd
}
