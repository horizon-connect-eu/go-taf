.DEFAULT_GOAL := build

format:
	go fmt ./...
	go vet ./...

.PHONY:format

build: format
	go build ./cmd/main.go
.PHONY:build

check: format
	go test -race $(shell go list ./... | grep -v /vendor/)
.PHONY:check

bench: format
	go test -bench . -run=^$$ -benchmem $(shell go list ./... | grep -v /vendor/) 
.PHONY:bench

run: format
	go run ./...
.PHONY:run
