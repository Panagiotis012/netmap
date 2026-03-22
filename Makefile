.PHONY: build run test dev

BINARY=netmap
VERSION=0.1.0

build:
	go build -ldflags "-X main.version=$(VERSION)" -o bin/$(BINARY) ./cmd/netmap

run: build
	./bin/$(BINARY)

test:
	go test ./... -v -race

dev:
	go run ./cmd/netmap
