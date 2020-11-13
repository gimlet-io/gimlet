[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/gimlet-io/gimlet-cli)

# Gimlet CLI

For an open-source Gitops workflow.

## Development

[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/gimlet-io/gimlet-cli)

```
make all
./build/gimlet
```

#### Dockerized development

Prepend make targets `_with-docker`:
```
➜  gimlet-cli git:(main) make _with-docker test                                   
make: '_with-docker' is up to date.
docker run --rm -v /home/laszlo/projects/gimlet-cli:/go/src/github.com/gimlet-io/gimlet-cli -w /go/src/github.com/gimlet-io/gimlet-cli golang:1.14.7 go test -race -timeout 30s github.com/gimlet-io/gimlet-cli/cmd 
go: downloading github.com/urfave/cli/v2 v2.3.0
go: downloading github.com/cpuguy83/go-md2man/v2 v2.0.0-20190314233015-f79a8a8ca69d
go: downloading github.com/russross/blackfriday/v2 v2.0.1
go: downloading github.com/shurcooL/sanitized_anchor_name v1.0.0
?       github.com/gimlet-io/gimlet-cli/cmd     [no test files]
``` 