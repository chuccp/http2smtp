package db

import (
	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/http2smtp/model"
)

func GetDb(config config.IConfig) (*db.DB, error) {
	dbType := config.GetString("core.dbtype")
	if dbType == "sqlite" {
		var sqliteConfig = model.DefaultSqliteConfig()
		err := config.UnmarshalKey("sqlite", &sqliteConfig)
		if err != nil {
			return nil, err
		}
		connection, err := db.ConnectionSQLite(sqliteConfig.Filename)
		if err != nil {
			return nil, err
		}
		return connection, nil
	} else if dbType == "mysql" {
		var mysqlConfig = model.DefaultMysqlConfig()
		err := config.UnmarshalKey("mysql", &mysqlConfig)
		if err != nil {
			return nil, err
		}
		connection, err := db.ConnectionMysql(mysqlConfig.Host, mysqlConfig.Port, mysqlConfig.Username, mysqlConfig.Password, mysqlConfig.Dbname, mysqlConfig.Charset)
		if err != nil {
			return nil, err
		}
		return connection, nil
	}
	return nil, db.NoConfigDBError
}
