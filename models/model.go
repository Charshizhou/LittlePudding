package models

import (
	"LittlePudding/modules/config"
	"log"
	"xorm.io/xorm"
)

type AddressType int8
type TriggerCode int16
type HandleCode int16
type Result int16

type CommonMap map[string]interface{}

var Db *xorm.Engine

const (
	Auto   AddressType = iota // 自动录入
	Manual                    // 手动录入
)

const (
	SuccessTrigger TriggerCode = 200
	FailTrigger    TriggerCode = 500
)

const ( // 完成
	Success Result = iota // 成功
	Failure               // 失败
	Expired               // 执行超时
	Overdue               // 任务过期
)

const (
	Page        = 1    // 当前页数
	PageSize    = 20   // 每页多少条数据
	MaxPageSize = 1000 // 每次最多取多少条
)

const DefaultTimeFormat = "2024-05-01 00:00:00"

type BaseModel struct {
	Page     int `xorm:"-"`
	PageSize int `xorm:"-"`
}

func (model *BaseModel) parsePageAndPageSize(params CommonMap) {
	page, ok := params["Page"]
	if ok {
		model.Page = page.(int)
	}
	pageSize, ok := params["PageSize"]
	if ok {
		model.PageSize = pageSize.(int)
	}
	if model.Page <= 0 {
		model.Page = Page
	}
	if model.PageSize <= 0 {
		model.PageSize = MaxPageSize
	}
}

func (model *BaseModel) pageLimitOffset() int {
	return (model.Page - 1) * model.PageSize
}

func InstallDb() error {
	Db = InitDb(&config.Setting{
		Db: config.DefaultDb,
	})
	if Db == nil {
		log.Fatal("数据库初始化失败")
	}
	err := Db.Sync2(new(Task))
	if err != nil {
		log.Fatal("数据库同步任务表失败", err)
	}
	err = Db.Sync2(new(TaskLog))
	if err != nil {
		log.Fatal("数据库同步任务日志表失败", err)
	}
	err = Db.Sync2(new(Executor))
	if err != nil {
		log.Fatal("数据库同步执行器表失败", err)
	}
	return err
}

func InitDb(setting *config.Setting) *xorm.Engine {
	var dbEngine, conf string
	if Db == nil {
		dbEngine, conf = config.NewDb(config.DefaultDb)
	} else {
		dbEngine, conf = config.NewDb(setting.Db)
	}
	engine, err := xorm.NewEngine(dbEngine, conf)
	if err != nil {
		log.Fatal("创建xorm引擎失败", err)
	}
	return engine
}
