package plugin

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const NameDNS = "dns"

func init() {
	RegisterDialPlugin(new(dnsPlugin))
	RegisterProxyPlugin(new(dnsPlugin))
}

const (
	DnsNetworkTypeUdp  = "udp"
	DnsNetworkTypeTcp  = "tcp"
	DnsNetworkTypeTls  = "tls"
	DnsNetworkTypeUrl  = "url"
	DnsNetworkTypeUrlQ = "url-q"
)

type DnsPluginCfg struct {
	Name string `json:"name" yaml:"name" comment:"plugin name"`

	Network     string `json:"network" yaml:"network" comment:"dns server network"`
	DNSServer   string `json:"dnsServer" yaml:"dnsServer" comment:"dns server"`
	InsecureTls bool   `json:"insecureTls" yaml:"insecureTls" comment:"dns tls insecure skip verify"`
	PreV6       bool   `json:"preV6" yaml:"preV6" comment:"enable and precedence ipv6"`
}

type dnsPlugin struct {
	cfg DnsPluginCfg
}

func (p *dnsPlugin) Name() string {
	return NameDNS
}

func (p *dnsPlugin) TempDialCfg() any {
	return p.tempCfg()
}

func (p *dnsPlugin) LoadDialCfg(str string) (DialPlugin, error) {
	var cfg DnsPluginCfg
	err := json.Unmarshal([]byte(str), &cfg)
	if err != nil {
		return nil, err
	}
	return &dnsPlugin{cfg: cfg}, nil
}

func (p *dnsPlugin) DialServe(ctx context.Context, ln net.Listener, drFunc DialFunc) error {
	return p.serve(ctx, ln, drFunc)
}

func (p *dnsPlugin) DialUpgrade(ctx context.Context, ln net.Listener, drFunc DialFunc) (context.Context, net.Listener, DialFunc, error) {
	return p.upgrade(ctx, ln, drFunc)
}

func (p *dnsPlugin) TempProxyCfg() any {
	return p.tempCfg()
}

func (p *dnsPlugin) LoadProxyCfg(str string) (ProxyPlugin, error) {
	var cfg DnsPluginCfg
	err := json.Unmarshal([]byte(str), &cfg)
	if err != nil {
		return nil, err
	}
	return &dnsPlugin{cfg: cfg}, nil
}

func (p *dnsPlugin) ProxyServe(ctx context.Context, ln net.Listener, drFunc DialFunc) error {
	return p.serve(ctx, ln, drFunc)
}

func (p *dnsPlugin) ProxyUpgrade(ctx context.Context, ln net.Listener, drFunc DialFunc) (context.Context, net.Listener, DialFunc, error) {
	return p.upgrade(ctx, ln, drFunc)
}

func (p *dnsPlugin) tempCfg() any {
	return DnsPluginCfg{
		Name: NameDNS,
	}
}

func (p *dnsPlugin) serve(ctx context.Context, ln net.Listener, drFunc DialFunc) error {
	return errors.New("no implement")
}

