package models

import "time"

type XxlJobLog struct {
	Id                     int         `json:"id" xorm:"bigint(20) notnull pk autoincr"`
	TaskGroup              int         `json:"job_group" xorm:"int(11) notnull"`                           // 执行器主键
	TaskId                 int         `json:"job_id" xorm:"int(11) notnull"`                              // 任务主键
	ExecutorAddress        string      `json:"executor_address" xorm:"varchar(255)"`                       // 执行器地址，本次执行的地址
	ExecutorHandler        string      `json:"executor_handler" xorm:"varchar(255)"`                       // 执行器任务handler
	ExecutorParam          string      `json:"executor_param" xorm:"varchar(512)"`                         // 执行器任务参数
	ExecutorShardingParam  string      `json:"executor_sharding_param" xorm:"varchar(20)"`                 // 执行器任务分片参数
	ExecutorFailRetryCount int         `json:"executor_fail_retry_count" xorm:"int(11) notnull default 0"` // 失败重试次数
	TriggerTime            time.Time   `json:"trigger_time" xorm:"datetime"`                               // 调度时间
	TriggerCode            int         `json:"trigger_code" xorm:"int(11) notnull"`                        // 调度结果
	TriggerMsg             string      `json:"trigger_msg" xorm:"text"`                                    // 调度消息
	HandleTime             time.Time   `json:"handle_time" xorm:"datetime"`                                // 执行时间
	HandleCode             int         `json:"handle_code" xorm:"int(11) notnull"`                         // 执行结果
	HandleMsg              string      `json:"handle_msg" xorm:"text"`                                     // 执行日志
	AlarmStatus            AlarmStatus `json:"alarm_status" xorm:"tinyint(4) notnull default 0"`           // 报警状态
}
