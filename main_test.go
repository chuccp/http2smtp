package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chuccp/go-web-frame/config"
)

func TestConfigureFileLoggingUsesStorageRoot(t *testing.T) {
	root := t.TempDir()
	cfg := config.NewConfig()

	if err := configureFileLogging(cfg, root); err != nil {
		t.Fatalf("configure file logging: %v", err)
	}
	if got := cfg.GetString("web.log.path"); got != filepath.Join(root, "logs", "http2smtp.log") {
		t.Errorf("log path = %q", got)
	}
	if !cfg.GetBoolOrDefault("web.log.write", false) {
		t.Error("file logging was not enabled")
	}
	info, err := os.Stat(filepath.Join(root, "logs"))
	if err != nil || !info.IsDir() {
		t.Fatalf("log directory was not created: %v", err)
	}
}

func TestConfigureFileLoggingKeepsDefaultStdoutOnly(t *testing.T) {
	cfg := config.NewConfig()
	if err := configureFileLogging(cfg, "./"); err != nil {
		t.Fatal(err)
	}
	if cfg.GetBoolOrDefault("web.log.write", false) {
		t.Error("the default local storage root should keep stdout-only logging")
	}
}
