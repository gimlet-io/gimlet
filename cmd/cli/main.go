package main

import (
	"fmt"
	"os"

	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimlet/pkg/commands"
	"github.com/gimlet-io/gimlet/pkg/commands/artifact"
	"github.com/gimlet-io/gimlet/pkg/commands/environment"
	"github.com/gimlet-io/gimlet/pkg/commands/manifest"
	"github.com/gimlet-io/gimlet/pkg/commands/release"
	"github.com/gimlet-io/gimlet/pkg/commands/stack"
	"github.com/gimlet-io/gimlet/pkg/version"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                 "gimlet",
		Version:              version.String(),
		Usage:                "",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			&manifest.Command,
			&artifact.Command,
			&release.Command,
			&stack.Command,
			&environment.Command,
			&commands.SyncCmd,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s\n", emoji.CrossMark, err.Error())
		os.Exit(1)
	}
}
