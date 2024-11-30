package client

import (
	"context"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"net"
	"sync"
)

func (c *Client) newListener(ctx context.Context, laddr net.Addr) *Listener {
	ln := &Listener{
		ch:    make(chan net.Conn),
		laddr: laddr,
	}
	ln.ctx, ln.cl = ctxtool.ContextsWithCancel(c.ctx, ctx)
	return ln
}

type Listener struct {
	ctx    context.Context
	cl     context.CancelFunc
	ch     chan net.Conn
	mux    sync.Mutex
	laddr  net.Addr
	closer sync.Once
}

func (ln *Listener) Accept() (net.Conn, error) {
	if ln.ctx.Err() != nil {
		return nil, ln.ctx.Err()
	}
	select {
	case conn := <-ln.ch:
		return conn, nil
	case <-ln.ctx.Done():
		return nil, ln.ctx.Err()
	}
}

func (ln *Listener) Close() error {
	err := ln.ctx.Err()
	ln.closer.Do(func() {
		ln.cl()
		err = nil
	})
	return err
}

func (ln *Listener) Addr() net.Addr {
	ln.mux.Lock()
	defer ln.mux.Unlock()
	return ln.laddr
}

func (ln *Listener) setAddr(addr net.Addr) {
	ln.mux.Lock()
	defer ln.mux.Unlock()
	ln.laddr = addr
}

func (ln *Listener) addConn(conn net.Conn) error {
	select {
	case ln.ch <- conn:
		return nil
	case <-ln.ctx.Done():
		_ = conn.Close()
		return ln.ctx.Err()
	}
}
