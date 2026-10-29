// Package compatibility guards the on-disk configuration and database against
// in-place upgrades from releases whose data format this build cannot read.
package compatibility

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/http2smtp/model"
)

// DataVersionKey is persisted in config.ini. It describes the complete application
// data format, not the display version of the binary.
const DataVersionKey = "core.dataVersion"

// CurrentDataVersion must be bumped whenever an installed instance cannot safely be
// opened by the old layout or cannot safely open data written by the old layout.
const CurrentDataVersion = 1

// CheckOptions contains the files needed to explain an incompatibility to the user.
type CheckOptions struct {
	Config        config.IConfig
	ConfigPath    string
	SQLitePath    string
	StorageRoot   string
	DatabaseType  string
	MySQLDatabase string
}

// IncompatibleError means startup must not attach the database. It is a typed error so
// main can serve the recovery page without a panic stack.
type IncompatibleError struct {
	Code            ProblemCode
	Reason          string
	RawVersion      string
	FoundVersion    int
	RequiredVersion int
	ConfigPath      string
	SQLitePath      string
	StorageRoot     string
	DatabaseType    string
	MySQLDatabase   string
}

// ProblemCode lets the web UI translate the reason without parsing error text.
type ProblemCode string

const (
	ProblemLegacyData      ProblemCode = "legacy"
	ProblemInvalidVersion  ProblemCode = "invalid"
	ProblemVersionMismatch ProblemCode = "version"
)

func (e *IncompatibleError) Error() string {
	var b strings.Builder
	b.WriteString("HTTP2SMTP 数据不兼容，已阻止连接数据库。\n")
	b.WriteString("原因：")
	b.WriteString(e.Reason)
	b.WriteString("\n\n此版本不能就地升级或读取现有数据，程序没有修改数据库。\n")
	b.WriteString("请先停止旧版本，并完整备份以下内容：\n")
	if e.ConfigPath != "" {
		b.WriteString("  1. 配置文件：")
		b.WriteString(e.ConfigPath)
		b.WriteByte('\n')
	}
	if e.SQLitePath != "" {
		b.WriteString("  2. SQLite 数据库：")
		b.WriteString(e.SQLitePath)
		b.WriteByte('\n')
	} else if strings.EqualFold(e.DatabaseType, "mysql") {
		b.WriteString("  2. MySQL 数据库：")
		b.WriteString(e.MySQLDatabase)
		b.WriteString("（请先执行完整数据库备份）\n")
	}
	if e.StorageRoot != "" {
		b.WriteString("  3. 附件和缓存目录：")
		b.WriteString(e.StorageRoot)
		b.WriteByte('\n')
	}
	b.WriteString("\n备份完成后，按以下方式重置：")
	switch {
	case e.SQLitePath != "":
		b.WriteString("将 config.ini 和 SQLite 数据库文件移出程序目录（不要删除），\n")
		b.WriteString("再启动此版本并重新完成初始化。")
	case strings.EqualFold(e.DatabaseType, "mysql"):
		b.WriteString("移走 config.ini，并在备份完成后重建一个空的 MySQL 数据库，\n")
		b.WriteString("再启动此版本并重新完成初始化。")
	default:
		b.WriteString("移走 config.ini，再启动此版本并重新完成初始化。")
	}
	b.WriteString("旧数据不会被自动删除。\n")
	return b.String()
}

// IsIncompatible reports whether err is a data-format refusal.
func IsIncompatible(err error) bool {
	_, ok := AsIncompatible(err)
	return ok
}

// AsIncompatible returns the typed refusal used by the management UI.
func AsIncompatible(err error) (*IncompatibleError, bool) {
	var incompatible *IncompatibleError
	return incompatible, errors.As(err, &incompatible)
}

// Check validates the persisted data generation. An unversioned but unused config is a
// fresh installation and is stamped. An unversioned initialized config, database, or
// v0.2.x config layout is rejected without changing the file.
func Check(opts CheckOptions) error {
	if opts.Config == nil {
		return errors.New("compatibility check requires a config")
	}

	raw := strings.TrimSpace(opts.Config.GetString(DataVersionKey))
	if raw == "" {
		if hasExistingInstallation(opts) {
			return opts.incompatible(
				ProblemLegacyData,
				"检测到旧版或没有数据版本标记的安装",
				"",
				0,
			)
		}
		opts.Config.Put(DataVersionKey, strconv.Itoa(CurrentDataVersion))
		if err := opts.Config.WriteConfig(); err != nil {
			return fmt.Errorf("write compatibility marker: %w", err)
		}
		return nil
	}

	version, err := strconv.Atoi(raw)
	if err != nil {
		return opts.incompatible(
			ProblemInvalidVersion,
			fmt.Sprintf("数据版本标记无效：%q", raw),
			raw,
			0,
		)
	}
	if version != CurrentDataVersion {
		return opts.incompatible(
			ProblemVersionMismatch,
			fmt.Sprintf("数据版本为 %d，当前程序需要版本 %d", version, CurrentDataVersion),
			raw,
			version,
		)
	}
	return nil
}

func hasExistingInstallation(opts CheckOptions) bool {
	state := model.CoreState(opts.Config)
	if state.Init || state.DbInit {
		return true
	}
	if sqliteHasData(opts.SQLitePath) {
		return true
	}

	// v0.2.x always wrote these settings, including when setup had not completed.
	// Their presence identifies an old layout even if core.init is false.
	for _, key := range []string{
		"manage.username",
		"manage.password",
		"manage.webpath",
		"schedule.start",
	} {
		if opts.Config.HasKey(key) {
			return true
		}
	}

	// A configured MySQL connection may already contain data. We cannot inspect it
	// without connecting, and connecting before this guard would defeat the purpose.
	if strings.EqualFold(opts.DatabaseType, "mysql") {
		return strings.TrimSpace(opts.Config.GetString("mysql.host")) != "" ||
			strings.TrimSpace(opts.Config.GetString("mysql.dbname")) != ""
	}
	return false
}

func sqliteHasData(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular() && info.Size() > 0
}

func (opts CheckOptions) incompatible(code ProblemCode, reason, rawVersion string, foundVersion int) error {
	return &IncompatibleError{
		Code:            code,
		Reason:          reason,
		RawVersion:      rawVersion,
		FoundVersion:    foundVersion,
		RequiredVersion: CurrentDataVersion,
		ConfigPath:      absolutePath(opts.ConfigPath),
		SQLitePath:      absolutePath(opts.SQLitePath),
		StorageRoot:     absolutePath(opts.StorageRoot),
		DatabaseType:    opts.DatabaseType,
		MySQLDatabase:   opts.MySQLDatabase,
	}
}

func absolutePath(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return filepath.Clean(absolute)
}
