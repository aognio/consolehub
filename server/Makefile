# ConsoleHub Server Makefile

BINARY_NAME ?= consolehub-server
BUILD_DIR ?= bin
INSTALL_DIR ?= /usr/local/bin

.PHONY: all build install test clean clean-data run

VERSION ?= v0.1.1
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS = -ldflags "-X consolehub/internal/version.Version=$(VERSION) -X consolehub/internal/version.GitCommit=$(GIT_COMMIT) -X consolehub/internal/version.BuildDate=$(BUILD_DATE)"

all: build

build:
	@echo "Building server binary $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/consolehub

install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	install -m 0755 $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)

test:
	@echo "Running server unit and integration tests..."
	go test -v ./...

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Clean build binaries only (preserves pb_data)
clean:
	@echo "Cleaning server build artifacts..."
	rm -rf $(BUILD_DIR) $(BINARY_NAME) coverage.out

# Clean local test database data (explicit wipe)
clean-data: clean
	@echo "Wiping local PocketBase database..."
	rm -rf pb_data
