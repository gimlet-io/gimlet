package main

import (
	"github.com/gimlet-io/gimlet-cli/commands/chart"
	"github.com/gimlet-io/gimlet-cli/version"
	"github.com/urfave/cli/v2"
	"os"
)

//go:generate go run ../scripts/includeWeb.go

func main() {
	app := &cli.App{
		Name:                 "gimlet",
		Version:              version.String(),
		Usage:                "for an open-source GitOps workflow",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			&chart.Command,
		},
	}
	app.Run(os.Args)
}
