package main

import (
	"github.com/peakedshout/anchorage-core/pkg/command"
	"github.com/spf13/cobra"
	"io"
	"os"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "output anchorage core runtime log information.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		readCloser, err := command.Call(ctx, command.CmdLog, nil)
		if err != nil {
			return err
		}
		defer readCloser.Close()
		_, err = io.Copy(os.Stdout, readCloser)
		return err
	},
}
