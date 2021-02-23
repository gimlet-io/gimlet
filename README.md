[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/gimlet-io/gimlet-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/gimlet-io/gimlet-cli)](https://goreportcard.com/report/github.com/gimlet-io/gimlet-cli)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# Gimlet CLI

A modular Gitops workflow for Kubernetes deployments.

## Installation

Linux / Mac

```
curl -L https://github.com/gimlet-io/gimlet-cli/releases/download/v0.5.0/gimlet-$(uname)-$(uname -m) -o gimlet
chmod +x gimlet
sudo mv ./gimlet /usr/local/bin/gimlet
gimlet --version
```

## Get Started

Gimlet CLI supports you throughout your Kubernetes deployment journey.

- If you are new to Kubernetes: [Deploy your app to Kubernetes without the boilerplate](https://gimlet.io/gimlet-cli/deploy-your-app-to-kubernetes-without-the-boilerplate/)

- If you want to modernize your CI pipeline: [Manage environments with Gimlet and GitOps](https://gimlet.io/gimlet-cli/manage-environments-with-gimlet-and-gitops/)

- If you want to manage services and environments at scale: [Manage environments with Gimlet and GitOps](https://gimlet.io/gimlet-cli/manage-environments-with-gimlet-and-gitops/)

## Reference

```
$ gimlet
NAME:
   gimlet - a modular Gitops workflow for Kubernetes deployments

USAGE:
   gimlet [global options] command [command options] [arguments...]

COMMANDS:
   chart     Manages Helm charts
   gitops    Manages the gitops repo
   seal      Seals secrets in the manifest
   manifest  Manages Gimlet manifests
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

## Development

[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/gimlet-io/gimlet-cli)

```
make all
./build/gimlet
```
