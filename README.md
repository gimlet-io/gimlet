[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/gimlet-io/gimlet-cli)

# Gimlet CLI

For an open-source Gitops workflow.

## Installation

Linux
```
curl -L https://github.com/gimlet-io/gimlet-cli/releases/download/v0.0.1/gimlet -o gimlet
chmod +x gimlet
sudo mv ./gimlet /usr/local/bin/gimlet
gimlet --version
```

Mac
```
curl -L https://github.com/gimlet-io/gimlet-cli/releases/download/v0.0.1/gimlet -o gimlet
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

### Writing values.yaml

```
gimlet chart configure onechart/onechart > values.yaml
```

### Using with Helm template and install

```
gimlet chart configure onechart/onechart | helm template myapp onechart/onechart -f -
```

## Development

[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/gimlet-io/gimlet-cli)

```
make all
./build/gimlet
```

#### Dockerized development

Prepend make targets `_with-docker`:
```
make _with-docker test build
``` 