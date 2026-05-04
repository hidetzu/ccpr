VERSION ?= $(shell git describe --tags --always --dirty)

.PHONY: build build-mcp test vet lint clean

build:
	go build -ldflags "-X main.version=$(VERSION)" -o bin/ccpr ./cmd/ccpr

build-mcp:
	go build -o bin/ccpr-mcp ./cmd/ccpr-mcp

test:
	go test ./... -v -race

vet:
	go vet ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/
