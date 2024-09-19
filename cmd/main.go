package main

import (
	"fmt"
	"os"

	"github.com/bang9ming9/bm-cli-tool/deploy"
	"github.com/bang9ming9/bm-cli-tool/eventlogger"
	"github.com/bang9ming9/bm-cli-tool/scan"
	"github.com/urfave/cli/v2"
)

const (
	AppVersion = "0.1.0"
)

var (
	app = cli.NewApp()
)

func init() {
	app.Name = "bm-cli-tool"
	app.Version = AppVersion
	app.Copyright = "bang9ming9"
	app.CommandNotFound = func(ctx *cli.Context, s string) {
		cli.ShowAppHelp(ctx)
		os.Exit(1)
	}
	app.OnUsageError = func(ctx *cli.Context, err error, isSubcommand bool) error {
		cli.ShowAppHelp(ctx)
		return err
	}

	app.Commands = append(app.Commands, []*cli.Command{
		deploy.Command,
		scan.Command,
		eventlogger.Command,
	}...)
}
func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
