# stackctl Makefile
# Common development and release tasks

.PHONY: help build test clean install release changelog version

# Default target
help: ## Show this help message
	@echo "stackctl - Development Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build commands
build: ## Build the binary
	@echo "Building stackctl..."
	go build -o bin/stackctl ./cmd/stackctl
	@echo "✓ Binary built at bin/stackctl"

build-dev: ## Build with version info for development
	@echo "Building stackctl with dev version..."
	go build -ldflags "\
		-X main.version=dev-$(shell git rev-parse --short HEAD) \
		-X main.commit=$(shell git rev-parse HEAD) \
		-X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ) \
		-X main.builtBy=make" \
		-o bin/stackctl ./cmd/stackctl
	@echo "✓ Built bin/stackctl (dev build)"

# Testing commands
test: ## Run all tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

# Installation commands
install: build ## Install to /usr/local/bin
	@echo "Installing stackctl to /usr/local/bin..."
	sudo cp bin/stackctl /usr/local/bin/stackctl
	@echo "✓ Installed. Run 'stackctl version' to verify"

install-dev: build-dev ## Install dev build to /usr/local/bin
	@echo "Installing dev build to /usr/local/bin..."
	sudo cp bin/stackctl /usr/local/bin/stackctl
	@echo "✓ Installed dev build"

# Cleanup
clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/ dist/ coverage.out coverage.html
	@echo "✓ Clean complete"

# Release management
version: ## Show current version and next suggested version
	@echo "Current version information:"
	@echo ""
	@echo "Latest git tag:"
	@git describe --tags --abbrev=0 2>/dev/null || echo "  (no tags yet)"
	@echo ""
	@echo "CHANGELOG.md status:"
	@if grep -q "## \[Unreleased\]" CHANGELOG.md; then \
		echo "  ✓ Unreleased section exists"; \
		echo ""; \
		echo "Changes in Unreleased:"; \
		sed -n '/## \[Unreleased\]/,/## \[/p' CHANGELOG.md | grep -E '^###' || echo "  (no changes documented yet)"; \
	else \
		echo "  ⚠ No Unreleased section found"; \
	fi
	@echo ""
	@echo "Commits since last tag:"
	@git log $(shell git describe --tags --abbrev=0 2>/dev/null || echo "HEAD")..HEAD --oneline 2>/dev/null || echo "  (no commits)"
	@echo ""
	@echo "Suggested next version:"
	@./scripts/suggest-version.sh 2>/dev/null || echo "  Run 'make setup-scripts' first"

changelog-prepare: ## Prepare CHANGELOG.md for a new release
	@echo "Preparing changelog for release..."
	@./scripts/changelog-prepare.sh

changelog-add: ## Add entry to CHANGELOG.md Unreleased section
	@./scripts/changelog-add.sh

setup-scripts: ## Set up release helper scripts
	@mkdir -p scripts
	@echo "Creating release helper scripts..."
	@chmod +x scripts/*.sh 2>/dev/null || true
	@echo "✓ Scripts ready in scripts/"

# GoReleaser integration
release-check: ## Check release configuration
	@echo "Checking GoReleaser configuration..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser check; \
	else \
		echo "⚠ GoReleaser not installed. Install: https://goreleaser.com/install/"; \
	fi

release-snapshot: ## Build a local snapshot release (no publish)
	@echo "Building snapshot release..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --clean --skip=publish; \
		echo "✓ Snapshot built in dist/"; \
	else \
		echo "⚠️  GoReleaser not installed (optional)"; \
		echo ""; \
		echo "Building with standard Go build instead..."; \
		mkdir -p dist; \
		GOOS=linux GOARCH=amd64 go build -o dist/stackctl_linux_amd64 ./cmd/stackctl; \
		GOOS=linux GOARCH=arm64 go build -o dist/stackctl_linux_arm64 ./cmd/stackctl; \
		echo "✓ Binaries built in dist/"; \
		echo ""; \
		echo "To get full release features (archives, checksums), install GoReleaser:"; \
		echo "  macOS: brew install goreleaser"; \
		echo "  Linux: https://goreleaser.com/install/"; \
	fi

release-dry-run: ## Dry run of release process
	@echo "Performing release dry run..."
	@./scripts/release-dry-run.sh

# GitHub Actions will use this
release: ## Create a release (CI only - use 'git tag vX.Y.Z && git push --tags')
	@echo "⚠ Releases should be created via git tags:"
	@echo ""
	@echo "  1. Update CHANGELOG.md with version and date"
	@echo "  2. git add CHANGELOG.md && git commit -m 'chore: prepare release vX.Y.Z'"
	@echo "  3. git tag -a vX.Y.Z -m 'Release vX.Y.Z'"
	@echo "  4. git push && git push --tags"
	@echo ""
	@echo "GitHub Actions will automatically build and publish the release."

# Development helpers
tidy: ## Run go mod tidy
	@echo "Tidying go modules..."
	go mod tidy
	@echo "✓ Modules tidied"

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	@echo "✓ Code formatted"

lint: ## Run linters (requires golangci-lint)
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠ golangci-lint not installed. Install: https://golangci-lint.run/usage/install/"; \
	fi

# Quick development cycle
dev: fmt build-dev install-dev ## Format, build, and install in one command
	@echo "✓ Ready for testing!"

# Show all targets (including ones without descriptions)
list: ## List all make targets
	@LC_ALL=C $(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'
