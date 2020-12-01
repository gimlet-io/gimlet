package commands

import (
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const File_RW_RW_R = 0664
const Dir_RWX_RX_R = 0754

func Run(command *cli.Command, args []string) error {
	app := &cli.App{
		Name:                 "gimlet",
		Version:              "test",
		Usage:                "for an open-source GitOps workflow",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			command,
		},
	}
	return app.Run(args)
}

func InputFiles(file string) (map[string]string, error) {
	files := map[string]string{}

	if strings.TrimSpace(file) == "-" {
		contents, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return files, err
		}
		files["manifest.yaml"] = string(contents)
		return files, err
	} else {
		file, err := filepath.Abs(file)
		if err != nil {
			return files, err
		}
		fd, err := os.Stat(file)
		if err != nil {
			return files, err
		}
		if fd.IsDir() {
			dir, err := ioutil.ReadDir(file)
			if err != nil {
				return files, err
			}
			for _, f := range dir {
				contents, err := ioutil.ReadFile(filepath.Join(file, f.Name()))
				if err != nil {
					return files, err
				}
				files[filepath.Join(file, f.Name())] = string(contents)
			}
			return files, nil
		} else {
			contents, err := ioutil.ReadFile(file)
			if err != nil {
				return files, err
			}
			files[file] = string(contents)
		}
	}
	return files, nil
}
