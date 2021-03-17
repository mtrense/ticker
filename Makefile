
# Makefile for ticker

VERSION ?= 0.0.0
BINARY_NAME ?= ticker

build:

test:

archive:

release:

build-prerequisites: generate-rpc
	mkdir -p bin dist

release-prerequisites:

test-prerequisites:

install-tools:

### BUILD ###################################################################

generate-rpc:
	protoc -Irpc --go_out=rpc --go_opt=paths=source_relative --go-grpc_out=rpc --go-grpc_opt=paths=source_relative rpc/admin.proto
	protoc -Irpc --go_out=rpc --go_opt=paths=source_relative --go-grpc_out=rpc --go-grpc_opt=paths=source_relative rpc/event_stream.proto

build-ticker: build-prerequisites
	go build \
		-ldflags "-X main.version=${VERSION} -X main.commit=$$(git rev-parse --short HEAD 2>/dev/null || echo \"none\")" \
		-o bin/$(OUTPUT_DIR)$(BINARY_NAME) \
		-tags memory,postgres\
		cli/main.go
build-ticker-linux_amd64: build-prerequisites
	$(MAKE) GOOS=linux GOARCH=amd64 OUTPUT_DIR=linux_amd64/ build
build-ticker-darwin_amd64: build-prerequisites
	$(MAKE) GOOS=darwin GOARCH=amd64 OUTPUT_DIR=darwin_amd64/ build
build-ticker-windows_amd64: build-prerequisites
	$(MAKE) GOOS=windows GOARCH=amd64 OUTPUT_DIR=windows_amd64/ build

build-linux_amd64: build-ticker-linux_amd64
build-darwin_amd64: build-ticker-darwin_amd64
build-windows_amd64: build-ticker-windows_amd64

build: build-ticker
build-all: build-linux_amd64 build-darwin_amd64 build-windows_amd64

### ARCHIVE #################################################################

archive-ticker-linux_amd64: build-ticker-linux_amd64
	tar czf dist/$(BINARY_NAME)-${VERSION}-linux_amd64.tar.gz -C bin/linux_amd64/ .
archive-ticker-darwin_amd64: build-ticker-darwin_amd64
	tar czf dist/$(BINARY_NAME)-${VERSION}-darwin_amd64.tar.gz -C bin/darwin_amd64/ .
archive-ticker-windows_amd64: build-ticker-windows_amd64
	tar czf dist/$(BINARY_NAME)-${VERSION}-windows_amd64.tar.gz -C bin/windows_amd64/ .

archive-linux_amd64: archive-ticker-linux_amd64
archive-darwin_amd64: archive-ticker-darwin_amd64
archive-windows_amd64: archive-ticker-windows_amd64

archive: archive-linux_amd64 archive-darwin_amd64 archive-windows_amd64

release: archive
	sha1sum dist/*.tar.gz > dist/$(BINARY_NAME)-${VERSION}.shasums

### TEST ####################################################################

test-ticker:
	ginkgo
test-ticker-watch:
	ginkgo watch
test: test-ticker
.PHONY: test-ticker
.PHONY: test

clean:
	rm -r bin/* dist/*

### DATABASE ################################################################

db-up:
	psql < db/up.sql

db-down:
	psql < db/down.sql

db-seed:
	psql < db/seed.sql

