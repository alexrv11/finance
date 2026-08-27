.PHONY: build run test tidy

build:
	go build -o bin/seed ./cmd/seed

run:
	go run ./cmd/seed

test:
	go test ./...

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

data-dir:
	mkdir -p data
