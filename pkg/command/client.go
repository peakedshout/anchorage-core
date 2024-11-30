package command

import (
	"errors"
	"github.com/peakedshout/anchorage-core/pkg/sdk"
	"github.com/peakedshout/go-pandorasbox/tool/dcopy"
	"github.com/peakedshout/go-pandorasbox/xnet/xtool/xhttp"
)

func (c *Cmd) addClient(ctx *xhttp.Context) error {
	var info sdk.ClientConfig
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	if info.ClientConfig == nil {
		return errors.New("nil data")
	}
	err = info.Check()
	if err != nil {
		return err
	}
	for _, config := range info.Listen {
		err = config.Check()
		if err != nil {
			return err
		}
	}
	for _, config := range info.Dial {
		err = config.Check()
		if err != nil {
			return err
		}
	}
	for _, config := range info.Proxy {
		err = config.Check()
		if err != nil {
			return err
		}
	}
	for _, config := range info.Plugin {
		err = config.Check()
		if err != nil {
			return err
		}
	}
	id, err := c._sdk.AddClient(&info)
	if err != nil {
		return err
	}
	return ctx.WriteAny(IdData[any]{Id: id})
}

func (c *Cmd) addClient2(ctx *xhttp.Context) error {
	var info sdk.ClientConfigUnit
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	if info.ClientConfig == nil {
		return errors.New("nil data")
	}
	err = info.Check()
	if err != nil {
		return err
	}
	id, err := c._sdk.AddClient2(&info)
	if err != nil {
		return err
	}
	return ctx.WriteAny(IdData[any]{Id: id})
}

func (c *Cmd) delClient(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	return c._sdk.DelClient(info.Id)
}

func (c *Cmd) startClient(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	return c._sdk.StartClient(info.Id)
}

func (c *Cmd) startClient2(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	return c._sdk.StartClient2(info.Id)
}

func (c *Cmd) stopClient(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	return c._sdk.StopClient(info.Id)
}

func (c *Cmd) reloadClient(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	return c._sdk.ReloadClient(info.Id)
}

func (c *Cmd) reloadClient2(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	return c._sdk.ReloadClient2(info.Id)
}

func (c *Cmd) updateClient(ctx *xhttp.Context) error {
	var info IdData[*sdk.ClientConfig]
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
	for _, config := range info.Data.Listen {
		err = config.Check()
		if err != nil {
			return err
		}
	}
	for _, config := range info.Data.Dial {
		err = config.Check()
		if err != nil {
			return err
		}
	}
	for _, config := range info.Data.Proxy {
		err = config.Check()
		if err != nil {
			return err
		}
	}
	for _, config := range info.Data.Plugin {
		err = config.Check()
		if err != nil {
			return err
		}
	}
	return c._sdk.UpdateClient(info.Id, func(cfg *sdk.ClientConfig) {
		_ = dcopy.CopySDT(*info.Data, cfg)
	})
}

func (c *Cmd) updateClient2(ctx *xhttp.Context) error {
	var info IdData[*sdk.ClientConfigUnit]
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
	return c._sdk.UpdateClient2(info.Id, func(cfg *sdk.ClientConfigUnit) {
		_ = dcopy.CopySDT(*info.Data, cfg)
	})
}

func (c *Cmd) configClient(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	cfg, err := c._sdk.GetConfigClient(info.Id)
	if err != nil {
		return err
	}
	return ctx.WriteAny(cfg)
}

func (c *Cmd) configClient2(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	cfg, err := c._sdk.GetConfigClient2(info.Id)
	if err != nil {
		return err
	}
	return ctx.WriteAny(cfg)
}
