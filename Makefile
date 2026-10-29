VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.Version=$(VERSION)

release-linux:
	mkdir -p build_assets
	GOOS=linux GOARCH=amd64 go build -tags release -v -ldflags "$(LDFLAGS)" -o build_assets/http2smtp  ./

release-osx:
	mkdir -p build_assets
	GOOS=darwin GOARCH=amd64 go build -tags release -v -ldflags "$(LDFLAGS)" -o build_assets/http2smtp  ./
