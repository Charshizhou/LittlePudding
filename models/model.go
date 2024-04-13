package models

import "xorm.io/xorm"

type Status int8
type ScheduleType int8
type AddressType int8
type AlarmStatus int8

var Db *xorm.Engine

const (
	Enabled  Status = 1
	Disabled Status = 2
)

const (
	None ScheduleType = iota
	Corn
	FixRate
	FixDelay
)

const (
	Auto   AddressType = iota // 自动录入
	Manual                    // 手动录入
)

const (
	Normal AlarmStatus = iota
	DoNotAlarm
	Success
	Fail
)
