package stack

import "github.com/urfave/cli/v2"

var Command = cli.Command{
	Name:  "stack",
	Usage: "Bootstrap curated Kubernetes stacks",
	Subcommands: []*cli.Command{
		&GenerateCmd,
		&ConfigureCmd,
		&LintCmd,
		&UpdateCmd,
	},
}
