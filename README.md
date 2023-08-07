# Gimlet

[![Go Report Card](https://goreportcard.com/badge/github.com/gimlet-io/gimlet-cli)](https://goreportcard.com/report/github.com/gimlet-io/gimlet-cli)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Overview

Kubernetes deployment tooling.

- You can start with as little Kubernetes as you want - even on your laptop.
- Then progress to a gitops based application delivery platform when you need to.

Gimlet composes the best of open-source out of the box.

## Installation

#### CLI
Linux / Mac

```console
curl -L "https://github.com/gimlet-io/gimlet/releases/download/cli-v0.23.4/gimlet-$(uname)-$(uname -m)" -o gimlet
chmod +x gimlet
sudo mv ./gimlet /usr/local/bin/gimlet
gimlet --version
```

#### Dashboard

Gimlet is installable by a kubectl apply:

```
kubectl apply -f https://raw.githubusercontent.com/gimlet-io/gimlet/main/deploy/gimlet.yaml
```

or with a Helm chart:

```
helm template gimlet onechart/onechart -f fixtures/gimlet-helm-values.yaml -n default
```

## Documentation

[https://gimlet.io/docs/installation](https://gimlet.io/docs/installation)
