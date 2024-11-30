package comm

import (
	"context"
	"crypto/tls"
	"github.com/peakedshout/anchorage-core/pkg/config"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"github.com/peakedshout/go-pandorasbox/logger"
	"github.com/peakedshout/go-pandorasbox/pcrypto"
	"github.com/peakedshout/go-pandorasbox/xnet"
	"github.com/peakedshout/go-pandorasbox/xnet/xmulti"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"github.com/peakedshout/go-pandorasbox/xrpc"
	"io"
	"net"
	"os"
	"time"
)

func MakeUpgrader(networks []config.ExNetworkConfig) (xnetutil.Upgrader, error) {
	nsList := make([]string, 0, len(networks))
	for _, networkConfig := range networks {
		nsList = append(nsList, networkConfig.Network)
	}

	upgrader, err := xnet.MakeNetworkUpgrader(func(index int, network string) (cfg *tls.Config, isClient bool, err error) {
		one := networks[index]
		if one.CertFile != "" || one.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(one.CertFile, one.KeyFile)
			if err != nil {
				return nil, false, err
			}
			cfg = &tls.Config{Certificates: []tls.Certificate{cert}}
		}
		if one.CertRaw != "" || one.KeyRaw != "" {
			cert, err := tls.X509KeyPair([]byte(one.CertRaw), []byte(one.KeyRaw))
			if err != nil {
				return nil, false, err
			}
			cfg = &tls.Config{Certificates: []tls.Certificate{cert}}
		}
		return cfg, false, nil
	}, nsList...)
	if err != nil {
		return nil, err
	}
	return upgrader, nil
}

func MakeCrypto(cryptos []config.CryptoConfig) ([]*xrpc.CryptoConfig, error) {
	var cryptoList []*xrpc.CryptoConfig
	for _, one := range cryptos {
		var keys [][]byte
		if len(one.KeyFiles) != 0 {
			for _, two := range one.KeyFiles {
				b, err := os.ReadFile(two)
				if err != nil {
					return nil, err
				}
				keys = append(keys, b)
			}
		} else if len(one.Keys) != 0 {
			for _, two := range one.Keys {
				keys = append(keys, []byte(two))
			}
		}
		pc, err := pcrypto.GetCrypto(one.Crypto, keys...)
		if err != nil {
			return nil, err
		}
		cryptoList = append(cryptoList, &xrpc.CryptoConfig{
			Name:     one.Name,
			Crypto:   pc,
			Priority: one.Priority,
		})
	}
	return cryptoList, nil
}

func MakeLogger(ctx context.Context, l logger.Logger, config config.LoggerConfig) logger.Logger {
	l.SetLoggerStack(config.NeedStack)
	l.SetLoggerLevel(logger.GetLogLevel(config.LogLevel))
	l.SetLoggerColor(config.NeedColor)

	if config.LogFile != "" {
		var fn func() (io.WriteCloser, error)
		if config.Clear {
			fn = func() (io.WriteCloser, error) {
				return os.OpenFile(config.LogFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
			}
		} else {
			fn = func() (io.WriteCloser, error) {
				return os.OpenFile(config.LogFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC|os.O_APPEND, 0666)
			}
		}
		go func() {
			_ = ctxtool.RunTimerFunc(ctx, 1*time.Second, func(ctx context.Context) error {
				writeCloser, err := fn()
				if err != nil {
					return nil
				}
				_ = l.SyncLoggerCopy(ctx, writeCloser)
				_ = writeCloser.Close()
				return nil
			})
		}()
	}
	return l
}

func MakeBaseAddress(list []config.BaseNetworkConfig) []net.Addr {
	res := make([]net.Addr, 0, len(list))
	for _, cfg := range list {
		res = append(res, xnetutil.NewNetAddr(cfg.Network, cfg.Address))
	}
	return res
}

type NodeUnit struct {
	Node string
	*xrpc.Client
	Dr    *xmulti.MultiAddrDialer
	Addrs []net.Addr
}

func MakeNodeUnit(ctx context.Context, nodes []config.NodeConfig) (map[string][]*NodeUnit, error) {
	nm := make(map[string][]*NodeUnit)
	for _, node := range nodes {
		upgrader, err := MakeUpgrader(node.ExNetworks)
		if err != nil {
			return nil, err
		}
		crypto, err := MakeCrypto(node.Crypto)
		if err != nil {
			return nil, err
		}
		addr := MakeBaseAddress(node.BaseNetwork)
		nCtx, _ := xrpc.SetSessionAuthInfoT[string](ctx, NodeName, node.NodeName)
		cfg := &xrpc.ClientConfig{
			Ctx:                      nCtx,
			CryptoList:               crypto,
			Upgrader:                 upgrader,
			SwitchNetworkSpeedTicker: true,
			CacheTime:                10 * time.Second,
			ShareDialFunc: func(ctx context.Context) (net.Conn, error) {
				return xmulti.DefaultMultiAddrDialer.MultiDialContext(ctx, addr...)
			},
		}
		if node.Auth != nil {
			cfg.SessionAuthInfo = xrpc.BuildUPAuth(node.Auth.UserName, node.Auth.Password)
		}
		unit := &NodeUnit{
			Node:   node.NodeName,
			Client: xrpc.NewClient(cfg),
			Dr:     xmulti.DefaultMultiAddrDialer,
			Addrs:  addr,
		}
		nm[node.NodeName] = append(nm[node.NodeName], unit)
	}
	return nm, nil
}
