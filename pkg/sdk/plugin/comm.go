package plugin

import (
	"context"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/logger"
	"net"
	"net/http"
	"runtime"
	"time"
)

var _noProxyHttpClientDialer = &net.Dialer{
	Timeout:   30 * time.Second,
	KeepAlive: 30 * time.Second,
}

var _noProxyHttpClient = http.Client{
	Transport: &http.Transport{
		DialContext:           _noProxyHttpClientDialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

func quickGC() func() {
	runtime.GC()
	runtime.GC()
	return func() {
		runtime.GC()
		runtime.GC()
	}
}

type pluginLog struct {
	log    logger.Logger
	prefix string
}

func (p *pluginLog) Debug(a ...any) {
	p.log.Debug(append([]any{p.prefix}, a...)...)
}

func (p *pluginLog) Info(a ...any) {
	p.log.Debug(append([]any{p.prefix}, a...)...)
}

func (p *pluginLog) Warn(a ...any) {
	p.log.Debug(append([]any{p.prefix}, a...)...)
}

func (p *pluginLog) Log(a ...any) {
	p.log.Debug(append([]any{p.prefix}, a...)...)
}

func (p *pluginLog) Error(a ...any) {
	p.log.Debug(append([]any{p.prefix}, a...)...)
}

func (p *pluginLog) Fatal(a ...any) {
	p.log.Debug(append([]any{p.prefix}, a...)...)
}

func getPluginLog(ctx context.Context, i interface{ Name() string }) *pluginLog {
	return &pluginLog{
		log:    logger.GetLogger(ctx),
		prefix: fmt.Sprintf("[>%s<]", i.Name()),
	}
}
