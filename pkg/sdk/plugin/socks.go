package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/peakedshout/go-pandorasbox/tool/dcopy"
	"github.com/peakedshout/go-pandorasbox/xnet/proxy/socks"
	"net"
)

func init() {
	RegisterDialPlugin(new(socksPlugin))
	RegisterProxyPlugin(new(socksPlugin))
}

const NameSocks = "socks"

type SocksPluginCfg struct {
	Name string `json:"name" yaml:"name" comment:"plugin name"`

	Version4              bool              `json:"v4" yaml:"v4" comment:"socks4 ?"`
	Version5              bool              `json:"v5" yaml:"v5" comment:"socks5 ?"`
	SwitchCMDCONNECT      bool              `json:"CMDCONNECT" yaml:"CMDCONNECT" comment:"CMDCONNECT ?"`
	SwitchCMDBIND         bool              `json:"CMDBIND" yaml:"CMDBIND" comment:"CMDBIND ?"`
	SwitchCMDUDPASSOCIATE bool              `json:"CMDUDPASSOCIATE" yaml:"CMDUDPASSOCIATE" comment:"CMDUDPASSOCIATE ?"`
	S4AuthIdList          []string          `json:"S4Auth" yaml:"S4Auth" comment:"S4Auth list"`
	S5AuthPasswordList    [][2]string       `json:"S5Auth" yaml:"S5Auth" comment:"S5Auth list"`
	Args                  map[string]string `json:"args" yaml:"args" comment:"plugin args"`
}

type socksPlugin struct {
	cfg SocksPluginCfg
}

func (p *socksPlugin) Name() string {
	return NameSocks
}

func (p *socksPlugin) TempProxyCfg() any {
	return p.tempCfg()
}

func (p *socksPlugin) LoadProxyCfg(str string) (ProxyPlugin, error) {
	var cfg SocksPluginCfg
	err := json.Unmarshal([]byte(str), &cfg)
	if err != nil {
		return nil, err
	}
	return &socksPlugin{cfg: cfg}, nil
}

func (p *socksPlugin) ProxyServe(ctx context.Context, ln net.Listener, drFunc DialFunc) error {
	return p.serve(ctx, ln, drFunc)
}

func (p *socksPlugin) ProxyUpgrade(ctx context.Context, ln net.Listener, drFunc DialFunc) (context.Context, net.Listener, DialFunc, error) {
	return p.upgrade(ctx, ln, drFunc)
}

func (p *socksPlugin) TempDialCfg() any {
	return p.tempCfg()
}

func (p *socksPlugin) LoadDialCfg(str string) (DialPlugin, error) {
	var cfg SocksPluginCfg
	err := json.Unmarshal([]byte(str), &cfg)
	if err != nil {
		return nil, err
	}
	return &socksPlugin{cfg: cfg}, nil
}

func (p *socksPlugin) DialServe(ctx context.Context, ln net.Listener, drFunc DialFunc) error {
	return p.serve(ctx, ln, drFunc)
}

func (p *socksPlugin) DialUpgrade(ctx context.Context, ln net.Listener, drFunc DialFunc) (context.Context, net.Listener, DialFunc, error) {
	return p.upgrade(ctx, ln, drFunc)
}

func (p *socksPlugin) tempCfg() any {
	return SocksPluginCfg{
		Name:             "socks",
		Version4:         true,
		Version5:         true,
		SwitchCMDCONNECT: true,
	}
}

func (p *socksPlugin) serve(ctx context.Context, ln net.Listener, drFunc DialFunc) error {
	cfg := p.cfg
	scfg := &socks.ServerConfig{
		VersionSwitch: socks.VersionSwitch{
			SwitchSocksVersion4: cfg.Version4,
			SwitchSocksVersion5: cfg.Version5,
		},
		CMDConfig: socks.CMDConfig{
			SwitchCMDCONNECT: cfg.SwitchCMDCONNECT,
			CMDCONNECTHandler: func(ctx context.Context, addr string) (net.Conn, error) {
				return drFunc(ctx, "tcp", addr)
			},
			SwitchCMDBIND:         false,
			CMDBINDHandler:        nil,
			SwitchCMDUDPASSOCIATE: cfg.SwitchCMDUDPASSOCIATE,
			CMDCMDUDPASSOCIATEHandler: func(ctx context.Context, addr net.Addr) (net.PacketConn, error) {
				//todo
				return nil, errors.New("todo")
			},
		},
	}
	if len(cfg.S4AuthIdList) != 0 {
		s4List := dcopy.CopyT(cfg.S4AuthIdList)
		scfg.Socks4AuthCb = socks.S4AuthCb{Socks4UserIdAuth: func(conn net.Conn, id socks.S4UserId) (net.Conn, socks.S4IdAuthCode) {
			for _, str := range s4List {
				if id.IsEqual2(socks.S4UserId(str)) == socks.CodeGranted {
					return conn, socks.CodeGranted
				}
			}
			return nil, socks.CodeRejectedDifferentUserId
		}}
	}
	if len(cfg.S5AuthPasswordList) == 0 {
		scfg.Socks5AuthCb = socks.S5AuthCb{
			Socks5AuthNOAUTHPriority: 0,
			Socks5AuthNOAUTH:         socks.DefaultAuthConnCb,
		}
	} else {
		s5List := dcopy.CopyT(cfg.S5AuthPasswordList)
		scfg.Socks5AuthCb = socks.S5AuthCb{
			Socks5AuthPASSWORDPriority: 0,
			Socks5AuthPASSWORD: func(conn net.Conn, auth socks.S5AuthPassword) net.Conn {
				for _, sl := range s5List {
					if auth.IsEqual(sl[0], sl[1]) {
						return conn
					}
				}
				return nil
			},
		}
	}
	server, err := socks.NewServerContext(ctx, scfg)
	if err != nil {
		return err
	}
	return server.Serve(ln)
}

func (p *socksPlugin) upgrade(ctx context.Context, ln net.Listener, drFunc DialFunc) (context.Context, net.Listener, DialFunc, error) {
	return ctx, ln, drFunc, nil
}
