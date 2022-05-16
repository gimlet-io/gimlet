# Gimlet CLI

[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/gimlet-io/gimlet-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/gimlet-io/gimlet-cli)](https://goreportcard.com/report/github.com/gimlet-io/gimlet-cli)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)


![Gimlet](https://raw.githubusercontent.com/gimlet-io/logos/main/Gimlet.svg?token=GHSAT0AAAAAABQITJ4YEEXMIO3CKHIUWHYGYUCAGHA | width=100)

## Overview

With Gimlet, you can build and run your developer platform on Kubernetes.

Gimlet is a command line tool and a dashboard that packages a set of conventions and matching workflows to manage a gitops developer platform effectively.

Caters for cluster admin and developer workflows.

For **cluster admins**:
- You can make an empty Kubernetes cluster a developer platform with ingress, observability, SSL, policies
- Then get a curated update stream of security and feature upgrades
- All delivered in a git repo with self-contained gitops automation

For **developers**:
- Configure your services without the Kubernetes yaml boilerplate
- Deploy, rollback from CLI or from a dashboard
- Focus on your own application, no need to navigate an inventory of Kubernetes resoure types

## Installation

Linux / Mac

```console
curl -L https://github.com/gimlet-io/gimlet/releases/download/cli-v0.15.0/gimlet-$(uname)-$(uname -m) -o gimlet
chmod +x gimlet
sudo mv ./gimlet /usr/local/bin/gimlet
gimlet --version
```
