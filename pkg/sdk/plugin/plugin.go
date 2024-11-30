package plugin

import (
	"context"
	"github.com/peakedshout/go-pandorasbox/tool/tmap"
	"net"
	"sort"
)

const Name = "name"

func RegisterDialPlugin(p DialPlugin) {
	dialPluginMap.Store(p.Name(), p)
}

func GetDialPlugin(name string) (DialPlugin, bool) {
	return dialPluginMap.Load(name)
}

func ListDialPlugin() []string {
	var sl []string
	dialPluginMap.Range(func(key string, _ DialPlugin) bool {
		sl = append(sl, key)
		return true
	})
	sort.Strings(sl)
	return sl
}

var (
	dialPluginMap tmap.SyncMap[string, DialPlugin]
)

func RegisterProxyPlugin(p ProxyPlugin) {
	proxyPluginMap.Store(p.Name(), p)
}

func GetProxyPlugin(name string) (ProxyPlugin, bool) {
	return proxyPluginMap.Load(name)
}

func ListProxyPlugin() []string {
	var sl []string
	proxyPluginMap.Range(func(key string, _ ProxyPlugin) bool {
		sl = append(sl, key)
		return true
	})
	sort.Strings(sl)
	return sl
}

var (
	proxyPluginMap tmap.SyncMap[string, ProxyPlugin]
)

type DialFunc = func(ctx context.Context, network string, address string) (net.Conn, error)

type DialPlugin interface {
	Name() string
	TempDialCfg() any
	LoadDialCfg(str string) (DialPlugin, error)
	DialServe(ctx context.Context, ln net.Listener, drFunc DialFunc) error
	DialUpgrade(ctx context.Context, ln net.Listener, drFunc DialFunc) (context.Context, net.Listener, DialFunc, error)
}

type ProxyPlugin interface {
	Name() string
	TempProxyCfg() any
	LoadProxyCfg(str string) (ProxyPlugin, error)
	ProxyServe(ctx context.Context, ln net.Listener, drFunc DialFunc) error
	ProxyUpgrade(ctx context.Context, ln net.Listener, drFunc DialFunc) (context.Context, net.Listener, DialFunc, error)
}

//type ListenPlugin interface {
//	Name() string
//	TempListenCfg() any
//	LoadListenCfg(str string) (ProxyPlugin, error)
//	ListenServe(ctx context.Context, rwc io.ReadWriteCloser, network string, address string) error
//}
