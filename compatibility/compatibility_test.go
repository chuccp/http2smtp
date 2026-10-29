package compatibility

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/chuccp/go-web-frame/config"
)

func TestCheckMarksFreshInstall(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.ini")
	if err := os.WriteFile(configPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := loadConfig(t, configPath)

	err := Check(CheckOptions{
		Config:       cfg,
		ConfigPath:   configPath,
		SQLitePath:   filepath.Join(dir, "d-mail.db"),
		StorageRoot:  dir,
		DatabaseType: "sqlite",
	})
	if err != nil {
		t.Fatalf("fresh install rejected: %v", err)
	}
	if got := cfg.GetString(DataVersionKey); got != strconv.Itoa(CurrentDataVersion) {
		t.Fatalf("data version = %q, want %d", got, CurrentDataVersion)
	}

	// The marker must survive a reload; otherwise every restart would look like a fresh
	// install until setup happens to write the config again.
	reloaded := loadConfig(t, configPath)
	if err := Check(CheckOptions{
		Config:       reloaded,
		ConfigPath:   configPath,
		SQLitePath:   filepath.Join(dir, "d-mail.db"),
		StorageRoot:  dir,
		DatabaseType: "sqlite",
	}); err != nil {
		t.Fatalf("versioned install rejected: %v", err)
	}
}

func TestCheckRejectsUnversionedInitializedInstall(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.ini")
	databasePath := filepath.Join(dir, "d-mail.db")
	const original = "[core]\ninit=true\ndbtype=sqlite\n[sqlite]\nfilename=d-mail.db\n"
	if err := os.WriteFile(configPath, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(databasePath, []byte("old sqlite data"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Check(CheckOptions{
		Config:       loadConfig(t, configPath),
		ConfigPath:   configPath,
		SQLitePath:   databasePath,
		StorageRoot:  dir,
		DatabaseType: "sqlite",
	})
	if err == nil {
		t.Fatal("unversioned initialized install should be rejected")
	}
	if !IsIncompatible(err) {
		t.Fatalf("error type = %T, want *IncompatibleError", err)
	}
	message := err.Error()
	for _, want := range []string{"数据不兼容", "备份", "重置", "config.ini", databasePath, "移出程序目录"} {
		if !strings.Contains(message, want) {
			t.Errorf("message does not contain %q:\n%s", want, message)
		}
	}

	// A refusal is read-only. In particular, do not stamp a legacy config and make the
	// next start look compatible.
	data, readErr := os.ReadFile(configPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(data) != original {
		t.Fatalf("legacy config changed during refusal:\n%s", data)
	}
	databaseData, readErr := os.ReadFile(databasePath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(databaseData) != "old sqlite data" {
		t.Fatalf("legacy database changed during refusal:\n%s", databaseData)
	}
}

func TestCheckRejectsOldConfigLayoutBeforeSetup(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.ini")
	raw := "[core]\ninit=false\n[sqlite]\nfilename=d-mail.db\n[manage]\nport=12566\nusername=\npassword=\nwebPath=web\n"
	if err := os.WriteFile(configPath, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Check(CheckOptions{
		Config:       loadConfig(t, configPath),
		ConfigPath:   configPath,
		SQLitePath:   filepath.Join(dir, "d-mail.db"),
		StorageRoot:  dir,
		DatabaseType: "sqlite",
	})
	if err == nil || !IsIncompatible(err) {
		t.Fatalf("old config layout should be incompatible, got %v", err)
	}
}

func TestCheckRejectsDatabaseWithoutMarker(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.ini")
	databasePath := filepath.Join(dir, "existing.db")
	if err := os.WriteFile(configPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(databasePath, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Check(CheckOptions{
		Config:       loadConfig(t, configPath),
		ConfigPath:   configPath,
		SQLitePath:   databasePath,
		StorageRoot:  dir,
		DatabaseType: "sqlite",
	})
	if err == nil || !IsIncompatible(err) {
		t.Fatalf("database without marker should be incompatible, got %v", err)
	}
}

func TestMySQLRefusalRecommendsDatabaseBackup(t *testing.T) {
	cfg, err := config.NewFromBytes([]byte(
		"[core]\ninit=true\ndbtype=mysql\n[mysql]\nhost=db.example\nport=3306\ndbname=http2smtp\n",
	), "ini")
	if err != nil {
		t.Fatal(err)
	}
	err = Check(CheckOptions{
		Config:        cfg,
		ConfigPath:    "config.ini",
		DatabaseType:  "mysql",
		MySQLDatabase: "http2smtp",
	})
	if err == nil || !IsIncompatible(err) {
		t.Fatalf("unversioned MySQL install should be incompatible, got %v", err)
	}
	for _, want := range []string{"MySQL 数据库", "完整数据库备份", "重建一个空的 MySQL 数据库", "config.ini"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("message does not contain %q:\n%s", want, err)
		}
	}
}

func TestCheckRejectsMismatchedDataVersion(t *testing.T) {
	for _, version := range []int{-1, 0, CurrentDataVersion + 1} {
		t.Run(strconv.Itoa(version), func(t *testing.T) {
			cfg, err := config.NewFromBytes(
				[]byte("[core]\ndataVersion="+strconv.Itoa(version)+"\n"),
				"ini",
			)
			if err != nil {
				t.Fatal(err)
			}
			err = Check(CheckOptions{
				Config:       cfg,
				ConfigPath:   "config.ini",
				DatabaseType: "sqlite",
			})
			if err == nil || !IsIncompatible(err) {
				t.Fatalf("version %d should be incompatible, got %v", version, err)
			}
		})
	}
}

func TestCheckRejectsInvalidDataVersion(t *testing.T) {
	cfg, err := config.NewFromBytes([]byte("[core]\ndataVersion=legacy\n"), "ini")
	if err != nil {
		t.Fatal(err)
	}
	err = Check(CheckOptions{Config: cfg, ConfigPath: "config.ini", DatabaseType: "sqlite"})
	if err == nil || !IsIncompatible(err) {
		t.Fatalf("invalid data version should be incompatible, got %v", err)
	}
}

func loadConfig(t *testing.T, path string) *config.SingleFileConfig {
	t.Helper()
	cfg, err := config.LoadSingleFileConfig(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	return cfg
}
