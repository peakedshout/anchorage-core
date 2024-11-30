package command

import (
	"errors"
	"github.com/peakedshout/anchorage-core/pkg/sdk"
	"github.com/peakedshout/go-pandorasbox/tool/dcopy"
	"github.com/peakedshout/go-pandorasbox/xnet/xtool/xhttp"
)

func (c *Cmd) addDial(ctx *xhttp.Context) error {
	var info IdSubData[string, *sdk.DialConfig]
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
	sid, err := c._sdk.AddDial(info.Id, info.Data)
	if err != nil {
		return err
	}
	return ctx.WriteAny(IdSubData[string, any]{Id: info.Id, Sub: sid})
}

func (c *Cmd) delDial(ctx *xhttp.Context) error {
	var index IdSubData[string, any]
	err := ctx.Bind(&index)
	if err != nil {
		return err
	}
	return c._sdk.DelDial(index.Id, index.Sub)
}

func (c *Cmd) startDial(ctx *xhttp.Context) error {
	var index IdSubData[string, any]
	err := ctx.Bind(&index)
	if err != nil {
		return err
	}
	return c._sdk.StartDial(index.Id, index.Sub)
}

func (c *Cmd) stopDial(ctx *xhttp.Context) error {
	var index IdSubData[string, any]
	err := ctx.Bind(&index)
	if err != nil {
		return err
	}
	return c._sdk.StopDial(index.Id, index.Sub)
}

func (c *Cmd) reloadDial(ctx *xhttp.Context) error {
	var index IdSubData[string, any]
	err := ctx.Bind(&index)
	if err != nil {
		return err
	}
	return c._sdk.ReloadDial(index.Id, index.Sub)
}

func (c *Cmd) updateDial(ctx *xhttp.Context) error {
	var index IdSubData[string, *sdk.DialConfig]
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
	return c._sdk.UpdateDial(index.Id, index.Sub, func(cfg *sdk.DialConfig) {
		_ = dcopy.CopySDT(*index.Data, cfg)
	})
}

func (c *Cmd) configDial(ctx *xhttp.Context) error {
	var index IdSubData[string, any]
	err := ctx.Bind(&index)
	if err != nil {
		return err
	}
	cfg, err := c._sdk.GetConfigDial(index.Id, index.Sub)
	if err != nil {
		return err
	}
	return ctx.WriteAny(cfg)
}
