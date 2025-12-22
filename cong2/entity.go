package config2

type System struct {
	HasInit  bool `json:"hasInit"`
	HasLogin bool `json:"hasLogin"`
	IsDocker bool `json:"isDocker"`
}

type Manage struct {
	Port            int    `json:"port"`
	WebPath         string `json:"webPath"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type Api struct {
	Port int `json:"port"`
}

type SetInfo struct {
	HasInit   bool      `json:"hasInit"`
	DbType    string    `json:"dbType"`
	CachePath string    `json:"cachePath"`
	Sqlite    *Sqlite   `json:"sqlite"`
	Mysql     *Mysql    `json:"mysql"`
	Manage    *Manage   `json:"manage"`
	Api       *Api      `json:"api"`
	IsDocker  bool      `json:"isDocker"`
	Schedule  *Schedule `json:"schedule"`
}

var defaultSetInfo = &SetInfo{
	HasInit:   false,
	DbType:    "sqlite",
	CachePath: ".cache",
	Sqlite:    &Sqlite{Filename: "d-mail.db"},
	Mysql:     &Mysql{Host: "", Port: 3306, Dbname: "d-main", Username: "", Password: "", Charset: "utf8"},
	Manage:    &Manage{WebPath: "web", Port: 12566, Username: "", Password: "", ConfirmPassword: ""},
	Api:       &Api{Port: 12567},
	Schedule:  &Schedule{Start: true},
}

type Sqlite struct {
	Filename string `json:"filename"`
}
type Mysql struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Dbname   string `json:"dbname"`
	Username string `json:"username"`
	Password string `json:"password"`
	Charset  string `json:"charset"`
}
type Schedule struct {
	Start bool `json:"start"`
}
