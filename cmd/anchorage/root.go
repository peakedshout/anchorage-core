/*
Copyright © 2024 peakedshout <peakedshout@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package main

import (
	"errors"
	"fmt"
	"github.com/peakedshout/anchorage-core/cmd/anchorage/internal"
	"github.com/peakedshout/anchorage-core/pkg/command"
	"github.com/peakedshout/anchorage-core/pkg/sdk"
	"github.com/peakedshout/go-pandorasbox/pcrypto"
	"github.com/peakedshout/go-pandorasbox/tool/hyaml"
	"github.com/peakedshout/go-pandorasbox/tool/xpprof"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io/fs"
	"net/http"
	"os"
	"time"
)

func main() {
	pcrypto.RegisterAll()
	bindAllCmdContext(rootCmd)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	startCmd.Flags().StringP("debug", "d", "", "Enable debug pprof and metrics")
	_ = internal.BindViper(startCmd).BindPFlag("debug", startCmd.Flags().Lookup("debug"))
	_ = startCmd.Flags().BoolP("init", "i", false, "If config not exist, will init config")
	_ = internal.BindViper(startCmd).BindPFlag("init", startCmd.Flags().Lookup("init"))
	bindTlsConfigCmdContext(startCmd)
	rootCmd.AddCommand(startCmd, stopCmd, reloadCmd, updateCmd, configCmd, initCmd, infoCmd, pingCmd)
	rootCmd.AddCommand(viewCmd)
	rootCmd.AddCommand(serverCmd, clientCmd)
	rootCmd.AddCommand(proxyCmd, listenCmd, dialCmd)
	rootCmd.AddCommand(pluginCmd)
	rootCmd.AddCommand(logCmd)
}

var rootCmd = &cobra.Command{
	Use:     "anchorage",
	Short:   "anchorage command tool.",
	Args:    cobra.NoArgs,
	Version: GetProjectInfo(),
}

var startCmd = &cobra.Command{
	Use:   "start [config]",
	Short: "start anchorage core.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fp := ConfigPath()
		if len(args) > 0 {
			fp = args[0]
		}
		i := internal.GetViper(cmd).GetBool("init")
		if i {
			_, err := os.Stat(fp)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					err = hyaml.SavePath(fp, new(sdk.Config))
				}
				if err != nil {
					return err
				}
			}
		}
		addr := internal.GetViper(cmd).GetString("debug")
		if addr != "" {
			s := &http.Server{}
			var mux http.ServeMux
			xpprof.InitMux(&mux)
			xpprof.InitCollector("anchorage", &mux)
			s.Handler = &mux
			s.Addr = addr
			go s.ListenAndServe()
		}
		ctx := getCmdContext(cmd)
		err := command.Serve(ctx, fp)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		return err
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop anchorage core.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdStop, nil, nil)
	},
}

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "reload anchorage core.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		return command.CallAny(ctx, command.CmdReload, nil, nil)
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update anchorage core config and reload. (requires vim, vi, nano, or emacs tool)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		var cfg sdk.Config
		err := command.CallAny(ctx, command.CmdConfig, nil, &cfg)
		if err != nil {
			return err
		}
		b, err := hyaml.MarshalWithComment(cfg)
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
		var n sdk.Config
		err = yaml.Unmarshal(tempFile, &n)
		if err != nil {
			return err
		}
		err = command.CallAny(ctx, command.CmdUpdate, n, nil)
		if err != nil {
			return err
		}
		return nil
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "print anchorage core config.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		var cfg sdk.Config
		err := command.CallAny(ctx, command.CmdConfig, nil, &cfg)
		if err != nil {
			return err
		}
		b, err := hyaml.MarshalWithCommentStr(cfg)
		if err != nil {
			return err
		}
		fmt.Println(b)
		return nil
	},
}

var initCmd = &cobra.Command{
	Use:   "init [config]",
	Short: "init anchorage core config file.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fp := ConfigPath()
		if len(args) > 0 {
			fp = args[0]
		}
		return hyaml.SavePath(fp, new(sdk.Config))
	},
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "print anchorage core info.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		var info command.Info
		err := command.CallAny(ctx, command.CmdInfo, nil, &info)
		if err != nil {
			return err
		}
		fmt.Printf("Core: %s\n", info.Core)
		fmt.Printf("Version: %s\n", info.Version)
		fmt.Print("Flag: ")
		for _, s := range info.Flag {
			fmt.Print(s, " ")
		}
		fmt.Println()
		return nil
	},
}

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "try ping anchorage core and print delay.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getCmdContext(cmd)
		now := time.Now()
		err := command.CallAny(ctx, command.CmdPing, nil, nil)
		if err != nil {
			return err
		}
		fmt.Printf("ping succeeded: %s\n", time.Since(now))
		return nil
	},
}
