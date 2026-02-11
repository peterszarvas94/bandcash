.PHONY: dev build run test clean fmt vet lint docker docker-run

# Development with hot reload
dev:
	air

# Build the binary
build:
	go build -o bin/server ./cmd/server

# Run the binary
run: build
	./bin/server

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/ tmp/ logs/

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Lint (requires golangci-lint)
lint:
	golangci-lint run

# All checks
check: fmt vet test

# Docker build
docker-build:
	docker build -t stargazing .

# Docker run
docker-run:docker-build
	docker run -p 8080:8080 stargazing
