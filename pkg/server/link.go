package server

import (
	"context"
	"errors"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"github.com/peakedshout/go-pandorasbox/tool/expired"
	"github.com/peakedshout/go-pandorasbox/xrpc"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

func (s *Server) newLinkManager(timeout uint) *linkManager {
	if timeout < 5*1000 {
		timeout = 5 * 1000
	}
	lm := &linkManager{
		s:           s,
		ctx:         s.server.Context(),
		boxMap:      make(map[uint64]*linkBox),
		linkTimeout: time.Duration(timeout) * time.Millisecond,
		expiredCtx:  expired.Init(s.server.Context(), runtime.NumCPU()),
	}
	lm.cache = expired.NewTODO(lm.expiredCtx)
	return lm
}

type linkManager struct {
	s      *Server
	ctx    context.Context
	lock   sync.Mutex
	ptr    uint64
	boxMap map[uint64]*linkBox

	linkTimeout time.Duration
	expiredCtx  *expired.ExpiredCtx
	cache       *expired.TODO
}

func (lm *linkManager) getBox(id uint64) *linkBox {
	if id == 0 {
		return nil
	}
	lm.lock.Lock()
	defer lm.lock.Unlock()
	return lm.boxMap[id&^(1<<63)]
}

func (lm *linkManager) newBox(rctx context.Context, info *linkInfo) *linkBox {
	var id uint64
	for id == 0 {
		id = atomic.AddUint64(&lm.ptr, 1)
	}
	lm.lock.Lock()
	defer lm.lock.Unlock()
	ctx, cl := context.WithCancel(rctx)
	box := &linkBox{
		lm:   lm,
		id:   id,
		info: info,
		ctx:  ctx,
		cl:   cl,
		run:  make(chan struct{}),
	}
	lm.boxMap[id] = box
	lm.expiredCtx.SetWithDuration(box, lm.linkTimeout)
	ctxtool.GWaitFunc(ctx, box.ExpiredFunc)
	return box
}

type linkInfo struct {
	link        string
	sit         [3]string // source itself target
	stSessId    [2]string // source target init sess id
	ppStreamId  [2]string // source target p sess_stream id
	initialized bool
	lid         string // identification bit instead of reliable uuid
	lList       []string
	nList       []string
}

type linkBox struct {
	lm *linkManager
	id uint64

	info *linkInfo

	ctx context.Context
	cl  context.CancelFunc

	run chan struct{}

	closer sync.Once
	mux    sync.Mutex
	p1     xrpc.Stream // from
	p2     xrpc.Stream // to
}

func (lb *linkBox) Id() any {
	return lb.getId(false)
}

func (lb *linkBox) getId(r bool) uint64 {
	if r {
		return lb.id | (1 << 63)
	} else {
		return lb.id &^ (1 << 63)
	}
}

func (lb *linkBox) ExpiredFunc() {
	lb.closer.Do(func() {
		lb.cl()
		lb.mux.Lock()
		if lb.p1 != nil {
			_ = lb.p1.Close()
		}
		if lb.p2 != nil {
			_ = lb.p2.Close()
		}
		lb.mux.Unlock()
		fn := func() {
			lb.lm.lock.Lock()
			delete(lb.lm.boxMap, lb.id)
			lb.lm.lock.Unlock()
		}
		if lb.lm.cache != nil {
			lb.lm.cache.Duration(10*time.Second, fn)
		} else {
			fn()
		}
	})
}

func (lb *linkBox) doExpired() {
	lb.lm.expiredCtx.Remove(lb.id, true)
}

func (lb *linkBox) delExpired() {
	lb.lm.expiredCtx.Remove(lb.id, false)
}

func (lb *linkBox) join(id uint64, lid string, sess xrpc.Stream) error {
	if lb.ctx.Err() != nil || lb.info.lid != lid {
		return ErrInvalidLinkId
	}
	s := false
	isP2 := (id>>63)&1 == 1
	lb.mux.Lock()
	select {
	case <-lb.run:
		lb.mux.Unlock()
		return ErrInvalidLinkId
	default:
	}
	if isP2 {
		if lb.p2 != nil {
			lb.mux.Unlock()
			return ErrInvalidLinkId
		}
		lb.info.ppStreamId[0] = sess.Id()
		lb.p2 = sess
		if lb.p1 != nil {
			lb.delExpired()
			s = true
			//close(lb.run)
		}
	} else {
		if lb.p1 != nil {
			lb.mux.Unlock()
			return ErrInvalidLinkId
		}
		lb.info.ppStreamId[1] = sess.Id()
		lb.p1 = sess
		if lb.p2 != nil {
			lb.delExpired()
			s = true
			//close(lb.run)
		}
	}
	lb.mux.Unlock()
	defer lb.ExpiredFunc()
	if s {
		// transmission link list
		err := lb.checkLinkList()
		if err != nil {
			return err
		}
		close(lb.run)
	}
	select {
	case <-sess.Context().Done():
		return sess.Context().Err()
	case <-lb.ctx.Done():
		return lb.ctx.Err()
	case <-lb.run:
	}

	var obj xrpc.Stream
	if isP2 {
		obj = lb.p1
	} else {
		obj = lb.p2
	}

	var b []byte
	for {
		err := sess.Recv(&b)
		if err != nil {
			return err
		}
		err = obj.Send(b)
		if err != nil {
			return err
		}
	}
}

func (lb *linkBox) checkLinkList() (err error) {
	info := linkCheckInfo{}
	defer func() {
		lb.mux.Lock()
		defer lb.mux.Unlock()
		lb.info.initialized = true
		lb.info.nList = info.NList
		lb.info.lList = info.LList
	}()
	if lb.info.sit[0] != "" {
		err = lb.p1.Recv(&info)
		if err != nil {
			return err
		}
		if info.Link != lb.info.link {
			return errors.New("link inconsistent goals")
		}
		defer func() {
			if err == nil {
				err = lb.p1.Send(info)
			}
		}()
	} else {
		defer func() {
			if err == nil {
				err = lb.p1.Send(nil)
			}
		}()
	}
	info.NList = append(info.NList, lb.lm.s.nodeName)
	info.LList = append(info.LList, lb.info.lid)
	info.Link = lb.info.link
	if lb.info.sit[2] != "" {
		err = lb.p2.Send(info)
		if err != nil {
			return err
		}
		err = lb.p2.Recv(&info)
		if err != nil {
			return err
		}
		if info.Link != lb.info.link {
			return errors.New("link inconsistent goals")
		}
		return nil
	} else {
		return lb.p2.Send(nil)
	}
}

type linkCheckInfo struct {
	Link  string
	NList []string
	LList []string
}
