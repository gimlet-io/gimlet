package commands

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"github.com/inancgumus/screen"
	"github.com/urfave/cli/v2"
)

var SyncCmd = cli.Command{
	Name:  "sync",
	Usage: "Sync files to Kubernetes pods",
	UsageText: `gimlet sync <folder-to-sync> <pod-name>[@<namespace>]:<path-in-pod>
	
	Example:

	gimlet sync folder-to-sync laszlo-debug@infrastructure:~/
	gimlet sync folder-to-sync laszlo-debug:~/
	`,
	Action: sync,
}

func sync(c *cli.Context) error {
	if c.Args().Len() != 2 {
		return fmt.Errorf("Usage: gimlet sync <folder-to-sync> <pod-name>[@<namespace>]:<path-in-pod>")
	}

	// check if kubectl installed
	fmt.Println("Verifying kubectl is installed..")
	cmd := exec.Command("which", "kubectl")
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("could not find kubectl")
	}

	// verifying rsync installed
	fmt.Println("Verifying rsync is installed..")
	podAndNamespace := os.Args[3]
	podAndNamespace = strings.Split(podAndNamespace, ":")[0]
	pod := ""
	namespaceClause := []string{}
	parts := strings.Split(podAndNamespace, "@")
	if len(parts) == 2 {
		pod = parts[0]
		namespaceClause = strings.Split("-n "+parts[1], " ")
	} else {
		pod = parts[0]
	}
	cmd = exec.Command("kubectl",
		append(namespaceClause, []string{"exec", "-i", pod, "--", "which", "rsync"}...)...,
	)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("rsync must be installed in the pod \n%s", err)
	}

	// prep krsync command
	krsyncPath, err := ioutil.TempFile("", "krsync")
	if err != nil {
		return err
	}

	_, err = krsyncPath.Write([]byte(krsync))
	if err != nil {
		return err
	}
	cmd = exec.Command("sh", "-c", "chmod 700 "+krsyncPath.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt)

	for {
		select {
		case <-ctrlC:
			return nil
		case <-ticker.C:
			var outb, errb bytes.Buffer
			cmd = exec.Command("bash", append([]string{krsyncPath.Name()}, os.Args[2:]...)...)
			cmd.Stdout = &outb
			cmd.Stderr = &errb
			cmd.Stdin = os.Stdin

			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "sync failed: %v\n", err)
				os.Exit(1)
			}

			screen.Clear()
			screen.MoveTopLeft()
			fmt.Println(outb.String())
			fmt.Println(errb.String())
		}
	}
}

const krsync = `
#!/bin/bash

if [ -z "$KRSYNC_STARTED" ]; then
    export KRSYNC_STARTED=true
    exec rsync -av --progress --stats --blocking-io --rsh "$0" $@
fi

# Running as --rsh
namespace=''
pod=$1
shift

# If user uses pod@namespace, rsync passes args as: {us} -l pod namespace ...
if [ "X$pod" = "X-l" ]; then
    pod=$1
    shift
    namespace="-n $1"
    shift
fi

exec kubectl $namespace exec -i $pod -- "$@"
`
