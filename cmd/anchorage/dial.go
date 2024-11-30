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
	dialCmd.AddCommand(
		dialAddCmd,
		dialDelCmd,
		dialStartCmd,
		dialStopCmd,
		dialReloadCmd,
		dialUpdateCmd,
		dialConfigCmd,
	)
}

var dialCmd = &cobra.Command{
	Use:   "dial",
	Short: "anchorage core dial command",
	Args:  cobra.NoArgs,
}

var dialAddCmd = &cobra.Command{
	Use:   "add id",
	Short: "add one anchorage core dial. (requires vim, vi, nano, or emacs tool)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		tempFile, err := internal.EditTempFile(ctx, internal.GetTmplDial())
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
		var n sdk.DialConfig
		err = hyaml.Unmarshal(tempFile, &n)
		if err != nil {
			return err
		}
		var info command.IdSubData[string, any]
		err = command.CallAny(ctx, command.CmdAddDial, command.IdSubData[string, *sdk.DialConfig]{
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

var dialDelCmd = &cobra.Command{
	Use:   "del id sub",
	Short: "del one anchorage core dial. (if it is in a working state, the related work will be stopped)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdDelDial, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, nil)
	},
}

var dialStartCmd = &cobra.Command{
	Use:   "start id sub",
	Short: "start one anchorage core dial.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdStartDial, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, nil)
	},
}

var dialStopCmd = &cobra.Command{
	Use:   "stop id sub",
	Short: "stop one anchorage core dial.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdStopDial, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, nil)
	},
}

var dialReloadCmd = &cobra.Command{
	Use:   "reload id sub",
	Short: "reload one anchorage core dial.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdReloadDial, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, nil)
	},
}

var dialUpdateCmd = &cobra.Command{
	Use:   "update id sub",
	Short: "update one anchorage core dial config and reload (if in working). (requires vim, vi, nano, or emacs tool)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		var cfg sdk.DialConfig
		err := command.CallAny(ctx, command.CmdConfigDial, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, &cfg)
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
		var n sdk.DialConfig
		err = hyaml.Unmarshal(tempFile, &n)
		if err != nil {
			return err
		}
		err = command.CallAny(ctx, command.CmdUpdateDial, command.IdSubData[string, *sdk.DialConfig]{
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

var dialConfigCmd = &cobra.Command{
	Use:   "config id sub",
	Short: "print one anchorage core dial config.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		var cfg sdk.DialConfig
		err := command.CallAny(ctx, command.CmdConfigDial, command.IdSubData[string, any]{Id: args[0], Sub: args[1]}, &cfg)
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
