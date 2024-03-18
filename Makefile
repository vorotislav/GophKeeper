PROGRAM_NAME = gophkeeper

BUILD_VERSION=$(shell git describe --tags)
BUILD_DATE=$(shell date +%FT%T%z)
BUILD_COMMIT=$(shell git rev-parse --short HEAD)

PKG_PATH=GophKeeper/cmd/util

LDFLAGS_CLIENT=-X ${PKG_PATH}.buildVersion=$(BUILD_VERSION) -X ${PKG_PATH}.buildDate=$(BUILD_DATE) -X ${PKG_PATH}.buildCommit=$(BUILD_COMMIT)
LDFLAGS_SERVER=-X ${PKG_PATH}.buildVersion=$(BUILD_VERSION) -X ${PKG_PATH}.buildDate=$(BUILD_DATE) -X ${PKG_PATH}.buildCommit=$(BUILD_COMMIT)

EXECUTABLE=gophkeeper_cli
WINDOWS=$(EXECUTABLE)_windows_amd64.exe
LINUX=$(EXECUTABLE)_linux_amd64
DARWIN=$(EXECUTABLE)_darwin_arm64

.PHONY: help dep fmt test

dep: ## Get the dependencies
	go mod download

fmt: ## Format the source files
	gofumpt -l -w .

test: dep ## Run tests
	go test -timeout 5m -race -covermode=atomic -coverprofile=.coverage.out ./... && \
	go tool cover -func=.coverage.out | tail -n1 | awk '{print "Total test coverage: " $$3}'
	@rm .coverage.out

cover: dep ## Run app tests with coverage report
	go test -timeout 5m -race -covermode=atomic -coverprofile=.coverage.out ./... && \
	go tool cover -html=.coverage.out -o .coverage.html
	## Open coverage report in default system browser
	xdg-open .coverage.html
	## Remove coverage report
	sleep 2 && rm -f .coverage.out .coverage.html

build: build/server	build/client

build/server:
	go build -ldflags "${LDFLAGS_SERVER}" -o ./bin/server ./cmd/server

build/client: build/client/darwin build/client/linux build/client/windows

build/client/windows:
	env GOOS=windows GOARCH=amd64 go build -v -o ./bin/client/win/$(WINDOWS) -ldflags "${LDFLAGS_CLIENT}" ./cmd/client

build/client/linux:
	env GOOS=linux GOARCH=amd64 go build -v -o ./bin/client/lin/$(LINUX) -ldflags "${LDFLAGS_CLIENT}" ./cmd/client

build/client/darwin:
	env GOOS=darwin GOARCH=arm64 go build -v -o ./bin/client/darwin/$(DARWIN) -ldflags "${LDFLAGS_CLIENT}"  ./cmd/client

windows: $(WINDOWS) ## Build for Windows

linux: $(LINUX) ## Build for Linux

darwin: $(DARWIN) ## Build for Darwin (macOS)

lint: lint/sources lint/openapi ## Run all linters

lint/sources: ## Lint the source files
	golangci-lint run --timeout 5m
	govulncheck ./...

lint/openapi: ## Lint openapi specifications
	@echo "Lint OpenAPI specifications"
	@for spec in $(OPENAPI_SPECS) ; do echo "* lint $$spec"; vacuum lint -t -q -x $$spec ; done