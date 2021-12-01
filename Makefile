fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code
	go vet ./...

staticcheck:
	staticcheck ./...

test: fmt vet staticcheck ## Run tests.
	go test ./... -race -coverprofile=coverage.out -covermode=atomic

coverage: test
	@echo "test coverage: $(shell go tool cover -func coverage.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}')"

htmlcov:
	go tool cover -html=coverage.out

cmddocs:
	go run cmd/docs/main.go
	
# complete path to iter8 binary
ITER8_BIN ?= /usr/local/bin/iter8
build:
	cd clibase && go build -o $(ITER8_BIN)
