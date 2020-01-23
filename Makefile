.PHONY: build docs fmt install test

build:
	go build

docs:
	go run docs/generate.go

fmt:
	gofmt -l -w -s .

install:
	go install

test:
	go test -cover -race -v ./...
