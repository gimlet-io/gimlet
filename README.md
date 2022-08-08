# Gimlet

[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/gimlet-io/gimlet-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/gimlet-io/gimlet-cli)](https://goreportcard.com/report/github.com/gimlet-io/gimlet-cli)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

<img src="Gimlet.svg" width="200"/>

## Overview

With Gimlet, you can build and run your developer platform on Kubernetes.

Gimlet is a command line tool and a dashboard that packages a set of conventions and matching workflows to manage a gitops developer platform effectively.

Caters for cluster admin and developer workflows.

## Installation

Linux / Mac

```console
curl -L https://github.com/gimlet-io/gimlet/releases/download/cli-v0.16.0/gimlet-$(uname)-$(uname -m) -o gimlet
chmod +x gimlet
sudo mv ./gimlet /usr/local/bin/gimlet
gimlet --version
```

## For cluster admins
- You can make an empty Kubernetes cluster a developer platform with ingress, observability, SSL, policies
- Then get a curated update stream of security and feature upgrades
- All delivered in a git repo with self-contained gitops automation

You can use the dashboard or the `gimlet stack` command for this.

### Tutorials
- [Make Kubernetes an application platform with Gimlet Stack](https://gimlet.io/docs/make-kubernetes-an-application-platform-with-gimlet-stack/)
- [Reconfiguring, upgrading and making custom changes to stacks](https://gimlet.io/docs/reconfiguring-upgrading-and-making-custom-changes-to-stacks/)

Don't forget to star the project. Your support goes a long way 🙏

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

- [Deploy your app to Kubernetes without the boilerplate](https://gimlet.io/docs/deploy-your-app-to-kubernetes-without-the-boilerplate/)
- [Manage your staging application configuration](https://gimlet.io/docs/manage-your-staging-application-configuration/)
- [Automatically deploy your application to staging](https://gimlet.io/docs/automatically-deploy-your-application-to-staging/)

Don't forget to star the project. Your support goes a long way 🙏

[![Star on GitHub](https://img.shields.io/github/stars/gimlet-io/gimlet.svg?style=social)](https://github.com/gimlet-io/gimlet/stargazers)

## Installing the dashboard

```
curl -L -s https://get.gimlet.io | bash -s <<your-domain.com>>
```

[See the full installation on this captioned video](https://youtu.be/HFjv7_08oP0)

## Youtube playlist on building a full developer platform

[Building a developer platform on CIVO Cloud with Gimlet and gitops](https://youtube.com/playlist?list=PLjJkiSWPwuPJeIEOn5BWMFdxSSpiQPQ4P)

