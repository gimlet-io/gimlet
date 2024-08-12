GOFILES = $(shell find . -type f -name '*.go' -not -path "./.git/*")
LDFLAGS = '-s -w -extldflags "-static" -X github.com/gimlet-io/gimlet/pkg/version.Version='${VERSION}

.PHONY: format test 
.PHONY: build-cli dist-cli fast-dist-cli fast-dist

format:
	@gofmt -w ${GOFILES}

test-prep:
	helm repo add onechart https://chart.onechart.dev

test: test-prep
	go test -timeout 60s $(shell go list ./...)
	go test -timeout 60s -tags=encryption github.com/gimlet-io/gimlet/pkg/dashboard/store -run TestEncryption
	go test -timeout 60s -tags=encryption github.com/gimlet-io/gimlet/pkg/dashboard/store -run TestReEncryption
	go test -timeout 60s -tags=encryption github.com/gimlet-io/gimlet/pkg/dashboard/store -run TestInitEncryption
	go test -timeout 60s -tags=encryption github.com/gimlet-io/gimlet/pkg/dashboard/store -run TestUnquote

test-dashboard-frontend:
	(cd web; npm install; npm run test)
test-with-postgres:
	docker stop testpostgres || true
	docker run --rm -e POSTGRES_PASSWORD=mysecretpassword -p 5432:5432 --name testpostgres -d postgres

	export DATABASE_DRIVER=postgres
	export DATABASE_CONFIG=postgres://postgres:mysecretpassword@127.0.0.1:5432/postgres?sslmode=disable

	go test -timeout 60s -tags=encryption github.com/gimlet-io/gimlet/pkg/dashboard/store -run TestEncryption
	go test -timeout 60s -tags=encryption github.com/gimlet-io/gimlet/pkg/dashboard/store -run TestReEncryption
	go test -timeout 60s -tags=encryption github.com/gimlet-io/gimlet/pkg/dashboard/store -run TestInitEncryption
	go test -timeout 60s -tags=encryption github.com/gimlet-io/gimlet/pkg/dashboard/store -run TestUnquote

	docker stop testpostgres || true

build-cli:
	CGO_ENABLED=0 go build -ldflags $(LDFLAGS) -o build/gimlet github.com/gimlet-io/gimlet/cmd/cli
build-agent:
	CGO_ENABLED=0 go build -ldflags $(LDFLAGS) -o build/gimlet-agent github.com/gimlet-io/gimlet/cmd/agent
build-dashboard:
	CGO_ENABLED=0 go build -ldflags $(LDFLAGS) -o build/gimlet-dashboard github.com/gimlet-io/gimlet/cmd/dashboard
build-image-builder:
	CGO_ENABLED=0 go build -ldflags $(LDFLAGS) -o build/image-builder github.com/gimlet-io/gimlet/cmd/image-builder

dist-dashboard:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/linux/amd64/gimlet-dashboard github.com/gimlet-io/gimlet/cmd/dashboard
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/linux/amd64/gimlet-agent github.com/gimlet-io/gimlet/cmd/agent

	GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/linux/arm64/gimlet-dashboard github.com/gimlet-io/gimlet/cmd/dashboard
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/linux/arm64/gimlet-agent github.com/gimlet-io/gimlet/cmd/agent
dist-cli:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimlet-darwin-x86_64 github.com/gimlet-io/gimlet/cmd/cli
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimlet-darwin-arm64 github.com/gimlet-io/gimlet/cmd/cli
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimlet-linux-arm64 github.com/gimlet-io/gimlet/cmd/cli
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimlet-linux-x86_64 github.com/gimlet-io/gimlet/cmd/cli
	CGO_ENABLED=0 GOOS=windows go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/gimlet.exe github.com/gimlet-io/gimlet/cmd/cli
dist-image-builder:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/image-builder-linux-x86_64 github.com/gimlet-io/gimlet/cmd/image-builder
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/image-builder-linux-arm64 github.com/gimlet-io/gimlet/cmd/image-builder

build-frontend:
	(cd web; npm install; npm run build)
	rm -rf cmd/dashboard/web/build
	mkdir -p cmd/dashboard/web/build
	@cp -r web/build/* cmd/dashboard/web/build
