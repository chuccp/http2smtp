package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/chuccp/go-web-frame/config"
)

// DefaultMaxBodySize is the request-body cap used when config.ini does not set
// core.maxBodySize: 10 MB, the framework's own default.
const DefaultMaxBodySize int64 = 10 << 20

// ByteSize is a byte count read from config.ini. Values may be written as plain bytes
// ("10485760") or with a K/M/G suffix ("10M"), so the default stays readable in the
// file; -1 means unlimited, matching web.ServerConfig.MaxBodySize.
type ByteSize int64

// ParseByteSize parses a size string. Units are binary (K = 1024).
func ParseByteSize(value string) (ByteSize, error) {
	text := strings.TrimSpace(strings.ToUpper(value))
	if text == "" {
		return 0, errors.New("empty size")
	}
	if text == "-1" {
		return ByteSize(-1), nil
	}
	multiplier := int64(1)
	for _, unit := range []struct {
		suffix string
		mult   int64
	}{
		{"KB", 1 << 10}, {"MB", 1 << 20}, {"GB", 1 << 30},
		{"K", 1 << 10}, {"M", 1 << 20}, {"G", 1 << 30}, {"B", 1},
	} {
		if strings.HasSuffix(text, unit.suffix) {
			multiplier = unit.mult
			text = strings.TrimSpace(strings.TrimSuffix(text, unit.suffix))
			break
		}
	}
	n, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q", value)
	}
	if n <= 0 {
		return 0, fmt.Errorf("size must be positive: %q", value)
	}
	if n > (1<<63-1)/multiplier {
		return 0, fmt.Errorf("size is too large: %q", value)
	}
	return ByteSize(n * multiplier), nil
}

// String renders a size back with the largest exact unit, so errors and logs stay
// human-readable.
func (b ByteSize) String() string {
	switch {
	case b < 0:
		return "-1"
	case b == 0:
		return "0"
	case b%(1<<30) == 0:
		return strconv.FormatInt(int64(b>>30), 10) + "G"
	case b%(1<<20) == 0:
		return strconv.FormatInt(int64(b>>20), 10) + "M"
	case b%(1<<10) == 0:
		return strconv.FormatInt(int64(b>>10), 10) + "K"
	default:
		return strconv.FormatInt(int64(b), 10)
	}
}

// UnmarshalText lets the INI decoder hand the string value straight to the parser.
func (b *ByteSize) UnmarshalText(text []byte) error {
	v, err := ParseByteSize(string(text))
	if err != nil {
		return err
	}
	*b = v
	return nil
}

// UnmarshalJSON accepts either "10M" or a plain number of bytes, so the JSON config
// endpoints and config files understand the same values.
func (b *ByteSize) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		return b.UnmarshalText([]byte(text))
	}
	var n int64
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	if n == 0 || n < -1 {
		return fmt.Errorf("invalid size %d", n)
	}
	*b = ByteSize(n)
	return nil
}

// MarshalJSON reports the byte count, which is what web.ServerConfig consumes.
func (b ByteSize) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(b))
}

type Config struct {
	Core   *CoreConfig   `json:"core"`
	Manage *ManageConfig `json:"manage"`
	Api    *ApiConfig    `json:"api"`
	Sqlite *SqliteConfig `json:"sqlite"`
	Mysql  *MysqlConfig  `json:"mysql"`
}

func DefaultConfig() *Config {
	return &Config{
		Core:   DefaultCoreConfig(),
		Manage: DefaultManageConfig(),
		Api:    DefaultApiConfig(),
		Sqlite: DefaultSqliteConfig(),
		Mysql:  DefaultMysqlConfig(),
	}
}

func GetConfig(config config.IConfig) (*Config, error) {
	cfg := DefaultConfig()
	// Pass the *Config itself, not &cfg: the decoder only applies its per-field
	// weakly-typed conversions when the target is a struct. A **Config falls back to
	// encoding/json, which cannot coerce the strings an INI file yields into
	// bool/int fields ("true" -> bool, "12566" -> int).
	err := config.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}
	// The decoder allocates a fresh nested struct for a pointer field, so defaults in
	// DefaultCoreConfig are lost whenever [core] exists but does not spell every key out.
	// maxBodySize is the one whose absence must still mean 10 MB.
	if cfg.Core == nil {
		cfg.Core = DefaultCoreConfig()
	} else if cfg.Core.MaxBodySize == 0 {
		cfg.Core.MaxBodySize = ByteSize(DefaultMaxBodySize)
	}
	return cfg, nil
}

