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
	listenCmd.AddCommand(
		listenAddCmd,
		listenDelCmd,
		listenStartCmd,
		listenStopCmd,
		listenReloadCmd,
		listenUpdateCmd,
		listenConfigCmd,
	)
}

var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "anchorage core listen command",
	Args:  cobra.NoArgs,
}

var listenAddCmd = &cobra.Command{
	Use:   "add id",
	Short: "add one anchorage core listen. (requires vim, vi, nano, or emacs tool)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		tempFile, err := internal.EditTempFile(ctx, internal.GetTmplListen())
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
		var n sdk.ListenConfig
		err = hyaml.Unmarshal(tempFile, &n)
		if err != nil {
			return err
		}
		var info command.IdSubData[string, any]
		err = command.CallAny(ctx, command.CmdAddListen, command.IdSubData[string, *sdk.ListenConfig]{
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

var listenDelCmd = &cobra.Command{
	Use:   "del id sub",
	Short: "del one anchorage core listen. (if it is in a working state, the related work will be stopped)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdDelListen, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, nil)
	},
}

var listenStartCmd = &cobra.Command{
	Use:   "start id sub",
	Short: "start one anchorage core listen.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdStartListen, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, nil)
	},
}

var listenStopCmd = &cobra.Command{
	Use:   "stop id sub",
	Short: "stop one anchorage core listen.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdStopListen, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, nil)
	},
}

var listenReloadCmd = &cobra.Command{
	Use:   "reload id sub",
	Short: "reload one anchorage core listen.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdReloadListen, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, nil)
	},
}

var listenUpdateCmd = &cobra.Command{
	Use:   "update id sub",
	Short: "update one anchorage core listen config and reload (if in working). (requires vim, vi, nano, or emacs tool)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		var cfg sdk.ListenConfig
		err := command.CallAny(ctx, command.CmdConfigListen, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, &cfg)
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
		var n sdk.ListenConfig
		err = hyaml.Unmarshal(tempFile, &n)
		if err != nil {
			return err
		}
		err = command.CallAny(ctx, command.CmdUpdateListen, command.IdSubData[string, *sdk.ListenConfig]{
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

var listenConfigCmd = &cobra.Command{
	Use:   "config id sub",
	Short: "print one anchorage core listen config.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		var cfg sdk.ListenConfig
		err := command.CallAny(ctx, command.CmdConfigListen, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, &cfg)
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
