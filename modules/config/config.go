package config

import (
	"LittlePudding/modules/utils"
	"fmt"
	"github.com/go-ini/ini"
	"strings"
)

const DefaultSection = "default"

type Db struct {
	Engine   string // 数据库引擎
	Host     string // 数据库地址
	Port     string // 数据库端口
	User     string // 数据库用户名
	Password string // 数据库密码
	Database string // 数据库名称
	Prefix   string // 数据库表前缀
	Charset  string // 数据库字符集
	Loc      string //
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
	Loc:      "Local",
}

var DefaultSetting = Setting{}

type Setting struct {
	Db Db

	AllowIps string // 允许访问的IP
	AppName  string // 应用名称

	ConcurrencyQueue int    // 并发队列，用于控制并发数
	AuthSecret       string // 授权密钥，用于验证请求是否合法

	ServerAddr string // 服务地址
	EnableTLS  bool   // 是否启用TLS
	CertFile   string // 证书文件
	KeyFile    string // 私钥文件
}

func NewDb(db Db) (engine, conf string) {
	return db.Engine, strings.Join([]string{db.User, ":", db.Password, "@tcp(", db.Host, ":", db.Port, ")/", db.Database, "?charset=", db.Charset, "&loc=", db.Loc}, "")
}

func ReadConfig(filename string) (*Setting, error) {
	conf, err := ini.Load(filename)
	if err != nil {
		return nil, err
	}
	section := conf.Section(DefaultSection)
	var s Setting
	s.Db.Engine = section.Key("db.engine").MustString(DefaultDb.Engine)
	s.Db.Host = section.Key("db.host").MustString(DefaultDb.Host)
	s.Db.Port = section.Key("db.port").MustString(DefaultDb.Port)
	s.Db.User = section.Key("db.user").MustString(DefaultDb.User)
	s.Db.Password = section.Key("db.password").MustString(DefaultDb.Password)
	s.Db.Database = section.Key("db.database").MustString(DefaultDb.Database)
	s.Db.Prefix = section.Key("db.prefix").MustString(DefaultDb.Prefix)
	s.Db.Charset = section.Key("db.charset").MustString(DefaultDb.Charset)

	s.AllowIps = section.Key("allow_ips").MustString("")
	s.AppName = section.Key("app.name").MustString("定时任务管理系统")
	s.AuthSecret = section.Key("auth_secret").MustString("")
	s.ServerAddr = section.Key("server_addr").MustString("localhost:50051")
	s.EnableTLS = section.Key("enable_tls").MustBool(false)
	s.CertFile = section.Key("cert_file").MustString("")
	s.KeyFile = section.Key("key_file").MustString("")

	if s.AuthSecret == "" {
		s.AuthSecret = utils.RandAuthToken()
	}
	return &s, nil
}

func WriteConfig(filename string, setting *Setting) error {
	conf, err := ini.Load(filename)
	if err != nil {
		return err
	}
	fmt.Println(conf)
	section := conf.Section(DefaultSection)
	//section.Key("db.engine").SetValue(setting.Db.Engine)
	//section.Key("db.host").SetValue(setting.Db.Host)
	//section.Key("db.port").SetValue(setting.Db.Port)
	//section.Key("db.user").SetValue(setting.Db.User)
	//section.Key("db.password").SetValue(setting.Db.Password)
	//section.Key("db.database").SetValue(setting.Db.Database)
	//section.Key("db.prefix").SetValue(setting.Db.Prefix)
	//section.Key("db.charset").SetValue(setting.Db.Charset)
	//section.Key("allow_ips").SetValue(setting.AllowIps)
	//section.Key("app.name").SetValue(setting.AppName)
	//section.Key("auth_secret").SetValue(setting.AuthSecret)
	section.Key("server_addr").SetValue(setting.ServerAddr)
	if setting.EnableTLS {
		section.Key("enable_tls").SetValue("true")
		section.Key("cert_file").SetValue(setting.CertFile)
		section.Key("key_file").SetValue(setting.KeyFile)
	} else {
		section.Key("enable_tls").SetValue("false")
		section.Key("cert_file").SetValue("")
		section.Key("key_file").SetValue("")
	}
	return conf.SaveTo(filename)
}
