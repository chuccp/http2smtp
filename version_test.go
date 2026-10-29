package main

import (
	"runtime"
	"strings"
	"testing"
)

func TestVersionLine(t *testing.T) {
	previous := Version
	Version = "v1.2.3-test"
	t.Cleanup(func() { Version = previous })

	got := versionLine()
	if !strings.Contains(got, "http2smtp version v1.2.3-test") {
		t.Fatalf("version line = %q", got)
	}
	platform := "(" + runtime.GOOS + "/" + runtime.GOARCH + ")"
	if !strings.Contains(got, platform) {
		t.Fatalf("version line %q does not contain platform %q", got, platform)
	}
}
