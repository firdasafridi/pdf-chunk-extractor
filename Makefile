# PDF Chunk Extractor Makefile

.PHONY: help build run clean setup check-deps test

# Default target
help:
	@echo "PDF Chunk Extractor - Available commands:"
	@echo ""
	@echo "  build      - Build the application"
	@echo "  run        - Run the application"
	@echo "  clean      - Clean build artifacts"
	@echo "  setup      - Setup project dependencies"
	@echo "  check-deps - Check if all dependencies are installed"
	@echo "  test       - Run tests"
	@echo "  quick-start - Setup and run in one command"
	@echo ""

# Build the application
build:
	@echo "🔨 Building PDF Chunk Extractor..."
	go build -o pdf-chunk-extractor main.go
	@echo "✅ Build complete! Run with: ./pdf-chunk-extractor"

# Run the application
run:
	@echo "🚀 Running PDF Chunk Extractor..."
	go run main.go

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f pdf-chunk-extractor
	@echo "✅ Clean complete!"

# Setup project dependencies
setup:
	@echo "📦 Setting up project dependencies..."
	go mod tidy
	@echo "✅ Dependencies updated!"

# Check if all dependencies are installed
check-deps:
	@echo "🔍 Checking dependencies..."
	@echo "Checking Go..."
	@which go > /dev/null || (echo "❌ Go is not installed" && exit 1)
	@echo "✅ Go is installed"
	@echo "Checking Tesseract..."
	@which tesseract > /dev/null || (echo "❌ Tesseract is not installed" && exit 1)
	@echo "✅ Tesseract is installed"
	@echo "Checking OpenAI API Key..."
	@if [ -z "$$OPENAI_API_KEY" ]; then echo "❌ OPENAI_API_KEY environment variable is not set"; exit 1; fi
	@echo "✅ OpenAI API Key is set"
	@echo "✅ All dependencies are ready!"

# Run tests
test:
	@echo "🧪 Running tests..."
	go test ./...
	@echo "✅ Tests complete!"

# Quick start - setup and run
quick-start: setup check-deps run

# Create necessary directories
init-dirs:
	@echo "📁 Creating necessary directories..."
	mkdir -p data output chunk
	@echo "✅ Directories created!"

# Show project status
status:
	@echo "📊 Project Status:"
	@echo "  Go version: $(shell go version)"
	@echo "  Tesseract: $(shell which tesseract 2>/dev/null || echo 'Not found')"
	@echo "  OpenAI API Key: $(if $(OPENAI_API_KEY),Set,Not set)"
	@echo "  Data files: $(shell ls data/*.pdf 2>/dev/null | wc -l | tr -d ' ') PDF files found"
	@echo "  Output files: $(shell ls output/*.txt 2>/dev/null | wc -l | tr -d ' ') text files generated"
	@echo "  Chunk files: $(shell find chunk -name "*.txt" 2>/dev/null | wc -l | tr -d ' ') chunks created"

# Install Tesseract (macOS)
install-tesseract-mac:
	@echo "📦 Installing Tesseract on macOS..."
	brew install tesseract tesseract-lang
	@echo "✅ Tesseract installed!"

# Install Tesseract (Ubuntu/Debian)
install-tesseract-ubuntu:
	@echo "📦 Installing Tesseract on Ubuntu/Debian..."
	sudo apt update
	sudo apt install -y tesseract-ocr tesseract-ocr-ind
	@echo "✅ Tesseract installed!"

# Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted!"

# Vet code
vet:
	@echo "🔍 Vetting code..."
	go vet ./...
	@echo "✅ Code vetted!"

# Lint code (requires golangci-lint)
lint:
	@echo "🔍 Linting code..."
	@which golangci-lint > /dev/null || (echo "❌ golangci-lint is not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run
	@echo "✅ Code linted!"

# Development build with race detection
build-dev:
	@echo "🔨 Building development version with race detection..."
	go build -race -o pdf-chunk-extractor-dev main.go
	@echo "✅ Development build complete!"

# Release build
build-release:
	@echo "🔨 Building release version..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o pdf-chunk-extractor-linux-amd64 main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o pdf-chunk-extractor-darwin-amd64 main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o pdf-chunk-extractor-darwin-arm64 main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o pdf-chunk-extractor-windows-amd64.exe main.go
	@echo "✅ Release builds complete!"

# Clean all output files
clean-output:
	@echo "🧹 Cleaning output files..."
	rm -rf output/* chunk/*
	@echo "✅ Output files cleaned!"

# Show help for environment setup
env-help:
	@echo "🔧 Environment Setup Help:"
	@echo ""
	@echo "1. Set your OpenAI API key:"
	@echo "   export OPENAI_API_KEY='your-api-key-here'"
	@echo ""
	@echo "2. Add to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
	@echo "   echo 'export OPENAI_API_KEY=\"your-api-key-here\"' >> ~/.bashrc"
	@echo ""
	@echo "3. Or create a .env file (not supported by default, use a tool like godotenv)"
	@echo ""
	@echo "4. Verify setup:"
	@echo "   make check-deps" 