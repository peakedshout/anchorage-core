package command

import (
	"bytes"
	"context"
	"errors"
	"github.com/peakedshout/anchorage-core/pkg/sdk"
	"github.com/peakedshout/go-pandorasbox/tool/dcopy"
	"github.com/peakedshout/go-pandorasbox/tool/hyaml"
	"github.com/peakedshout/go-pandorasbox/xnet/xdummy"
	"github.com/peakedshout/go-pandorasbox/xnet/xtool/xcmd"
	"github.com/peakedshout/go-pandorasbox/xnet/xtool/xhttp"
	"io"
	"time"
)

const (
	Args            = "args"
	ArgNetwork      = "network"
	ArgAddress      = "address"
	ArgUsername     = "username"
	ArgPassword     = "password"
	ArgCertFile     = "certfile"
	ArgKeyFile      = "keyfile"
	ArgTlsType      = "tlstype"
	ArgDummyNetWork = "dummynetwork"

	_prefix  = "anchorage"
	_network = "tcp"
	_address = "127.0.0.1:2024"
)

var _xCmdConfig = &xcmd.Config{
	Type:    xhttp.TypeMuxWait,
	Prefix:  _prefix,
	Network: _network,
	Address: _address,
}

// Serve It should only be called once. If called repeatedly, the global variable will become invalid.
func Serve(ctx context.Context, fp string) (err error) {
	if fp == "" {
		return errors.New("fp is nil")
	}
	cmd, err := toInit(ctx, fp)
	if err != nil {
		return err
	}
	defer cmd.Close()
	err = cmd.Serve()
	if err != nil {
		return err
	}
	return nil
}

func Client(ctx context.Context) (*xhttp.Client, error) {
	cmd, err := toInit(ctx, "")
	if err != nil {
		return nil, err
	}
	return cmd.Client(), nil
}

func Call(ctx context.Context, name string, data any) (io.ReadCloser, error) {
	cmd, err := toInit(ctx, "")
	if err != nil {
		return nil, err
	}
	defer cmd.Close()
	return cmd.Call(ctx, name, data)
}

func CallBytes(ctx context.Context, name string, data any) ([]byte, error) {
	cmd, err := toInit(ctx, "")
	if err != nil {
		return nil, err
	}
	defer cmd.Close()
	return cmd.CallBytes(ctx, name, data)
}

func CallAny(ctx context.Context, name string, data any, out any) error {
	cmd, err := toInit(ctx, "")
	if err != nil {
		return err
	}
	defer cmd.Close()
	return cmd.CallAny(ctx, name, data, out)
}

func Init(ctx context.Context, fp string) (*Cmd, error) {
	return toInit(ctx, fp)
}

func toInit(ctx context.Context, fp string) (*Cmd, error) {
	cfg := dcopy.CopyT(_xCmdConfig)
	if m, ok := ctx.Value(Args).(map[string]string); ok {
		if str := m[ArgNetwork]; str != "" {
			cfg.Network = str
		}
		if str := m[ArgAddress]; str != "" {
			cfg.Address = str
		}
		if str := m[ArgUsername]; str != "" {
			cfg.Auth = &xcmd.Auth{
				UserName: str,
				Password: m[ArgPassword],
			}
		}
		cfg.TlsType = xcmd.TlsType(m[ArgTlsType])
		cfg.CertFile = m[ArgCertFile]
		cfg.KeyFile = m[ArgKeyFile]
		if str := m[ArgDummyNetWork]; str == "xdummy" {
			ln, dr := xdummy.NewDummyListenDialer(ctx)
			cfg.Dr = dr
			cfg.Ln = ln
		}
	}
	cfg.OnlyClient = fp == ""
	newXCmd, err := xcmd.NewXCmd(ctx, cfg)
	if err != nil {
		return nil, err
	}
	cmd := &Cmd{
		XCmd: newXCmd,
	}
	if !cfg.OnlyClient {
		cmd.fp = fp
		cmd._sdk, err = sdk.NewSdkFromFile(ctx, fp)
		if err != nil {
			_ = cmd.XCmd.Close()
			return nil, err
		}
		cmd.setHandler()
	}
	return cmd, nil
}

type Cmd struct {
	*xcmd.XCmd
	_sdk    *sdk.Sdk
	fp      string
	closing bool
}

func (c *Cmd) Close() error {
	if c._sdk != nil {
		c._sdk.Stop()
	}
	return c.XCmd.Close()
}

