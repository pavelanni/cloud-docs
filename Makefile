# Cloud Docs Makefile
# Builds all CLI tools and server for multiple architectures

# Variables
DIST_DIR = dist
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS = -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

# Targets
TARGETS = server token iframe

# Platforms
PLATFORMS = \
	darwin/arm64 \
	linux/amd64

.PHONY: all clean build test fmt vet deps help

# Default target
all: clean build

# Help target
help:
	@echo "Available targets:"
	@echo "  all      - Clean and build all targets for all platforms"
	@echo "  build    - Build all targets for all platforms"
	@echo "  clean    - Remove dist directory"
	@echo "  test     - Run tests"
	@echo "  fmt      - Format code"
	@echo "  vet      - Vet code"
	@echo "  deps     - Download and tidy dependencies"
	@echo "  help     - Show this help message"

# Clean dist directory
clean:
	rm -rf $(DIST_DIR)

# Create dist directory structure
$(DIST_DIR):
	mkdir -p $(DIST_DIR)

# Build all targets for all platforms
build: $(DIST_DIR)
	@echo "Building all targets for all platforms..."
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		echo "Building for $$os/$$arch..."; \
		mkdir -p $(DIST_DIR)/$$os-$$arch; \
		for target in $(TARGETS); do \
			echo "  Building $$target for $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) \
				-o $(DIST_DIR)/$$os-$$arch/$$target \
				./cmd/$$target; \
		done; \
	done
	@echo "Build complete!"

# Individual platform targets
build-darwin-arm64: $(DIST_DIR)
	@echo "Building for macOS ARM64..."
	@mkdir -p $(DIST_DIR)/darwin-arm64
	@for target in $(TARGETS); do \
		echo "  Building $$target..."; \
		GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) \
			-o $(DIST_DIR)/darwin-arm64/$$target \
			./cmd/$$target; \
	done

build-linux-amd64: $(DIST_DIR)
	@echo "Building for Linux AMD64..."
	@mkdir -p $(DIST_DIR)/linux-amd64
	@for target in $(TARGETS); do \
		echo "  Building $$target..."; \
		GOOS=linux GOARCH=amd64 go build $(LDFLAGS) \
			-o $(DIST_DIR)/linux-amd64/$$target \
			./cmd/$$target; \
	done

# Development targets
test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

deps:
	go mod download
	go mod tidy

# List built binaries
list:
	@echo "Built binaries:"
	@find $(DIST_DIR) -type f -executable 2>/dev/null | sort || echo "No binaries found. Run 'make build' first."

# Archive binaries for distribution
archive: build
	@echo "Creating archives..."
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		archive_name="cloud-docs-$(VERSION)-$$os-$$arch.tar.gz"; \
		echo "Creating $$archive_name..."; \
		tar -czf $(DIST_DIR)/$$archive_name -C $(DIST_DIR)/$$os-$$arch .; \
	done
	@echo "Archives created in $(DIST_DIR)/"