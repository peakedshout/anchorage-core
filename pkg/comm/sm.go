package comm

import (
	"context"
	"errors"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"github.com/peakedshout/go-pandorasbox/tool/mslice"
	"github.com/peakedshout/go-pandorasbox/xrpc"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

func NewSessionManager(ctx context.Context, max int32, nm map[string][]*NodeUnit) *SessionManager {
	if max < 1 {
		max = 1
	}
	sm := &SessionManager{
		ctx:     ctx,
		max:     max,
		nm:      nm,
		mux:     sync.Mutex{},
		sessMap: make(map[string][]*sessionUnit),
	}
	go sm.run()
	return sm
}

type SessionManager struct {
	ctx     context.Context
	max     int32
	nm      map[string][]*NodeUnit
	mux     sync.Mutex
	sessMap map[string][]*sessionUnit
}

func (sm *SessionManager) GetView() map[string][]xrpc.SessionView {
	m := make(map[string][]xrpc.SessionView)
	for node, units := range sm.nm {
		for _, unit := range units {
			m[node] = append(m[node], unit.SessionView()...)
		}
	}
	for _, list := range m {
		sort.SliceStable(list, func(i, j int) bool {
			return list[i].Id < list[j].Id
		})
	}
	return m
}

func (sm *SessionManager) GetStream(ctx context.Context, node string, header string, req any) (xrpc.Stream, error) {
	return sm.getStream(ctx, node, header, req)
}

func (sm *SessionManager) run() {
	_ = ctxtool.RunTimerFunc(sm.ctx, 30*time.Second, func(ctx context.Context) error {
		sm.mux.Lock()
		defer sm.mux.Unlock()
		for node, units := range sm.sessMap {
			l := make([]*sessionUnit, 0, len(units))
			for _, unit := range units {
				if unit.streamNum == 0 {
					_ = unit.ClientSession.Close()
				} else {
					l = append(l, unit)
				}
			}
			sm.sessMap[node] = l
		}
		return nil
	})
}

type sessionUnit struct {
	*xrpc.ClientSession
	streamNum int32
}

type streamUnit struct {
	xrpc.Stream
	closer  sync.Once
	closeFn func()
}

func (su *streamUnit) Close() error {
	su.closer.Do(su.closeFn)
	return su.Stream.Close()
}

func (sm *SessionManager) getStream(ctx context.Context, node string, header string, req any) (xrpc.Stream, error) {
	var session *sessionUnit
	var err error
	if node == "" {
		session, err = sm.getSession2(ctx)
	} else {
		session, err = sm.getSession(ctx, node)
	}
	if err != nil {
		return nil, err
	}
	fn := func() {
		atomic.AddInt32(&session.streamNum, -1)
	}
	stream, err := session.Stream(ctx, header)
	if err != nil {
		fn()
		return nil, err
	}
	err = stream.Send(req)
	if err != nil {
		fn()
		return nil, err
	}
	return &streamUnit{
		Stream:  stream,
		closer:  sync.Once{},
		closeFn: fn,
	}, nil
}

func (sm *SessionManager) getSession2(ctx context.Context) (*sessionUnit, error) {
	sm.mux.Lock()
	defer sm.mux.Unlock()
	for _, units := range sm.sessMap {
		for _, unit := range units {
			now := atomic.AddInt32(&unit.streamNum, 1)
			if now > sm.max || unit.ClientSession.Context().Err() != nil {
				atomic.AddInt32(&unit.streamNum, -1)
				continue
			}
			return unit, nil
		}
	}
	for _, units := range sm.nm {
		list := mslice.MakeRandRangeSlice(0, len(units))
		for _, i := range list {
			if sm.ctx.Err() != nil {
				return nil, sm.ctx.Err()
			}
			nu := units[i]
			session, err := sm.newSession(ctx, nu)
			if err == nil {
				unit := &sessionUnit{
					ClientSession: session,
					streamNum:     1,
				}
				return unit, nil
			}
		}
	}
	return nil, errors.New("get session failed")
}

func (sm *SessionManager) getSession(ctx context.Context, node string) (*sessionUnit, error) {
	sm.mux.Lock()
	defer sm.mux.Unlock()
	units, ok := sm.sessMap[node]
	if ok {
		for _, unit := range units {
			now := atomic.AddInt32(&unit.streamNum, 1)
			if now > sm.max || unit.ClientSession.Context().Err() != nil {
				atomic.AddInt32(&unit.streamNum, -1)
				continue
			}
			return unit, nil
		}
	}
	nodeUnits, ok := sm.nm[node]
	if ok {
		list := mslice.MakeRandRangeSlice(0, len(nodeUnits))
		for _, i := range list {
			if sm.ctx.Err() != nil {
				return nil, sm.ctx.Err()
			}
			nu := nodeUnits[i]
			session, err := sm.newSession(ctx, nu)
			if err == nil {
				unit := &sessionUnit{
					ClientSession: session,
					streamNum:     1,
				}
				return unit, nil
			}
		}
	}
	return nil, errors.New("get session failed")
}

func (sm *SessionManager) newSession(ctx context.Context, nu *NodeUnit) (*xrpc.ClientSession, error) {
	conn, err := nu.Dr.MultiDialContext(ctx, nu.Addrs...)
	if err != nil {
		return nil, err
	}
	sCtx, _ := xrpc.SetSessionAuthInfoT[string](sm.ctx, NodeName, nu.Node)
	session, err := nu.WithConn(sCtx, conn)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return session, nil
}
