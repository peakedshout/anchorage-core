package main

import (
	"fmt"
	"github.com/peakedshout/anchorage-core/pkg/command"
	"github.com/peakedshout/anchorage-core/pkg/sdk"
	"github.com/spf13/cobra"
	"strings"
)

func init() {
	viewCmd.AddCommand(viewServerCmd, viewClientCmd)
}

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "anchorage core runtime view information.",
	Args:  cobra.NoArgs,
}

var vsList = []string{"default", "session", "route", "link", "sync", "proxy"}

var viewServerCmd = &cobra.Command{
	Use:   "server [ default { id } | session { id } | route { id } | link { id } | sync { id } | proxy { id } ]",
	Short: "print anchorage core server runtime view information. ([default session route link sync proxy])",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var data any
		call := command.CmdViewServer
		if len(args) > 0 {
			switch args[0] {
			case vsList[0]:
				if len(args) > 1 {
					call = command.CmdViewServerById
					data = command.IdData[any]{Id: args[1]}
				}
			case vsList[1], vsList[2], vsList[3], vsList[4], vsList[5]:
				call += "_" + args[0]
				if len(args) == 1 {
					return fmt.Errorf("invaild args: num")
				}
				data = command.IdData[any]{Id: args[1]}
			default:
				return fmt.Errorf("invaild arg: '%s' must be %v", args[0], vsList)
			}
		}
		ctx := getCmdContext(cmd)
		bytes, err := command.CallBytes(ctx, call, data)
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var vcList = []string{"default", "session", "proxyT"}

var viewClientCmd = &cobra.Command{
	Use:   "client [ default { id sub } | session { id } | proxyT { id } ]",
	Short: "print anchorage core server runtime view information. ([default session proxyT])",
	Args:  cobra.MaximumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		var data any
		call := command.CmdViewClient
		if len(args) > 0 {
			switch args[0] {
			case vcList[0]:
				switch len(args) {
				case 2:
					call = command.CmdViewClientById
					data = command.IdData[any]{Id: args[1]}
				case 3:
					switch {
					case strings.HasPrefix(args[2], sdk.IdPrefixDial):
						call = command.CmdViewClientDialById
					case strings.HasPrefix(args[2], sdk.IdPrefixListen):
						call = command.CmdViewClientListenById
					case strings.HasPrefix(args[2], sdk.IdPrefixProxy):
						call = command.CmdViewClientProxyById
					default:
						return fmt.Errorf("invaild sid: %s", args[2])
					}
					data = command.IdSubData[string, any]{Id: args[1], Sub: args[2]}
				default:
				}
			case vcList[1], vcList[2]:
				call += "_" + args[0]
				if len(args) != 2 {
					return fmt.Errorf("invaild args: num")
				}
				data = command.IdData[any]{Id: args[1]}
			default:
				return fmt.Errorf("invaild arg: '%s' must be %v", args[0], vcList)
			}
		}
		ctx := getCmdContext(cmd)
		bytes, err := command.CallBytes(ctx, call, data)
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}
