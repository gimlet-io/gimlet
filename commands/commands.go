package commands

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

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
