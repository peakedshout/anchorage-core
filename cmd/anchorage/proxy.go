package main

import (
	"errors"
	"fmt"
	"github.com/peakedshout/anchorage-core/cmd/anchorage/internal"
	"github.com/peakedshout/anchorage-core/pkg/command"
	"github.com/peakedshout/anchorage-core/pkg/sdk"
	"github.com/peakedshout/go-pandorasbox/tool/hyaml"
	"github.com/spf13/cobra"
)

func init() {
	proxyCmd.AddCommand(
		proxyAddCmd,
		proxyDelCmd,
		proxyStartCmd,
		proxyStopCmd,
		proxyReloadCmd,
		proxyUpdateCmd,
		proxyConfigCmd,
	)
}

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "anchorage core proxy command",
	Args:  cobra.NoArgs,
}

var proxyAddCmd = &cobra.Command{
	Use:   "add id",
	Short: "add one anchorage core proxy. (requires vim, vi, nano, or emacs tool)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		tempFile, err := internal.EditTempFile(ctx, internal.GetTmplProxy())
		if err != nil {
			if errors.Is(err, internal.ErrNotFindEditor) {
				err = ErrNotSupportCfg
			}
			if errors.Is(err, internal.ErrNoChanges) {
				fmt.Println("cancel edit")
				return nil
			}
			return err
		}
		var n sdk.ProxyConfig
		err = hyaml.Unmarshal(tempFile, &n)
		if err != nil {
			return err
		}
		var info command.IdSubData[string, any]
		err = command.CallAny(ctx, command.CmdAddProxy, command.IdSubData[string, *sdk.ProxyConfig]{
			Id:   args[0],
			Data: &n,
		}, &info)
		if err != nil {
			return err
		}
		fmt.Printf("Id: %s\n", info.Sub)
		return nil
	},
}

var proxyDelCmd = &cobra.Command{
	Use:   "del id sub",
	Short: "del one anchorage core proxy. (if it is in a working state, the related work will be stopped)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdDelProxy, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, nil)
	},
}

var proxyStartCmd = &cobra.Command{
	Use:   "start id sub",
	Short: "start one anchorage core proxy.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdStartProxy, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, nil)
	},
}

var proxyStopCmd = &cobra.Command{
	Use:   "stop id sub",
	Short: "stop one anchorage core proxy.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdStopProxy, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, nil)
	},
}

var proxyReloadCmd = &cobra.Command{
	Use:   "reload id sub",
	Short: "reload one anchorage core proxy.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdReloadProxy, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, nil)
	},
}

var proxyUpdateCmd = &cobra.Command{
	Use:   "update id sub",
	Short: "update one anchorage core proxy config and reload (if in working). (requires vim, vi, nano, or emacs tool)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		var cfg sdk.ProxyConfig
		err := command.CallAny(ctx, command.CmdConfigProxy, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, &cfg)
		if err != nil {
			return err
		}
		b, err := hyaml.MarshalWithCommentT(cfg)
		if err != nil {
			return err
		}
		tempFile, err := internal.EditTempFile(ctx, b)
		if err != nil {
			if errors.Is(err, internal.ErrNotFindEditor) {
				err = ErrNotSupportCfg
			}
			if errors.Is(err, internal.ErrNoChanges) {
				fmt.Println("cancel edit")
				return nil
			}
			return err
		}
		var n sdk.ProxyConfig
		err = hyaml.Unmarshal(tempFile, &n)
		if err != nil {
			return err
		}
		err = command.CallAny(ctx, command.CmdUpdateProxy, command.IdSubData[string, *sdk.ProxyConfig]{
			Id:   args[0],
			Sub:  args[1],
			Data: &n,
		}, nil)
		if err != nil {
			return err
		}
		return nil
	},
}

var proxyConfigCmd = &cobra.Command{
	Use:   "config id sub",
	Short: "print one anchorage core proxy config.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		var cfg sdk.ProxyConfig
		err := command.CallAny(ctx, command.CmdConfigProxy, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, &cfg)
		if err != nil {
			return err
		}
		b, err := hyaml.MarshalWithCommentStrT(cfg)
		if err != nil {
			return err
		}
		fmt.Println(b)
		return nil
	},
}
