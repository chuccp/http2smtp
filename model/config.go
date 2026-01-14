package model

type Config struct {
	Core   *CoreConfig   `json:"core"`
	Manage *ManageConfig `json:"manage"`
	Api    *ApiConfig    `json:"api"`
}

func DefaultConfig() *Config {
	return &Config{
		Core:   DefaultCoreConfig(),
		Manage: DefaultManageConfig(),
		Api:    DefaultApiConfig(),
	}
}

type CoreConfig struct {
	Init      string `json:"init"`
	CachePath string `json:"cachePath"`
	DbType    string `json:"dbType"`
	LogLevel  string `json:"logLevel"`
	IsDocker  string `json:"isDocker"`
}

func DefaultCoreConfig() *CoreConfig {
	return &CoreConfig{
		Init:      "false",
		CachePath: ".cache",
		DbType:    "sqlite",
		LogLevel:  "info",
		IsDocker:  "false",
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
