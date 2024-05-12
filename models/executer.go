package models

import "time"

// Executor 执行组模型
type Executor struct {
	Id            int         `json:"id" xorm:"int pk autoincr"`
	ExecutorName  string      `json:"executor_name" xorm:"varchar(64) notnull"`           // 执行器名称
	ExecutorTitle string      `json:"executor_title" xorm:"varchar(32) notnull"`          // 执行器标题
	AddressType   AddressType `json:"address_type" xorm:"tinyint(4) notnull default (0)"` // 执行器地址类型 0：自动录入 1：手动录入
	AddressList   string      `json:"address_list" xorm:"text"`                           // 执行器地址列表
	UpdateTime    time.Time   `json:"update_time" xorm:"datetime default null"`           // 更新时间
}
