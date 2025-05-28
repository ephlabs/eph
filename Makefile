.PHONY: build test clean lint

VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X github.com/ephlabs/eph/pkg/version.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/eph ./cmd/eph
	go build $(LDFLAGS) -o bin/ephd ./cmd/ephd

test:
	go test ./... -v -cover

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

run-daemon:
	go run ./cmd/ephd

run-cli:
	go run ./cmd/eph
