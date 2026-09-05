.PHONY: build test vet fmt install clean

build:
	go build -o bin/yuiop ./cmd/yuiop

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

install:
	go install ./cmd/yuiop

clean:
	rm -rf bin