.DEFAULT_GOAL := build

format:
	go fmt ./...
	go vet ./...
.PHONY:format

build: cmd/plugin_loader.go format
	mkdir -p out
	go build -o out ./cmd/main.go ./cmd/plugin_loader.go
.PHONY:build

build-with-webui: cmd/plugin_loader.go format
	mkdir -p out
	go build -tags webui -o out ./cmd/main.go ./cmd/plugin_loader.go
.PHONY:build-with-webui

generate-generic-structs:
	mkdir -p pkg/message/generic
	quicktype res/schemas/GENERIC_*.json --src-lang schema -l golang --package genericmsg -o pkg/message/generic/message.go

generate-aiv-structs:
	mkdir -p pkg/message/aiv
	quicktype res/schemas/AIV_*.json --src-lang schema -l golang --package aivmsg -o pkg/message/aiv/message.go

generate-mbd-structs:
	mkdir -p pkg/message/mbd
	quicktype res/schemas/MBD_*.json --src-lang schema -l golang --package mbdmsg -o pkg/message/mbd/message.go

generate-tas-structs:
	mkdir -p pkg/message/tas
	quicktype res/schemas/TAS_*.json --src-lang schema -l golang --package tasmsg -o pkg/message/tas/message.go

generate-v2x-structs:
	mkdir -p pkg/message/v2x
	quicktype res/schemas/V2X_*.json --src-lang schema -l golang --package v2xmsg -o pkg/message/v2x/message.go

generate-tch-structs:
	mkdir -p pkg/message/tch
	quicktype res/schemas/TCH_*.json --src-lang schema -l golang --package tchmsg -o pkg/message/tch/message.go

generate-taqi-structs:
	mkdir -p pkg/message/taqi
	quicktype res/schemas/TAQI_*.json --src-lang schema -l golang --package taqimsg -o pkg/message/taqi/message.go

generate-structs: generate-aiv-structs generate-generic-structs generate-mbd-structs generate-tas-structs generate-v2x-structs generate-tch-structs generate-taqi-structs
.PHONY:generate-structs

remove-generic-structs:
	rm pkg/message/generic/message.go

remove-aiv-structs:
	rm pkg/message/aiv/message.go

remove-mbd-structs:
	rm pkg/message/mbd/message.go

remove-tas-structs:
	rm pkg/message/tas/message.go

remove-v2x-structs:
	rm pkg/message/v2x/message.go

remove-tch-structs:
	rm pkg/message/tch/message.go

clean-structs: remove-generic-structs remove-aiv-structs remove-mbd-structs remove-tas-structs remove-v2x-structs remove-tch-structs


check: format
	go test -race $(shell go list ./... | grep -v /vendor/)
.PHONY:check

bench: format
	go test -bench . -run=^$$ -benchmem $(shell go list ./... | grep -v /vendor/) 
.PHONY:bench

PLUGIN_FILES = $(shell find plugins/ -type f -name '*.go')

cmd/plugin_loader.go: $(PLUGIN_FILES)
	GOARCH= GOOS= go generate cmd/main.go

PROJECT_NAME = go-taf

run: build
	TAF_CONFIG=res/taf.json out/main
.PHONY:run

run-with-webui: build-with-webui
	TAF_CONFIG=res/taf.json out/main
.PHONY:run-with-webui
