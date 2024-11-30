package command

import (
	"errors"
	"github.com/peakedshout/anchorage-core/pkg/sdk"
	"github.com/peakedshout/go-pandorasbox/tool/dcopy"
	"github.com/peakedshout/go-pandorasbox/xnet/xtool/xhttp"
)

func (c *Cmd) addServer(ctx *xhttp.Context) error {
	var info sdk.ServerConfig
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	if info.ServerConfig == nil {
		return errors.New("nil data")
	}
	err = info.Check()
	if err != nil {
		return err
	}
	id, err := c._sdk.AddServer(&info)
	if err != nil {
		return err
	}
	return ctx.WriteAny(IdData[any]{Id: id})
}

func (c *Cmd) delServer(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	return c._sdk.DelServer(info.Id)
}

func (c *Cmd) startServer(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	return c._sdk.StartServer(info.Id)
}

func (c *Cmd) stopServer(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	return c._sdk.StopServer(info.Id)
}

func (c *Cmd) reloadServer(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	return c._sdk.ReloadServer(info.Id)
}

func (c *Cmd) updateServer(ctx *xhttp.Context) error {
	var info IdData[*sdk.ServerConfig]
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
	return c._sdk.UpdateServer(info.Id, func(cfg *sdk.ServerConfig) {
		_ = dcopy.CopySDT(*info.Data, cfg)
	})
}

func (c *Cmd) configServer(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	cfg, err := c._sdk.GetConfigServer(info.Id)
	if err != nil {
		return err
	}
	return ctx.WriteAny(cfg)
}
