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

docs:
	pkgsite & sleep 5
	wget -r -N -q -p -k -E --regex-type pcre --accept-regex '^.*/(static|third_party|gitlab-vs.informatik.uni-ulm.de)/.*$$' http://localhost:8080/gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/ || true
	mv localhost:8080 docs
	mkdir -p docs/third_party/dialog-polyfill
	wget -q http://localhost:8080/third_party/dialog-polyfill/dialog-polyfill.js -O docs/third_party/dialog-polyfill/dialog-polyfill.js
	wget -q http://localhost:8080/static/frontend/frontend.js -O docs/static/frontend/frontend.js
	wget -q http://localhost:8080/static/frontend/unit/unit.js -O docs/static/frontend/unit/unit.js
	wget -q http://localhost:8080/static/frontend/unit/main/main.js -O docs/static/frontend/unit/main/main.js
	for i in `find docs -type f -name "*\?*"`; do mv $$i `echo $$i | cut -d? -f1`; done
	for i in `find docs -type f -name "*.html"`; do sed -i 's/%3F[a-z=]*\.css//' $$i; done
	pkill pkgsite
PHONY:docs

run: build
	out/main
.PHONY:run
