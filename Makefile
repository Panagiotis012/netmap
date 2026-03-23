.PHONY: build run test dev frontend clean

BINARY=netmap
VERSION=0.1.0

frontend:
	cd web && npm ci && npm run build
	rm -rf cmd/netmap/dist && cp -r web/dist cmd/netmap/dist

build: frontend
	CGO_ENABLED=1 go build -ldflags "-X main.version=$(VERSION)" -o bin/$(BINARY) ./cmd/netmap

run: build
	./bin/$(BINARY)

test:
	go test ./... -v -race

dev:
	go run ./cmd/netmap

clean:
	rm -rf bin/ cmd/netmap/dist web/dist
