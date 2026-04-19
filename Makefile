.PHONY: build test tidy

build:
	go build -o warehouse .

test:
	go test ./...

tidy:
	go mod tidy
