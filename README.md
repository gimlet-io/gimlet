<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Gimlet CLI](#gimlet-cli)
  - [Installation](#installation)
  - [Usage](#usage)
    - [Configuring a Helm chart](#configuring-a-helm-chart)
      - [Using with Helm template and install](#using-with-helm-template-and-install)
    - [Writing manifests to the gitops repo](#writing-manifests-to-the-gitops-repo)
      - [Configuring and writing a chart to gitops](#configuring-and-writing-a-chart-to-gitops)
    - [Handling secrets](#handling-secrets)
      - [Sealing secrets in `values.yaml`](#sealing-secrets-in-valuesyaml)
  - [Development](#development)
    - [Housekeeping README.md](#housekeeping-readmemd)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/gimlet-io/gimlet-cli)

# Gimlet CLI

For a modular Gitops workflow.

## Installation

Linux / Mac

```
curl -L https://github.com/gimlet-io/gimlet-cli/releases/download/v0.0.1/gimlet-$(uname)-$(uname -m) -o gimlet
chmod +x gimlet
sudo mv ./gimlet /usr/local/bin/gimlet
gimlet --version
```

## Usage

### Configuring a Helm chart

```
➜  ~ gimlet chart configure onechart/onechart
👩‍💻 Configure on http://127.0.0.1:28955
👩‍💻 Close the browser when you are done
Browser opened
Browser closed
📁 Generating values..

---
image:
  repository: myapp
  tag: 1.0.0
ingress:
  host: myapp.local
  tlsEnabled: true
replicas: 2
```

Saving values.yaml

```
gimlet chart configure onechart/onechart > values.yaml
```

Updating values.yaml in place

```
gimlet chart configure -f values.yaml -o values.yaml onechart/onechart
```

#### Using with Helm template and install

One-liner:

```
gimlet chart configure onechart/onechart | helm template myapp onechart/onechart -f -
```

Keeping values.yaml for versioning:

```
gimlet chart configure onechart/onechart > values.yaml
helm template myapp onechart/onechart -f values.yaml
```

### Writing manifests to the gitops repo

```
NAME:
   gimlet gitops write - Writes app manifests to a gitops environment

USAGE:
   gimlet gitops write -f my-app.yaml \
     --env staging \
     --app my-app \
     -m "Releasing Bugfix 345"

OPTIONS:
   --file value, -f value     manifest file,folder or "-" for stdin to write (mandatory)
   --env value                environment to write to (mandatory)
   --app value                name of the application that you configure (mandatory)
   --gitops-repo-path value   path to the working copy of the gitops repo
   --message value, -m value  gitops commit message
   --help, -h                 show help (default: false)
```

#### Configuring and writing a chart to gitops

```
gimlet chart configure onechart/onechart | \
  helm template myapp onechart/onechart -f - | \
  gimlet gitops write -f - \
    --env staging \
    --app my-app \
    -m "First deploy"
```

### Handling secrets

You can store encrypted secrets in the GitOps repo with Gimlet CLI as it has built-in support for [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets).

#### Sealing secrets in `values.yaml`

Sealing values under the `sealedSecrets` field using Bitnami's Sealed Secrets sealing key:

```
gimlet seal -f values.yaml -o values.yaml -p sealedSecrets -k sealingKey.crt
```

Configuring and sealing Helm charts:

```
gimlet chart configure onechart/onechart |
    gimlet seal -p sealedSecrets -k sealingKey.crt -f - > values.yaml
```

## Development

[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/gimlet-io/gimlet-cli)

```
make all
./build/gimlet
```

### Housekeeping README.md

```
npx doctoc README.md && npx prettier --write "**/*.md" "*.md"
```
