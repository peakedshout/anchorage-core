package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/peakedshout/anchorage-core/pkg/comm"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"github.com/peakedshout/go-pandorasbox/control"
	"github.com/peakedshout/go-pandorasbox/pcrypto"
	"github.com/peakedshout/go-pandorasbox/tool/gpool"
	"github.com/peakedshout/go-pandorasbox/tool/hjson"
	"github.com/peakedshout/go-pandorasbox/tool/mslice"
	"github.com/peakedshout/go-pandorasbox/xnet"
	"github.com/peakedshout/go-pandorasbox/xnet/xmulti"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"github.com/peakedshout/go-pandorasbox/xnet/xquic"
	"github.com/peakedshout/go-pandorasbox/xnet/xtls"
	"github.com/peakedshout/go-pandorasbox/xrpc"
	"github.com/quic-go/quic-go"
	"math/rand"
	"net"
	"runtime"
	"sync"
	"time"
)

func (c *Client) newLauncher(nm map[string][]*comm.NodeUnit) {
	var list []*comm.NodeUnit
	for _, units := range nm {
		list = append(list, units...)
	}
	c.launcher = &launcher{
		c:    c,
		nm:   nm,
		list: list,
	}
}

type launcher struct {
	c    *Client
	nm   map[string][]*comm.NodeUnit
	list []*comm.NodeUnit
}

func (l *launcher) rpcByNode(ctx context.Context, node string, header string, send, recv any) (*comm.NodeUnit, error) {
	var nodes []*comm.NodeUnit
	if node == "" {
		nodes = l.list
	} else {
		units, ok := l.nm[node]
		if !ok {
			return nil, errors.New("invalid node")
		}
		nodes = units
	}
	return l.rpc(ctx, nodes, header, send, recv)
}

func (l *launcher) rpc(ctx context.Context, nodes []*comm.NodeUnit, header string, send, recv any) (*comm.NodeUnit, error) {
	tmpCtx, tmpCl := context.WithCancel(ctx)
	defer tmpCl()
	var wg sync.WaitGroup
	var mux sync.Mutex
	var b *comm.NodeUnit
	lens := len(nodes)
	ints := mslice.MakeRandRangeSlice(0, lens)
	errs := make([]error, 0, lens)
	if lens > runtime.NumCPU() {
		lens = runtime.NumCPU()
	}
	pool := gpool.NewGPool(tmpCtx, lens)
	for _, i := range ints {
		wg.Add(1)
		nu := nodes[i]
		pool.Do(func() {
			defer wg.Done()
			_ = nu.RpcCallback(tmpCtx, func(ctx context.Context, fn xrpc.RpcFunc) error {
				mux.Lock()
				defer mux.Unlock()
				if tmpCtx.Err() != nil {
					return tmpCtx.Err()
				}
				err := fn(tmpCtx, header, send, recv)
				if err == nil {
					tmpCl()
					b = nu
				} else {
					errs = append(errs, err)
				}
				return err
			})
		})
	}
	wg.Wait()
	if b == nil {
		return nil, fmt.Errorf("rpc failed err: %w", errors.Join(errs...))
	}
	return b, nil
}

func (l *launcher) selectNodeUnit(s *comm.LinkRequest) []*comm.NodeUnit {
	var list []*comm.NodeUnit
	node := s.GetNode()
	if node != "" {
		list = l.nm[node]
	} else {
		for _, units := range l.nm {
			list = append(list, units...)
		}
	}
	if !s.SwitchTP2P && !s.SwitchUP2P {
		return list
	}
	nl := make([]*comm.NodeUnit, 0, len(list))
	for _, unit := range list {
		for _, addr := range unit.Addrs {
			if s.PP2PNetwork != "" {
				if s.PP2PNetwork == addr.Network() {
					nl = append(nl, unit)
					break
				}
			} else {
				if s.SwitchTP2P && xnet.GetStdBaseNetwork(addr.Network()) == "tcp" {
					nl = append(nl, unit)
					break
				}
				if s.SwitchUP2P && xnet.GetStdBaseNetwork(addr.Network()) == "udp" {
					nl = append(nl, unit)
					break
				}
			}
		}
	}
	return nl
}

func (l *launcher) nodeCallBack(ctx context.Context, td time.Duration, node string, fn func(ctx context.Context, nu *comm.NodeUnit) error) error {
	if node == "" {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		return ctxtool.RunTimerFunc(ctx, td, func(ctx context.Context) error {
			var units []*comm.NodeUnit
			for _, one := range l.nm {
				units = one
				break
			}
			i := r.Intn(len(units))
			return fn(ctx, units[i])
		})
	} else {
		units, ok := l.nm[node]
		if !ok {
			return errors.New("invalid node")
		}
		tmpCtx, tmpCl := context.WithCancelCause(ctx)
		defer tmpCl(context.Canceled)
		for i := 0; i < len(units); i++ {
			nu := units[i]
			go func() {
				err := ctxtool.RunTimerFunc(tmpCtx, td, func(ctx context.Context) error {
					return fn(ctx, nu)
				})
				if err != nil {
					tmpCl(err)
				}
			}()
		}
		return ctxtool.Wait(tmpCtx)
	}
}

