BINDIR      := $(CURDIR)/bin
INSTALL_PATH ?= /usr/local/bin
DIST_DIRS   := find * -type d -exec
TARGETS     := darwin/amd64 darwin/arm64 linux/amd64 linux/386 linux/arm linux/arm64 linux/ppc64le linux/s390x windows/amd64

BINNAME     ?= iter8
ITER8_IMG ?= iter8/iter8:latest

GOBIN         = $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN         = $(shell go env GOPATH)/bin
endif
GOX           = $(GOBIN)/gox

# go options
TAGS       :=
LDFLAGS    := -w -s
GOFLAGS    :=

# Rebuild the binary if any of these files change
SRC := $(shell find . -type f -name '*.go' -print) go.mod go.sum

# Required for globs to work correctly
SHELL      = /usr/bin/env bash

GIT_TAG    = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)

ifdef VERSION
	BINARY_VERSION = $(VERSION)
endif
BINARY_VERSION ?= ${GIT_TAG}

# Only set Version if building a tag or VERSION is set
ifneq ($(BINARY_VERSION),)
	LDFLAGS += -X github.com/iter8-tools/iter8/basecli/RootCmd.Version=${BINARY_VERSION}
endif

LDFLAGS += $(EXT_LDFLAGS)

.PHONY: all
all: build

# ------------------------------------------------------------------------------
#  build

.PHONY: build
build: $(BINDIR)/$(BINNAME)

$(BINDIR)/$(BINNAME): $(SRC)
	GO111MODULE=on go build $(GOFLAGS) -trimpath -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o '$(BINDIR)'/$(BINNAME) ./

.PHONY: cmddocs
cmddocs:
	go run cmd/docs/main.go	

# ------------------------------------------------------------------------------
#  install

.PHONY: install
install: build
	@install "$(BINDIR)/$(BINNAME)" "$(INSTALL_PATH)/$(BINNAME)"

# ------------------------------------------------------------------------------
#  dependencies

# If go get is run from inside the project directory it will add the dependencies
# to the go.mod file. To avoid that we change to a directory without a go.mod file
# when downloading the following dependencies

$(GOX):
	(cd /; GO111MODULE=on go get -u github.com/mitchellh/gox)

# ------------------------------------------------------------------------------
#  release

.PHONY: build-cross
build-cross: LDFLAGS += -extldflags "-static"
build-cross: $(GOX)
	GOFLAGS="-trimpath" GO111MODULE=on CGO_ENABLED=0 $(GOX) -parallel=3 -output="_dist/{{.OS}}-{{.Arch}}/$(BINNAME)" -osarch='$(TARGETS)' $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' ./

.PHONY: dist
dist: build-cross
	( \
		cd _dist && \
		$(DIST_DIRS) cp ../LICENSE {} \; && \
		$(DIST_DIRS) cp ../README.md {} \; && \
		$(DIST_DIRS) tar -zcf iter8-${VERSION}-{}.tar.gz {} \; && \
		$(DIST_DIRS) zip -r iter8-${VERSION}-{}.zip {} \; \
	)

.PHONY: clean
clean:
	@rm -rf '$(BINDIR)' ./_dist

.PHONY: docker-build
docker-build:
	docker build -f Dockerfile -t $(ITER8_IMG) .

.PHONY: docker-push
docker-push:
	docker push $(ITER8_IMG)

# ------------------------------------------------------------------------------
#  test

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code
	go vet ./...

.PHONY: staticcheck
staticcheck:
	staticcheck ./...

.PHONY: test
test: fmt vet staticcheck ## Run tests.
	go test ./... -race -coverprofile=coverage.out -covermode=atomic	

.PHONY: coverage
coverage: test
	@echo "test coverage: $(shell go tool cover -func coverage.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}')"

.PHONY: htmlcov
htmlcov: coverage
	go tool cover -html=coverage.out
