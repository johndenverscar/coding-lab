.PHONY: help build run test test-unit test-integration docker-build deploy clean fmt vet lint deps

# Variables
APP_NAME := qbert
DOCKER_IMAGE := $(APP_NAME):latest
DEPLOYMENT_DIR := deploy
TEST_DIR := test

# Default target
.DEFAULT_GOAL := help

## help: Display this help message
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build the Go binary
build:
	@echo "Building $(APP_NAME)..."
	CGO_ENABLED=0 go build -o bin/server cmd/server/main.go
	@echo "✓ Build complete: bin/server"

## run: Run the server locally
run: build
	@echo "Running $(APP_NAME)..."
	./bin/server

## test: Run all tests
test: test-unit test-integration

## test-unit: Run unit tests
test-unit:
	@echo "Running unit tests..."
	go test -v -race -coverprofile=coverage.out ./pkg/...
	@echo "✓ Unit tests complete"

## test-integration: Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v ./$(TEST_DIR)/

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE)..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "✓ Docker image built: $(DOCKER_IMAGE)"

## deploy: Deploy to Kubernetes cluster
deploy: docker-build
	@echo "Deploying to Kubernetes..."
	kubectl config use-context colima
	kubectl apply -f $(DEPLOYMENT_DIR)/
	kubectl wait --for=condition=available deployment/$(APP_NAME) --timeout=60s
	@echo "✓ Deployment complete"

## redeploy: Rebuild and redeploy to Kubernetes
redeploy: docker-build
	@echo "Redeploying to Kubernetes..."
	kubectl config use-context colima
	kubectl rollout restart deployment/$(APP_NAME)
	kubectl wait --for=condition=available deployment/$(APP_NAME) --timeout=60s
	@echo "✓ Redeployment complete"

## status: Check deployment status
status:
	@echo "Checking deployment status..."
	kubectl get pods -l app=$(APP_NAME)
	kubectl get svc $(APP_NAME)

## logs: Show application logs
logs:
	kubectl logs -l app=$(APP_NAME) --tail=100 -f

## clean: Clean up build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf bin/
	rm -f coverage.out
	@echo "✓ Clean complete"

## clean-deploy: Delete Kubernetes resources
clean-deploy:
	@echo "Deleting Kubernetes resources..."
	kubectl delete -f $(DEPLOYMENT_DIR)/ --ignore-not-found=true
	@echo "✓ Resources deleted"

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "✓ Format complete"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "✓ Vet complete"

## lint: Run golangci-lint (requires golangci-lint to be installed)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
		echo "✓ Lint complete"; \
	else \
		echo "⚠ golangci-lint not installed. Install with: brew install golangci-lint"; \
	fi

## deps: Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "✓ Dependencies updated"

## check: Run fmt, vet, and tests
check: fmt vet test-unit
	@echo "✓ All checks passed"

## dev: Quick development cycle - format, build, and run
dev: fmt build run

## coverage: Run tests with coverage report
coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./pkg/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"
