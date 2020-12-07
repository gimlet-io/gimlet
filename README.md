<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Gimlet CLI](#gimlet-cli)
  - [Installation](#installation)
  - [Usage](#usage)
    - [Bootstrapping GitOps](#bootstrapping-gitops)
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
curl -L https://github.com/gimlet-io/gimlet-cli/releases/download/v0.1.0/gimlet-$(uname)-$(uname -m) -o gimlet
chmod +x gimlet
sudo mv ./gimlet /usr/local/bin/gimlet
gimlet --version
```

## Usage

### Bootstrapping GitOps

```
$ gimlet gitops bootstrap \
  --env staging
  --gitops-repo-url "git@github.com:<user>/<repo>.git"

⏳ Generating manifests
⏳ Generating deploy key
✔️ GitOps configuration written to gitops/staging/flux

👉 1) Push the configuration to git
👉 2) Add the following deploy key to your Git provider

ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDK0t17JqpvvciRNMj1tQ0pZdHLRIi/o/dNaI4Stdc8kaKci3DlL3P8BPu0tXt78OH2CHSEUaMNpoQcKpnvZrgomBQikTHGwdqM89o0C7MSjMdI1V4Lp8V7ZJ
jY3nT2WCUgCYB3TDvQps/ficr5wXNB7Y0+nkNSf0q3sbtsnz9LL0vSFhK0Uj3b7p9eNdkEB8gYvedmBRW8GljDk/s5oKrHaz1eHwQqTMseTdPSgRuB6W1kFBFnOxMERpyRgfrcjkipiS/q8Or+eQQ7ghzHJ5GD
30OicdBpdukJJ3fIymgxnuMDrJdh1x/rvoAN76MXKkGcApiUYTTPvNFhKMkmjjtLieUXCyigKIZOsA1Qh4eUhDEs4f7OAKgFU77KiGU73Lm0XYYEiwcupGR4nY9sW5BvaLDKSXuUXNIsVROKOFUOrUIMRT6pXD
jlC92QkOo2Y10qfDazUoCkZ2i4mtMUpVLEWThVAg8h6yzcwpwOvR23ISMb6YWiHU4UQe29AJuatW0nWxpx7ks6+dqhP9LL2z10BiEpHehEYrOMf+H5iUxklRXNanvDoGWy9srRFOLG4uPaLDOLAj6DXcFySlda
MPC3rWWiUPCsdzrdmI4AbAK3xEBUVw7dipGzZtkQQa+Vgb/F9QQuIXgOWcZkMhQBnfzebJsWP9simgEzPjYS+l5sWw==

👉 3) Apply the gitops manifests on the cluster to start the gitops loop:

kubectl apply -f gitops/staging/flux/flux.yaml
kubectl apply -f gitops/staging/flux/deploy-key.yaml
kubectl wait --for condition=established --timeout=60s crd/gitrepositories.source.toolkit.fluxcd.io
kubectl wait --for condition=established --timeout=60s crd/kustomizations.kustomize.toolkit.fluxcd.io
kubectl apply -f gitops/staging/flux/gitops-repo.yaml

🎊 Happy Gitopsing
```

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

Seal secret values with [Bitnami's Sealed Secret project](https://github.com/bitnami-labs/sealed-secrets).

```bash
# Fetch the sealing key first
kubeseal --fetch-cert --controller-namespace=infrastructure > sealing-key.pub

# Configure a Helm chart then seal all secrets in one go
gimlet chart configure onechart/onechart |
    gimlet seal -p sealedSecrets -k sealingKey.pub -f - > values.yaml
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
