GOFILES = $(shell find . -type f -name '*.go' -not -path "./.git/*")
LDFLAGS = '-s -w -extldflags "-static" -X github.com/gimlet-io/gimlet-cli/pkg.version.Version='${VERSION}

.PHONY: all format test build-cli build-stack dist build-cli-frontend build-stack-frontend fast-dist-cli fast-dist-stack

format:
	@gofmt -w ${GOFILES}

test-prep:
	touch pkg/commands/stack/web/bundle.js
	touch pkg/commands/stack/web/bundle.js.LICENSE.txt
	touch pkg/commands/stack/web/index.html
	touch pkg/commands/chart/bundle.js
	touch pkg/commands/chart/bundle.js.LICENSE.txt
	touch pkg/commands/chart/index.html

test: test-prep
	go test -timeout 60s $(shell go list ./... )

build-cli:
	CGO_ENABLED=0 go build -ldflags $(LDFLAGS) -o build/gimlet github.com/gimlet-io/gimlet-cli/cmd/cli
build-stack:
	CGO_ENABLED=0 go build -ldflags $(LDFLAGS) -o build/stack github.com/gimlet-io/gimlet-cli/cmd/stack

dist:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimlet-darwin-x86_64 github.com/gimlet-io/gimlet-cli/cmd/cli
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimlet-darwin-arm64 github.com/gimlet-io/gimlet-cli/cmd/cli
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimlet-linux-arm64 github.com/gimlet-io/gimlet-cli/cmd/cli
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimlet-linux-x86_64 github.com/gimlet-io/gimlet-cli/cmd/cli
	CGO_ENABLED=0 GOOS=windows go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimlet.exe github.com/gimlet-io/gimlet-cli/cmd/cli

fast-dist-cli:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimlet-linux-x86_64 github.com/gimlet-io/gimlet-cli/cmd/cli
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimlet-darwin-x86_64 github.com/gimlet-io/gimlet-cli/cmd/cli
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimlet-darwin-arm64 github.com/gimlet-io/gimlet-cli/cmd/cli

fast-dist-stack:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/stack-linux-x86_64 github.com/gimlet-io/gimlet-stack/cmd
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/stack-darwin-x86_64 github.com/gimlet-io/gimlet-stack/cmd

build-cli-frontend:
	(cd web/cli; npm install; npm run build)
	@cp web/cli/dist/bundle.js pkg/commands/chart/
	@cp web/cli/dist/bundle.js.LICENSE.txt pkg/commands/chart/
	@cp web/cli/dist/index.html pkg/commands/chart/

build-stack-frontend:
	(cd web/stack; npm install; npm run build)
	@cp web/stack/dist/bundle.js pkg/commands/stack/web/
	@cp web/stack/dist/bundle.js.LICENSE.txt pkg/commands/stack/web/
	@cp web/stack/dist/index.html pkg/commands/stack/web/
