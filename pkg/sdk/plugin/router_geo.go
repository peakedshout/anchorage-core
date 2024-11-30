package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
)

const NameGeoRouter = "geo_router"

func init() {
	RegisterDialPlugin(new(geoRouterPlugin))
	RegisterProxyPlugin(new(geoRouterPlugin))
}

const (
	GeoTypeGeoip   = "geoip"
	GeoTypeGeostie = "geosite"
)

type GeoRouterPluginCfg struct {
	Name string `json:"name" yaml:"name" comment:"plugin name"`

	DefaultMethod string      `json:"default" yaml:"default" comment:"default method"`
	ProxyList     [][2]string `json:"proxy" yaml:"proxy" comment:"proxy list"`
	DirectList    [][2]string `json:"direct" yaml:"direct" comment:"direct list"`
	BlockList     [][2]string `json:"block" yaml:"block" comment:"block list"`

	GeoipPath   string `json:"geoipPath" yaml:"geoipPath" comment:"geoip.dat path"`
	GeositePath string `json:"geositePath" yaml:"geositePath" comment:"geosite.dat path"`
}

type geoRouterPlugin struct {
	cfg GeoRouterPluginCfg
}

func (p *geoRouterPlugin) Name() string {
	return NameGeoRouter
}

func (p *geoRouterPlugin) TempDialCfg() any {
	return p.tempCfg()
}

func (p *geoRouterPlugin) LoadDialCfg(str string) (DialPlugin, error) {
	var cfg GeoRouterPluginCfg
	err := json.Unmarshal([]byte(str), &cfg)
	if err != nil {
		return nil, err
	}
	return &geoRouterPlugin{cfg: cfg}, nil
}

func (p *geoRouterPlugin) DialServe(ctx context.Context, ln net.Listener, drFunc DialFunc) error {
	return p.serve(ctx, ln, drFunc)
}

func (p *geoRouterPlugin) DialUpgrade(ctx context.Context, ln net.Listener, drFunc DialFunc) (context.Context, net.Listener, DialFunc, error) {
	return p.upgrade(ctx, ln, drFunc)
}

func (p *geoRouterPlugin) TempProxyCfg() any {
	return p.tempCfg()
}

func (p *geoRouterPlugin) LoadProxyCfg(str string) (ProxyPlugin, error) {
	var cfg GeoRouterPluginCfg
	err := json.Unmarshal([]byte(str), &cfg)
	if err != nil {
		return nil, err
	}
	return &geoRouterPlugin{cfg: cfg}, nil
}

func (p *geoRouterPlugin) ProxyServe(ctx context.Context, ln net.Listener, drFunc DialFunc) error {
	return p.serve(ctx, ln, drFunc)
}

func (p *geoRouterPlugin) ProxyUpgrade(ctx context.Context, ln net.Listener, drFunc DialFunc) (context.Context, net.Listener, DialFunc, error) {
	return p.upgrade(ctx, ln, drFunc)
}

func (p *geoRouterPlugin) tempCfg() any {
	return GeoRouterPluginCfg{
		Name: NameGeoRouter,
	}
}

func (p *geoRouterPlugin) serve(ctx context.Context, ln net.Listener, drFunc DialFunc) error {
	return errors.New("no implement")
}

func (p *geoRouterPlugin) upgrade(ctx context.Context, ln net.Listener, drFunc DialFunc) (context.Context, net.Listener, DialFunc, error) {
	log := getPluginLog(ctx, p)
	cfg := p.cfg
	df := cfg.DefaultMethod
	if df == "" {
		df = MethodProxy
	}
	ipm := make(map[string]struct{})
	stiem := make(map[string]struct{})
	var err error
	err = p.classify(cfg.ProxyList, ipm, stiem)
	if err != nil {
		log.Debug(err)
		return nil, nil, nil, err
	}
	err = p.classify(cfg.DirectList, ipm, stiem)
	if err != nil {
		log.Debug(err)
		return nil, nil, nil, err
	}
	err = p.classify(cfg.BlockList, ipm, stiem)
	if err != nil {
		log.Debug(err)
		return nil, nil, nil, err
	}
	var geoip map[string][]*ipCIDR
	var geostie map[string][]matcher
	if len(ipm) != 0 {
		geoip, err = loadGeoipFromFile(cfg.GeoipPath, ipm)
		if err != nil {
			log.Debug(err)
			return nil, nil, nil, err
		}
	}
	if len(stiem) != 0 {
		geostie, err = loadGeositeFromFile(cfg.GeositePath, stiem)
		if err != nil {
			log.Debug(err)
			return nil, nil, nil, err
		}
	}
	var drFuncX func(ctx context.Context, network string, address string) (net.Conn, error)
	switch df {
	case MethodProxy:
		drFuncX = func(ctx context.Context, network string, address string) (net.Conn, error) {
			if p.matchNK(cfg.BlockList, geoip, geostie, network, address) {
				log.Debug(fmt.Sprintf("Blocking: %s_%s", network, address))
				return nil, fmt.Errorf("router: blocking [%s]%s", network, address)
			}
			if p.matchNK(cfg.DirectList, geoip, geostie, network, address) {
				log.Debug(fmt.Sprintf("Directing: %s_%s", network, address))
				return new(net.Dialer).DialContext(ctx, network, address)
			}
			log.Debug(fmt.Sprintf("Proxying: %s_%s", network, address))
			return drFunc(ctx, network, address)
		}
	case MethodDirect:
		drFuncX = func(ctx context.Context, network string, address string) (net.Conn, error) {
			if p.matchNK(cfg.BlockList, geoip, geostie, network, address) {
				log.Debug(fmt.Sprintf("Blocking: %s_%s", network, address))
				return nil, fmt.Errorf("router: blocking [%s]%s", network, address)
			}
			if p.matchNK(cfg.ProxyList, geoip, geostie, network, address) {
				log.Debug(fmt.Sprintf("Proxying: %s_%s", network, address))
				return drFunc(ctx, network, address)
			}
			log.Debug(fmt.Sprintf("Directing: %s_%s", network, address))
			return new(net.Dialer).DialContext(ctx, network, address)
		}
	case MethodBlock:
		drFuncX = func(ctx context.Context, network string, address string) (net.Conn, error) {
			if p.matchNK(cfg.ProxyList, geoip, geostie, network, address) {
				log.Debug(fmt.Sprintf("Proxying: %s_%s", network, address))
				return drFunc(ctx, network, address)
			}
			if p.matchNK(cfg.DirectList, geoip, geostie, network, address) {
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

func (p *geoRouterPlugin) matchNK(rl [][2]string, geoip map[string][]*ipCIDR, geosite map[string][]matcher, network, address string) bool {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.To4() != nil {
		ip = ip.To4()
	}
	for _, sl := range rl {
		switch sl[0] {
		case GeoTypeGeoip:
			if ip == nil {
				continue
			}
			for _, cidr := range geoip[sl[1]] {
				if cidr.Contains(ip) {
					return true
				}
			}
		case GeoTypeGeostie:
			for _, m := range geosite[sl[1]] {
				if m.match(host) {
					return true
				}
			}
		}
	}
	return false
}

func (p *geoRouterPlugin) classify(list [][2]string, ipm map[string]struct{}, stiem map[string]struct{}) error {
	for _, sl := range list {
		switch sl[0] {
		case GeoTypeGeoip:
			ipm[sl[1]] = struct{}{}
		case GeoTypeGeostie:
			stiem[sl[1]] = struct{}{}
		default:
			return fmt.Errorf("invaild geo type: %s", sl[0])
		}
	}
	return nil
}
