package server

import (
	"context"
	"github.com/peakedshout/anchorage-core/pkg/comm"
	"github.com/peakedshout/anchorage-core/pkg/config"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"github.com/peakedshout/go-pandorasbox/xrpc"
	"time"
)

type syncManager struct {
	s        *Server
	nm       map[string][]*comm.NodeUnit
	interval time.Duration
}

func (s *Server) newNodeManager(interval time.Duration, nodes []config.NodeConfig) error {
	nm, err := comm.MakeNodeUnit(s.server.Context(), nodes)
	if err != nil {
		return err
	}
	if interval < 1*time.Second {
		interval = 1 * time.Second
	}
	s.sm = &syncManager{
		s:        s,
		nm:       nm,
		interval: interval,
	}
	return nil
}

func (sm *syncManager) run() context.CancelFunc {
	ctx, cl := context.WithCancel(sm.s.server.Context())
	for _, units := range sm.nm {
		for _, unit := range units {
			go sm.syncNode(ctx, unit)
		}
	}
	return cl
}

func (sm *syncManager) syncNode(ctx context.Context, nu *comm.NodeUnit) {
	for ctx.Err() == nil {
		_ = sm.syncNodeOnce(ctx, nu)
		time.Sleep(1 * time.Second)
	}
}

func (sm *syncManager) syncNodeOnce(nCtx context.Context, nu *comm.NodeUnit) error {
	xCtx := xrpc.SetClientShareStreamClass(nCtx, xrpc.ClientNotShareStream)
	err := nu.ReverseRpc(xCtx, comm.CallSync, syncInfo{
		SourceNode: sm.s.nodeName,
		TargetNode: nu.Node,
	}, map[string]xrpc.ClientReverseRpcHandler{
		comm.CallLinkReq: func(ctx xrpc.ClientReverseRpcContext) (any, error) {
			var info comm.LinkRequest
			err := ctx.Bind(&info)
			if err != nil {
				return nil, err
			}
			rrpc := sm.s.route.get(&info)
			if rrpc == nil {
				return nil, ErrNotFoundService.Errorf(info.Link)
			}
			sourceId := info.BoxId
			sourceLId := info.BoxLId

			sid, _ := xrpc.GetSessionAuthInfoT[string](ctx.Context(), xrpc.SessionId)
			tid, _ := xrpc.GetSessionAuthInfoT[string](rrpc.Context(), xrpc.SessionId)
			binfo := &linkInfo{link: info.Link, stSessId: [2]string{sid, tid}, lid: uuid.NewId(1)}
			binfo.sit[0] = nu.Node
			binfo.sit[1] = sm.s.nodeName
			if unit, ok := rrpc.(*remoteRouteUnit); ok {
				binfo.sit[2] = unit.node
			}

			box := sm.s.lm.newBox(rrpc.Context(), binfo)
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

			sCtx, err := xrpc.SetStreamAuthInfoT[uint64](nCtx, comm.KeyLinkId, sourceId)
			if err != nil {
				return nil, err
			}
			sCtx, err = xrpc.SetStreamAuthInfoT[string](sCtx, comm.KeyLinkLId, sourceLId)
			if err != nil {
				return nil, err
			}
			stream, err := nu.Stream(sCtx, comm.CallLink)
			if err != nil {
				return nil, err
			}
			err = stream.Send(nil)
			if err != nil {
				_ = stream.Close()
				return nil, err
			}
			go func() {
				defer stream.Close()
				_ = box.join(box.getId(false), binfo.lid, stream)
			}()
			return recv, nil
		},
		comm.CallSyncMap: func(context xrpc.ClientReverseRpcContext) (any, error) {
			return sm.s.route.getLocalMap(), nil
		},
	})
	return err
}

type syncInfo struct {
	SourceNode string
	TargetNode string
}