func (l *launcher) handleLink(ctx context.Context, nu *comm.NodeUnit, pinfo *p2pInfo) (xrpc.Stream, error) {
	if pinfo.network == "" {
		return l.handleLinkStream(ctx, nu)
	}
	addrs := make([]net.Addr, 0, len(nu.Addrs))
	for _, addr := range nu.Addrs {
		if addr.Network() == pinfo.network {
			addrs = append(addrs, addr)
		}
	}
	if len(addrs) == 0 {
		return l.handleLinkStream(ctx, nu)
	}
	return l.handleP2PStream(ctx, nu, pinfo, addrs)
}

func (l *launcher) handleLinkStream(ctx context.Context, nu *comm.NodeUnit) (xrpc.Stream, error) {
	stream, err := nu.Stream(ctx, comm.CallLink)
	if err != nil {
		return nil, err
	}
	err = stream.Send(nil)
	if err != nil {
		_ = stream.Close()
		return nil, err
	}
	return stream, nil
}

func (l *launcher) handleP2PStream(ctx context.Context, nu *comm.NodeUnit, pinfo *p2pInfo, addrs []net.Addr) (xrpc.Stream, error) {
	var x map[string]xnetutil.Dialer
	if pinfo.network == "tcp" {
		pinfo.tr = &net.Dialer{Control: control.PortReuseControl}
		x = map[string]xnetutil.Dialer{
			"tcp": pinfo.tr,
		}
	} else {
		udp, err := net.ListenUDP("udp", &net.UDPAddr{})
		if err != nil {
			return nil, err
		}
		tr := &quic.Transport{Conn: udp}
		defer func() {
			if err != nil {
				_ = tr.Close()
			}
		}()
		pinfo.ur = xquic.NewQuicTransportDialer(tr, nil)
		x = map[string]xnetutil.Dialer{
			"udp": pinfo.ur,
		}
	}
	dr := xmulti.NewMultiAddrDialer(xmulti.MultiAddrDialTypeGo, x)
	conn, err := dr.MultiDialContext(ctx, addrs...)
	if err != nil {
		return nil, err
	}
	session, err := nu.WithConn(ctx, conn)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	stream, err := session.Stream(ctx, comm.CallLink)
	if err != nil {
		_ = session.Close()
		return nil, err
	}
	err = stream.Send(nil)
	if err != nil {
		_ = stream.Close()
		_ = session.Close()
		return nil, err
	}
	if pinfo.network == "tcp" {
		pinfo.tr.LocalAddr = conn.LocalAddr()
	}
	pinfo.enable = true
	return &streamUnit{
		Stream: stream,
		sess:   session,
	}, nil
}

func (l *launcher) handleConn(stream xrpc.Stream, pinfo *p2pInfo) (net.Conn, error) {
	err := stream.Recv(nil)
	if err != nil {
		return nil, err
	}
	publaddr, _ := xrpc.GetSessionAuthInfoT[string](stream.Context(), xrpc.LocalPubAddress)
	publnk, _ := xrpc.GetSessionAuthInfoT[string](stream.Context(), xrpc.LocalPubNetwork)
	priladdr, _ := xrpc.GetSessionAuthInfoT[string](stream.Context(), xrpc.LocalPriAddress)
	prilnk, _ := xrpc.GetSessionAuthInfoT[string](stream.Context(), xrpc.LocalPriNetwork)
	if pinfo.isListener {
		return l.handleConnListen(stream, pinfo, publaddr, publnk, priladdr, prilnk)
	} else {
		return l.handleConnDial(stream, pinfo, publaddr, publnk, priladdr, prilnk)
	}
}

func (l *launcher) handleConnListen(stream xrpc.Stream, pinfo *p2pInfo, publaddr, publnk, priladdr, prilnk string) (net.Conn, error) {
	var rinfo connInfo
	var b []byte
	err := stream.Recv(&b)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &rinfo)
	if err != nil {
		return nil, err
	}
	laddr := xnetutil.NewNetAddr(prilnk, priladdr)
	raddr := xnetutil.NewNetAddr(rinfo.PublicNetwork, rinfo.PublicAddress)
	rinfo.PublicNetwork = publnk
	rinfo.PublicAddress = publaddr
	if pinfo.enable {
		cfg, cert, key, err := pcrypto.NewDefaultTlsConfigWithRaw()
		if err != nil {
			return nil, err
		}
		pinfo.cfg = cfg
		rinfo.Cert = cert
		rinfo.Key = key
	}
	err = stream.Send(hjson.MustMarshal(rinfo))
	if err != nil {
		return nil, err
	}
	if pinfo.enable {
		time.Sleep(1 * time.Second)
		return l.handleConnP2P(stream, raddr.String(), pinfo)
	} else {
		return newConn(stream, laddr, raddr, func() {}), nil
	}
}

