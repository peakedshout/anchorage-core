package server

import (
	"github.com/peakedshout/anchorage-core/pkg/comm"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"github.com/peakedshout/go-pandorasbox/xrpc"
	"sort"
)

func (s *Server) GetLinkView() []comm.LinkView {
	s.lm.lock.Lock()
	defer s.lm.lock.Unlock()
	list := make([]comm.LinkView, 0, len(s.lm.boxMap))
	for id, box := range s.lm.boxMap {
		box.mux.Lock()
		lv := comm.LinkView{
			Id:              id,
			Link:            box.info.link,
			Status:          comm.LinkStatusWork,
			SIT:             box.info.sit,
			InitSTSessionId: box.info.stSessId,
			WorkSTStreamId:  box.info.ppStreamId,
			LinkId:          box.info.lid,
			LinkList:        box.info.lList,
			NodeList:        box.info.nList,
		}
		if box.p1 == nil || box.p2 == nil {
			lv.Status = comm.LinkStatusWait
		} else if !box.info.initialized {
			lv.Status = comm.LinkStatusInit
		} else if box.p1.Context().Err() != nil || box.p2.Context().Err() != nil {
			lv.Status = comm.LinkStatusDead
		}
		if box.p1 != nil {
			lv.WorkSTMonitorInfo[0] = xnetutil.FormatMonitorInfo(box.p1)
		}
		if box.p2 != nil {
			lv.WorkSTMonitorInfo[1] = xnetutil.FormatMonitorInfo(box.p2)
		}
		list = append(list, lv)
		box.mux.Unlock()
	}
	return list
}

func (s *Server) GetRouteView() *comm.ServiceRouteView {
	return s.route.getView(false)
}

func (s *Server) GetServerSessionView() []xrpc.SessionView {
	return s.server.SessionView()
}

func (s *Server) GetSyncView() map[string][]xrpc.SessionView {
	m := make(map[string][]xrpc.SessionView)
	for node, units := range s.sm.nm {
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
	FromNetwork   string
	FromAddress   string
	ToNetwork     string
	ToAddress     string
	TargetNetwork string
	TargetAddress string
}

func (s *Server) GetProxyView() map[string][]ProxyUnitView {
	return s.proxy.View()
}
