package command

import (
	"errors"
	"github.com/peakedshout/anchorage-core/pkg/sdk"
	"github.com/peakedshout/go-pandorasbox/tool/dcopy"
	"github.com/peakedshout/go-pandorasbox/xnet/xtool/xhttp"
)

func (c *Cmd) addProxy(ctx *xhttp.Context) error {
	var info IdSubData[string, *sdk.ProxyConfig]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	if info.Data == nil {
		return errors.New("nil data")
	}
	err = info.Data.Check()
	if err != nil {
		return err
	}
	sid, err := c._sdk.AddProxy(info.Id, info.Data)
	if err != nil {
		return err
	}
	return ctx.WriteAny(IdSubData[string, any]{Id: info.Id, Sub: sid})
}

func (c *Cmd) delProxy(ctx *xhttp.Context) error {
	var index IdSubData[string, any]
	err := ctx.Bind(&index)
	if err != nil {
		return err
	}
	return c._sdk.DelProxy(index.Id, index.Sub)
}

func (c *Cmd) startProxy(ctx *xhttp.Context) error {
	var index IdSubData[string, any]
	err := ctx.Bind(&index)
	if err != nil {
		return err
	}
	return c._sdk.StartProxy(index.Id, index.Sub)
}

func (c *Cmd) stopProxy(ctx *xhttp.Context) error {
	var index IdSubData[string, any]
	err := ctx.Bind(&index)
	if err != nil {
		return err
	}
	return c._sdk.StopProxy(index.Id, index.Sub)
}

func (c *Cmd) reloadProxy(ctx *xhttp.Context) error {
	var index IdSubData[string, any]
	err := ctx.Bind(&index)
	if err != nil {
		return err
	}
	return c._sdk.ReloadProxy(index.Id, index.Sub)
}

func (c *Cmd) updateProxy(ctx *xhttp.Context) error {
	var index IdSubData[string, *sdk.ProxyConfig]
	err := ctx.Bind(&index)
	if err != nil {
		return err
	}
	if index.Data == nil {
		return errors.New("nil data")
	}
	err = index.Data.Check()
	if err != nil {
		return err
	}
	return c._sdk.UpdateProxy(index.Id, index.Sub, func(cfg *sdk.ProxyConfig) {
		_ = dcopy.CopySDT(*index.Data, cfg)
	})
}

func (c *Cmd) configProxy(ctx *xhttp.Context) error {
	var index IdSubData[string, any]
	err := ctx.Bind(&index)
	if err != nil {
		return err
	}
	cfg, err := c._sdk.GetConfigProxy(index.Id, index.Sub)
	if err != nil {
		return err
	}
	return ctx.WriteAny(cfg)
}
