-include .env

# Variables
# Priority: 1. Command line, 2. .env file, 3. Directory name
PROJECT_NAME ?= $(shell basename $(CURDIR))
VERSION ?= $(shell cat VERSION 2>/dev/null || echo "0.0.0")
BINARY_NAME ?= $(PROJECT_NAME)
MAIN_PATH := ./cmd/main.go
BIN_DIR := bin
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# Colors for output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[36m

# =============================================================================
# Build & Run
# =============================================================================

.PHONY: build
build: ## Build the application binary
	@echo "$(COLOR_GREEN)Building $(BINARY_NAME)...$(COLOR_RESET)"
	@mkdir -p $(BIN_DIR)
	go build -ldflags="-X main.Version=$(VERSION)" -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_PATH)

.PHONY: run
run: ## Run the application
	@echo "$(COLOR_GREEN)Running $(BINARY_NAME)...$(COLOR_RESET)"
	go run $(MAIN_PATH)

# =============================================================================
# Dependencies
# =============================================================================

.PHONY: deps
deps: ## Download and verify dependencies
	@echo "$(COLOR_GREEN)Downloading dependencies...$(COLOR_RESET)"
	go mod download
	go mod verify

.PHONY: tidy
tidy: ## Clean up go.mod and go.sum
	@echo "$(COLOR_GREEN)Tidying go.mod...$(COLOR_RESET)"
	go mod tidy

.PHONY: update-dependencies
update-dependencies: ## Update all dependencies to latest versions
	@echo "$(COLOR_YELLOW)Updating dependencies...$(COLOR_RESET)"
	go get -u ./...
	go mod tidy

# =============================================================================
# Code Quality
# =============================================================================

.PHONY: fmt
fmt: ## Format code with gofmt and goimports
	@echo "$(COLOR_GREEN)Formatting code...$(COLOR_RESET)"
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "$(COLOR_YELLOW)goimports not installed, skipping...$(COLOR_RESET)"; \
	fi

.PHONY: vet
vet: ## Run go vet
	@echo "$(COLOR_GREEN)Running go vet...$(COLOR_RESET)"
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	@echo "$(COLOR_GREEN)Running golangci-lint...$(COLOR_RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "$(COLOR_YELLOW)golangci-lint not installed. Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(COLOR_RESET)"; \
	fi

.PHONY: staticcheck
staticcheck: ## Run staticcheck
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "$(COLOR_YELLOW)staticcheck not installed. Install: go install honnef.co/go/tools/cmd/staticcheck@latest$(COLOR_RESET)"; \
	fi

# =============================================================================
# Testing
# =============================================================================

.PHONY: test
test: ## Run all tests with coverage
	@echo "$(COLOR_GREEN)Running tests...$(COLOR_RESET)"
	go test ./... -v -race -coverprofile=$(COVERAGE_FILE)

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "$(COLOR_GREEN)Running unit tests...$(COLOR_RESET)"
	go test -v -race -short ./...

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(COLOR_GREEN)Running integration tests...$(COLOR_RESET)"
	go test -v -race -tags=integration ./...

.PHONY: test-coverage
test-coverage: test ## Generate and open coverage report
	@echo "$(COLOR_GREEN)Generating coverage report...$(COLOR_RESET)"
	go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(COLOR_BLUE)Coverage report: $(COVERAGE_HTML)$(COLOR_RESET)"

.PHONY: test-coverage-func
test-coverage-func: test ## Show coverage per function
	go tool cover -func=$(COVERAGE_FILE)

.PHONY: bench
bench: ## Run benchmarks
	@echo "$(COLOR_GREEN)Running benchmarks...$(COLOR_RESET)"
	go test -bench=. -benchmem ./...

# =============================================================================
# Mocks
# =============================================================================

.PHONY: generate-mocks
generate-mocks: ## Generate mocks using mockery
	@echo "$(COLOR_GREEN)Generating mocks...$(COLOR_RESET)"
	mockery

# =============================================================================
# Security
# =============================================================================

.PHONY: vuln-check
vuln-check: ## Check for known vulnerabilities
	@echo "$(COLOR_GREEN)Checking for vulnerabilities...$(COLOR_RESET)"
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "$(COLOR_YELLOW)govulncheck not installed. Install: go install golang.org/x/vuln/cmd/govulncheck@latest$(COLOR_RESET)"; \
	fi

.PHONY: sec-scan
sec-scan: ## Run security scanner (gosec)
	@echo "$(COLOR_GREEN)Running security scan...$(COLOR_RESET)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -fmt=json -out=gosec-report.json ./...; \
		gosec ./...; \
	else \
		echo "$(COLOR_YELLOW)gosec not installed. Install: go install github.com/securego/gosec/v2/cmd/gosec@latest$(COLOR_RESET)"; \
	fi

# =============================================================================
# Docker
# =============================================================================

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(COLOR_GREEN)Building Docker image...$(COLOR_RESET)"
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest

