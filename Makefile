.PHONY: build test vet lint clean

build:
	go build -o bin/ccpr ./cmd/ccpr

test:
	go test ./... -v -race

vet:
	go vet ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/
