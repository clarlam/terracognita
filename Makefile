SHELL := /bin/bash
BIN := terracognita
BIN_DIR := $(GOPATH)/bin

GOLANGCI_LINT := $(BIN_DIR)/golangci-lint
GOLINT := $(BIN_DIR)/golint
MOCKGEN := $(BIN_DIR)/mockgen

VERSION= $(shell git describe --tags --always)
PLATFORMS=darwin linux windows
ARCHITECTURES=386 amd64
BUILD_PATH := builds

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-X github.com/cycloidio/terracognita/cmd.Version=${VERSION}"

.PHONY: help
help: Makefile ## This help dialog
	@IFS=$$'\n' ; \
	help_lines=(`fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//'`); \
	for help_line in $${help_lines[@]}; do \
		IFS=$$'#' ; \
		help_split=($$help_line) ; \
		help_command=`echo $${help_split[0]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		help_info=`echo $${help_split[2]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		printf "%-30s %s\n" $$help_command $$help_info ; \
	done

$(GOLANGCI_LINT):
	@go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

$(GOLINT):
	@go get -u golang.org/x/lint/golint

$(MOCKGEN):
	@go get -u github.com/golang/mock/mockgen

.PHONY: lint
lint: $(GOLANGCI_LINT) $(GOLINT) ## Runs the linter
	@GO111MODULE=on golangci-lint run -D errcheck -E goimports ./... && golint -set_exit_status ./...

.PHONY: generate
generate: $(MOCKGEN) ## Generates the needed code
	@GO111MODULE=on go generate ./...

.PHONY: test
test: ## Runs the tests
	@docker run --rm \
		-v $$(pwd):/app \
		-w /app \
		-u $(shell id -u):$(shell id -g) \
		-v $(shell go env GOCACHE):/tmp/gocach \
		-e "GOCACHE=/tmp/gocach" \
		-v $(GOPATH)/pkg/mod:/go/pkg/mod golang:1.12 \
		go test ./...

.PHONY: ci
ci: lint test ## Runs the linter and the tests

.PHONY: dbuild
dbuild: ## Builds the docker image with same name as the binary
	@docker build -t $(BIN) .

.PHONY: build
build: ## Builds the binary
	GO111MODULE=on CGO_ENABLED=0 GOARCH=amd64 go build -o $(BIN) ${LDFLAGS}

.PHONY: build-all build-compress
build-all: ## Builds the binaries
	$(foreach GOOS, $(PLATFORMS),\
	$(foreach GOARCH, $(ARCHITECTURES),\
	$(shell export GO111MODULE=on; export CGO_ENABLED=0; export GOOS=$(GOOS); export GOARCH=$(GOARCH); go build -v -o $(BUILD_PATH)/$(BIN)-$(GOOS)-$(GOARCH) ${LDFLAGS})))

build-compress: build-all ## Builds and compress the binaries
	$(foreach GOOS, $(PLATFORMS),\
	$(foreach GOARCH, $(ARCHITECTURES),\
	$(shell export GOOS=$(GOOS); export GOARCH=$(GOARCH); tar -C $(BUILD_PATH) -czf $(BUILD_PATH)/$(BIN)-$(GOOS)-$(GOARCH).tar.gz $(BIN)-$(GOOS)-$(GOARCH))))

.PHONY: install
install: ## Install the binary
	GO111MODULE=on CGO_ENABLED=0 GOARCH=amd64 go install ${LDFLAGS}

.PHONY: clean
clean: ## Removes binary and/or docker image
	rm -f $(BIN)
	docker rmi -f $(BIN)
