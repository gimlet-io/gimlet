package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Reads all files in the web/dist folder
// and encodes them as strings literals in web.go
func main() {
	workDir, _ := os.Getwd()
	fmt.Printf("Generating binaries from web/dist to %s\n", workDir+"/../commands/chart/web.go")
	out, err := os.Create("../commands/chart/web.go")
	if err != nil {
		panic(err)
	}

	defer out.Close()

	write(out, "package chart\n")

	write(out, "var web = map[string]string{\n")

	const webDist = "../web/dist/"
	var entries []string

	err = filepath.Walk(webDist, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		entry := "\"" + info.Name() + "\": `"
		content, _ := ioutil.ReadFile(webDist + info.Name())
		escapedBackticks := strings.ReplaceAll(string(content), "`", "` + \"`\" + `")
		entry += escapedBackticks
		entry += "`,"
		entries = append(entries, entry)
		return nil
	})
	if err != nil {
		panic(err)
	}

	write(out, strings.Join(entries, "\n"))

	write(out, "\n}\n")

	out.Sync()

}

func write(out *os.File, content string) {
	_, err := out.Write([]byte(content))
	if err != nil {
		panic(err)
	}
}
