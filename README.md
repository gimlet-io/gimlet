# Gimlet CLI

[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/gimlet-io/gimlet-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/gimlet-io/gimlet-cli)](https://goreportcard.com/report/github.com/gimlet-io/gimlet-cli)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Overview

Gimlet CLI is a command line tool that packages a set of conventions and matching workflows to manage the GitOps repository effectively. A modular Gitops workflow for Kubernetes deployments.

## Installation

Linux / Mac

```console
curl -L https://github.com/gimlet-io/gimlet-cli/releases/download/v0.9.2/gimlet-$(uname)-$(uname -m) -o gimlet
chmod +x gimlet
sudo mv ./gimlet /usr/local/bin/gimlet
gimlet --version
```

## Get Started

Gimlet CLI supports you throughout your Kubernetes deployment journey.

- If you are new to Kubernetes: [Deploy your app to Kubernetes without the boilerplate](https://gimlet.io/gimlet-cli/deploy-your-app-to-kubernetes-without-the-boilerplate/)

- If you want to modernize your CI pipeline: [Manage environments with Gimlet and GitOps](https://gimlet.io/gimlet-cli/manage-environments-with-gimlet-and-gitops/)

- If you want to manage services and environments at scale: [Manage environments with Gimlet and GitOps](https://gimlet.io/gimlet-cli/manage-environments-with-gimlet-and-gitops/)

Visit [gimlet.io](https://gimlet.io/) for the full documentation, examples and guides.

## Contribution Guidelines

Welcome to the Gimlet project! 🤗 

We are excited to see your interest, and appreciate your support! We welcome contributions from people of all backgrounds who are interested in making great software with us. If you have any difficulties getting involved or finding answers to your questions, please don't hesitate to ask your questions.

### Issues

If you encounter any issues or have any relevant questions, please add an issue to [GitHub issues](https://github.com/gimlet-io/gimlet-cli/issues/new).

### New Features / Components

If you have any ideas on new features or want to improve the existing features, you can suggest it by opening a GitHub issue. Make sure to include detailed information about the feature requests, use cases, and any other information that could be helpful.

### Pull Request Process

* Fork the repository.
* Create a new branch and make your changes.
* Open a pull request with detailed commit message and reference issue number if applicable.
* A maintainer will review your pull request, and help you throughout the process.

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
