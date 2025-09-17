package db

import (
	"fmt"
	"github.com/chuccp/http2smtp/config"
	"github.com/chuccp/http2smtp/sqlite"
	"github.com/chuccp/http2smtp/util"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"path"
)

var NoDatabase = &NoDatabaseError{}

type DB struct {
	db          *gorm.DB
	err         error
	storageRoot string
}

func (d *DB) GetSMTPModel() *STMPModel {
	return NewSMTPModel(d.db, "t_SMTP")
}
func (d *DB) GetMailModel() *MailModel {
	return NewMailModel(d.db, "t_mail")
}
func (d *DB) GetLogModel() *LogModel {
	return NewLogModel(d.db, "t_log")
}
func (d *DB) GetTokenModel() *TokenModel {
	return NewTokenModel(d.db, "t_token")
}
func (d *DB) GetScheduleModel() *ScheduleModel {
	return NewScheduleModel(d.db, "t_schedule")
}
func CreateDB(storageRoot string) *DB {
	return &DB{storageRoot: storageRoot}
}

// InitBySetInfo setInfo *config.SetInfo
func (d *DB) InitBySetInfo(setInfo *config.SetInfo) error {
	var err error
	dbType := setInfo.DbType
	if util.EqualsAnyIgnoreCase(dbType, "sqlite") {
		dbName := path.Join(d.storageRoot, setInfo.Sqlite.Filename)
		d.db, err = gorm.Open(sqlite.Open(dbName), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
		if err != nil {
			d.err = err
			return err
		}
		d.err = err
		return err
	} else if util.EqualsAnyIgnoreCase(dbType, "mysql") {
		username := setInfo.Mysql.Username
		password := setInfo.Mysql.Password
		host := setInfo.Mysql.Host
		port := setInfo.Mysql.Port
		dbname := setInfo.Mysql.Dbname
		if len(setInfo.Mysql.Charset) == 0 {
			setInfo.Mysql.Charset = "utf8"
		}
		charset := setInfo.Mysql.Charset
		d.db, err = CreateMysqlConnection(username, password, host, port, dbname, charset)
		if err != nil {
			d.err = err
			return err
		}
		d.err = err
		return err
	}
	d.err = NoDatabase
	return d.err
}
func (d *DB) Init(config *config.Config) error {
	var err error
	dbType := config.GetString("core", "dbType")
	if util.EqualsAnyIgnoreCase(dbType, "sqlite") {
		dbName := config.GetStringOrDefault("sqlite", "filename", "d-mail.db")

		d.db, err = gorm.Open(sqlite.Open(dbName), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
		if err != nil {
			d.err = err
			return err
		}
		d.err = err
		return err
	} else if util.EqualsAnyIgnoreCase(dbType, "mysql") {
		username := config.GetString("mysql", "username")
		password := config.GetString("mysql", "password")
		host := config.GetString("mysql", "host")
		port := config.GetIntOrDefault("mysql", "port", 3306)
		dbname := config.GetString("mysql", "dbname")
		charset := config.GetStringOrDefault("mysql", "charset", "UTF8")
		//dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local", username, password, host, port, dbname, charset)
		d.db, err = CreateMysqlConnection(username, password, host, port, dbname, charset)
		if err != nil {
			d.err = err
			return err
		}
		d.err = err
		return err
	}
	d.err = NoDatabase
	return d.err
}

func CreateMysqlConnection(username string, password string, host string, port int, dbname string, charset string) (db *gorm.DB, err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", username, password, host, port, dbname, charset)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
}

type NoDatabaseError struct {
	error
}

func (error *NoDatabaseError) Error() string {
	return "No database type selected"
}
