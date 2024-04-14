package config

import "strings"

type Db struct {
	Engine   string // 数据库引擎
	Host     string // 数据库地址
	Port     string // 数据库端口
	User     string // 数据库用户名
	Password string // 数据库密码
	Database string // 数据库名称
	Prefix   string // 数据库表前缀
	Charset  string // 数据库字符集
}

var DefaultDb = Db{
	Engine:   "mysql",
	Host:     "localhost",
	Port:     "3306",
	User:     "root",
	Password: "123456",
	Database: "littlepudding",
	Prefix:   "",
	Charset:  "utf8",
}

func NewDb(db Db) (engine, conf string) {
	return db.Engine, strings.Join([]string{db.User, ":", db.Password, "@tcp(", db.Host, ":", db.Port, ")/", db.Database, "?charset=", db.Charset}, "")
}
