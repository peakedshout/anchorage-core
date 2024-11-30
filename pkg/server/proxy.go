package server

import (
	"context"
	"errors"
	"github.com/peakedshout/anchorage-core/pkg/comm"
	"github.com/peakedshout/go-pandorasbox/tool/expired"
	"github.com/peakedshout/go-pandorasbox/tool/tmap"
	"github.com/peakedshout/go-pandorasbox/xnet/xquic"
	"github.com/peakedshout/go-pandorasbox/xrpc"
	"net"
	"sort"
	"time"
)

func (s *Server) dialProxy(ctx context.Context, req *comm.ProxyRequest) (net.Conn, error) {
	switch req.Network {
	case "tcp", "udp":
		dr := new(net.Dialer)
		return dr.DialContext(ctx, req.Network, req.Address)
	case "quic":
		return xquic.DialContext(ctx, req.Network, req.Address)
	default:
		return nil, errors.New("invalid network")
	}
}

func newProxyManager(nm map[string][]*comm.NodeUnit, cache *expired.TODO, multi int) *proxyManager {
	return &proxyManager{
		StreamManager: comm.NewStreamManager(nm),
		cache:         cache,
		timeout:       10 * time.Second,
		cfg:           xrpc.NewTmpShareStreamConfig(false, multi),
	}
}

type proxyManager struct {
	*comm.StreamManager
	m       tmap.SyncMap[xrpc.Stream, *proxyInfo]
	cache   *expired.TODO
	timeout time.Duration
	cfg     *xrpc.ShareStreamConfig
}

type proxyInfo struct {
	nodes []string
	node  string

	lnk, laddr, rnk, raddr, onk, oaddr string
}

func (pm *proxyManager) Record(l xrpc.Stream, node string, nodes []string, req *comm.ProxyRequest, t xrpc.Stream, conn net.Conn) func() {
	info := &proxyInfo{node: node, nodes: nodes, onk: req.Network, oaddr: req.Address}
	info.laddr, _ = xrpc.GetSessionAuthInfoT[string](l.Context(), xrpc.LocalPubAddress)
	info.lnk, _ = xrpc.GetSessionAuthInfoT[string](l.Context(), xrpc.LocalPubNetwork)
	if t != nil {
		info.raddr, _ = xrpc.GetSessionAuthInfoT[string](l.Context(), xrpc.RemotePubAddress)
		info.rnk, _ = xrpc.GetSessionAuthInfoT[string](l.Context(), xrpc.RemotePubNetwork)
	}
	if conn != nil {
		info.raddr = conn.RemoteAddr().String()
		info.rnk = conn.RemoteAddr().Network()
	}
	pm.m.Store(l, info)
	return func() {
		fn := func() {
			pm.m.Delete(l)
		}
		if pm.cache != nil {
			pm.cache.Duration(10*time.Second, fn)
		} else {
			fn()
		}
	}
}

func (pm *proxyManager) View() map[string][]ProxyUnitView {
	m := make(map[string][]ProxyUnitView)
	pm.m.Range(func(stream xrpc.Stream, info *proxyInfo) bool {
		m[info.node] = append(m[info.node], ProxyUnitView{
			StreamView:    xrpc.GetStreamView(stream),
			Nodes:         info.nodes,
			Node:          info.node,
			FromNetwork:   info.lnk,
			FromAddress:   info.laddr,
			ToNetwork:     info.rnk,
			ToAddress:     info.raddr,
			TargetNetwork: info.onk,
			TargetAddress: info.oaddr,
		})
		return true
	})
	for _, list := range m {
		sort.SliceStable(list, func(i, j int) bool {
			return list[i].MonitorInfo.CreateTime.Before(list[j].MonitorInfo.CreateTime)
		})
	}
	return m
}
