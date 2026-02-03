# Auth service — build/lint/test
.PHONY: run lint test build
run:
	go run .
lint:
	golangci-lint run ./...
test:
	go test ./...
build:
	CGO_ENABLED=0 go build -o bin/auth .
