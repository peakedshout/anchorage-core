package client

import (
	"context"
	"fmt"
	"github.com/peakedshout/anchorage-core/pkg/comm"
	"github.com/peakedshout/anchorage-core/pkg/config"
	"github.com/peakedshout/anchorage-core/pkg/server"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"github.com/peakedshout/go-pandorasbox/xnet/fasttool"
	"io"
	"math/rand"
	"net"
	"testing"
	"time"
)

func newAddr() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("127.0.0.1:%d", r.Intn(10000)+10000)
}

func testServer(ctx context.Context, t *testing.T) (string, string, func()) {
	scfg := &config.ServerConfig{
		NodeInfo: config.NodeConfig{
			NodeName: "node1",
			BaseNetwork: []config.BaseNetworkConfig{{
				Network: "tcp",
				Address: newAddr(),
			}, {
				Network: "udp",
				Address: newAddr(),
			}},
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		},
		SyncNodes: nil,
	}
	s, err := server.NewServerContext(ctx, scfg)
	if err != nil {
		t.Fatal(err)
	}
	go s.Serve()
	fn := func() {
		_ = s.Close()
	}
	go ctxtool.RunTimerFunc(ctx, 1*time.Second, func(ctx context.Context) error {
		fmt.Println(s.GetRouteView())
		return nil
	})
	return scfg.NodeInfo.BaseNetwork[0].Address, scfg.NodeInfo.BaseNetwork[1].Address, fn
}

