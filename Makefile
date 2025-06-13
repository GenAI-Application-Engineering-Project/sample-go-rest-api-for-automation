# Makefile for products

# Name of the binary executable
BINARY_NAME = products
BIN_DIR = bin

# Go environment variables
GO_CMD = go
GO_BUILD = $(GO_CMD) build
GO_RUN = $(GO_CMD) run
GO_TEST = $(GO_CMD) test

# Default target: build the CLI
all: build

# Build the CLI
build:
	@mkdir -p $(BIN_DIR)
	$(GO_BUILD) -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/server

# Clean up the build artifacts
clean:
	rm -f $(BINARY_NAME)

run: build
	@ $(BIN_DIR)/$(BINARY_NAME)

# Run unit tests (if you have any)
test:
	$(GO_TEST) -coverprofile=coverage.out ./...

# Generate and view test coverage report (HTML format)
test-rpt: test
	@go tool cover -html=coverage.out -o coverage.html
	@xdg-open coverage.html || open coverage.html

# Generate test coverage in a concise format (for CI)
ci-coverage: test
	@go tool cover -func=coverage.out
	@echo "Coverage report generated."

lint:
	golangci-lint run

lint-v:
	golangci-lint run -v

# Help message to describe the targets
help:
	@echo "Makefile commands:"
	@echo "  make            - Build the CLI"
	@echo "  make build      - Build the CLI binary"
	@echo "  make run        - Run the CLI"
	@echo "  make clean      - Clean up the build"
	@echo "  make test       - Run unit tests"
	@echo "  make test-rpt   - Generate and open the coverage report (HTML)"
	@echo "  make ci-coverage - Generate test coverage for CI (concise format)"
	@echo "  make help       - Show this help message"
