.DEFAULT_GOAL := build

format:
	go fmt ./...
	go vet ./...
.PHONY:format

build: cmd/plugin_loader.go format
	mkdir -p out
	go build -o out ./cmd/main.go ./cmd/plugin_loader.go
.PHONY:build

check: format
	go test -race $(shell go list ./... | grep -v /vendor/)
.PHONY:check

bench: format
	go test -bench . -run=^$$ -benchmem $(shell go list ./... | grep -v /vendor/) 
.PHONY:bench

PLUGIN_FILES = $(shell find plugins/ -type f -name '*.go')

cmd/plugin_loader.go: $(PLUGIN_FILES)
	go generate cmd/main.go

run: build
	TAF_CONFIG=res/taf.json out/main
.PHONY:run
