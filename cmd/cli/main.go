package main

import (
	"fmt"
	"os"

	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimlet-cli/pkg/commands/artifact"
	"github.com/gimlet-io/gimlet-cli/pkg/commands/chart"
	"github.com/gimlet-io/gimlet-cli/pkg/commands/gitops"
	"github.com/gimlet-io/gimlet-cli/pkg/commands/manifest"
	"github.com/gimlet-io/gimlet-cli/pkg/commands/release"
	"github.com/gimlet-io/gimlet-cli/pkg/commands/stack"
	"github.com/gimlet-io/gimlet-cli/pkg/version"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                 "gimlet",
		Version:              version.String(),
		Usage:                "a modular Gitops workflow for Kubernetes deployments",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			&chart.Command,
			&gitops.Command,
			&manifest.Command,
			&artifact.Command,
			&release.Command,
			&stack.Command,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s\n", emoji.CrossMark, err.Error())
		os.Exit(1)
	}
}
