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
	pluginCmd.AddCommand(
		pluginGetUnitTempCfg,
		pluginListUnit,
		pluginAddCmd,
		pluginDelCmd,
		pluginUpdateCmd,
		pluginConfigCmd,
	)
}

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "anchorage core plugin command",
	Args:  cobra.NoArgs,
}

var pluginGetUnitTempCfg = &cobra.Command{
	Use:   "get_ut type name",
	Short: "get one anchorage core plugin unit temp config. (type: dial, listen, proxy)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		bs, err := command.CallBytes(ctx, command.CmdGetPluginUnit, [2]string{args[0], args[1]})
		if err != nil {
			return err
		}
		fmt.Println(string(bs))
		return nil
	},
}

var pluginListUnit = &cobra.Command{
	Use:   "list_ut type",
	Short: "get anchorage core plugin unit list. (type: dial, listen, proxy)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		var sl []string
		err := command.CallAny(ctx, command.CmdListPluginUnit, args[0], &sl)
		if err != nil {
			return err
		}
		fmt.Println(sl)
		return nil
	},
}

var pluginAddCmd = &cobra.Command{
	Use:   "add id",
	Short: "add one anchorage core plugin. (requires vim, vi, nano, or emacs tool)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		tempFile, err := internal.EditTempFile(ctx, internal.GetTmplPlugin())
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
		var n sdk.PluginConfig
		err = hyaml.Unmarshal(tempFile, &n)
		if err != nil {
			return err
		}
		err = command.CallAny(ctx, command.CmdAddPlugin, command.IdSubData[string, *sdk.PluginConfig]{
			Id:   args[0],
			Data: &n,
		}, nil)
		if err != nil {
			return err
		}
		return nil
	},
}

var pluginDelCmd = &cobra.Command{
	Use:   "del id sub",
	Short: "del one anchorage core plugin. (if it is in a working state, the related work will be stopped)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdDelPlugin, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, nil)
	},
}

var pluginUpdateCmd = &cobra.Command{
	Use:   "update id sub",
	Short: "update one anchorage core plugin config. (requires vim, vi, nano, or emacs tool)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		var cfg sdk.PluginConfig
		err := command.CallAny(ctx, command.CmdConfigPlugin, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, &cfg)
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
		var n sdk.PluginConfig
		err = hyaml.Unmarshal(tempFile, &n)
		if err != nil {
			return err
		}
		err = command.CallAny(ctx, command.CmdUpdatePlugin, command.IdSubData[string, *sdk.PluginConfig]{
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

var pluginConfigCmd = &cobra.Command{
	Use:   "config id sub",
	Short: "print one anchorage core listen config.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		var cfg sdk.PluginConfig
		err := command.CallAny(ctx, command.CmdConfigPlugin, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, &cfg)
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
