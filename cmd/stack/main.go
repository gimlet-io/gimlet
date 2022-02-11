package main

import (
	"fmt"
	"os"

	"github.com/gimlet-io/gimlet-cli/pkg/commands/stack"
	"github.com/urfave/cli/v2"

	"github.com/enescakir/emoji"
	"github.com/gimlet-io/gimlet-cli/pkg/version"
)

func main() {
	app := &cli.App{
		Name:                 "stack",
		Version:              version.String(),
		Usage:                "bootstrap curated Kubernetes stacks",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			&stack.BootstrapCmd,
			&stack.GenerateCmd,
			&stack.ConfigureCmd,
			&stack.LintCmd,
			&stack.UpdateCmd,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s\n", emoji.CrossMark, err.Error())
		os.Exit(1)
	}
}
