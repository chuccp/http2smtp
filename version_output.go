package main

import (
	"fmt"
	"runtime"
)

func versionLine() string {
	return fmt.Sprintf("http2smtp version %s (%s/%s)", Version, runtime.GOOS, runtime.GOARCH)
}

func printVersion() {
	fmt.Println(versionLine())
}
