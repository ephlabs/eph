# Eph - Enhanced Makefile with better testing integration
.PHONY: build test test-ci test-integration test-watch coverage coverage-html clean lint run-daemon run-cli install-test-tools help

# Build configuration
VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X github.com/ephlabs/eph/pkg/version.Version=$(VERSION)"

# Default target
help:
	@echo "🌊 Eph - Ephemeral Environment Controller"
	@echo ""
	@echo "Available targets:"
	@echo "  build            - Build eph CLI and ephd daemon binaries"
	@echo "  test             - Run tests with enhanced output (development)"
	@echo "  test-ci          - Run tests exactly like CI (includes coverage and XML)"
	@echo "  test-integration - Run integration tests"
	@echo "  test-watch       - Run tests in watch mode for development"
	@echo "  coverage         - Check coverage thresholds"
	@echo "  coverage-html    - Generate HTML coverage report"
	@echo "  lint             - Run linter"
	@echo "  clean            - Clean build artifacts"
	@echo "  run-daemon       - Run ephd daemon"
	@echo "  run-cli          - Run eph CLI"
	@echo "  install-test-tools - Install enhanced testing tools"
	@echo ""
	@echo "Quick start:"
	@echo "  make install-test-tools  # Install testing tools"
	@echo "  make test               # Run tests with beautiful output"
	@echo "  make build              # Build binaries"

# Build targets
build:
	@echo "🔨 Building Eph binaries..."
	go build $(LDFLAGS) -o bin/eph ./cmd/eph
	go build $(LDFLAGS) -o bin/ephd ./cmd/ephd
	@echo "✅ Build complete!"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.txt coverage.html
	rm -f unit-tests.xml integration-tests.xml
	rm -f unit-tests.json integration-tests.json
	@echo "✅ Clean complete!"

# Development targets
run-daemon:
	@echo "🚀 Starting ephd daemon..."
	LOG_PRETTY=true LOG_LEVEL=debug go run $(LDFLAGS) ./cmd/ephd

run-cli:
	@echo "💻 Running eph CLI..."
	LOG_PRETTY=true LOG_LEVEL=debug go run $(LDFLAGS) ./cmd/eph

# Linting
lint:
	@echo "🔍 Running linter..."
	golangci-lint run ./...
	@echo "✅ Linting complete!"

# Testing tools installation
install-test-tools:
	@echo "📦 Installing enhanced testing tools..."
	@command -v gotestsum >/dev/null 2>&1 || { \
		echo "Installing gotestsum..."; \
		go install gotest.tools/gotestsum@latest; \
	}
	@command -v gotestfmt >/dev/null 2>&1 || { \
		echo "Installing gotestfmt..."; \
		go install github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest; \
	}
	@echo "✅ Testing tools installed!"

# Enhanced testing targets

# Run tests with beautiful output (development) - replaces your old test target
test: install-test-tools
	@echo "🧪 Running tests with enhanced output..."
	gotestsum --format=pkgname-and-test-fails -- -race -tags='!integration' ./...

# Run tests exactly like CI (includes coverage and XML output)
test-ci: install-test-tools
	@echo "🚀 Running tests in CI mode..."
	gotestsum \
		--format=pkgname-and-test-fails \
		--junitfile=unit-tests.xml \
		--jsonfile=unit-tests.json \
		-- \
		-race \
		-coverprofile=coverage.txt \
		-covermode=atomic \
		-coverpkg=./... \
		-tags='!integration' \
		./...

# Run integration tests
test-integration: install-test-tools
	@echo "🔧 Running integration tests..."
	gotestsum \
		--format=testname \
		--junitfile=integration-tests.xml \
		--jsonfile=integration-tests.json \
		-- \
		-race \
		-coverprofile=integration-coverage.out \
		-covermode=atomic \
		-tags=integration \
		./internal/server/...

# Watch mode for development
test-watch: install-test-tools
	@echo "👁️  Running tests in watch mode..."
	@echo "Press 'r' to rerun tests, 'q' to quit"
	gotestsum --watch --format=pkgname-and-test-fails -- -race -tags='!integration' ./...

# Check coverage thresholds (basic coverage check)
coverage: test-ci
	@echo "📊 Checking basic coverage..."
	go tool cover -func=coverage.txt | tail -1

# Generate HTML coverage report
coverage-html: test-ci
	@echo "🌐 Generating HTML coverage report..."
	go tool cover -html=coverage.txt -o coverage.html
	@echo "✅ Open coverage.html in your browser to view coverage details"

# Fallback to old test behavior if tools not available
test-basic:
	@echo "🧪 Running basic tests (fallback)..."
	go test ./... -v -cover

# Run the full test suite (like CI)
test-full: test-ci test-integration coverage
	@echo "✅ All tests completed successfully!"
	@echo ""
	@echo "📊 Test artifacts generated:"
	@echo "  - unit-tests.xml (JUnit format)"
	@echo "  - integration-tests.xml (JUnit format)"
	@echo "  - coverage.txt (Go coverage profile)"
	@echo "  - coverage.html (HTML coverage report)"

# Alias for backwards compatibility with old simple test
test-simple:
	go test ./... -v -cover