func (l *launcher) handleConnDial(stream xrpc.Stream, pinfo *p2pInfo, publaddr, publnk, priladdr, prilnk string) (net.Conn, error) {
	rinfo := connInfo{
		PublicNetwork: publnk,
		PublicAddress: publaddr,
		TimeStamp:     time.Now(),
	}
	err := stream.Send(hjson.MustMarshal(rinfo))
	if err != nil {
		return nil, err
	}
	var b []byte
	err = stream.Recv(&b)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &rinfo)
	if err != nil {
		return nil, err
	}
	laddr := xnetutil.NewNetAddr(prilnk, priladdr)
	raddr := xnetutil.NewNetAddr(rinfo.PublicNetwork, rinfo.PublicAddress)
	if pinfo.enable {
		cfg, err := pcrypto.MakeTlsConfig(rinfo.Cert, rinfo.Key)
		if err != nil {
			return nil, err
		}
		pinfo.cfg = cfg
		duration := 1*time.Second - (time.Now().Sub(rinfo.TimeStamp) / 2)
		if duration > 0 {
			time.Sleep(duration)
		}
		return l.handleConnP2P(stream, raddr.String(), pinfo)
	} else {
		// turn
		return newConn(stream, laddr, raddr, func() {}), nil
	}
}

func (l *launcher) handleConnP2P(stream xrpc.Stream, addr string, pinfo *p2pInfo) (conn net.Conn, err error) {
	ctx := stream.Context()
	defer func() {
		if conn != nil {
			ctxtool.GWaitFunc(ctx, func() {
				_ = conn.Close()
			})
		}
	}()
	if pinfo.network == "tcp" {
		for i := 0; i < 3; i++ {
			conn, err = pinfo.tr.DialContext(ctx, pinfo.network, addr)
			if err != nil {
				time.Sleep(100 * time.Millisecond)
			}
		}
		if conn == nil {
			return nil, err
		}
		u := xtls.TLSUpgrader(pinfo.cfg, !pinfo.isListener)
		conn, err = u.UpgradeContext(ctx, conn)
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		conn = &p2pConn{
			Conn: conn,
			closeFn: func() {
				_ = stream.Close()
			},
		}
		return conn, nil
	} else {
		if pinfo.isListener {
			tmpLn := xquic.NewQuicListenConfigWithTransport(pinfo.ur.Transport(), pinfo.cfg)
			ctxx, cl := context.WithCancel(ctx)
			defer cl()
			ln, err := tmpLn.ListenContext(ctxx, "", "")
			if err != nil {
				_ = pinfo.ur.Close()
				return nil, err
			}
			defer ln.Close()
			go func() {
				defer cl()
				tmpDr := xquic.NewQuicTransportDialer(pinfo.ur.Transport(), pinfo.cfg)
				for i := 0; i < 3; i++ {
					xconn, _ := tmpDr.DialContext(ctxx, pinfo.network, addr)
					if xconn != nil {
						_ = xconn.Close()
						return
					}
					time.Sleep(100 * time.Millisecond)
				}
			}()
			accept, err := ln.Accept()
			if err != nil {
				_ = pinfo.ur.Close()
				return nil, err
			}
			conn = accept
		} else {
			tmpDr := xquic.NewQuicTransportDialer(pinfo.ur.Transport(), pinfo.cfg)
			for i := 0; i < 3; i++ {
				conn, err = tmpDr.DialContext(ctx, pinfo.network, addr)
				if err != nil {
					time.Sleep(100 * time.Millisecond)
				} else {
					break
				}
			}
			if conn == nil {
				_ = pinfo.ur.Close()
				return nil, err
			}
		}
		conn = &p2pConn{
			Conn: conn,
			closeFn: func() {
				_ = stream.Close()
				_ = pinfo.ur.Close()
			},
		}
		return conn, nil
	}
}

type streamUnit struct {
	xrpc.Stream
	sess *xrpc.ClientSession
}

func (su *streamUnit) Close() error {
	defer su.sess.Close()
	return su.Stream.Close()
}

type p2pConn struct {
	net.Conn
	closeFn func()
}

func (pc *p2pConn) Close() error {
	if pc.closeFn != nil {
		pc.closeFn()
	}
	return pc.Conn.Close()
}

type p2pInfo struct {
	network    string
	isListener bool
	enable     bool
	tr         *net.Dialer
	ur         *xquic.QuicTransportDialer
	cfg        *tls.Config
}

func selectP2PNetwork(nu *comm.NodeUnit, r *comm.RegisterListenerInfo, s *comm.LinkRequest) string {
	u, t := false, false
	for _, addr := range nu.Addrs {
		switch addr.Network() {
		case "tcp":
			t = true
		case "udp":
			u = true
		}
	}
	t = t && s.SwitchTP2P && r.Settings.SwitchTP2P
	u = u && s.SwitchUP2P && r.Settings.SwitchUP2P
	if !t && !u {
		return ""
	}
	if s.PP2PNetwork == "tcp" {
		if t {
			return "tcp"
		} else {
			return "udp"
		}
	} else {
		if u {
			return "udp"
		} else {
			return "tcp"
		}
	}
}