func (p *dnsPlugin) upgrade(ctx context.Context, ln net.Listener, drFunc DialFunc) (context.Context, net.Listener, DialFunc, error) {
	cfg := p.cfg
	c := &dnsClient{
		network: cfg.Network,
		server:  cfg.DNSServer,
		preV6:   cfg.PreV6,
		rMap:    map[string]*resolveUnit{},
	}
	if c.network == "" {
		c.network = DnsNetworkTypeUdp
	}
	switch c.network {
	case DnsNetworkTypeUdp, DnsNetworkTypeTcp, DnsNetworkTypeTls:
		c.dnsC = new(dns.Client)
		c.dnsC.Net = c.network
		if c.network == DnsNetworkTypeTls {
			c.tlsCfg = &tls.Config{InsecureSkipVerify: cfg.InsecureTls}
			c.dnsC.Net = "tcp-" + c.network
			c.dnsC.TLSConfig = c.tlsCfg
		}
	case DnsNetworkTypeUrl:
		c.tlsCfg = &tls.Config{InsecureSkipVerify: cfg.InsecureTls}
		c.dohC = new(http.Client)
		c.dohC.Transport = &http.Transport{
			DialContext:           _noProxyHttpClientDialer.DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSClientConfig:       c.tlsCfg,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	case DnsNetworkTypeUrlQ:
		c.tlsCfg = &tls.Config{InsecureSkipVerify: cfg.InsecureTls, NextProtos: []string{http3.NextProtoH3}}
		c.doqC = new(http3.SingleDestinationRoundTripper)
	default:
		return nil, nil, nil, errors.New("invalid dns network type")
	}

	log := getPluginLog(ctx, p)
	drFuncX := func(ctx context.Context, network string, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		ip := net.ParseIP(host)
		if ip != nil {
			return drFunc(ctx, network, address)
		}
		for i := 0; i < 2; i++ {
			if ctx.Err() != nil {
				log.Debug(fmt.Sprintf("Resolve %s domain ip failed: %s", host, ctx.Err()))
				return nil, ctx.Err()
			}
			addr, efunc, err := c.getAddress(ctx, host)
			if err != nil {
				log.Debug(fmt.Sprintf("Resolve %s domain ip failed: %s", host, err))
				return nil, err
			}
			conn, err := drFunc(ctx, network, fmt.Sprintf("%s:%s", addr, port))
			if err != nil {
				efunc()
			} else {
				log.Debug(fmt.Sprintf("Resolve %s domain ip successed: %s", host, addr))
				return conn, nil
			}
		}
		err = fmt.Errorf("%s does not have an IP address available", host)
		log.Debug(fmt.Sprintf("Resolve %s domain ip failed: %s", host, err))
		return nil, err
	}
	return ctx, ln, drFuncX, nil
}

type dnsClient struct {
	dnsC   *dns.Client
	dohC   *http.Client
	doqC   *http3.SingleDestinationRoundTripper
	tlsCfg *tls.Config

	network string
	server  string
	preV6   bool

	rMux sync.Mutex
	rMap map[string]*resolveUnit
}

func (d *dnsClient) getAddress(ctx context.Context, host string) (string, func(), error) {
	d.rMux.Lock()
	d.gcResolve()
	r, ok := d.rMap[host]
	if !ok {
		r = &resolveUnit{
			name: host,
		}
		d.rMap[host] = r
	}
	d.rMux.Unlock()
	for {
		addr := r.get(d.preV6)
		if addr != "" {
			return addr, func() {
				r.del(d.preV6, addr)
			}, nil
		}
		err := d.newResolve(ctx, host, r)
		if err != nil {
			return "", nil, err
		}
	}
}

func (d *dnsClient) newResolve(ctx context.Context, host string, r *resolveUnit) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	var resolveFunc func(ctx context.Context, host string, t uint16, ipm map[string]*time.Time) error
	if d.dnsC != nil {
		resolveFunc = d.dnsResolve
	} else if d.dohC != nil {
		resolveFunc = d.dohResolve
	} else if d.doqC != nil {
		resolveFunc = d.doqResolve
	}
	r.lastT = time.Time{}
	if d.preV6 {
		r.ipv6Map = make(map[string]*time.Time)
		err := resolveFunc(ctx, host, dns.TypeAAAA, r.ipv6Map)
		if err != nil {
			return err
		}
		for _, t := range r.ipv6Map {
			if t.After(r.lastT) {
				r.lastT = *t
			}
		}
	}
	r.ipv4Map = make(map[string]*time.Time)
	err := resolveFunc(ctx, host, dns.TypeA, r.ipv4Map)
	if err != nil {
		return err
	}
	for _, t := range r.ipv4Map {
		if t.After(r.lastT) {
			r.lastT = *t
		}
	}
	return nil
}

