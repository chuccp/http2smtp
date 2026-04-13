release-linux:
	mkdir -p build_assets
	GOOS=linux GOARCH=amd64 go build  -v -o build_assets/http2smtp  ./

release-osx:
	mkdir -p build_assets
	GOOS=darwin GOARCH=amd64 go build  -v -o build_assets/http2smtp  ./