func TestClient_Serve(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cl()
	addr, _, sx := testServer(ctx, t)
	defer sx()
	cfg := &config.ClientConfig{
		Nodes: []config.NodeConfig{{
			NodeName: "node1",
			BaseNetwork: []config.BaseNetworkConfig{{
				Network: "tcp",
				Address: addr,
			}},
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		}},
	}
	cc, err := NewClientContext(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()
	lc := comm.RegisterListenerInfo{
		Name:  "tl",
		Notes: "",
		Auth:  nil,
		Settings: comm.Settings{
			SwitchHide: false,
			SwitchLink: true,
			SwitchUP2P: false,
			SwitchTP2P: false,
		},
	}
	_ = cc.Serve(ctx, lc, func(conn net.Conn) error {
		_, err := io.Copy(conn, conn)
		return err
	})
}

func TestClient_Dial(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	addr, _, sx := testServer(ctx, t)
	defer sx()
	cfg := &config.ClientConfig{
		Nodes: []config.NodeConfig{{
			NodeName: "node1",
			BaseNetwork: []config.BaseNetworkConfig{{
				Network: "tcp",
				Address: addr,
			}},
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		}},
	}
	cc, err := NewClientContext(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()
	lc := comm.RegisterListenerInfo{
		Name:  "tl",
		Notes: "",
		Auth:  nil,
		Settings: comm.Settings{
			SwitchHide: false,
			SwitchLink: true,
			SwitchUP2P: false,
			SwitchTP2P: false,
		},
	}
	go func() {
		_ = cc.Serve(ctx, lc, func(conn net.Conn) error {
			_, err := io.Copy(conn, conn)
			return err
		})
	}()
	time.Sleep(2 * time.Second)
	//view, err := cc.GetView(ctx)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//fmt.Println("client", view)
	lr := comm.LinkRequest{
		Node:        nil,
		Link:        "tl",
		PP2PNetwork: "",
		SwitchUP2P:  false,
		SwitchTP2P:  false,
		ForceP2P:    false,
		Auth:        nil,
		BoxId:       0,
	}
	conn, err := cc.Dial(ctx, lr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < 10000; i++ {
		b := []byte(uuid.NewIdn(4096))
		_, err = conn.Write(b)
		if err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 4096)
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != string(buf) {
			t.Fatal()
		}
	}
}

func TestClient_Dial2(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	addr, _, sx := testServer(ctx, t)
	defer sx()
	cfg := &config.ClientConfig{
		Nodes: []config.NodeConfig{{
			NodeName: "node1",
			BaseNetwork: []config.BaseNetworkConfig{{
				Network: "tcp",
				Address: addr,
			}},
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		}},
	}
	cc, err := NewClientContext(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()
	lc := comm.RegisterListenerInfo{
		Name:  "tl",
		Notes: "",
		Auth: &config.AuthInfo{
			UserName: "testu",
			Password: "testp",
		},
		Settings: comm.Settings{
			SwitchHide: false,
			SwitchLink: true,
			SwitchUP2P: false,
			SwitchTP2P: false,
		},
	}
	go func() {
		_ = cc.Serve(ctx, lc, func(conn net.Conn) error {
			_, err := io.Copy(conn, conn)
			return err
		})
	}()
	time.Sleep(2 * time.Second)
	lr := comm.LinkRequest{
		Node:        nil,
		Link:        "tl",
		PP2PNetwork: "",
		SwitchUP2P:  false,
		SwitchTP2P:  false,
		ForceP2P:    false,
		Auth: &config.AuthInfo{
			UserName: "testu",
			Password: "testp",
		},
		BoxId: 0,
	}
	conn, err := cc.Dial(ctx, lr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < 10000; i++ {
		b := []byte(uuid.NewIdn(4096))
		_, err = conn.Write(b)
		if err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 4096)
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != string(buf) {
			t.Fatal()
		}
	}
}

func TestClient_DialP2PT(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	addr, _, sx := testServer(ctx, t)
	defer sx()
	cfg := &config.ClientConfig{
		Nodes: []config.NodeConfig{{
			NodeName: "node1",
			BaseNetwork: []config.BaseNetworkConfig{{
				Network: "tcp",
				Address: addr,
			}},
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		}},
	}
	cc, err := NewClientContext(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()
	lc := comm.RegisterListenerInfo{
		Name:  "tl",
		Notes: "",
		Auth:  nil,
		Settings: comm.Settings{
			SwitchHide: false,
			SwitchLink: true,
			SwitchUP2P: false,
			SwitchTP2P: true,
		},
	}
	go func() {
		_ = cc.Serve(ctx, lc, func(conn net.Conn) error {
			_, err := io.Copy(conn, conn)
			return err
		})
	}()
	time.Sleep(2 * time.Second)
	lr := comm.LinkRequest{
		Node:        nil,
		Link:        "tl",
		PP2PNetwork: "",
		SwitchUP2P:  false,
		SwitchTP2P:  true,
		ForceP2P:    true,
		Auth:        nil,
		BoxId:       0,
	}
	conn, err := cc.Dial(ctx, lr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < 10000; i++ {
		b := []byte(uuid.NewIdn(4096))
		_, err = conn.Write(b)
		if err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 4096)
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != string(buf) {
			t.Fatal()
		}
	}
}

func TestClient_DialP2PU(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	_, addr, sx := testServer(ctx, t)
	defer sx()
	cfg := &config.ClientConfig{
		Nodes: []config.NodeConfig{{
			NodeName: "node1",
			BaseNetwork: []config.BaseNetworkConfig{{
				Network: "udp",
				Address: addr,
			}},
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		}},
	}
	cc, err := NewClientContext(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()
	lc := comm.RegisterListenerInfo{
		Name:  "tl",
		Notes: "",
		Auth:  nil,
		Settings: comm.Settings{
			SwitchHide: false,
			SwitchLink: true,
			SwitchUP2P: true,
			SwitchTP2P: false,
		},
	}
	go func() {
		_ = cc.Serve(ctx, lc, func(conn net.Conn) error {
			_, err := io.Copy(conn, conn)
			return err
		})
	}()
	time.Sleep(2 * time.Second)
	lr := comm.LinkRequest{
		Node:        nil,
		Link:        "tl",
		PP2PNetwork: "",
		SwitchUP2P:  true,
		SwitchTP2P:  false,
		ForceP2P:    true,
		Auth:        nil,
		BoxId:       0,
	}
	conn, err := cc.Dial(ctx, lr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < 10000; i++ {
		b := []byte(uuid.NewIdn(4096))
		_, err = conn.Write(b)
		if err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 4096)
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != string(buf) {
			t.Fatal()
		}
	}
}

func TestClient_Listen(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	addr, _, sx := testServer(ctx, t)
	defer sx()
	cfg := &config.ClientConfig{
		Nodes: []config.NodeConfig{{
			NodeName: "node1",
			BaseNetwork: []config.BaseNetworkConfig{{
				Network: "tcp",
				Address: addr,
			}},
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		}},
	}
	cc, err := NewClientContext(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()
	lc := comm.RegisterListenerInfo{
		Name:  "tl",
		Notes: "",
		Auth: &config.AuthInfo{
			UserName: "testu",
			Password: "testp",
		},
		Settings: comm.Settings{
			SwitchHide: false,
			SwitchLink: true,
			SwitchUP2P: false,
			SwitchTP2P: false,
		},
	}
	go func() {
		ln := cc.Listen(ctx, lc)
		defer ln.Close()
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer conn.Close()
				_, _ = io.Copy(conn, conn)
			}(conn)
		}
	}()
	time.Sleep(2 * time.Second)
	lr := comm.LinkRequest{
		Node:        nil,
		Link:        "tl",
		PP2PNetwork: "",
		SwitchUP2P:  false,
		SwitchTP2P:  false,
		ForceP2P:    false,
		Auth: &config.AuthInfo{
			UserName: "testu",
			Password: "testp",
		},
		BoxId: 0,
	}
	conn, err := cc.Dial(ctx, lr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < 10000; i++ {
		b := []byte(uuid.NewIdn(4096))
		_, err = conn.Write(b)
		if err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 4096)
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != string(buf) {
			t.Fatal()
		}
	}
}

func TestClient_ListenP2PU(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	_, addr, sx := testServer(ctx, t)
	defer sx()
	cfg := &config.ClientConfig{
		Nodes: []config.NodeConfig{{
			NodeName: "node1",
			BaseNetwork: []config.BaseNetworkConfig{{
				Network: "udp",
				Address: addr,
			}},
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		}},
	}
	cc, err := NewClientContext(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()
	lc := comm.RegisterListenerInfo{
		Name:  "tl",
		Notes: "",
		Auth:  nil,
		Settings: comm.Settings{
			SwitchHide: false,
			SwitchLink: true,
			SwitchUP2P: true,
			SwitchTP2P: false,
		},
	}
	go func() {
		ln := cc.Listen(ctx, lc)
		defer ln.Close()
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer conn.Close()
				_, _ = io.Copy(conn, conn)
			}(conn)
		}
	}()
	time.Sleep(2 * time.Second)
	lr := comm.LinkRequest{
		Node:        nil,
		Link:        "tl",
		PP2PNetwork: "",
		SwitchUP2P:  true,
		SwitchTP2P:  false,
		ForceP2P:    true,
		Auth:        nil,
		BoxId:       0,
	}
	conn, err := cc.Dial(ctx, lr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < 10000; i++ {
		b := []byte(uuid.NewIdn(4096))
		_, err = conn.Write(b)
		if err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 4096)
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != string(buf) {
			t.Fatal()
		}
	}
}

func TestClient_DialB(t *testing.T) {
	p2p := true
	ctx, cl := context.WithTimeout(context.Background(), 100*time.Second)
	defer cl()
	_, addr, sx := testServer(ctx, t)
	defer sx()
	cfg := &config.ClientConfig{
		Nodes: []config.NodeConfig{{
			NodeName: "node1",
			BaseNetwork: []config.BaseNetworkConfig{{
				Network: "udp",
				Address: addr,
			}},
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		}},
	}
	cc, err := NewClientContext(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()
	lc := comm.RegisterListenerInfo{
		Name:  "tl",
		Notes: "",
		Auth:  nil,
		Settings: comm.Settings{
			SwitchHide: false,
			SwitchLink: true,
			SwitchUP2P: p2p,
			SwitchTP2P: false,
		},
	}
	go func() {
		_ = cc.Serve(ctx, lc, func(conn net.Conn) error {
			_, err := io.Copy(conn, conn)
			return err
		})
	}()
	time.Sleep(2 * time.Second)
	lr := comm.LinkRequest{
		Node:        nil,
		Link:        "tl",
		PP2PNetwork: "",
		SwitchUP2P:  p2p,
		SwitchTP2P:  false,
		ForceP2P:    p2p,
		Auth:        nil,
		BoxId:       0,
	}
	conn, err := cc.Dial(ctx, lr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	b := []byte(uuid.NewIdn(32 * 1024))
	now := time.Now()
	defer func() {
		d := time.Since(now)
		fmt.Println(320 / d.Seconds())
	}()
	for i := 0; i < 1024*10; i++ {
		_, err = conn.Write(b)
		if err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 32*1024)
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != string(buf) {
			t.Fatal()
		}
	}
}

func testSyncServer(ctx context.Context, t *testing.T) ([]string, []string, func()) {
	scfg1 := &config.ServerConfig{
		NodeInfo: config.NodeConfig{
			NodeName: "node1",
			BaseNetwork: []config.BaseNetworkConfig{{
				Network: "tcp",
				Address: newAddr(),
			}, {
				Network: "udp",
				Address: newAddr(),
			}},
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		},
		SyncTimeInterval: 1000,
	}
	scfg2 := &config.ServerConfig{
		NodeInfo: config.NodeConfig{
			NodeName: "node2",
			BaseNetwork: []config.BaseNetworkConfig{{
				Network: "tcp",
				Address: newAddr(),
			}, {
				Network: "udp",
				Address: newAddr(),
			}},
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		},
		SyncTimeInterval: 1000,
	}
	scfg1.SyncNodes = []config.NodeConfig{
		{
			NodeName:         "node2",
			BaseNetwork:      scfg2.NodeInfo.BaseNetwork,
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		},
	}
	scfg2.SyncNodes = []config.NodeConfig{
		{
			NodeName:         "node1",
			BaseNetwork:      scfg1.NodeInfo.BaseNetwork,
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		},
	}
	s1, err := server.NewServerContext(ctx, scfg1)
	if err != nil {
		t.Fatal(err)
	}
	s2, err := server.NewServerContext(ctx, scfg2)
	if err != nil {
		t.Fatal(err)
	}
	go s1.Serve()
	go s2.Serve()
	fn := func() {
		_ = s1.Close()
		_ = s2.Close()
	}
	go ctxtool.RunTimerFunc(ctx, 1*time.Second, func(ctx context.Context) error {
		fmt.Println("s1------------------------")
		fmt.Println(s1.GetRouteView())
		fmt.Println("s2------------------------")
		fmt.Println(s2.GetRouteView())
		return nil
	})
	return []string{scfg1.NodeInfo.BaseNetwork[0].Address, scfg1.NodeInfo.BaseNetwork[1].Address},
		[]string{scfg2.NodeInfo.BaseNetwork[0].Address, scfg2.NodeInfo.BaseNetwork[1].Address}, fn
}

func TestServerSync(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cl()
	s1, _, fn := testSyncServer(ctx, t)
	defer fn()
	cfg := &config.ClientConfig{
		Nodes: []config.NodeConfig{{
			NodeName: "node1",
			BaseNetwork: []config.BaseNetworkConfig{{
				Network: "tcp",
				Address: s1[0],
			}},
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		}},
	}
	cc, err := NewClientContext(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()
	lc := comm.RegisterListenerInfo{
		Name:  "tl",
		Notes: "",
		Auth:  nil,
		Settings: comm.Settings{
			SwitchHide: false,
			SwitchLink: true,
			SwitchUP2P: false,
			SwitchTP2P: false,
		},
	}
	_ = cc.Serve(ctx, lc, func(conn net.Conn) error {
		_, err := io.Copy(conn, conn)
		return err
	})
}

func TestClient_DialSync(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	s1, s2, fn := testSyncServer(ctx, t)
	defer fn()
	cfg1 := &config.ClientConfig{
		Nodes: []config.NodeConfig{{
			NodeName: "node1",
			BaseNetwork: []config.BaseNetworkConfig{{
				Network: "tcp",
				Address: s1[0],
			}},
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		}},
	}
	cfg2 := &config.ClientConfig{
		Nodes: []config.NodeConfig{{
			NodeName: "node2",
			BaseNetwork: []config.BaseNetworkConfig{{
				Network: "udp",
				Address: s2[1],
			}},
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		}},
	}
	cc1, err := NewClientContext(ctx, cfg1)
	if err != nil {
		t.Fatal(err)
	}
	cc2, err := NewClientContext(ctx, cfg2)
	if err != nil {
		t.Fatal(err)
	}
	defer cc1.Close()
	defer cc2.Close()
	lc := comm.RegisterListenerInfo{
		Name:  "tl",
		Notes: "",
		Auth:  nil,
		Settings: comm.Settings{
			SwitchHide: false,
			SwitchLink: true,
			SwitchUP2P: false,
			SwitchTP2P: false,
		},
	}
	go func() {
		_ = cc1.Serve(ctx, lc, func(conn net.Conn) error {
			_, err := io.Copy(conn, conn)
			return err
		})
	}()
	time.Sleep(2 * time.Second)
	lr := comm.LinkRequest{
		Node:        nil,
		Link:        "tl",
		PP2PNetwork: "",
		SwitchUP2P:  false,
		SwitchTP2P:  false,
		ForceP2P:    false,
		Auth:        nil,
		BoxId:       0,
	}
	conn, err := cc2.Dial(ctx, lr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < 10000; i++ {
		b := []byte(uuid.NewIdn(4096))
		_, err = conn.Write(b)
		if err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 4096)
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != string(buf) {
			t.Fatal()
		}
	}
}

func TestClient_DialP2PUSync(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()
	s1, s2, fn := testSyncServer(ctx, t)
	defer fn()
	cfg1 := &config.ClientConfig{
		Nodes: []config.NodeConfig{{
			NodeName: "node1",
			BaseNetwork: []config.BaseNetworkConfig{{
				Network: "tcp",
				Address: s1[0],
			}, {
				Network: "udp",
				Address: s1[1],
			}},
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		}},
	}
	cfg2 := &config.ClientConfig{
		Nodes: []config.NodeConfig{{
			NodeName: "node2",
			BaseNetwork: []config.BaseNetworkConfig{{
				Network: "udp",
				Address: s2[1],
			}},
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		}},
	}
	cc1, err := NewClientContext(ctx, cfg1)
	if err != nil {
		t.Fatal(err)
	}
	cc2, err := NewClientContext(ctx, cfg2)
	if err != nil {
		t.Fatal(err)
	}
	defer cc1.Close()
	defer cc2.Close()
	lc := comm.RegisterListenerInfo{
		Name:  "tl",
		Notes: "",
		Auth:  nil,
		Settings: comm.Settings{
			SwitchHide: false,
			SwitchLink: true,
			SwitchUP2P: true,
			SwitchTP2P: false,
		},
	}
	go func() {
		_ = cc1.Serve(ctx, lc, func(conn net.Conn) error {
			_, err := io.Copy(conn, conn)
			return err
		})
	}()
	time.Sleep(2 * time.Second)
	lr := comm.LinkRequest{
		Node:        nil,
		Link:        "tl",
		PP2PNetwork: "",
		SwitchUP2P:  true,
		SwitchTP2P:  false,
		ForceP2P:    true,
		Auth:        nil,
		BoxId:       0,
	}
	conn, err := cc2.Dial(ctx, lr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < 10000; i++ {
		b := []byte(uuid.NewIdn(4096))
		_, err = conn.Write(b)
		if err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 4096)
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != string(buf) {
			t.Fatal()
		}
	}
}

func TestClient_Proxy(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()

	ln, err := fasttool.EchoTcpListenerContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	addr, _, sx := testServer(ctx, t)
	defer sx()
	cfg := &config.ClientConfig{
		Nodes: []config.NodeConfig{{
			NodeName: "node1",
			BaseNetwork: []config.BaseNetworkConfig{{
				Network: "tcp",
				Address: addr,
			}},
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		}},
	}

	cc, err := NewClientContext(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	conn, err := cc.Proxy(8).DialContext(ctx, ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < 10000; i++ {
		b := []byte(uuid.NewIdn(4096))
		_, err = conn.Write(b)
		if err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 4096)
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != string(buf) {
			t.Fatal()
		}
	}
}

func TestClient_ProxySync(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cl()

	ln, err := fasttool.EchoTcpListenerContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	s1, _, fn := testSyncServer(ctx, t)
	defer fn()
	cfg := &config.ClientConfig{
		Nodes: []config.NodeConfig{{
			NodeName: "node1",
			BaseNetwork: []config.BaseNetworkConfig{{
				Network: "tcp",
				Address: s1[0],
			}, {
				Network: "udp",
				Address: s1[1],
			}},
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		}},
	}

	nCtx := context.WithValue(ctx, ProxyNodes, []string{"node1", "node2"})
	cc, err := NewClientContext(nCtx, cfg)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	conn, err := cc.Proxy(8).DialContext(nCtx, ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < 10000; i++ {
		b := []byte(uuid.NewIdn(4096))
		_, err = conn.Write(b)
		if err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 4096)
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != string(buf) {
			t.Fatal()
		}
	}
}
