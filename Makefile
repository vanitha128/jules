.PHONY: help build run test test-unit test-integration lint fmt docker-up docker-down docker-build docker-logs air

# Variables
APP_NAME=go-moon-app
BUILD_DIR=build
MAIN_GO_PATH=cmd/server/main.go
INTEGRATION_TEST_DIR=tests/integration

# Go commands
GO=go
GO_BUILD=$(GO) build
GO_RUN=$(GO) run
GO_TEST=$(GO) test
GO_FMT=$(GO) fmt
GO_LIST=$(GO) list

# Docker commands
DOCKER_COMPOSE=docker-compose

# Linters & Dev tools
# Ensure these are installed. See comments below for installation instructions.
GOLANGCI_LINT=golangci-lint
AIR=air

help:
	@echo "Available targets:"
	@echo "  build             - Build the Go application"
	@echo "  run               - Run the Go application (after building)"
	@echo "  run-dev           - Run the Go application directly with go run (for quick dev)"
	@echo "  test              - Run all tests (unit and integration)"
	@echo "  test-unit         - Run unit tests"
	@echo "  test-integration  - Run integration tests (ensure Docker services are up if needed)"
	@echo "  lint              - Lint the Go code"
	@echo "  fmt               - Format the Go code"
	@echo "  docker-up         - Start services using Docker Compose (detached mode)"
	@echo "  docker-down       - Stop services using Docker Compose"
	@echo "  docker-build      - Build Docker images"
	@echo "  docker-logs       - Show logs from Docker Compose services (app service by default)"
	@echo "  air               - Run the application with live reload using Air"
	@echo "  install-tools     - Install golangci-lint and air (if not already installed)"

# Build the Go application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO_BUILD) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_GO_PATH)
	@echo "$(APP_NAME) built in $(BUILD_DIR)/"

# Run the Go application (after building)
run: build
	@echo "Running $(APP_NAME)..."
	@./$(BUILD_DIR)/$(APP_NAME)

# Run the Go application directly (for quick dev)
run-dev:
	@echo "Running $(APP_NAME) with go run..."
	$(GO_RUN) $(MAIN_GO_PATH)

# Run all tests
test: test-unit test-integration
	@echo "All tests completed."

# Run unit tests (exclude integration tests)
# This uses a simple grep; a more robust method might involve build tags.
test-unit:
	@echo "Running unit tests..."
	$(GO_TEST) -v -cover $(shell $(GO_LIST) ./... | grep -v /tests/integration)

# Run integration tests
# Note: These tests might require Docker Compose services to be running.
test-integration:
	@echo "Running integration tests..."
	@echo "Make sure dependent services (DB, Redis) are running, e.g., via 'make docker-up'."
	cd $(INTEGRATION_TEST_DIR) && $(GO_TEST) -v .
	# Alternatively, to run from root: $(GO_TEST) -v ./tests/integration/...


# Lint the Go code
# To install golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
lint:
	@echo "Linting code..."
	$(GOLANGCI_LINT) run ./...

# Format the Go code
fmt:
	@echo "Formatting code..."
	$(GO_FMT) ./...

# Docker Compose targets
docker-up:
	@echo "Starting Docker services in detached mode..."
	$(DOCKER_COMPOSE) up -d --remove-orphans
	@echo "Services started. Use 'make docker-logs' to see app logs."

docker-down:
	@echo "Stopping Docker services..."
	$(DOCKER_COMPOSE) down
	@echo "To remove volumes, run 'docker-compose down -v'."

docker-build:
	@echo "Building Docker images..."
	$(DOCKER_COMPOSE) build

docker-logs:
	@echo "Showing logs for 'app' service (Ctrl+C to stop)..."
	$(DOCKER_COMPOSE) logs -f app

# Run with Air (live reload)
# To install Air: go install github.com/cosmtrek/air@latest
# Ensure .air.toml is configured.
air:
	@echo "Starting with Air (live reload)..."
	$(AIR)

# Install developer tools
install-tools:
	@echo "Installing/Updating golangci-lint and air..."
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install github.com/cosmtrek/air@latest
	@echo "Tools installation/update attempt complete. Ensure they are in your PATH."

# Default target (optional, can be 'help' or 'build')
default: help