// CoreState reads the core section through the typed config.
//
// Do not reach for IConfig.GetBoolOrDefault("core.init") instead: values loaded from an
// INI file are strings, and that getter only accepts a real bool, so it silently reports
// the default. Every setup guard and the startup database attach depend on this.
func CoreState(cfg config.IConfig) *CoreConfig {
	c, err := GetConfig(cfg)
	if err != nil || c.Core == nil {
		return DefaultCoreConfig()
	}
	return c.Core
}

// MaskSecrets blanks credential fields so the config can be returned to clients.
// Callers that persist config treat an empty password as "keep the existing one".
func (cfg *Config) MaskSecrets() *Config {
	if cfg != nil && cfg.Mysql != nil {
		cfg.Mysql.Password = ""
	}
	return cfg
}

type CoreConfig struct {
	Init      bool   `json:"init"`
	DbInit    bool   `json:"dbInit"`
	CachePath string `json:"cachePath"`
	DbType    string `json:"dbType"`
	LogLevel  string `json:"logLevel"`
	IsDocker  bool   `json:"isDocker"`
	Debug     bool   `json:"debug"`
	// MaxBodySize caps request bodies on both servers; 0 means unset (10 MB). Written by
	// the settings page and applied on the next restart, like the ports.
	MaxBodySize ByteSize `json:"maxBodySize"`
}

func DefaultCoreConfig() *CoreConfig {
	return &CoreConfig{
		Init:      false,
		DbInit:    false,
		CachePath: ".cache",
		DbType:    "sqlite",
		LogLevel:  "info",
		IsDocker:  false,
		Debug:     false,
		// 0 is "unset": GetConfig fills in DefaultMaxBodySize, and a settings save that
		// does not carry the field leaves an existing config.ini value alone.
		MaxBodySize: 0,
	}
}

// MaxBodySizeValue returns the effective cap for both the framework's JSON body limit
// and the form/multipart path: a negative value means unlimited.
func (c *CoreConfig) MaxBodySizeValue() int64 {
	if c == nil || c.MaxBodySize == 0 {
		return DefaultMaxBodySize
	}
	return int64(c.MaxBodySize)
}

type ManageConfig struct {
	Port    int    `json:"port"`
	WebPath string `json:"webPath"`
}

var ManagePort = 12566
var ApiPort = 12567

func DefaultManageConfig() *ManageConfig {
	return &ManageConfig{
		Port:    12566,
		WebPath: "web",
	}
}

type ApiConfig struct {
	Port int `json:"port"`
}

func DefaultApiConfig() *ApiConfig {
	return &ApiConfig{
		Port: 12567,
	}
}

type SqliteConfig struct {
	Filename string `json:"filename"`
}

func DefaultSqliteConfig() *SqliteConfig {
	return &SqliteConfig{
		Filename: "d-mail.db",
	}
}

type MysqlConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Dbname   string `json:"dbname"`
	Username string `json:"username"`
	Password string `json:"password"`
	Charset  string `json:"charset"`
}

func DefaultMysqlConfig() *MysqlConfig {
	return &MysqlConfig{
		Host:     "",
		Port:     3306,
		Dbname:   "d-main",
		Username: "",
		Password: "",
		Charset:  "utf8",
	}
}

type System struct {
	HasInit   bool `json:"hasInit"`
	HasDbInit bool `json:"hasDbInit"`
	HasAdmin  bool `json:"hasAdmin"`
	HasLogin  bool `json:"hasLogin"`
	IsDocker  bool `json:"isDocker"`
	// Populated only when the request carries a valid session. The session itself is an
	// HttpOnly cookie the client cannot read, so the UI needs these to render the
	// username and decide whether to show admin-only navigation.
	Username string `json:"username"`
	IsAdmin  bool   `json:"isAdmin"`
}
