GOFILES = $(shell find . -type f -name '*.go' -not -path "./.git/*")
LDFLAGS = '-s -w -extldflags "-static" -X github.com/gimlet-io/gimlet-cli/pkg.version.Version='${VERSION}

.PHONY: format test 
.PHONY: build-cli dist-cli build-cli-frontend build-stack-frontend fast-dist-cli
.PHONY: build-gimletd dist-gilmetd

format:
	@gofmt -w ${GOFILES}

test-prep:
	touch pkg/commands/stack/web/bundle.js
	touch pkg/commands/stack/web/bundle.js.LICENSE.txt
	touch pkg/commands/stack/web/index.html
	touch pkg/commands/chart/bundle.js
	touch pkg/commands/chart/bundle.js.LICENSE.txt
	touch pkg/commands/chart/index.html
	git config --global user.email "git@gimlet.io"
	git config --global user.name "Github Actions"

test: test-prep
	go test -timeout 60s $(shell go list ./... )
test-dashboard-frontend:
	(cd web/dashboard; npm install; npm run test)
test-with-postgres:
	docker run --rm -e POSTGRES_PASSWORD=mysecretpassword -p 5432:5432 -d postgres

	export DATABASE_DRIVER=postgres
	export DATABASE_CONFIG=postgres://postgres:mysecretpassword@127.0.0.1:5432/postgres?sslmode=disable
	go test -timeout 60s github.com/gimlet-io/gimlet-cli/pkg/gimletd/store/...

build-cli:
	CGO_ENABLED=0 go build -ldflags $(LDFLAGS) -o build/gimlet github.com/gimlet-io/gimlet-cli/cmd/cli
build-gimletd:
	go build -ldflags $(LDFLAGS) -o build/gimletd github.com/gimlet-io/gimlet-cli/cmd/gimletd
build-agent:
	CGO_ENABLED=0 go build -ldflags $(LDFLAGS) -o build/gimlet-agent github.com/gimlet-io/gimlet-cli/cmd/agent
build-dashboard:
	CGO_ENABLED=0 go build -ldflags $(LDFLAGS) -o build/gimlet-dashboard github.com/gimlet-io/gimlet-cli/cmd/dashboard

dist-gimletd:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimletd-linux-x86_64 github.com/gimlet-io/gimlet-cli/cmd/gimletd
dist-dashboard:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimlet-dashboard-linux-x86_64 github.com/gimlet-io/gimlet-cli/cmd/dashboard
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimlet-agent-linux-x86_64 github.com/gimlet-io/gimlet-cli/cmd/agent
dist-cli:
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
build-dashboard-frontend:
	(cd web/dashboard; npm install; npm run build)

start-local-env:
	docker-compose -f fixtures/k3s/docker-compose.yml up -d
stop-local-env:
	docker-compose -f fixtures/k3s/docker-compose.yml stop
clean-local-env:
	docker-compose -f fixtures/k3s/docker-compose.yml down
	docker volume rm k3s_k3s-gimlet
