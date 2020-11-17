GO_VERSION=1.14.7
GOFILES = $(shell find . -type f -name '*.go' -not -path "./.git/*")
LDFLAGS = '-extldflags "-static" -X github.com/gimlet-io/gimlet-cli/version.Version='${VERSION}

DOCKER_RUN?=
_with-docker:
	$(eval DOCKER_RUN=docker run --rm -v $(shell pwd):/go/src/github.com/gimlet-io/gimlet-cli -w /go/src/github.com/gimlet-io/gimlet-cli golang:$(GO_VERSION))

.PHONY: all format-backend test-backend generate-backend build-backend dist build-frontend

all: build-frontend generate-backend test-backend build-backend

format-backend:
	@gofmt -w ${GOFILES}

test-backend:
	$(DOCKER_RUN) go test -race -timeout 30s github.com/gimlet-io/gimlet-cli/cmd $(go list ./... )

generate-backend:
	$(DOCKER_RUN) go generate github.com/gimlet-io/gimlet-cli/cmd

build-backend:
	$(DOCKER_RUN) go build -ldflags $(LDFLAGS) -o build/gimlet github.com/gimlet-io/gimlet-cli/cmd

dist:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimlet github.com/gimlet-io/gimlet-cli/cmd
	CGO_ENABLED=0 GOOS=darwin go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimlet-darwin github.com/gimlet-io/gimlet-cli/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimlet-armhf github.com/gimlet-io/gimlet-cli/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimlet-arm64 github.com/gimlet-io/gimlet-cli/cmd
	CGO_ENABLED=0 GOOS=windows go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimlet.exe github.com/gimlet-io/gimlet-cli/cmd

build-frontend:
	(cd web/; npm install; npm run build)
