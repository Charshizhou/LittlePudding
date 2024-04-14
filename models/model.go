package models

import (
	"LittlePudding/modules/config"
	"log"
	"xorm.io/xorm"
)

type AddressType int8
type TriggerCode int16
type HandleCode int16

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

const (
	SuccessHandle HandleCode = 200
	FailHandle    HandleCode = 500
	Running       HandleCode = 0
)

func InitDb() error {
	Db = CreateDb()
	if Db == nil {
		log.Fatal("数据库初始化失败")
	}
	err := Db.Sync2(new(User))
	if err != nil {
		log.Fatal("数据库同步用户表失败", err)
	}
	err = Db.Sync2(new(Task))
	if err != nil {
		log.Fatal("数据库同步任务表失败", err)
	}
	err = Db.Sync2(new(TaskLog))
	if err != nil {
		log.Fatal("数据库同步任务日志表失败", err)
	}
	err = Db.Sync2(new(TaskLogGlue))
	if err != nil {
		log.Fatal("数据库同步任务日志关联表失败", err)
	}
	err = Db.Sync2(new(Executor))
	if err != nil {
		log.Fatal("数据库同步执行器表失败", err)
	}
	err = Db.Sync2(new(Registry))
	if err != nil {
		log.Fatal("数据库同步注册中心表失败", err)
	}
	err = Db.Sync2(new(TaskLogReport))
	if err != nil {
		log.Fatal("数据库同步任务日志报表表失败", err)
	}
	return err
}

func CreateDb() *xorm.Engine {
	dbEngine, conf := config.NewDb(config.DefaultDb)
	engine, err := xorm.NewEngine(dbEngine, conf)
	if err != nil {
		log.Fatal("创建xorm引擎失败", err)
	}
	return engine
}
