.PHONY: build test clean run install lint

# Переменные
BINARY_NAME=hsanalyzer
MAIN_PATH=./cmd/hsanalyzer
BUILD_DIR=./bin

# Сборка
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME).exe $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME).exe"

# Запуск
run:
	@go run $(MAIN_PATH)

# Установка зависимостей
install:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

# Тестирование
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

# Unit тесты
test-unit:
	@echo "Running unit tests..."
	@go test -v -race ./internal/...
	@echo "Unit tests complete"

# Integration тесты
test-integration:
	@echo "Running integration tests..."
	@go test -v -race ./tests/...
	@echo "Integration tests complete"

# Покрытие тестами
coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Линтинг
lint:
	@echo "Running linter..."
	@golangci-lint run ./...
	@echo "Linting complete"

# Форматирование
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Formatting complete"

# Очистка
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Сборка для разных платформ
build-all: build-windows build-linux build-mac

build-windows:
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

build-linux:
	@echo "Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

build-mac:
	@echo "Building for macOS..."
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)

# Помощь
help:
	@echo "Available targets:"
	@echo "  build            - Build the application"
	@echo "  run              - Run the application"
	@echo "  install          - Install dependencies"
	@echo "  test             - Run all tests"
	@echo "  test-unit        - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  coverage         - Generate test coverage report"
	@echo "  lint             - Run linter"
	@echo "  fmt              - Format code"
	@echo "  clean            - Clean build artifacts"
	@echo "  build-all        - Build for all platforms"
	@echo "  help             - Show this help message"
