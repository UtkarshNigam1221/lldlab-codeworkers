.PHONY: build-sandbox build run dev test test-unit test-integration test-all coverage lint lint-fix fmt vet clean help

build-sandbox:
	docker build -t code-sandbox -f docker/Dockerfile.sandbox .

build:
	docker compose build

run: build-sandbox
	docker compose up

dev:
	go run ./cmd/worker

# Testing targets
test-unit:
	go test -v -short -race -coverprofile=coverage.out ./...

test-integration: build-sandbox
	go test -v -race -run Integration ./internal/api/...

test-all: test-unit test-integration

test: test-unit

coverage:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

coverage-report:
	go tool cover -func=coverage.out

# Linting and formatting targets
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run --timeout=5m

lint-fix:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run --fix --timeout=5m

fmt:
	go fmt ./...
	gofmt -s -w .

vet:
	go vet ./...

# Static analysis
staticcheck:
	@which staticcheck > /dev/null || (echo "Installing staticcheck..." && go install honnef.co/go/tools/cmd/staticcheck@latest)
	staticcheck ./...

# Security check
security:
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securego/gosec/v2/cmd/gosec@latest)
	gosec -quiet ./...

# CI target - runs all checks
ci: fmt vet lint test-unit

# API testing
test-api:
	chmod +x scripts/test_api.sh
	./scripts/test_api.sh

clean:
	docker compose down
	docker rmi code-sandbox codeworkers-worker -f
	rm -f codeworkers coverage.out coverage.html

help:
	@echo "Available targets:"
	@echo "  build-sandbox      - Build the code execution sandbox image"
	@echo "  build              - Build the worker service"
	@echo "  run                - Start the service with docker-compose"
	@echo "  dev                - Run locally for development"
	@echo ""
	@echo "Testing:"
	@echo "  test-unit          - Run unit tests"
	@echo "  test-integration   - Run integration tests (requires Docker)"
	@echo "  test-all           - Run all tests"
	@echo "  test               - Alias for test-unit"
	@echo "  coverage           - Generate test coverage report"
	@echo "  coverage-report    - Show coverage statistics"
	@echo "  test-api           - Run API integration tests"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint               - Run golangci-lint"
	@echo "  lint-fix           - Run golangci-lint with auto-fix"
	@echo "  fmt                - Format code with gofmt"
	@echo "  vet                - Run go vet"
	@echo "  staticcheck        - Run staticcheck"
	@echo "  security           - Run gosec security scanner"
	@echo "  ci                 - Run all CI checks (fmt, vet, lint, test)"
	@echo ""
	@echo "  clean              - Stop and remove containers/images"
