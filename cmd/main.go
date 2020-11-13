package main

import (
	"github.com/urfave/cli/v2"
	"os"
	"github.com/gimlet-io/gimlet-cli/version"
)

func main() {
	app := &cli.App{
		Name: "gimlet",
		Version: version.String(),
		EnableBashCompletion: true,
	}
	app.Run(os.Args)
}
