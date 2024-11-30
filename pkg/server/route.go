package server

import (
	"context"
	"github.com/peakedshout/anchorage-core/pkg/comm"
	"github.com/peakedshout/go-pandorasbox/tool/mslice"
	"github.com/peakedshout/go-pandorasbox/xrpc"
	"sync"
	"time"
)

type remoteRouteInfo struct {
	comm.RegisterListenerInfo
	Delay time.Duration
}

type localRouteUnit struct {
	xrpc.ReverseRpc
	info comm.RegisterListenerInfo
}

func (lru *localRouteUnit) isLocal() bool {
	return true
}

type remoteRouteUnit struct {
	node string
	mux  sync.Mutex
	xrpc.ReverseRpc
	m map[string][]remoteRouteInfo
}

func (rru *remoteRouteUnit) isLocal() bool {
	return false
}

func (rru *remoteRouteUnit) update(m map[string][]remoteRouteInfo) {
	rru.mux.Lock()
	defer rru.mux.Unlock()
	rru.m = m
}

type routeManager struct {
	s         *Server
	localName string
	lMux      sync.Mutex
	lMap      map[string][]*localRouteUnit
	rMux      sync.Mutex
	rMap      map[string][]*remoteRouteUnit
}

func (s *Server) newRoute(localName string) *routeManager {
	r := &routeManager{
		s:         s,
		localName: localName,
		lMap:      make(map[string][]*localRouteUnit),
		rMap:      make(map[string][]*remoteRouteUnit),
	}
	return r
}

func (r *routeManager) get(info *comm.LinkRequest) xrpc.ReverseRpc {
	defer info.RemoveNode()
	if info.GetNode() == r.localName {
		return r.getLocal(info)
	}
	if info.GetNode() != "" {
		return r.getRemote(info)
	}
	local := r.getLocal(info)
	if local != nil {
		return local
	}
	return r.getRemote(info)
}

func (r *routeManager) getLocalMap() map[string][]remoteRouteInfo {
	r.lMux.Lock()
	defer r.lMux.Unlock()
	m := make(map[string][]remoteRouteInfo, len(r.lMap))
	for k, units := range r.lMap {
		l := make([]remoteRouteInfo, 0, len(units))
		for _, unit := range units {
			l = append(l, remoteRouteInfo{
				RegisterListenerInfo: unit.info,
				Delay:                r.getDelayByCtx(unit.ReverseRpc.Context()),
			})
		}
		m[k] = l
	}
	return m
}

func (r *routeManager) getLocal(info *comm.LinkRequest) xrpc.ReverseRpc {
	r.lMux.Lock()
	ll, ok := r.lMap[info.Link]
	if ok {
		list := mslice.MakeRandRangeSlice(0, len(ll))
		for _, i := range list {
			ctx := ll[i]
			if ctx.Context().Err() == nil && ctx.info.Equal(info) {
				r.lMux.Unlock()
				return ctx
			}
		}
	}
	r.lMux.Unlock()
	return nil
}

func (r *routeManager) setLocal(name string, ctx *localRouteUnit) {
	r.lMux.Lock()
	defer r.lMux.Unlock()
	r.lMap[name] = append(r.lMap[name], ctx)
}

func (r *routeManager) delLocal(name string, ctx *localRouteUnit) {
	r.lMux.Lock()
	defer r.lMux.Unlock()
	for i, l := range r.lMap[name] {
		if l == ctx {
			r.lMap[name] = append(r.lMap[name][:i], r.lMap[name][i+1:]...)
			break
		}
	}
}

func (r *routeManager) getRemote(info *comm.LinkRequest) xrpc.ReverseRpc {
	if info.GetNode() != "" {
		r.rMux.Lock()
		defer r.rMux.Unlock()
		units, ok := r.rMap[info.GetNode()]
		if ok {
			return r.getRemote2(units, info)
		}
		return nil
	}
	r.rMux.Lock()
	defer r.rMux.Unlock()
	for _, units := range r.rMap {
		ctx := r.getRemote2(units, info)
		if ctx != nil {
			return ctx
		}
	}
	return nil
}

func (r *routeManager) getRemote2(units []*remoteRouteUnit, info *comm.LinkRequest) xrpc.ReverseRpc {
	list := mslice.MakeRandRangeSlice(0, len(units))
	for _, i := range list {
		unit := units[i]
		unit.mux.Lock()
		if lnInfos, ok := unit.m[info.Link]; ok {
			for _, lnInfo := range lnInfos {
				if lnInfo.Equal(info) && unit.ReverseRpc.Context().Err() == nil {
					unit.mux.Unlock()
					return unit
				}
			}
		}
		unit.mux.Unlock()
	}
	return nil
}

func (r *routeManager) setRemote(node string, unit *remoteRouteUnit) {
	r.rMux.Lock()
	defer r.rMux.Unlock()
	r.rMap[node] = append(r.rMap[node], unit)
}

func (r *routeManager) delRemote(node string, unit *remoteRouteUnit) {
	r.rMux.Lock()
	defer r.rMux.Unlock()
	for i, l := range r.rMap[node] {
		if l == unit {
			r.rMap[node] = append(r.rMap[node][:i], r.rMap[node][i+1:]...)
			break
		}
	}
}

func (r *routeManager) getView(hide bool) *comm.ServiceRouteView {
	nv := make(map[string]map[string][]comm.ServiceRouteViewUnit)
	sv := make(map[string][]comm.ServiceRouteViewUnit)
	r.lMux.Lock()
	nv[r.localName] = make(map[string][]comm.ServiceRouteViewUnit)
	for _, units := range r.lMap {
		for _, unit := range units {
			if hide && unit.info.Settings.SwitchHide {
				continue
			}
			info := comm.ServiceRouteViewUnit{
				Name:     unit.info.Name,
				Node:     r.localName,
				Notes:    unit.info.Notes,
				Auth:     unit.info.Auth != nil,
				Settings: unit.info.Settings,
				Delay:    r.getDelayByCtx(unit.ReverseRpc.Context()),
			}
			nv[r.localName][info.Name] = append(nv[r.localName][info.Name], info)
			sv[info.Name] = append(sv[info.Name], info)
		}
	}
	r.lMux.Unlock()
	r.rMux.Lock()
	for node, units := range r.rMap {
		nv[node] = make(map[string][]comm.ServiceRouteViewUnit)
		for _, unit := range units {
			baseDelay := r.getDelayByCtx(unit.ReverseRpc.Context())
			unit.mux.Lock()
			for _, infos := range unit.m {
				for _, one := range infos {
					if hide && one.Settings.SwitchHide {
						continue
					}
					info := comm.ServiceRouteViewUnit{
						Name:     one.Name,
						Node:     node,
						Notes:    one.Notes,
						Auth:     one.Auth != nil,
						Settings: one.Settings,
						Delay:    baseDelay + one.Delay,
					}
					nv[node][info.Name] = append(nv[node][info.Name], info)
					sv[info.Name] = append(sv[info.Name], info)
				}
			}
			unit.mux.Unlock()
		}
	}
	r.rMux.Unlock()
	return &comm.ServiceRouteView{
		NodeView:    nv,
		ServiceView: sv,
	}
}

func (r *routeManager) getDelayByCtx(ctx context.Context) time.Duration {
	sid, err := xrpc.GetSessionAuthInfoT[string](ctx, xrpc.SessionId)
	if err != nil {
		return 0
	}
	return r.s.server.GetDelay(sid)
}
