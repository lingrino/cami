.PHONY: build fmt install test

build:
	go build

fmt:
	gofmt -l -w -s .

install:
	go install

test:
	go test -cover -race -v ./...
