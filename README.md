# Gimlet

[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/gimlet-io/gimlet-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/gimlet-io/gimlet-cli)](https://goreportcard.com/report/github.com/gimlet-io/gimlet-cli)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Roadmap](https://img.shields.io/badge/roadmap-gimlet-blue)](https://github.com/orgs/gimlet-io/projects/1/views/2)
[![Milestones](https://img.shields.io/badge/milestones-gimlet-blue)](https://github.com/gimlet-io/gimlet/milestones)

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://github.com/gimlet-io/gimlet-documentation/blob/main/public/logo-dark.svg">
  <img alt="Gimlet" src="https://github.com/gimlet-io/gimlet-documentation/blob/main/public/logo.svg" width="200">
</picture>

## Overview

With Gimlet, you can build and run your developer platform on Kubernetes.

Gimlet is a command line tool and a dashboard that packages a set of conventions and matching workflows to manage a gitops developer platform effectively.

Caters for cluster admin and developer workflows.

## Installation

### CLI
Linux / Mac

```console
curl -L "https://github.com/gimlet-io/gimlet/releases/download/cli-v0.20.0/gimlet-$(uname)-$(uname -m)" -o gimlet
chmod +x gimlet
sudo mv ./gimlet /usr/local/bin/gimlet
gimlet --version
```

### Dashboard

The following oneliner kickstarts the Gimlet installer.

```bash
curl -L -s https://get.gimlet.io | bash -s staging.mycompany.com [my-github-org]
```

The installer initiates a gitops environment and puts Gimlet into its gitops repository. This way Gimlet itself is managed by gitops.

[Installer documentaion](https://gimlet.io/docs/installation)

## For cluster admins
- You can make an empty Kubernetes cluster a developer platform with ingress, observability, SSL, policies
- Then get a curated update stream of security and feature upgrades
- All delivered in a git repo with self-contained gitops automation

You can use the dashboard or the `gimlet stack` command for this.

### Tutorials
- [Make Kubernetes an application platform](https://gimlet.io/docs/make-kubernetes-an-application-platform)
- [Managing infrastructure components](https://gimlet.io/docs/managing-infrastructure-components)

[![Star on GitHub](https://img.shields.io/github/stars/gimlet-io/gimlet.svg?style=social)](https://github.com/gimlet-io/gimlet/stargazers)

## For developers
- Configure your services without the Kubernetes yaml boilerplate
- Deploy, rollback from CLI or from a dashboard
- Focus on your own application, no need to navigate an inventory of Kubernetes resoure types


### For zero config deploys, try this

```bash

cat << EOF > staging.yaml
app: myapp
env: staging
namespace: my-team
chart:
  repository: https://chart.onechart.dev
  name: onechart
  version: 0.36.0
values:
  image:
    repository: myapp
    tag: 1.1.0
  ingress:
    host: myapp.staging.mycompany.com
    tlsEnabled: true
EOF

gimlet manifest template -f staging.yaml -o manifests.yaml
```

### Tutorials

- [Deploy your first app to Kubernetes](https://gimlet.io/docs/deploy-your-first-app-to-kubernetes)
- [How to manage deployment configs](https://gimlet.io/docs/how-to-manage-deployment-configs)

[![Star on GitHub](https://img.shields.io/github/stars/gimlet-io/gimlet.svg?style=social)](https://github.com/gimlet-io/gimlet/stargazers)

## Youtube playlist on building a full developer platform

[Building a developer platform on CIVO Cloud with Gimlet and gitops](https://youtube.com/playlist?list=PLjJkiSWPwuPJeIEOn5BWMFdxSSpiQPQ4P)

