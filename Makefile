include config.mk

.PHONY: install-deps
install-deps: ## Install development dependencies
	@echo "--->  Installing Dependencies"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: lint
lint: ## Run linter
	@echo "--->  Linting"
	@golangci-lint run -v

.PHONY: lint-fix
lint-fix: ## Auto-fix linting issues
	@echo "---> Lint-Fixing code"
	@golangci-lint run --fix

.PHONY: test
test: ## Run tests with coverage
	@echo "--->  Running tests"
	@go test -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

.PHONY: test-cov
test-cov: test ## Run tests and open coverage in browser
	@go tool cover -html=coverage.out

.PHONY: build
build: ## Build for current platform
	@echo "---> Building for $(GOOS)/$(GOARCH) with binary name $(BINARY_NAME)"
	@mkdir -p $(OUTPUT_DIR)
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w" -buildvcs=false -o $(OUTPUT_DIR)/$(BINARY_NAME) ./

.PHONY: build-linux
build-linux: ## Build for Linux (amd64)
	@$(MAKE) build GOOS=linux GOARCH=amd64

.PHONY: build-linux-arm64
build-linux-arm64: ## Build for Linux (arm64)
	@$(MAKE) build GOOS=linux GOARCH=arm64

.PHONY: build-macos
build-macos: ## Build for macOS (amd64)
	@$(MAKE) build GOOS=darwin GOARCH=amd64

.PHONY: build-macos-arm64
build-macos-arm64: ## Build for macOS (arm64, Apple Silicon)
	@$(MAKE) build GOOS=darwin GOARCH=arm64

.PHONY: build-windows
build-windows: ## Build for Windows (amd64)
	@$(MAKE) build GOOS=windows GOARCH=amd64 BINARY_NAME=$(BINARY_NAME).exe

.PHONY: build-all
build-all: ## Build for all platforms
	@$(MAKE) build-linux
	@$(MAKE) build-linux-arm64
	@$(MAKE) build-macos
	@$(MAKE) build-macos-arm64
	@$(MAKE) build-windows

.PHONY: clean
clean: ## Remove build artifacts
	@echo "--->  Cleaning"
	@rm -rf bin/
	@rm -f coverage.out

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sed 's/Makefile://' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
