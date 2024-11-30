package plugin

import (
	"context"
	"encoding/json"
	"github.com/peakedshout/go-pandorasbox/tool/dcopy"
	"github.com/peakedshout/go-pandorasbox/xnet/proxy/httpproxy"
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"net"
	"net/http"
)

const NameHttpProxy = "httpproxy"

func init() {
	RegisterDialPlugin(new(httpProxyPlugin))
	RegisterProxyPlugin(new(httpProxyPlugin))
}

type HttpProxyPluginCfg struct {
	Name string `json:"name" yaml:"name" comment:"plugin name"`

	HttpAuthPasswordList [][2]string `json:"Auth" yaml:"Auth" comment:"Auth list"`
	OSProxySettings      bool        `json:"OSProxySettings" yaml:"OSProxySettings" comment:"OS proxy settings"`
}

type httpProxyPlugin struct {
	cfg HttpProxyPluginCfg
}

func (p *httpProxyPlugin) Name() string {
	return NameHttpProxy
}

func (p *httpProxyPlugin) TempProxyCfg() any {
	return p.cfg
}

func (p *httpProxyPlugin) LoadProxyCfg(str string) (ProxyPlugin, error) {
	var cfg HttpProxyPluginCfg
	err := json.Unmarshal([]byte(str), &cfg)
	if err != nil {
		return nil, err
	}
	return &httpProxyPlugin{cfg: cfg}, nil
}

func (p *httpProxyPlugin) ProxyServe(ctx context.Context, ln net.Listener, drFunc DialFunc) error {
	return p.serve(ctx, ln, drFunc)
}

func (p *httpProxyPlugin) ProxyUpgrade(ctx context.Context, ln net.Listener, drFunc DialFunc) (context.Context, net.Listener, DialFunc, error) {
	return p.upgrade(ctx, ln, drFunc)
}

func (p *httpProxyPlugin) TempDialCfg() any {
	return p.tempCfg()
}

func (p *httpProxyPlugin) LoadDialCfg(str string) (DialPlugin, error) {
	var cfg HttpProxyPluginCfg
	err := json.Unmarshal([]byte(str), &cfg)
	if err != nil {
		return nil, err
	}
	return &httpProxyPlugin{cfg: cfg}, nil
}

func (p *httpProxyPlugin) DialServe(ctx context.Context, ln net.Listener, drFunc DialFunc) error {
	return p.serve(ctx, ln, drFunc)
}

func (p *httpProxyPlugin) DialUpgrade(ctx context.Context, ln net.Listener, drFunc DialFunc) (context.Context, net.Listener, DialFunc, error) {
	return p.upgrade(ctx, ln, drFunc)
}

func (p *httpProxyPlugin) tempCfg() any {
	return HttpProxyPluginCfg{
		Name: "httpproxy",
	}
}

func (p *httpProxyPlugin) serve(ctx context.Context, ln net.Listener, drFunc DialFunc) error {
	cfg := p.cfg
	scfg := &httpproxy.ServerConfig{
		Forward: xnetutil.NewCallBackDialer(drFunc),
	}
	if len(cfg.HttpAuthPasswordList) > 0 {
		sl := dcopy.CopyT(cfg.HttpAuthPasswordList)
		scfg.ReqAuthCb = func(req *http.Request) bool {
			user, pass, ok := proxyBasicAuth(req)
			if ok {
				for _, up := range sl {
					if up[0] == user && up[1] == pass {
						return true
					}
				}
			}
			return false
		}
	}
	server, err := httpproxy.NewServerContext(ctx, scfg)
	if err != nil {
		return err
	}

	if cfg.OSProxySettings && httpproxy_osProxySetting != nil {
		log := getPluginLog(ctx, p)
		dFunc, err := httpproxy_osProxySetting(ln)
		if err != nil {
			log.Warn("os proxy setting enable failed: " + err.Error())
			return err
		}
		log.Info("os proxy setting enable successes")
		defer func() {
			err = dFunc()
			if err != nil {
				log.Warn("os proxy setting disable failed: " + err.Error())
			} else {
				log.Info("os proxy setting disable successes")
			}
		}()
	}

	return server.Serve(ln)
}

func (p *httpProxyPlugin) upgrade(ctx context.Context, ln net.Listener, drFunc DialFunc) (context.Context, net.Listener, DialFunc, error) {
	return ctx, ln, drFunc, nil
}

func proxyBasicAuth(req *http.Request) (username, password string, ok bool) {
	username, password, ok = httpproxy.ProxyBasicAuth(req)
	if ok {
		req.Header.Del("Proxy-Authorization")
	}
	return username, password, ok
}
