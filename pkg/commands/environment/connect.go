package environment

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/gimlet-io/gimlet-cli/pkg/client"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
)

var environmentConnectCmd = cli.Command{
	Name:      "connect",
	Usage:     "Applies the environment gitops manifests to the kubernetes client",
	UsageText: `gimlet environment connect --env staging`,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "env",
			Usage:    "environment to connect to the kubernetes client",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "server",
			Usage:    "Gimlet server URL, GIMLET_SERVER environment variable alternatively",
			EnvVars:  []string{"GIMLET_SERVER"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "token",
			Usage:    "Gimlet server api token, GIMLET_TOKEN environment variable alternatively",
			EnvVars:  []string{"GIMLET_TOKEN"},
			Required: true,
		},
	},
	Action: connect,
}

func connect(c *cli.Context) error {
	envName := c.String("env")
	serverURL := c.String("server")
	token := c.String("token")

	config := new(oauth2.Config)
	auth := config.Client(
		context.Background(),
		&oauth2.Token{
			AccessToken: token,
		},
	)

	client := client.NewClient(serverURL, auth)

	files, err := client.GitopsManifestsGet(envName)
	if err != nil {
		return err
	}

	tmpDir, err := ioutil.TempDir("/tmp", "gimlet")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	applyManifests(files["infra"], tmpDir)
	applyManifests(files["apps"], tmpDir)

	return nil
}

func applyManifests(files map[string]string, filesPath string) {
	// TODO we want to apply the flux.yaml first
	//sortFiles, have to apply flux first
	// Extract the keys from the map
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}

	// Define the custom sorting logic
	sort.SliceStable(keys, func(i, j int) bool {
		// Check if one of the keys is "flux.yaml"
		if keys[i] == "flux.yaml" {
			return true // "flux.yaml" should always come first
		} else if keys[j] == "flux.yaml" {
			return false // "flux.yaml" should always come first
		}
		// For other keys, use the default sorting order
		return keys[i] < keys[j]
	})

	// Create a new sorted map
	sortedFiles := make(map[string]string)
	for _, k := range keys {
		sortedFiles[k] = files[k]
	}

	for fileName, content := range sortedFiles {
		filePath := filepath.Join(filesPath, fileName)
		err := ioutil.WriteFile(filePath, []byte(fmt.Sprintf("%v", content)), 0644)
		if err != nil {
			logrus.Warnf("cannot write files to %s", filePath)
		}
		infos, err := getObjects(filePath)
		if err != nil {
			logrus.Warnf("cannot get objects: %s", err)
			continue
		}
		for _, info := range infos {
			res, err := applyObject(info)
			if err != nil {
				logrus.Warnf("cannot apply object: %s", err)
				continue
			}
			fmt.Println(res)
		}
		if fileName == "flux.yaml" {
			// TODO
			// kubectl wait --for condition=established --timeout=60s crd/gitrepositories.source.toolkit.fluxcd.io
			// kubectl wait --for condition=established --timeout=60s crd/kustomizations.kustomize.toolkit.fluxcd.io
		}
	}
}