func (c *Cmd) stateHandler(ctx *xhttp.Context) error {
	if c.closing {
		return errors.New("closing")
	}
	ver := ctx.RHeader().Get(CmdHClientVersion)
	if ver != "" && ver != CmdCoreVersion {
		return errors.New("the requested version is inconsistent with the service version")
	}
	return nil
}

func (c *Cmd) stateHandlerNoCheck(ctx *xhttp.Context) error {
	if c.closing {
		return errors.New("closing")
	}
	return nil
}

func (c *Cmd) setOtherHandler() {
	c.XCmd.Set(CmdPing, c.stateHandlerNoCheck, func(c *xhttp.Context) error {
		return nil
	})
	c.XCmd.Set(CmdInfo, c.stateHandlerNoCheck, func(c *xhttp.Context) error {
		return c.WriteAny(&Info{
			Core:    CmdCore,
			Version: CmdCoreVersion,
			Flag:    FlagList,
		})
	})
}

func (c *Cmd) setHandler() {
	c.setOtherHandler()

	c.XCmd.Set(CmdStop, c.stateHandler, func(ctx *xhttp.Context) error {
		c.closing = true
		go func() {
			time.Sleep(1 * time.Second)
			_ = c.Close()
		}()
		return nil
	})
	c.XCmd.Set(CmdReload, c.stateHandler, func(ctx *xhttp.Context) error {
		c._sdk.Stop()
		time.Sleep(1 * time.Second)
		_sdk, err := sdk.NewSdkFromFile(ctx, c.fp)
		if err != nil {
			c.closing = true
			go func() {
				time.Sleep(1 * time.Second)
				_ = c.Close()
			}()
			return err
		}
		c._sdk = _sdk
		return nil
	})
	c.XCmd.Set(CmdConfig, c.stateHandler, func(ctx *xhttp.Context) error {
		cfg := c._sdk.GetConfig()
		return ctx.WriteAny(cfg)
	})
	c.XCmd.Set(CmdUpdate, c.stateHandler, func(ctx *xhttp.Context) error {
		var cfg sdk.Config
		err := ctx.Bind(&cfg)
		if err != nil {
			return err
		}
		srcCfg := c._sdk.GetConfig()
		out1, err := hyaml.Marshal(srcCfg)
		if err != nil {
			return err
		}
		out2, err := hyaml.Marshal(cfg)
		if err != nil {
			return err
		}
		if bytes.Equal(out1, out2) {
			return errors.New("no difference")
		}
		c._sdk.Stop()
		time.Sleep(1 * time.Second)
		err = hyaml.SavePath(c.fp, cfg)
		if err != nil {
			return err
		}
		_sdk, err := sdk.NewSdkFromFile(ctx, c.fp)
		if err != nil {
			c.closing = true
			go func() {
				time.Sleep(1 * time.Second)
				_ = c.Close()
			}()
			return err
		}
		c._sdk = _sdk
		return nil
	})

	c.XCmd.Set(CmdViewServer, c.stateHandler, c.serverView)
	c.XCmd.Set(CmdViewServerById, c.stateHandler, c.serverViewById)
	c.XCmd.Set(CmdViewServerSession, c.stateHandler, c.serverSessionView)
	c.XCmd.Set(CmdViewServerRoute, c.stateHandler, c.serverRouteView)
	c.XCmd.Set(CmdViewServerLink, c.stateHandler, c.serverLinkView)
	c.XCmd.Set(CmdViewServerSync, c.stateHandler, c.serverSyncView)
	c.XCmd.Set(CmdViewServerProxy, c.stateHandler, c.serverProxyView)
	c.XCmd.Set(CmdViewClient, c.stateHandler, c.clientView)
	c.XCmd.Set(CmdViewClientUnit, c.stateHandler, c.clientView2)
	c.XCmd.Set(CmdViewClientById, c.stateHandler, c.clientViewById)
	c.XCmd.Set(CmdViewClientUnitById, c.stateHandler, c.clientViewById2)
	c.XCmd.Set(CmdViewClientListenById, c.stateHandler, c.clientListenViewById)
	c.XCmd.Set(CmdViewClientDialById, c.stateHandler, c.clientDialViewById)
	c.XCmd.Set(CmdViewClientProxyById, c.stateHandler, c.clientProxyViewById)
	c.XCmd.Set(CmdViewClientSession, c.stateHandler, c.clientSessionView)
	c.XCmd.Set(CmdViewClientProxyT, c.stateHandler, c.clientProxyView)
	c.XCmd.Set(CmdViewClientProxyTUnit, c.stateHandler, c.clientProxyUnitView)

	c.XCmd.Set(CmdAddServer, c.stateHandler, c.addServer)
	c.XCmd.Set(CmdDelServer, c.stateHandler, c.delServer)
	c.XCmd.Set(CmdStartServer, c.stateHandler, c.startServer)
	c.XCmd.Set(CmdStopServer, c.stateHandler, c.stopServer)
	c.XCmd.Set(CmdReloadServer, c.stateHandler, c.reloadServer)
	c.XCmd.Set(CmdUpdateServer, c.stateHandler, c.updateServer)
	c.XCmd.Set(CmdConfigServer, c.stateHandler, c.configServer)

	c.XCmd.Set(CmdAddClient, c.stateHandler, c.addClient)
	c.XCmd.Set(CmdAddClientUnit, c.stateHandler, c.addClient2)
	c.XCmd.Set(CmdDelClient, c.stateHandler, c.delClient)
	c.XCmd.Set(CmdStartClient, c.stateHandler, c.startClient)
	c.XCmd.Set(CmdStartClientUnit, c.stateHandler, c.startClient2)
	c.XCmd.Set(CmdStopClient, c.stateHandler, c.stopClient)
	c.XCmd.Set(CmdReloadClient, c.stateHandler, c.reloadClient)
	c.XCmd.Set(CmdReloadClientUnit, c.stateHandler, c.reloadClient2)
	c.XCmd.Set(CmdUpdateClient, c.stateHandler, c.updateClient)
	c.XCmd.Set(CmdUpdateClientUnit, c.stateHandler, c.updateClient2)
	c.XCmd.Set(CmdConfigClient, c.stateHandler, c.configClient)
	c.XCmd.Set(CmdConfigClientUnit, c.stateHandler, c.configClient2)

	c.XCmd.Set(CmdAddProxy, c.stateHandler, c.addProxy)
	c.XCmd.Set(CmdDelProxy, c.stateHandler, c.delProxy)
	c.XCmd.Set(CmdStartProxy, c.stateHandler, c.startProxy)
	c.XCmd.Set(CmdStopProxy, c.stateHandler, c.stopProxy)
	c.XCmd.Set(CmdReloadProxy, c.stateHandler, c.reloadProxy)
	c.XCmd.Set(CmdUpdateProxy, c.stateHandler, c.updateProxy)
	c.XCmd.Set(CmdConfigProxy, c.stateHandler, c.configProxy)

	c.XCmd.Set(CmdAddListen, c.stateHandler, c.addListen)
	c.XCmd.Set(CmdDelListen, c.stateHandler, c.delListen)
	c.XCmd.Set(CmdStartListen, c.stateHandler, c.startListen)
	c.XCmd.Set(CmdStopListen, c.stateHandler, c.stopListen)
	c.XCmd.Set(CmdReloadListen, c.stateHandler, c.reloadListen)
	c.XCmd.Set(CmdUpdateListen, c.stateHandler, c.updateListen)
	c.XCmd.Set(CmdConfigListen, c.stateHandler, c.configListen)

	c.XCmd.Set(CmdAddDial, c.stateHandler, c.addDial)
	c.XCmd.Set(CmdDelDial, c.stateHandler, c.delDial)
	c.XCmd.Set(CmdStartDial, c.stateHandler, c.startDial)
	c.XCmd.Set(CmdStopDial, c.stateHandler, c.stopDial)
	c.XCmd.Set(CmdReloadDial, c.stateHandler, c.reloadDial)
	c.XCmd.Set(CmdUpdateDial, c.stateHandler, c.updateDial)
	c.XCmd.Set(CmdConfigDial, c.stateHandler, c.configDial)

	c.XCmd.Set(CmdGetPluginUnit, c.stateHandler, c.getPluginTempCfg)
	c.XCmd.Set(CmdListPluginUnit, c.stateHandler, c.listPluginUnit)
	c.XCmd.Set(CmdAddPlugin, c.stateHandler, c.addPlugin)
	c.XCmd.Set(CmdDelPlugin, c.stateHandler, c.delPlugin)
	c.XCmd.Set(CmdUpdatePlugin, c.stateHandler, c.updatePlugin)
	c.XCmd.Set(CmdConfigPlugin, c.stateHandler, c.configPlugin)

	c.XCmd.Set(CmdLog, c.stateHandler, c.getLogger)
}

type IdData[T any] struct {
	Id   string `json:"id"`
	Data T      `json:"data,omitempty"`
}

type IdSubData[T1, T2 any] struct {
	Id   string `json:"id"`
	Sub  T1     `json:"sub"`
	Data T2     `json:"data,omitempty"`
}
