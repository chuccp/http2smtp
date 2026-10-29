package main

import (
	"fmt"
	"runtime"
)

// Version is overwritten from the release tag at build time. Local development
// builds keep the default so they cannot be mistaken for a published release.
var Version = "dev"

func versionLine() string {
	return fmt.Sprintf("http2smtp version %s (%s/%s)", Version, runtime.GOOS, runtime.GOARCH)
}

func printVersion() {
	fmt.Println(versionLine())
}
