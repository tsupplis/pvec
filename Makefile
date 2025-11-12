.PHONY: all build test test-verbose test-coverage clean install lint fmt help docker docker-build docker-run test-client analyze

# Binary name
BINARY_NAME=pvec
BUILD_DIR=./bin
CMD_DIR=.
DOCKER_IMAGE=pvec
DOCKER_TAG=latest
EXAMPLE_DIR=./examples/test-client

# Build flags
LDFLAGS=-ldflags="-s -w"

all: clean fmt lint test build ## Run fmt, lint, test, and build

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

install: ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) $(CMD_DIR)

test: ## Run tests
	@echo "Running tests..."
	go test ./pkg/... -race -coverprofile=coverage.out

test-verbose: ## Run tests with verbose output
	@echo "Running tests (verbose)..."
	go test ./pkg/... -v -race -coverprofile=coverage.out

test-coverage: test ## Run tests and show coverage
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Run linter (requires golangci-lint)
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install it from https://golangci-lint.run/usage/install/" && exit 1)
	@echo "Running linter..."
	golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

clean: ## Remove build artifacts
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	rm -f $(BINARY_NAME)

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: ## Run application in Docker container
	@echo "Running Docker container..."
	docker run --rm -it \
		-v $(HOME)/.pvecrc:/root/.pvecrc:ro \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker: docker-build ## Build and show Docker info

test-client: ## Build and run the test client example
	@echo "Building test client..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/test-client $(EXAMPLE_DIR)
	@echo "Test client built: $(BUILD_DIR)/test-client"
	@echo ""
	@echo "Running test client..."
	@$(BUILD_DIR)/test-client

analyze: ## Run code analysis and generate metrics report
	@echo "Running code analysis..."
	@which python3 > /dev/null || (echo "python3 not found" && exit 1)
	python3 scripts/analyze_code.py
	@echo ""
	@echo "✅ Analysis complete! Report saved to docs/code_analysis.md"
	@echo ""
	@echo "📊 Quick Summary:"
	@grep "Total Lines" docs/code_analysis.md | head -1 || true
	@grep "Go Files" docs/code_analysis.md | head -1 || true
	@grep "Packages" docs/code_analysis.md | head -1 || true
