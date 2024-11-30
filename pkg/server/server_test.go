package server

import (
	"context"
	"github.com/peakedshout/anchorage-core/pkg/config"
	"testing"
	"time"
)

func TestNewServerContext(t *testing.T) {
	ctx, cl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cl()
	scfg := &config.ServerConfig{
		NodeInfo: config.NodeConfig{
			NodeName: "node1",
			BaseNetwork: []config.BaseNetworkConfig{{
				Network: "tcp",
				Address: "127.0.0.1:0",
			}},
			ExNetworks:       nil,
			Crypto:           nil,
			HandshakeTimeout: 0,
			HandleTimeout:    0,
		},
		SyncNodes: nil,
	}
	s, err := NewServerContext(ctx, scfg)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	_ = s.Serve()
}
