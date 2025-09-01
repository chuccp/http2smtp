package config

import (
	"github.com/chuccp/http2smtp/util"
	"os"
)

type Config struct {
	filename string
	config   *util.Config
}

func NewConfig() *Config {
	return &Config{filename: "config.ini"}
}

func (config *Config) UpdateSetInfo(setInfo *SetInfo) error {
	config.config.SetBoolean("core", "init", setInfo.HasInit)
	config.config.SetString("core", "cachePath", setInfo.CachePath)
	if setInfo.DbType == "sqlite" {
		config.config.SetString("core", "dbType", setInfo.DbType)
		config.config.SetString("sqlite", "filename", setInfo.Sqlite.Filename)
	} else if setInfo.DbType == "mysql" {
		config.config.SetString("core", "dbType", setInfo.DbType)
		config.config.SetString("mysql", "host", setInfo.Mysql.Host)
		config.config.SetInt("mysql", "port", setInfo.Mysql.Port)
		config.config.SetString("mysql", "dbname", setInfo.Mysql.Dbname)
		config.config.SetString("mysql", "charset", setInfo.Mysql.Charset)
		config.config.SetString("mysql", "username", setInfo.Mysql.Username)
		config.config.SetString("mysql", "password", setInfo.Mysql.Password)
	}
	config.config.SetInt("manage", "port", setInfo.Manage.Port)
	config.config.SetString("manage", "username", setInfo.Manage.Username)
	config.config.SetString("manage", "password", setInfo.Manage.Password)
	config.config.SetString("manage", "webPath", setInfo.Manage.WebPath)
	config.config.SetInt("api", "port", setInfo.Api.Port)
	//config.config.SetBoolean("schedule", "start", setInfo.Schedule.Start)

	//Manage
	err := config.config.Save()
	if err != nil {
		return err
	}
	return nil
}

func (config *Config) Init(webPort int, apiPort int) error {
	if !util.FileExists(config.filename) {
		_, err := os.Create(config.filename)
		if err != nil {
			return err
		}
		fig, err := util.LoadFile(config.filename)
		if err != nil {
			return err
		}
		config.config = fig
		if apiPort > 0 {
			defaultSetInfo.Api.Port = apiPort
		}
		if webPort > 0 {
			defaultSetInfo.Manage.Port = webPort
		}
		config.UpdateSetInfo(defaultSetInfo)
		return nil
	} else {
		fig, err := util.LoadFile(config.filename)
		if err != nil {
			return err
		}
		if apiPort > 0 {
			fig.SetInt("api", "port", apiPort)
		}
		if webPort > 0 {
			fig.SetInt("manage", "port", webPort)
		}
		config.config = fig
		return nil
	}
}
func (config *Config) ReadSetInfo() *SetInfo {
	var setInfo SetInfo
	setInfo.HasInit = config.config.GetBooleanOrDefault("core", "init", false)
	setInfo.DbType = config.config.GetStringOrDefault("core", "dbType", defaultSetInfo.DbType)
	var sqlite Sqlite
	sqlite.Filename = config.config.GetStringOrDefault("sqlite", "filename", defaultSetInfo.Sqlite.Filename)
	setInfo.Sqlite = &sqlite
	var mysql Mysql
	mysql.Host = config.config.GetString("mysql", "host")
	mysql.Port = config.config.GetIntOrDefault("mysql", "port", defaultSetInfo.Mysql.Port)
	mysql.Dbname = config.config.GetString("mysql", "dbname")
	mysql.Username = config.config.GetString("mysql", "username")
	mysql.Password = config.config.GetString("mysql", "password")
	mysql.Charset = config.config.GetString("mysql", "charset")
	setInfo.Mysql = &mysql
	var manage Manage
	manage.WebPath = config.config.GetStringOrDefault("manage", "webPath", defaultSetInfo.Manage.WebPath)
	manage.Port = config.config.GetIntOrDefault("manage", "port", defaultSetInfo.Manage.Port)
	manage.Username = config.config.GetString("manage", "username")
	manage.Password = config.config.GetString("manage", "password")
	setInfo.Manage = &manage
	var api Api
	api.Port = config.config.GetIntOrDefault("api", "port", defaultSetInfo.Api.Port)
	setInfo.Api = &api
	return &setInfo
}

func (config *Config) GetInt(section, name string) int {
	getInt, err := config.config.GetInt(section, name)
	if err != nil {
		return 0
	} else {
		return getInt
	}
}

func (config *Config) GetString(section, name string) string {
	return config.config.GetString(section, name)
}

func (config *Config) GetStringOrDefault(section string, name string, defaultValue string) string {
	return config.config.GetStringOrDefault(section, name, defaultValue)
}
func (config *Config) GetIntOrDefault(section string, name string, defaultValue int) int {
	getInt, err := config.config.GetInt(section, name)
	if err != nil {
		return defaultValue
	} else {
		return getInt
	}
}
func (config *Config) GetBooleanOrDefault(section string, name string, defaultValue bool) bool {
	return config.config.GetBooleanOrDefault(section, name, defaultValue)
}
