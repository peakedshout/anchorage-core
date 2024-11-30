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
	serverCmd.AddCommand(
		serverAddCmd,
		serverDelCmd,
		serverStartCmd,
		serverStopCmd,
		serverReloadCmd,
		serverUpdateCmd,
		serverConfigCmd,
	)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "anchorage core server command",
	Args:  cobra.NoArgs,
}

var serverAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add one anchorage core server. (requires vim, vi, nano, or emacs tool)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		tempFile, err := internal.EditTempFile(ctx, internal.GetTmplServer())
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
		var n sdk.ServerConfig
		err = hyaml.Unmarshal(tempFile, &n)
		if err != nil {
			return err
		}
		var info command.IdData[any]
		err = command.CallAny(ctx, command.CmdAddServer, n, &info)
		if err != nil {
			return err
		}
		fmt.Printf("Id: %s\n", info.Id)
		return nil
	},
}

var serverDelCmd = &cobra.Command{
	Use:   "del id",
	Short: "del one anchorage core server. (if it is in a working state, the related work will be stopped)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdDelServer, command.IdData[any]{Id: args[0]}, nil)
	},
}

var serverStartCmd = &cobra.Command{
	Use:   "start id",
	Short: "start one anchorage core server.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdStartServer, command.IdData[any]{Id: args[0]}, nil)
	},
}

var serverStopCmd = &cobra.Command{
	Use:   "stop id",
	Short: "stop one anchorage core server.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdStopServer, command.IdData[any]{Id: args[0]}, nil)
	},
}

var serverReloadCmd = &cobra.Command{
	Use:   "reload id",
	Short: "reload one anchorage core server.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdReloadServer, command.IdData[any]{Id: args[0]}, nil)
	},
}

var serverUpdateCmd = &cobra.Command{
	Use:   "update id",
	Short: "update one anchorage core server config and reload (if in working). (requires vim, vi, nano, or emacs tool)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		var cfg sdk.ServerConfig
		err := command.CallAny(ctx, command.CmdConfigServer, command.IdData[any]{Id: args[0]}, &cfg)
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
		var n sdk.ServerConfig
		err = hyaml.Unmarshal(tempFile, &n)
		if err != nil {
			return err
		}
		err = command.CallAny(ctx, command.CmdUpdateServer, command.IdData[*sdk.ServerConfig]{
			Id:   args[0],
			Data: &n,
		}, nil)
		if err != nil {
			return err
		}
		return nil
	},
}

var serverConfigCmd = &cobra.Command{
	Use:   "config id",
	Short: "print one anchorage core server config.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		var cfg sdk.ServerConfig
		err := command.CallAny(ctx, command.CmdConfigServer, command.IdData[any]{Id: args[0]}, &cfg)
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
