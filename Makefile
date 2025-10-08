.PHONY: build server cli qrgen install clean test run

# Build all binaries
build: server cli qrgen

# Build server binary
server:
	@echo "Building server..."
	@mkdir -p bin
	@go build -o bin/sourdough-server ./cmd/server

# Build CLI binary
cli:
	@echo "Building CLI..."
	@mkdir -p bin
	@go build -o bin/sourdough ./cmd/sourdough

# Build QR code generator
qrgen:
	@echo "Building QR generator..."
	@mkdir -p bin
	@go build -o bin/qrgen ./cmd/qrgen

# Install CLI to system path
install: cli
	@echo "Installing sourdough CLI..."
	@sudo ln -sf $(PWD)/bin/sourdough /usr/local/bin/sourdough

# Install systemd service
install-service: server
	@echo "Installing systemd service..."
	@sudo cp sourdough.service /etc/systemd/system/
	@sudo systemctl daemon-reload
	@sudo systemctl enable sourdough
	@sudo systemctl start sourdough
	@echo "Service installed and started!"
	@echo "Check status with: sudo systemctl status sourdough"

# Uninstall systemd service
uninstall-service:
	@echo "Uninstalling systemd service..."
	@sudo systemctl stop sourdough || true
	@sudo systemctl disable sourdough || true
	@sudo rm -f /etc/systemd/system/sourdough.service
	@sudo systemctl daemon-reload

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf qrcodes/

# Run server in foreground (for testing)
run: server
	@echo "Starting server..."
	@./bin/sourdough-server

# Run unit tests
test:
	@echo "Running unit tests..."
	@go test -v ./...

# Run unit tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Generate test data
test-data:
	@echo "Generating test data..."
	@./test/generate_test_data.sh

# Run integration tests (requires server to be running)
test-integration:
	@echo "Running integration tests..."
	@./test/integration_test.sh

# Run all tests (unit + integration)
test-all: test
	@echo ""
	@echo "Running integration tests..."
	@./test/integration_test.sh

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Show help
help:
	@echo "Sourdough Build Commands:"
	@echo "  make build            - Build all binaries"
	@echo "  make server           - Build server binary"
	@echo "  make cli              - Build CLI binary"
	@echo "  make qrgen            - Build QR generator"
	@echo "  make install          - Install CLI to /usr/local/bin"
	@echo "  make install-service  - Install and start systemd service"
	@echo "  make uninstall-service - Stop and remove systemd service"
	@echo "  make run              - Run server in foreground"
	@echo "  make clean            - Remove build artifacts"
	@echo "  make test             - Run unit tests"
	@echo "  make test-coverage    - Run tests with coverage report"
	@echo "  make test-data        - Generate test dataset"
	@echo "  make test-integration - Run integration tests"
	@echo "  make test-all         - Run all tests (unit + integration)"
	@echo "  make deps             - Download dependencies"
