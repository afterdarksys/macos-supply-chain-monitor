.PHONY: build install clean test run

# Build variables
BINARY_NAME=scm
BUILD_DIR=build
INSTALL_DIR=/usr/local/bin

# Go build flags
LDFLAGS=-ldflags "-s -w"

build:
	@echo "🔨 Building ${BINARY_NAME}..."
	@mkdir -p ${BUILD_DIR}
	@go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} .
	@echo "✅ Build complete: ${BUILD_DIR}/${BINARY_NAME}"

install: build
	@echo "📦 Installing ${BINARY_NAME} to ${INSTALL_DIR}..."
	@cp ${BUILD_DIR}/${BINARY_NAME} ${INSTALL_DIR}/
	@echo "✅ Installed successfully"
	@echo "💡 Run 'scm --help' to get started"

clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf ${BUILD_DIR}
	@echo "✅ Clean complete"

test:
	@echo "🧪 Running tests..."
	@go test -v ./...

run: build
	@echo "🚀 Running ${BINARY_NAME}..."
	@./${BUILD_DIR}/${BINARY_NAME}

deps:
	@echo "📥 Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies ready"

fmt:
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@echo "✅ Formatting complete"

vet:
	@echo "🔍 Vetting code..."
	@go vet ./...
	@echo "✅ Vet complete"

lint: fmt vet
	@echo "✅ Linting complete"

dev: clean build
	@./${BUILD_DIR}/${BINARY_NAME} --help

# Initialize config directory and sample config
init-config:
	@mkdir -p ~/.config/scm
	@echo "📝 Creating sample config..."
	@cat config.example.yaml > ~/.config/scm/config.yaml
	@echo "✅ Config created at ~/.config/scm/config.yaml"

help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  install      - Install to ${INSTALL_DIR}"
	@echo "  clean        - Remove build artifacts"
	@echo "  test         - Run tests"
	@echo "  run          - Build and run"
	@echo "  deps         - Download dependencies"
	@echo "  fmt          - Format code"
	@echo "  vet          - Vet code"
	@echo "  lint         - Run fmt and vet"
	@echo "  init-config  - Create sample config"
