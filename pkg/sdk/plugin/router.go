package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"regexp"
)

const NameRouter = "router"

func init() {
	RegisterDialPlugin(new(routerPlugin))
	RegisterProxyPlugin(new(routerPlugin))
}

const (
	MethodProxy  = "proxy"
	MethodDirect = "direct"
	MethodBlock  = "block"
)

type RouterPluginCfg struct {
	Name string `json:"name" yaml:"name" comment:"plugin name"`

	DefaultMethod string      `json:"default" yaml:"default" comment:"default method"`
	ProxyList     [][2]string `json:"proxy" yaml:"proxy" comment:"proxy list"`
	DirectList    [][2]string `json:"direct" yaml:"direct" comment:"direct list"`
	BlockList     [][2]string `json:"block" yaml:"block" comment:"block list"`
}

type routerPlugin struct {
	cfg RouterPluginCfg
}

func (p *routerPlugin) Name() string {
	return NameRouter
}

func (p *routerPlugin) TempDialCfg() any {
	return p.tempCfg()
}

func (p *routerPlugin) LoadDialCfg(str string) (DialPlugin, error) {
	var cfg RouterPluginCfg
	err := json.Unmarshal([]byte(str), &cfg)
	if err != nil {
		return nil, err
	}
	return &routerPlugin{cfg: cfg}, nil
}

func (p *routerPlugin) DialServe(ctx context.Context, ln net.Listener, drFunc DialFunc) error {
	return p.serve(ctx, ln, drFunc)
}

func (p *routerPlugin) DialUpgrade(ctx context.Context, ln net.Listener, drFunc DialFunc) (context.Context, net.Listener, DialFunc, error) {
	return p.upgrade(ctx, ln, drFunc)
}

func (p *routerPlugin) TempProxyCfg() any {
	return p.tempCfg()
}

func (p *routerPlugin) LoadProxyCfg(str string) (ProxyPlugin, error) {
	var cfg RouterPluginCfg
	err := json.Unmarshal([]byte(str), &cfg)
	if err != nil {
		return nil, err
	}
	return &routerPlugin{cfg: cfg}, nil
}

func (p *routerPlugin) ProxyServe(ctx context.Context, ln net.Listener, drFunc DialFunc) error {
	return p.serve(ctx, ln, drFunc)
}

func (p *routerPlugin) ProxyUpgrade(ctx context.Context, ln net.Listener, drFunc DialFunc) (context.Context, net.Listener, DialFunc, error) {
	return p.upgrade(ctx, ln, drFunc)
}

func (p *routerPlugin) tempCfg() any {
	return RouterPluginCfg{
		Name: NameRouter,
	}
}

func (p *routerPlugin) serve(ctx context.Context, ln net.Listener, drFunc DialFunc) error {
	return errors.New("no implement")
}

func (p *routerPlugin) upgrade(ctx context.Context, ln net.Listener, drFunc DialFunc) (context.Context, net.Listener, DialFunc, error) {
	pl, dl, bl, err := p.makeRegexp()
	if err != nil {
		return nil, nil, nil, err
	}
	log := getPluginLog(ctx, p)
	df := p.cfg.DefaultMethod
	if df == "" {
		df = MethodProxy
	}
	var drFuncX func(ctx context.Context, network string, address string) (net.Conn, error)
	switch df {
	case MethodProxy:
		drFuncX = func(ctx context.Context, network string, address string) (net.Conn, error) {
			if p.matchNK(bl, network, address) {
				log.Debug(fmt.Sprintf("Blocking: %s_%s", network, address))
				return nil, fmt.Errorf("router: blocking [%s]%s", network, address)
			}
			if p.matchNK(dl, network, address) {
				log.Debug(fmt.Sprintf("Directing: %s_%s", network, address))
				return new(net.Dialer).DialContext(ctx, network, address)
			}
			log.Debug(fmt.Sprintf("Proxying: %s_%s", network, address))
			return drFunc(ctx, network, address)
		}
	case MethodDirect:
		drFuncX = func(ctx context.Context, network string, address string) (net.Conn, error) {
			if p.matchNK(bl, network, address) {
				log.Debug(fmt.Sprintf("Blocking: %s_%s", network, address))
				return nil, fmt.Errorf("router: blocking [%s]%s", network, address)
			}
			if p.matchNK(pl, network, address) {
				log.Debug(fmt.Sprintf("Proxying: %s_%s", network, address))
				return drFunc(ctx, network, address)
			}
			log.Debug(fmt.Sprintf("Directing: %s_%s", network, address))
			return new(net.Dialer).DialContext(ctx, network, address)
		}
	case MethodBlock:
		drFuncX = func(ctx context.Context, network string, address string) (net.Conn, error) {
			if p.matchNK(pl, network, address) {
				log.Debug(fmt.Sprintf("Proxying: %s_%s", network, address))
				return drFunc(ctx, network, address)
			}
			if p.matchNK(dl, network, address) {
				log.Debug(fmt.Sprintf("Directing: %s_%s", network, address))
				return new(net.Dialer).DialContext(ctx, network, address)
			}
			log.Debug(fmt.Sprintf("Blocking: %s_%s", network, address))
			return nil, fmt.Errorf("router: blocking [%s]%s", network, address)
		}
	default:
		log.Debug(fmt.Sprintf("bad default method: %s", df))
		return nil, nil, nil, errors.New("router: bad default method")
	}
	return ctx, ln, drFuncX, nil
}

func (p *routerPlugin) matchNK(rl [][2]*regexp.Regexp, network, address string) bool {
	for _, nk := range rl {
		if nk[0].MatchString(network) && nk[1].MatchString(address) {
			return true
		}
	}
	return false
}

func (p *routerPlugin) makeRegexp() (pl, dl, bl [][2]*regexp.Regexp, err error) {
	cfg := p.cfg
	pl = make([][2]*regexp.Regexp, 0, len(cfg.ProxyList))
	for _, nk := range cfg.ProxyList {
		nc, err := regexp.Compile(nk[0])
		if err != nil {
			return nil, nil, nil, err
		}
		ac, err := regexp.Compile(nk[1])
		if err != nil {
			return nil, nil, nil, err
		}
		pl = append(pl, [2]*regexp.Regexp{nc, ac})
	}
	dl = make([][2]*regexp.Regexp, 0, len(cfg.DirectList))
	for _, nk := range cfg.DirectList {
		nc, err := regexp.Compile(nk[0])
		if err != nil {
			return nil, nil, nil, err
		}
		ac, err := regexp.Compile(nk[1])
		if err != nil {
			return nil, nil, nil, err
		}
		dl = append(dl, [2]*regexp.Regexp{nc, ac})
	}
	bl = make([][2]*regexp.Regexp, 0, len(cfg.BlockList))
	for _, nk := range cfg.BlockList {
		nc, err := regexp.Compile(nk[0])
		if err != nil {
			return nil, nil, nil, err
		}
		ac, err := regexp.Compile(nk[1])
		if err != nil {
			return nil, nil, nil, err
		}
		bl = append(bl, [2]*regexp.Regexp{nc, ac})
	}
	return pl, dl, bl, nil
}
