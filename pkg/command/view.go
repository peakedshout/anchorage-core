package command

import "github.com/peakedshout/go-pandorasbox/xnet/xtool/xhttp"

func (c *Cmd) serverView(ctx *xhttp.Context) error {
	view := c._sdk.GetServerView()
	return ctx.WriteAny(view)
}

func (c *Cmd) serverViewById(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	view, err := c._sdk.GetServerViewById(info.Id)
	if err != nil {
		return err
	}
	return ctx.WriteAny(view)
}

func (c *Cmd) serverSessionView(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	view, err := c._sdk.GetServerSessionView(info.Id)
	if err != nil {
		return err
	}
	return ctx.WriteAny(view)
}

func (c *Cmd) serverRouteView(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	view, err := c._sdk.GetServerRouteView(info.Id)
	if err != nil {
		return err
	}
	return ctx.WriteAny(view)
}

func (c *Cmd) serverLinkView(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	view, err := c._sdk.GetServerLinkView(info.Id)
	if err != nil {
		return err
	}
	return ctx.WriteAny(view)
}

func (c *Cmd) serverSyncView(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	view, err := c._sdk.GetServerSyncView(info.Id)
	if err != nil {
		return err
	}
	return ctx.WriteAny(view)
}

func (c *Cmd) serverProxyView(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	view, err := c._sdk.GetServerProxyView(info.Id)
	if err != nil {
		return err
	}
	return ctx.WriteAny(view)
}

func (c *Cmd) clientView(ctx *xhttp.Context) error {
	view := c._sdk.GetClientView()
	return ctx.WriteAny(view)
}

func (c *Cmd) clientView2(ctx *xhttp.Context) error {
	view := c._sdk.GetClientView2()
	return ctx.WriteAny(view)
}

func (c *Cmd) clientViewById(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	view, err := c._sdk.GetClientViewById(info.Id)
	if err != nil {
		return err
	}
	return ctx.WriteAny(view)
}

func (c *Cmd) clientViewById2(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	view, err := c._sdk.GetClientViewById2(info.Id)
	if err != nil {
		return err
	}
	return ctx.WriteAny(view)
}

func (c *Cmd) clientListenViewById(ctx *xhttp.Context) error {
	var info IdSubData[string, any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	view, err := c._sdk.GetListenViewById(info.Id, info.Sub)
	if err != nil {
		return err
	}
	return ctx.WriteAny(view)
}

func (c *Cmd) clientDialViewById(ctx *xhttp.Context) error {
	var info IdSubData[string, any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	view, err := c._sdk.GetDialViewById(info.Id, info.Sub)
	if err != nil {
		return err
	}
	return ctx.WriteAny(view)
}

func (c *Cmd) clientProxyViewById(ctx *xhttp.Context) error {
	var info IdSubData[string, any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	view, err := c._sdk.GetProxyViewById(info.Id, info.Sub)
	if err != nil {
		return err
	}
	return ctx.WriteAny(view)
}

func (c *Cmd) clientSessionView(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	view, err := c._sdk.GetClientSessionView(info.Id)
	if err != nil {
		return err
	}
	return ctx.WriteAny(view)
}

func (c *Cmd) clientProxyView(ctx *xhttp.Context) error {
	var info IdData[any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	view, err := c._sdk.GetClientProxyView(info.Id)
	if err != nil {
		return err
	}
	return ctx.WriteAny(view)
}

func (c *Cmd) clientProxyUnitView(ctx *xhttp.Context) error {
	var info IdSubData[string, any]
	err := ctx.Bind(&info)
	if err != nil {
		return err
	}
	view, err := c._sdk.GetProxyUnitView(info.Id, info.Sub)
	if err != nil {
		return err
	}
	return ctx.WriteAny(view)
}
