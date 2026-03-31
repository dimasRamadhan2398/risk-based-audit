.PHONY: build run test clean docker-build docker-up docker-down docker-logs deps fmt lint

# Variables
APP_NAME=auth-service
CMD_DIR=./cmd/auth
CONFIG_FILE=./configs/config.yaml

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Build the application
build:
	$(GOBUILD) -o bin/$(APP_NAME) $(CMD_DIR)/main.go

# Run the application
run:
	$(GORUN) $(CMD_DIR)/main.go

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build files
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	$(GOFMT) -s -w .

# Run linter
lint:
	$(GOLINT) run ./...

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Docker cleanup
docker-clean:
	docker-compose down -v
	docker system prune -f

# Migrate database
migrate-up:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/rb_audit?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/rb_audit?sslmode=disable" down

# Create new migration
migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

# Install tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest

# Development
dev: deps fmt build
	./bin/$(APP_NAME)
