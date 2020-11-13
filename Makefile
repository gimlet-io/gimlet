GO_VERSION=1.14.7
GOFILES = $(shell find . -type f -name '*.go' -not -path "./.git/*")

DOCKER_RUN?=
_with-docker:
	$(eval DOCKER_RUN=docker run --rm -v $(shell pwd):/go/src/github.com/gimlet-io/gimlet-cli -w /go/src/github.com/gimlet-io/gimlet-cli golang:$(GO_VERSION))

all: test build

format:
	@gofmt -w ${GOFILES}

test:
	$(DOCKER_RUN) go test -race -timeout 30s github.com/gimlet-io/gimlet-cli/cmd $(go list ./... )

build:
	$(DOCKER_RUN) go build -o build/gimlet github.com/gimlet-io/gimlet-cli/cmd
