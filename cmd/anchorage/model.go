package main

import (
	"context"
	"github.com/peakedshout/anchorage-core/cmd/anchorage/internal"
	"github.com/peakedshout/anchorage-core/pkg/command"
	"github.com/peakedshout/go-pandorasbox/xnet/xtool/xcmd"
	"github.com/peakedshout/go-pandorasbox/xnet/xtool/xhttp"
	"github.com/spf13/cobra"
)

type cmdTlsConfig struct {
	CertFile string `Barg:"cmd.cert" Harg:"cmd tls cert file" Garg:"ck"`
	KeyFile  string `Barg:"cmd.key" Harg:"cmd tls key file" Garg:"ck"`
}

type cmdConfig struct {
	Network  string `Barg:"cmd.nk" Harg:"cmd network"`
	Address  string `Barg:"cmd.addr" Harg:"cmd address"`
	Username string `Barg:"cmd.u" Harg:"cmd username" Garg:"up"`
	Password string `Barg:"cmd.p" Harg:"cmd password" Garg:"up"`
	Tls      bool   `Barg:"cmd.tls" Harg:"cmd tls enable"`
	Insecure bool   `Barg:"cmd.i" Harg:"cmd tls insecure"`
}

func getCmdContext(cmd *cobra.Command) context.Context {
	cfg := internal.GetKeyT[cmdConfig](cmd, "cmd")
	tt := xcmd.TlsTypeDefault
	if cfg.Tls {
		tt = xcmd.TlsTypeEnable
		if cfg.Insecure {
			tt = xcmd.TlsTypeInsecure
		}
	}
	tlsCfg := internal.GetKeyT[cmdTlsConfig](cmd, "cmd.tls")
	if tlsCfg == nil {
		tlsCfg = &cmdTlsConfig{}
	}
	ctx := xhttp.SetArgsHeader(cmd.Context(), map[string][]string{
		command.CmdHClientName:    {"anchorage-cli"},
		command.CmdHClientVersion: {command.CmdCoreVersion},
	})
	return context.WithValue(ctx, command.Args, map[string]string{
		command.ArgNetwork:  cfg.Network,
		command.ArgAddress:  cfg.Address,
		command.ArgUsername: cfg.Username,
		command.ArgPassword: cfg.Password,
		command.ArgTlsType:  string(tt),
		command.ArgCertFile: tlsCfg.CertFile,
		command.ArgKeyFile:  tlsCfg.KeyFile,
	})
}

func bindAllCmdContext(cmd *cobra.Command) {
	var cfg cmdConfig
	internal.BindKey(cmd, "cmd", &cfg)
	for _, sub := range cmd.Commands() {
		if sub == initCmd {
			continue
		}
		bindAllCmdContext(sub)
	}
}

func bindTlsConfigCmdContext(cmd *cobra.Command) {
	var cfg cmdTlsConfig
	internal.BindKey(cmd, "cmd.tls", &cfg)
}
