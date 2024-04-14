package models

import "time"

type Status int8
type ScheduleType int8
type AlarmStatus int8

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
	Normal AlarmStatus = iota
	DoNotAlarm
	Success
	Fail
)

type Task struct {
	Id                     int          `json:"id" xorm:"int(11) pk autoincr"`
	TaskCorn               string       `json:"jobcorn" xorm:"varchar(64) notnull"`                                // crontab
	TaskDesc               string       `json:"task_desc" xorm:"varchar(255) notnull"`                             // 任务描述
	Author                 string       `json:"author" xorm:"varchar(64) default ('')"`                            // 任务创建者
	AlarmStatus            AlarmStatus  `json:"alarm_status" xorm:"tinyint(4) notnull default (0)"`                // 报警状态
	AlarmEmail             string       `json:"alarm_email" xorm:"varchar(256) default ('')"`                      // 报警邮件
	ScheduleType           ScheduleType `json:"schedule_type" xorm:"tinyint(4) notnull default (0)"`               // 调度类型 corn,FIX_RATE,FIX_DELAY
	ScheduleConf           string       `json:"schedule_conf" xorm:"varchar(128) default ('')"`                    // 调度配置
	MisfireStrategy        string       `json:"misfire_strategy xorm::varchar(50) notnull default ('DO_NOTHING')"` // 调度过期策略
	ExecutorRouteStrategy  string       `json:"executor_route_strategy" xorm:"varchar(50) default ('')"`           // 执行器路由策略
	ExecutorHandler        string       `json:"executor_handler" xorm:"varchar(255) default ('')"`                 // 执行器任务处理器
	ExecutorParam          string       `json:"executer_param" xorm:"varchar(512) default ('')"`                   // 执行器任务参数
	ExecutorBlockStrategy  string       `json:"executor_block_strategy" xorm:"varchar(50) default ('')"`           // 执行器阻塞策略
	ExecutorTimeout        int          `json:"executor_timeout" xorm:"int(11) notnull default (0)"`               // 执行器任务超时时间
	ExecutorFailRetryCount int          `json:"executor_fail_retry_count" xorm:"int(11) notnull default (0)"`      // 失败重试次数
	GlueType               string       `json:"glue_type" xorm:"varchar(50) notnull"`                              // 任务代码方式
	GlueSource             string       `json:"glue_source" xorm:"varchar(256)"`                                   // 任务代码源
	GlueRemark             string       `json:"glue_remark" xorm:"varchar(256) default ('')"`                      // 任务代码备注
	GlueUpdatetime         time.Time    `json:"glue_updatetime" xorm:"datetime default null"`                      // GLUE更新时间
	ChildTaskid            string       `json:"child_taskid" xorm:"varchar(255) default null"`                     // 子任务id，多个逗号分隔
	Status                 Status       `json:"status" xorm:"tinyint(4) notnull index default (0)"`                // 状态 1:运行 0:停止
	TriggerLastTime        time.Time    `json:"trigger_last_time" xorm:"bigint(13) notnull default (0)"`           // 上次调度时间
	TriggerNextTime        time.Time    `json:"trigger_next_time" xorm:"bigint(13) notnull default (0)"`           // 下次调度时间
	AddTime                time.Time    `json:"add_time" xorm:"datetime created"`                                  // 任务添加时间
	UpdateTime             time.Time    `json:"update_time" xorm:"datetime default null"`                          // 任务更新时间
}

func taskTableName() string {
	return "task"
}

func (task *Task) Create() (insertId int, err error) {
	_, err = Db.Insert(task)
	if err != nil {
		insertId = task.Id
	}
	return
}

func (task *Task) Update(id int, data CommonMap) (int64, error) {
	return Db.Table(task).Update(data)
}

//func (task *Task) Delelte(id int) (int64, error) {
//}
