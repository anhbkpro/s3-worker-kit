# S3 Worker Kit - Development Makefile

.PHONY: help build test test-integration test-unit clean setup dev-up dev-down dev-logs bucket-create bucket-delete run run-server run-cli fmt lint deps tidy

# Default target
help: ## Show this help message
	@echo "S3 Worker Kit - Development Commands"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build the application binary
	@echo "🔨 Building application..."
	go build -o worker cmd/worker/main.go

# Test targets
test: ## Run all tests
	@echo "🧪 Running all tests..."
	go test ./...

test-unit: ## Run unit tests only
	@echo "🧪 Running unit tests..."
	go test -short ./...

test-integration: ## Run integration tests (requires LocalStack)
	@echo "🧪 Running integration tests..."
	@echo "⚠️  Make sure LocalStack is running: make dev-up"
	@export AWS_ENDPOINT=http://localhost:4566 && \
	export AWS_REGION=us-east-1 && \
	export AWS_ACCESS_KEY_ID=test && \
	export AWS_SECRET_ACCESS_KEY=test && \
	go test -tags=integration ./...

# Development environment targets
setup: ## Setup development environment (install dependencies, create bucket)
	@echo "🚀 Setting up development environment..."
	@make deps
	@make dev-up
	@sleep 5
	@make bucket-create

dev-up: ## Start development services (LocalStack, Tempo, OTLP Collector, Grafana)
	@echo "🐳 Starting development services..."
	docker-compose up -d
	@echo "⏳ Waiting for services to be ready..."
	@sleep 10
	@echo "✅ Services started!"
	@echo "  - LocalStack S3: http://localhost:4566"
	@echo "  - Tempo: http://localhost:3200"
	@echo "  - Grafana: http://localhost:3000 (admin/admin)"
	@echo "  - OTLP gRPC: localhost:4317"
	@echo "  - OTLP HTTP: localhost:4318"

dev-down: ## Stop development services
	@echo "🛑 Stopping development services..."
	docker-compose down

dev-logs: ## Show logs from development services
	docker-compose logs -f

dev-status: ## Show status of development services
	docker-compose ps

# Bucket management
bucket-create: ## Create test bucket in LocalStack
	@echo "🪣 Creating test bucket..."
	@export AWS_ENDPOINT=http://localhost:4566 && \
	export AWS_REGION=us-east-1 && \
	export AWS_ACCESS_KEY_ID=test && \
	export AWS_SECRET_ACCESS_KEY=test && \
	aws --endpoint-url=http://localhost:4566 s3 mb s3://test-bucket 2>/dev/null || echo "Bucket already exists or LocalStack not running"

bucket-delete: ## Delete test bucket from LocalStack
	@echo "🗑️  Deleting test bucket..."
	@export AWS_ENDPOINT=http://localhost:4566 && \
	export AWS_REGION=us-east-1 && \
	export AWS_ACCESS_KEY_ID=test && \
	export AWS_SECRET_ACCESS_KEY=test && \
	aws --endpoint-url=http://localhost:4566 s3 rb s3://test-bucket --force 2>/dev/null || echo "Bucket doesn't exist or LocalStack not running"

bucket-list: ## List objects in test bucket
	@echo "📋 Listing test bucket contents..."
	@export AWS_ENDPOINT=http://localhost:4566 && \
	export AWS_REGION=us-east-1 && \
	export AWS_ACCESS_KEY_ID=test && \
	export AWS_SECRET_ACCESS_KEY=test && \
	aws --endpoint-url=http://localhost:4566 s3 ls s3://test-bucket/ || echo "Bucket empty or LocalStack not running"

# Run targets
run: run-server ## Run the application (alias for run-server)

run-server: ## Run the application in server mode
	@echo "🚀 Starting server..."
	@export AWS_ENDPOINT=http://localhost:4566 && \
	export AWS_REGION=us-east-1 && \
	export AWS_ACCESS_KEY_ID=test && \
	export AWS_SECRET_ACCESS_KEY=test && \
	export WORKER_POOL_SIZE=5 && \
	export OTEL_ENABLED=true && \
	export OTEL_EXPORTER=otlp && \
	export OTEL_ENDPOINT=localhost:4317 && \
	export OTEL_INSECURE=true && \
	./worker -mode=server -port=4000

run-cli: ## Run the application in CLI mode (upload single file)
	@echo "📤 Running CLI upload..."
	@export AWS_ENDPOINT=http://localhost:4566 && \
	export AWS_REGION=us-east-1 && \
	export AWS_ACCESS_KEY_ID=test && \
	export AWS_SECRET_ACCESS_KEY=test && \
	export WORKER_POOL_SIZE=5 && \
	export OTEL_ENABLED=true && \
	export OTEL_EXPORTER=otlp && \
	export OTEL_ENDPOINT=localhost:4317 && \
	export OTEL_INSECURE=true && \
	echo '{"bucket":"test-bucket","key":"test-file.txt","data":"SGVsbG8gV29ybGQ="}' | ./worker -mode=cli

# Code quality targets
fmt: ## Format Go code
	@echo "🎨 Formatting code..."
	go fmt ./...

lint: ## Run linter (requires golangci-lint)
	@echo "🔍 Running linter..."
	golangci-lint run

# Dependency management
deps: ## Download dependencies
	@echo "📦 Downloading dependencies..."
	go mod download

tidy: ## Clean up go.mod and go.sum
	@echo "🧹 Tidying dependencies..."
	go mod tidy

# Cleanup
clean: ## Clean build artifacts and test cache
	@echo "🧽 Cleaning up..."
	go clean
	go clean -testcache
	rm -f worker
	@echo "✅ Cleanup complete!"

# Development workflow targets
dev: ## Full development setup (setup + build + test)
	@echo "🚀 Setting up full development environment..."
	@make setup
	@make build
	@make test-unit
	@echo "✅ Development environment ready!"

# Integration test workflow
test-all: ## Run full test suite (unit + integration)
	@echo "🧪 Running full test suite..."
	@make test-unit
	@make dev-up
	@sleep 5
	@make bucket-create
	@make test-integration
	@echo "✅ All tests passed!"

# Quick start for new developers
quickstart: ## Quick start guide for new developers
	@echo "🚀 Quick Start Guide"
	@echo ""
	@echo "1. Setup environment:"
	@echo "   make setup"
	@echo ""
	@echo "2. Build application:"
	@echo "   make build"
	@echo ""
	@echo "3. Run tests:"
	@echo "   make test"
	@echo ""
	@echo "4. Start server:"
	@echo "   make run-server"
	@echo ""
	@echo "5. Test API (in another terminal):"
	@echo "   curl -X POST -H 'Content-Type: application/json' \\"
	@echo "        -d '{\"tasks\":[{\"bucket\":\"test-bucket\",\"key\":\"hello.txt\",\"data\":\"SGVsbG8gV29ybGQ=\"}]}' \\"
	@echo "        http://localhost:4000/api/v1/upload/tasks"
	@echo ""
	@echo "6. View traces in Grafana:"
	@echo "   http://localhost:3000 (admin/admin)"
	@echo ""
	@echo "Available commands:"
	@make help