.PHONY: docker-run
docker-run: ## Run Docker container
	docker run --rm -p 8080:8080 --env-file .env $(BINARY_NAME):$(VERSION)

.PHONY: docker-push
docker-push: ## Push Docker image to registry
	docker push $(BINARY_NAME):$(VERSION)
	docker push $(BINARY_NAME):latest

# =============================================================================
# Profiling
# =============================================================================

.PHONY: profile-cpu
profile-cpu: ## Run CPU profiling
	@echo "$(COLOR_GREEN)Running CPU profiling...$(COLOR_RESET)"
	go test -cpuprofile=cpu.prof -bench=. ./...
	go tool pprof -http=:8081 cpu.prof

.PHONY: profile-mem
profile-mem: ## Run memory profiling
	@echo "$(COLOR_GREEN)Running memory profiling...$(COLOR_RESET)"
	go test -memprofile=mem.prof -bench=. ./...
	go tool pprof -http=:8081 mem.prof

.PHONY: profile-trace
profile-trace: ## Run execution trace
	go test -trace=trace.out -bench=. ./...
	go tool trace trace.out

# =============================================================================
# Git & Hooks
# =============================================================================

.PHONY: pre-commit
pre-commit: fmt lint vet test-unit ## Run pre-commit checks

.PHONY: install-hooks
install-hooks: ## Install git pre-commit hooks
	@echo "$(COLOR_GREEN)Installing git hooks...$(COLOR_RESET)"
	@mkdir -p .git/hooks
	@echo '#!/bin/sh\nmake pre-commit' > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "$(COLOR_BLUE)Git hooks installed!$(COLOR_RESET)"

# =============================================================================
# CI/CD
# =============================================================================

.PHONY: ci
ci: deps fmt lint vet test vuln-check ## Run full CI pipeline

.PHONY: ci-local
ci-local: clean ci ## Run CI pipeline locally with clean state

# =============================================================================
# Utilities
# =============================================================================

.PHONY: clean
clean: ## Clean build artifacts and caches
	@echo "$(COLOR_GREEN)Cleaning...$(COLOR_RESET)"
	rm -rf $(BIN_DIR)/
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	rm -f *.prof *.out
	rm -f gosec-report.json
	go clean -cache -testcache

.PHONY: clean-all
clean-all: clean ## Clean everything including global module cache
	@echo "$(COLOR_YELLOW)Cleaning global module cache...$(COLOR_RESET)"
	go clean -modcache

.PHONY: todo
todo: ## Show TODO and FIXME comments in code
	@grep -rnw . -e 'TODO' -e 'FIXME' --include="*.go" || echo "No TODOs found!"

.PHONY: check-updates
check-updates: ## Check for outdated dependencies
	@echo "$(COLOR_GREEN)Checking for outdated dependencies...$(COLOR_RESET)"
	@if command -v go-mod-outdated >/dev/null 2>&1; then \
		go list -u -m -json all | go-mod-outdated -update -direct; \
	else \
		echo "$(COLOR_YELLOW)go-mod-outdated not installed. Install: go install github.com/psampaz/go-mod-outdated@latest$(COLOR_RESET)"; \
		go list -u -m all; \
	fi

.PHONY: version
version: ## Show current version
	@echo "$(COLOR_BLUE)Version: $(VERSION)$(COLOR_RESET)"

.PHONY: bump-version
bump-version: ## Bump version (use TYPE=major|minor|patch)
	@if [ -z "$(TYPE)" ]; then \
		echo "$(COLOR_YELLOW)Usage: make bump-version TYPE=major|minor|patch$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_GREEN)Bumping $(TYPE) version...$(COLOR_RESET)"
	@NEW_VERSION=$$(echo $(VERSION) | awk -F. -v type=$(TYPE) '{ \
		if (type == "major") printf "%d.0.0", $$1+1; \
		else if (type == "minor") printf "%d.%d.0", $$1, $$2+1; \
		else if (type == "patch") printf "%d.%d.%d", $$1, $$2, $$3+1; \
	}'); \
	echo $$NEW_VERSION > VERSION; \
	echo "$(COLOR_BLUE)Version updated: $(VERSION) -> $$NEW_VERSION$(COLOR_RESET)"

# =============================================================================
# Help
# =============================================================================

.PHONY: help
help: ## Show this help message
	@printf "\033[1m%s - Available targets:\033[0m\n\n" "$(PROJECT_NAME)"
	@awk 'BEGIN {FS = ":.*?## "; category = ""} \
		/^# =+$$/ {getline; if ($$0 ~ /^# /) {gsub(/^# /, "", $$0); gsub(/ *$$/, "", $$0); category = $$0}} \
		/^[a-zA-Z_-]+:.*?## / { \
			if (category != last_category) { \
				if (last_category != "") printf "\n"; \
				printf "\033[1;33m%s:\033[0m\n", category; \
				last_category = category \
			} \
			printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 \
		}' $(MAKEFILE_LIST)
	@echo ""

.DEFAULT_GOAL := help
