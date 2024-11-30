package command

import (
	"errors"
	"github.com/peakedshout/anchorage-core/pkg/sdk"
	"github.com/peakedshout/go-pandorasbox/tool/dcopy"
	"github.com/peakedshout/go-pandorasbox/xnet/xtool/xhttp"
)

func (c *Cmd) addListen(ctx *xhttp.Context) error {
	var info IdSubData[string, *sdk.ListenConfig]
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
	sid, err := c._sdk.AddListen(info.Id, info.Data)
	if err != nil {
		return err
	}
	return ctx.WriteAny(IdSubData[string, any]{Id: info.Id, Sub: sid})
}

func (c *Cmd) delListen(ctx *xhttp.Context) error {
	var index IdSubData[string, any]
	err := ctx.Bind(&index)
	if err != nil {
		return err
	}
	return c._sdk.DelListen(index.Id, index.Sub)
}

func (c *Cmd) startListen(ctx *xhttp.Context) error {
	var index IdSubData[string, any]
	err := ctx.Bind(&index)
	if err != nil {
		return err
	}
	return c._sdk.StartListen(index.Id, index.Sub)
}

func (c *Cmd) stopListen(ctx *xhttp.Context) error {
	var index IdSubData[string, any]
	err := ctx.Bind(&index)
	if err != nil {
		return err
	}
	return c._sdk.StopListen(index.Id, index.Sub)
}

func (c *Cmd) reloadListen(ctx *xhttp.Context) error {
	var index IdSubData[string, any]
	err := ctx.Bind(&index)
	if err != nil {
		return err
	}
	return c._sdk.ReloadListen(index.Id, index.Sub)
}

func (c *Cmd) updateListen(ctx *xhttp.Context) error {
	var index IdSubData[string, *sdk.ListenConfig]
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
	return c._sdk.UpdateListen(index.Id, index.Sub, func(cfg *sdk.ListenConfig) {
		_ = dcopy.CopySDT(*index.Data, cfg)
	})
}

func (c *Cmd) configListen(ctx *xhttp.Context) error {
	var index IdSubData[string, any]
	err := ctx.Bind(&index)
	if err != nil {
		return err
	}
	cfg, err := c._sdk.GetConfigListen(index.Id, index.Sub)
	if err != nil {
		return err
	}
	return ctx.WriteAny(cfg)
}
