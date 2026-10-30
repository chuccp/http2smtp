package configapply

import (
	"context"
	"testing"

	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/http2smtp/model"
)

func newConfigContext(t *testing.T) *core.Context {
	t.Helper()
	return core.NewContext(config.NewConfig(), context.Background())
}

// Saving the settings form while one database type is selected must not touch the
// stored settings of the other: both sections arrive filled with model.DefaultConfig()
// values the form never asked about, and writing the inactive section would replace its
// stored host/dbname/port with those defaults — exactly the wipe the empty-field rule
// in putDbSettings is meant to prevent.
func TestPutDbSettingsKeepsTheInactiveSection(t *testing.T) {
	t.Run("sqlite save keeps the mysql section", func(t *testing.T) {
		ctx := newConfigContext(t)
		cfg := ctx.GetConfig()
		cfg.Put("core.dbtype", "mysql")
		cfg.Put("mysql.host", "db.example.com")
		cfg.Put("mysql.port", 3307)
		cfg.Put("mysql.dbname", "prod")
		cfg.Put("mysql.username", "app")
		cfg.Put("mysql.charset", "utf8mb4")

		// Exactly what the settings form posts while the sqlite option is selected.
		setInfo := model.DefaultConfig()
		setInfo.Core.DbType = "sqlite"
		setInfo.Sqlite.Filename = "new.db"

		putDbSettings(ctx, setInfo)

		if got := cfg.GetStringOrDefault("core.dbtype", ""); got != "sqlite" {
			t.Errorf("core.dbtype = %q, want sqlite", got)
		}
		if got := cfg.GetStringOrDefault("sqlite.filename", ""); got != "new.db" {
			t.Errorf("sqlite.filename = %q, want new.db", got)
		}
		for key, want := range map[string]any{
			"mysql.host":     "db.example.com",
			"mysql.port":     3307,
			"mysql.dbname":   "prod",
			"mysql.username": "app",
			"mysql.charset":  "utf8mb4",
		} {
			switch want := want.(type) {
			case string:
				if got := cfg.GetStringOrDefault(key, ""); got != want {
					t.Errorf("%s = %q after a sqlite-mode save, want %q — the inactive section was overwritten with defaults", key, got, want)
				}
			case int:
				if got := cfg.GetIntOrDefault(key, 0); got != want {
					t.Errorf("%s = %d after a sqlite-mode save, want %d — the inactive section was overwritten with defaults", key, got, want)
				}
			}
		}
	})

	t.Run("mysql save keeps the sqlite filename", func(t *testing.T) {
		ctx := newConfigContext(t)
		cfg := ctx.GetConfig()
		cfg.Put("core.dbtype", "sqlite")
		cfg.Put("sqlite.filename", "custom.db")

		setInfo := model.DefaultConfig()
		setInfo.Core.DbType = "mysql"
		setInfo.Mysql.Host = "db.example.com"
		setInfo.Mysql.Port = 3306
		setInfo.Mysql.Dbname = "prod"

		putDbSettings(ctx, setInfo)

		if got := cfg.GetStringOrDefault("sqlite.filename", ""); got != "custom.db" {
			t.Errorf("sqlite.filename = %q after a mysql-mode save, want custom.db — the inactive section was overwritten with defaults", got)
		}
		if got := cfg.GetStringOrDefault("mysql.dbname", ""); got != "prod" {
			t.Errorf("mysql.dbname = %q, want prod", got)
		}
	})
}

// Empty fields mean "the caller did not send a value" and leave what config.ini holds
// alone; only the fields that carry a value are written.
func TestPutDbSettingsEmptyFieldsLeftAlone(t *testing.T) {
	ctx := newConfigContext(t)
	cfg := ctx.GetConfig()
	cfg.Put("core.dbtype", "sqlite")
	cfg.Put("sqlite.filename", "keep.db")
	cfg.Put("core.maxbodysize", "5M")

	setInfo := model.DefaultConfig()
	setInfo.Core.DbType = ""
	setInfo.Core.MaxBodySize = 0
	setInfo.Sqlite.Filename = ""
	setInfo.Mysql = &model.MysqlConfig{}

	putDbSettings(ctx, setInfo)

	if got := cfg.GetStringOrDefault("core.dbtype", ""); got != "sqlite" {
		t.Errorf("core.dbtype = %q, want the stored value", got)
	}
	if got := cfg.GetStringOrDefault("sqlite.filename", ""); got != "keep.db" {
		t.Errorf("sqlite.filename = %q, want the stored value", got)
	}
	if got := cfg.GetStringOrDefault("core.maxbodysize", ""); got != "5M" {
		t.Errorf("core.maxbodysize = %q, want the stored value", got)
	}
}

// A value that is sent replaces the stored one, on both the size and the ports.
func TestPutSettingsWritesSentValues(t *testing.T) {
	ctx := newConfigContext(t)
	cfg := ctx.GetConfig()
	cfg.Put("manage.port", 12566)
	cfg.Put("api.port", 12567)

	setInfo := model.DefaultConfig()
	setInfo.Core.MaxBodySize = 10 << 20
	setInfo.Manage.Port = 20001
	setInfo.Api.Port = 20002

	putDbSettings(ctx, setInfo)
	putPortSettings(ctx, setInfo)

	if got := cfg.GetStringOrDefault("core.maxbodysize", ""); got != "10M" {
		t.Errorf("core.maxbodysize = %q, want 10M", got)
	}
	if got := cfg.GetIntOrDefault("manage.port", 0); got != 20001 {
		t.Errorf("manage.port = %d, want 20001", got)
	}
	if got := cfg.GetIntOrDefault("api.port", 0); got != 20002 {
		t.Errorf("api.port = %d, want 20002", got)
	}
}

// Port fields that are not sent (or arrive as 0 from a failed parse) must not move the
// listener out from under the instance.
func TestPutPortSettingsIgnoresUnsetPorts(t *testing.T) {
	ctx := newConfigContext(t)
	cfg := ctx.GetConfig()
	cfg.Put("manage.port", 20001)
	cfg.Put("api.port", 20002)

	setInfo := model.DefaultConfig()
	setInfo.Manage.Port = 0
	setInfo.Api.Port = 0

	putPortSettings(ctx, setInfo)

	if got := cfg.GetIntOrDefault("manage.port", 0); got != 20001 {
		t.Errorf("manage.port = %d, want the stored 20001", got)
	}
	if got := cfg.GetIntOrDefault("api.port", 0); got != 20002 {
		t.Errorf("api.port = %d, want the stored 20002", got)
	}
}
