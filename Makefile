.PHONY: all build install uninstall clean distclean test test-coverage lint format vendor deps tidy help deps-list deps-indirect

# Default target
all: help

# Compiler settings
GO?=go
GOBUILD=$(GO) build
GOTEST=$(GO) test
GOCLEAN=$(GO) clean
GOFMT=$(GO)fmt

# Application binary name
BINARY_NAME=skulker
BINARY_PATH=/usr/local/bin/$(BINARY_NAME)

# Project directory
PROJECT_NAME=skulker

# --- Release & Packaging Variables ---
VERSION ?= v0.1.0
# Strip the 'v' for the RPM package versioning (e.g., v1.0.0 -> 1.0.0)
VERSION_CLEAN = $(patsubst v%,%,$(VERSION))
RPM_DIR = $(PWD)/build/rpm
RELEASE_DIR = $(PWD)/build/release

# Build raw binaries for common architectures
build-release: clean
	@echo "Building raw binaries for $(VERSION)..."
	mkdir -p $(RELEASE_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(RELEASE_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(RELEASE_DIR)/$(BINARY_NAME)-linux-arm64 main.go
	@echo "Raw binaries built in $(RELEASE_DIR)/"

# Build the RPM package locally
rpm: clean
	@echo "Building RPM for version $(VERSION_CLEAN)..."
	mkdir -p $(RPM_DIR)/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
	# Create source tarball from current git HEAD
	git archive --format=tar.gz --prefix=$(PROJECT_NAME)-$(VERSION_CLEAN)/ \
		-o $(RPM_DIR)/SOURCES/$(PROJECT_NAME)-$(VERSION_CLEAN).tar.gz HEAD
	# Build RPM using isolated workspace
	rpmbuild -ba \
		--define "_topdir $(RPM_DIR)" \
		--define "pkg_version $(VERSION_CLEAN)" \
		packaging/skulker.spec
	@echo "RPM built successfully in $(RPM_DIR)/RPMS/"

# The Grand Release Target: Tag, Push, Build Artifacts, and Release
release: build-release rpm
	@echo "Creating release $(VERSION)..."
	git tag $(VERSION)
	git push origin $(VERSION)
	gh release create $(VERSION) \
		$(RELEASE_DIR)/$(BINARY_NAME)-linux-amd64 \
		$(RELEASE_DIR)/$(BINARY_NAME)-linux-arm64 \
		$$(find $(RPM_DIR)/RPMS -name "*.rpm") \
		--title "$(VERSION)" \
		--generate-notes
	@echo "Release $(VERSION) created and all artifacts uploaded successfully."
	
release:
	@echo "Creating release $(VERSION)..."
	git tag $(VERSION)
	git push origin @(VERSION)
	gh release create $(VERSION) --title "$(VERSION)" --generate-notes
	@echo "Release $(VERSION) created successfully"


# Build: Compile the Go project
build:
	$(GOBUILD) -o $(BINARY_NAME) main.go

# Install: Copy the binary to system path (requires sudo)
install: build
	@echo "Installing $(BINARY_NAME) to $(BINARY_PATH)..."
	sudo install -D -m 755 $(BINARY_NAME) $(BINARY_PATH)
	@echo "Installation complete!"

# Uninstall: Remove the binary from system path
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@echo "This will remove $(BINARY_PATH)..."
	@echo -n "Are you sure? [y/N] "
	read response
	if [ "$$response" = "y" ]; then \
		sudo rm -f $(BINARY_PATH); \
		echo "Uninstalled!"; \
	else \
		echo "Uninstall cancelled."; \
	fi

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Full clean (build artifacts, binary, and cache)
distclean: clean
	rm -f coverage.out coverage.html
	rm -rf ~/pkg/mod/*
	rm -rf build/
	rm -rf vendor/

# Test: Run all tests
test:
	$(GOTEST) -v ./.

# Test with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint: Run linters (golangci-lint if available, otherwise vet)
lint:
	@command -v golangci-lint > /dev/null 2>&1 && \
		golangci-lint run || \
		$(GO) vet ./.

# Format: Format Go code
format:
	$(GOFMT) -w .

# Vendor: Download and cache dependencies
vendor:
	$(GO) mod vendor

# Download dependencies
deps:
	$(GO) mod download

# Tidy: Clean up module cache and unused dependencies
tidy:
	$(GO) mod tidy

# List all dependencies
deps-list:
	@echo "== All Dependencies =="
	$(GO) list -deps -m=false ./.

# List indirect dependencies  
deps-indirect:
	@echo "== Indirect Dependencies =="
	$(GO) list -deps -m=false ./... | grep -v "^skulker"

# Help: Show available targets
help:
	@echo "Skulker - Makefile Targets"
	@echo ""
	@echo "Basic Targets:"
	@echo "  make              Show this help message"
	@echo "  make build        Build the application"
	@echo "  make install      Build and install the application (requires sudo)"
	@echo "  make uninstall    Uninstall the application"
	@echo "  make clean        Clean build artifacts"
	@echo "  make distclean    Full clean (including cache)"
	@echo ""
	@echo "Testing:"
	@echo "  make test             Run tests"
	@echo "  make test-coverage    Run tests with coverage report"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint     Run linters (golangci-lint or go vet)"
	@echo "  make format   Format Go code"
	@echo "  make tidy     Clean up dependencies"
	@echo ""
	@echo "Dependencies:"
	@echo "  make deps      Download dependencies"
	@echo "  make vendor    Create vendor directory"
	@echo "  make deps-list List all dependencies"
	@echo "  make deps-indirect  List indirect dependencies"
	@echo ""
	@echo "Variables:"
	@echo "  PROJECT_NAME - Project name (default: skulker)"
	@echo "  BINARY_NAME  - Binary name (default: skulker)"
	@echo "  BINARY_PATH  - Install path (default: /usr/local/bin/skulker)"
	@echo "  GO           - Go command (default: go)"
	@echo ""