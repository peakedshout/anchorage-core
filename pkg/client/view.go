package client

import (
	"github.com/peakedshout/go-pandorasbox/xrpc"
	"sort"
)

func (c *Client) GetClientSessionView() map[string][]xrpc.SessionView {
	m := make(map[string][]xrpc.SessionView)
	for node, units := range c.launcher.nm {
		for _, unit := range units {
			m[node] = append(m[node], unit.SessionView()...)
		}
	}
	for _, list := range m {
		sort.SliceStable(list, func(i, j int) bool {
			return list[i].MonitorInfo.CreateTime.Before(list[j].MonitorInfo.CreateTime)
		})
	}
	return m
}

type ProxyUnitView struct {
	xrpc.StreamView
	Nodes         []string
	Node          string
	LocalNetwork  string
	LocalAddress  string
	RemoteNetwork string
	RemoteAddress string
}

func (c *Client) GetProxyView() map[string][]ProxyUnitView {
	m := make(map[string][]ProxyUnitView)
	c.proxy.Range(func(_ any, value *ProxyDialer) bool {
		view := value.View()
		if len(view) == 0 {
			return true
		}
		m[view[0].Node] = view
		return true
	})
	return m
}
