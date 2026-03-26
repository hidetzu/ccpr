VERSION ?= $(shell git describe --tags --always --dirty)

.PHONY: build test vet lint clean

build:
	go build -ldflags "-X main.version=$(VERSION)" -o bin/ccpr ./cmd/ccpr

test:
	go test ./... -v -race

vet:
	go vet ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/
