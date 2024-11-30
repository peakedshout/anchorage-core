package command

import (
	"errors"
	"github.com/peakedshout/anchorage-core/pkg/sdk"
	"github.com/peakedshout/go-pandorasbox/tool/dcopy"
	"github.com/peakedshout/go-pandorasbox/tool/hyaml"
	"github.com/peakedshout/go-pandorasbox/xnet/xtool/xhttp"
)

func (c *Cmd) getPluginTempCfg(ctx *xhttp.Context) error {
	var sl [2]string
	err := ctx.Bind(&sl)
	if err != nil {
		return err
	}
	tempCfg, err := c._sdk.GetPluginUnitTempCfg(sl[0], sl[1])
	if err != nil {
		return err
	}
	out, err := hyaml.MarshalWithCommentT(tempCfg)
	if err != nil {
		return err
	}
	_, err = ctx.Write(out)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cmd) listPluginUnit(ctx *xhttp.Context) error {
	var t string
	err := ctx.Bind(&t)
	if err != nil {
		return err
	}
	sl, err := c._sdk.ListPluginUnit(t)
	if err != nil {
		return err
	}
	return ctx.WriteAny(sl)
}

func (c *Cmd) addPlugin(ctx *xhttp.Context) error {
	var info IdSubData[string, *sdk.PluginConfig]
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
	return c._sdk.AddPlugin(info.Id, info.Data)
}

func (c *Cmd) delPlugin(ctx *xhttp.Context) error {
	var index IdSubData[string, any]
	err := ctx.Bind(&index)
	if err != nil {
		return err
	}
	return c._sdk.DelPlugin(index.Id, index.Sub)
}

func (c *Cmd) updatePlugin(ctx *xhttp.Context) error {
	var index IdSubData[string, *sdk.PluginConfig]
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
	return c._sdk.UpdatePlugin(index.Id, index.Sub, func(cfg *sdk.PluginConfig) {
		_ = dcopy.CopySDT(*index.Data, cfg)
	})
}

func (c *Cmd) configPlugin(ctx *xhttp.Context) error {
	var index IdSubData[string, any]
	err := ctx.Bind(&index)
	if err != nil {
		return err
	}
	cfg, err := c._sdk.GetConfigPlugin(index.Id, index.Sub)
	if err != nil {
		return err
	}
	return ctx.WriteAny(cfg)
}
