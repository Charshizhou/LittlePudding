package models

import "time"

type XxlJobLogReport struct {
	Id           int       `json:"id" xorm:"int(11) notnull pk autoincr"`
	TriggerDay   time.Time `json:"trigger_day" xorm:"datetime"`                    // 调度时间
	RunningCount int       `json:"running_count" xorm:"int(11) notnull default 0"` // 运行中 日志数量
	SucCount     int       `json:"suc_count" xorm:"int(11) notnull default 0"`     // 执行成功 数量
	FailCount    int       `json:"fail_count" xorm:"int(11) notnull default 0"`    // 执行失败数量
	UpdateTime   time.Time `json:"update_time" xorm:"datetime default null"`       // 更新时间
}
