.PHONY: build run test dev clean

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

clean:
	rm -rf bin/
