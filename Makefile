# ConsoleHub Monorepo Root Makefile

.PHONY: all build install test clean run-server run-demo-go

# Default target builds all server binaries and libraries
all: build

build:
	@echo "=== Building ConsoleHub Server ==="
	$(MAKE) -C server build
	@echo "=== Building Go Demo Agent ==="
	@cd demos/go-demo && go build -o ../../bin/demo-agent main.go

install:
	@echo "=== Installing ConsoleHub Server ==="
	$(MAKE) -C server install

test:
	@echo "=== Testing ConsoleHub Server ==="
	$(MAKE) -C server test
	@echo "=== Testing Go Client Library ==="
	@cd libraries/go/consolehub && go test -v ./...

run-server:
	$(MAKE) -C server run

run-demo-go: build
	./bin/demo-agent

clean:
	@echo "=== Cleaning Monorepo Build Artifacts ==="
	$(MAKE) -C server clean
	rm -rf bin/
