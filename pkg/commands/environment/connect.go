package environment

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"sort"

	"github.com/gimlet-io/gimlet/pkg/client"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
)

var environmentConnectCmd = cli.Command{
	Name:      "connect",
	Usage:     "Applies the environment gitops manifests on the cluster",
	UsageText: `gimlet environment connect --env staging`,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "env",
			Usage:    "environment to connect with the cluster",
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
	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt)
	go func() {
		<-ctrlC
		os.Exit(0)
	}()

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

	infraRepoManifests := files["infra"]
	appsRepoManifests := files["apps"]
	applyManifests(infraRepoManifests, tmpDir)
	applyManifests(appsRepoManifests, tmpDir)

	return check(c)
}

func applyManifests(files map[string]string, filesPath string) {
	sortedKeys := sortByFluxFirst(files)

	for _, fileName := range sortedKeys {
		content := files[fileName]
		filePath := filepath.Join(filesPath, fileName)
		err := ioutil.WriteFile(filePath, []byte(fmt.Sprintf("%v", content)), 0666)
		if err != nil {
			panic(err)
		}

		infos, err := getObjects(filePath)
		if err != nil {
			panic(err)
		}

		for _, info := range infos {
			response, err := applyObject(info)
			if err != nil {
				panic(err)
			}
			fmt.Println(response)
		}

		if fileName == "flux.yaml" {
			err := waitFor("crd/gitrepositories.source.toolkit.fluxcd.io")
			if err != nil {
				panic(err)
			}
			err = waitFor("crd/kustomizations.kustomize.toolkit.fluxcd.io")
			if err != nil {
				panic(err)
			}
		}
	}
}

func sortByFluxFirst(files map[string]string) []string {
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		// Check if one of the keys is "flux.yaml"
		if keys[i] == "flux.yaml" {
			return true // "flux.yaml" should always come first
		} else if keys[j] == "flux.yaml" {
			return false
		}
		// For other keys, use the default sorting order
		return keys[i] < keys[j]
	})

	return keys
}
