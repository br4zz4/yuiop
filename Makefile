.PHONY: build test vet fmt install clean

VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)

build:
	go build -ldflags "-X github.com/br4zz4/yuiop/internal/cli.Version=$(VERSION)" -o bin/yuiop ./cmd/yuiop

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

install:
	go install -ldflags "-X github.com/br4zz4/yuiop/internal/cli.Version=$(VERSION)" ./cmd/yuiop

clean:
	rm -rf bin