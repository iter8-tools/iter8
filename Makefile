# Rebuild the binary if any of these files change
SRC := $(shell find . -type f -name '*.(go|proto|tpl)' -print) go.mod go.sum cmd/gitcommit.txt

# ------------------------------------------------------------------------------
#  build and install

.PHONY: gitcommit
gitcommit:
	./gitcommit.sh

.PHONY: build
build: gitcommit
	go build

.PHONY: install
install: gitcommit
	go install

# ------------------------------------------------------------------------------
#  test

.PHONY: fmt
fmt: gitcommit
	go fmt ./...

.PHONY: vet
vet: gitcommit
	go vet ./...

.PHONY: golangci-lint
golangci-lint: gitcommit
	golangci-lint run ./...

.PHONY: test
test: fmt vet golangci-lint
	go test -v ./... -coverprofile=coverage.out

.PHONY: coverage
coverage: test
	@echo "test coverage: $(SHELL go tool cover -func coverage.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}')"

.PHONY: htmlcov
htmlcov: coverage
	go tool cover -html=coverage.out