func (d *dnsClient) dnsResolve(ctx context.Context, host string, t uint16, ipm map[string]*time.Time) error {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(host), t)
	r, since, err := d.dnsC.ExchangeContext(ctx, m, d.server)
	if err != nil {
		return err
	}
	for _, rr := range r.Answer {
		et := time.Now().Add(time.Duration(rr.Header().Ttl)*time.Second - since)
		switch x := rr.(type) {
		case *dns.A:
			ipm[x.A.String()] = &et
		case *dns.AAAA:
			ipm[x.AAAA.String()] = &et
		case *dns.CNAME:
			err = d.dnsResolve(ctx, x.Target, t, ipm)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *dnsClient) dohResolve(ctx context.Context, host string, t uint16, ipm map[string]*time.Time) error {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(host), t)
	msg, err := m.Pack()
	if err != nil {
		return err
	}
	encodedMsg := base64.RawURLEncoding.EncodeToString(msg)
	requestURL := fmt.Sprintf("%s?dns=%s", d.server, encodedMsg)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}
	now := time.Now()
	resp, err := d.dohC.Do(req)
	if err != nil {
		return err
	}
	bs, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return err
	}
	err = m.Unpack(bs)
	if err != nil {
		return err
	}
	since := time.Since(now)
	for _, rr := range m.Answer {
		et := time.Now().Add(time.Duration(rr.Header().Ttl)*time.Second - since)
		switch x := rr.(type) {
		case *dns.A:
			ipm[x.A.String()] = &et
		case *dns.AAAA:
			ipm[x.AAAA.String()] = &et
		case *dns.CNAME:
			err = d.dohResolve(ctx, x.Target, t, ipm)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *dnsClient) doqResolve(ctx context.Context, host string, t uint16, ipm map[string]*time.Time) error {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(host), t)
	msg, err := m.Pack()
	if err != nil {
		return err
	}
	encodedMsg := base64.RawURLEncoding.EncodeToString(msg)
	requestURL := fmt.Sprintf("%s?dns=%s", d.server, encodedMsg)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}

	if d.doqC.Connection == nil || d.doqC.Connection.Context().Err() != nil {
		u, err := url.ParseRequestURI(d.server)
		if err != nil {
			return err
		}
		connection, err := quic.DialAddr(context.Background(), u.Host, d.tlsCfg, nil)
		if err != nil {
			return err
		}
		d.doqC = &http3.SingleDestinationRoundTripper{Connection: connection}
		//d.doqC.Connection = connection
	}

	now := time.Now()
	stream, err := d.doqC.OpenRequestStream(ctx)
	if err != nil {
		return err
	}
	defer stream.Close()
	err = stream.SendRequestHeader(req)
	if err != nil {
		return err
	}
	resp, err := stream.ReadResponse()
	if err != nil {
		return err
	}
	bs, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return err
	}
	err = m.Unpack(bs)
	if err != nil {
		return err
	}
	since := time.Since(now)
	for _, rr := range m.Answer {
		et := time.Now().Add(time.Duration(rr.Header().Ttl)*time.Second - since)
		switch x := rr.(type) {
		case *dns.A:
			ipm[x.A.String()] = &et
		case *dns.AAAA:
			ipm[x.AAAA.String()] = &et
		case *dns.CNAME:
			err = d.doqResolve(ctx, x.Target, t, ipm)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *dnsClient) gcResolve() {
	now := time.Now()
	var dl []string
	for host, unit := range d.rMap {
		if now.After(unit.lastT) {
			dl = append(dl, host)
		}
	}
	for _, host := range dl {
		delete(d.rMap, host)
	}
}

type resolveUnit struct {
	mux     sync.Mutex
	name    string
	ipv4Map map[string]*time.Time
	ipv6Map map[string]*time.Time
	lastT   time.Time
}

func (r *resolveUnit) get(v6 bool) string {
	r.mux.Lock()
	defer r.mux.Unlock()
	now := time.Now()
	var xaddr string
	if v6 {
		var dl []string
		for addr, t := range r.ipv6Map {
			if t.After(now) {
				xaddr = addr
			} else {
				dl = append(dl, addr)
			}
		}
		for _, s := range dl {
			delete(r.ipv6Map, s)
		}
		if xaddr != "" {
			return xaddr
		}
	}
	var dl []string
	for addr, t := range r.ipv4Map {
		if t.After(now) {
			xaddr = addr
		} else {
			dl = append(dl, addr)
		}
	}
	for _, s := range dl {
		delete(r.ipv6Map, s)
	}
	return xaddr
}

func (r *resolveUnit) del(v6 bool, address string) {
	r.mux.Lock()
	defer r.mux.Unlock()
	if v6 {
		delete(r.ipv6Map, address)
	}
	delete(r.ipv4Map, address)
}
