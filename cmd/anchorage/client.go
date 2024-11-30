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
	clientCmd.AddCommand(
		clientAddCmd,
		clientDelCmd,
		clientStartCmd,
		clientStopCmd,
		clientReloadCmd,
		clientUpdateCmd,
		clientConfigCmd,
	)
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "anchorage core client command",
	Args:  cobra.NoArgs,
}

var clientAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add one anchorage core client. (requires vim, vi, nano, or emacs tool)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		tempFile, err := internal.EditTempFile(ctx, internal.GetTmplClient())
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
		var n sdk.ClientConfigUnit
		err = hyaml.Unmarshal(tempFile, &n)
		if err != nil {
			return err
		}
		var info command.IdData[any]
		err = command.CallAny(ctx, command.CmdAddClientUnit, n, &info)
		if err != nil {
			return err
		}
		fmt.Printf("Id: %s\n", info.Id)
		return nil
	},
}

var clientDelCmd = &cobra.Command{
	Use:   "del id",
	Short: "del one anchorage core client. (if it is in a working state, the related work will be stopped)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdDelClient, command.IdData[any]{Id: args[0]}, nil)
	},
}

var clientStartCmd = &cobra.Command{
	Use:   "start id",
	Short: "start one anchorage core client.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdStartClient, command.IdData[any]{Id: args[0]}, nil)
	},
}

var clientStopCmd = &cobra.Command{
	Use:   "stop id",
	Short: "stop one anchorage core client.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdStopClient, command.IdData[any]{Id: args[0]}, nil)
	},
}

var clientReloadCmd = &cobra.Command{
	Use:   "reload id",
	Short: "reload one anchorage core client.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdReloadClient, command.IdData[any]{Id: args[0]}, nil)
	},
}

var clientUpdateCmd = &cobra.Command{
	Use:   "update id",
	Short: "update one anchorage core client config and reload (if in working). (requires vim, vi, nano, or emacs tool)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		var cfg sdk.ClientConfigUnit
		err := command.CallAny(ctx, command.CmdConfigClientUnit, command.IdData[any]{Id: args[0]}, &cfg)
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
		var n sdk.ClientConfigUnit
		err = hyaml.Unmarshal(tempFile, &n)
		if err != nil {
			return err
		}
		err = command.CallAny(ctx, command.CmdUpdateClientUnit, command.IdData[*sdk.ClientConfigUnit]{
			Id:   args[0],
			Data: &n,
		}, nil)
		if err != nil {
			return err
		}
		return nil
	},
}

var clientConfigCmd = &cobra.Command{
	Use:   "config id",
	Short: "print one anchorage core client config.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		var cfg sdk.ClientConfigUnit
		err := command.CallAny(ctx, command.CmdConfigClientUnit, command.IdData[any]{Id: args[0]}, &cfg)
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
