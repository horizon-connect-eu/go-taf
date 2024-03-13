.DEFAULT_GOAL := build

format:
	go fmt ./...
	go vet ./...
.PHONY:format

buildplugins:
	mkdir -p plugins/bin
	go build -buildmode=plugin -o plugins/bin ./plugins/tam
.PHONY:buildplugins

build: format buildplugins
	mkdir -p out
	go build -o out ./cmd/main.go
.PHONY:build

check: format
	go test -race $(shell go list ./... | grep -v /vendor/)
.PHONY:check

bench: format
	go test -bench . -run=^$$ -benchmem $(shell go list ./... | grep -v /vendor/) 
.PHONY:bench

run: format buildplugins
	TAF_CONFIG=res/taf.json go run ./cmd/
.PHONY:run
