package model

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

type CoreConfig struct {
	Init      string `json:"init"`
	CachePath string `json:"cachePath"`
	DbType    string `json:"dbType"`
	LogLevel  string `json:"logLevel"`
	IsDocker  string `json:"isDocker"`
	Debug     bool   `json:"debug"`
}

func DefaultCoreConfig() *CoreConfig {
	return &CoreConfig{
		Init:      "",
		CachePath: ".cache",
		DbType:    "sqlite",
		LogLevel:  "info",
		IsDocker:  "false",
		Debug:     false,
	}
}

type ManageConfig struct {
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	WebPath  string `json:"webPath"`
}

func DefaultManageConfig() *ManageConfig {
	return &ManageConfig{
		Port:     12566,
		Username: "admin",
		Password: "admin",
		WebPath:  "web",
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
