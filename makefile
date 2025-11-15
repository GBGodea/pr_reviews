.PHONY: build run test clean docker-up docker-down help

build:
	go build -o pr-service ./cmd/main.go

run:
	@echo "Make sure PostgreSQL is running on localhost:5432"
	go run ./cmd/main.go

test:
	go test -v ./...

clean:
	rm -f pr-service

docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f app

docker-restart:
	docker-compose restart

setup:
	go mod download
	go mod tidy

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

check: fmt lint test

help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-up     - Start containers"
	@echo "  docker-down   - Stop containers"
	@echo "  docker-logs   - View app logs"
	@echo "  docker-restart- Restart containers"
	@echo "  setup         - Setup dependencies"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  check         - Run all checks"