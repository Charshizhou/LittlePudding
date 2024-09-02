package models

import (
	"time"
	"xorm.io/xorm"
)

type TaskLog struct {
	Id                     int64     `json:"id" xorm:"bigint(20) notnull pk autoincr"`
	TaskExecutor           int       `json:"task_executor" xorm:"int(11) notnull"`                       // 执行器主键
	TaskId                 int       `json:"task_id" xorm:"int(11) notnull"`                             // 任务主键
	ExecutorAddress        string    `json:"executor_address" xorm:"varchar(255)"`                       // 执行器地址
	ExecutorParam          string    `json:"executor_param" xorm:"varchar(512)"`                         // 执行器任务参数
	ExecutorFailRetryCount int       `json:"executor_fail_retry_count" xorm:"int(11) notnull default 0"` // 失败重试次数
	DispatchTime           time.Time `json:"dispatch_time" xorm:"datetime"`                              // 调度时间
	ExecTime               time.Time `json:"exec_time" xorm:"datetime"`                                  // 执行时间
	ExecResult             Result    `json:"exec_result" xorm:"int(11) notnull"`                         // 执行结果
	BaseModel              `xorm:"-"`
}

func (taskLog *TaskLog) Create() (insertId int64, err error) {
	_, err = Db.Insert(taskLog)
	if err == nil {
		insertId = taskLog.Id
	}

	return
}

// Update 更新
func (taskLog *TaskLog) Update(id int64, data CommonMap) (int64, error) {
	return Db.Table(taskLog).ID(id).Update(data)
}

func (taskLog *TaskLog) List(params CommonMap) ([]TaskLog, error) {
	taskLog.parsePageAndPageSize(params)
	list := make([]TaskLog, 0)
	session := Db.Desc("id")
	taskLog.parseWhere(session, params)
	err := session.Limit(taskLog.PageSize, taskLog.pageLimitOffset()).Find(&list)

	return list, err
}

// 删除指定任务的日志
func (taskLog *TaskLog) DeleteByTaskId(taskId int) (int64, error) {
	return Db.Where("task_id = ?", taskId).Delete(taskLog)
}

// 清空表
func (taskLog *TaskLog) Clear() (int64, error) {
	return Db.Where("1=1").Delete(taskLog)
}

func (taskLog *TaskLog) Total(params CommonMap) (int64, error) {
	session := Db.NewSession()
	defer session.Close()
	taskLog.parseWhere(session, params)
	return session.Count(taskLog)
}

// 解析where
func (taskLog *TaskLog) parseWhere(session *xorm.Session, params CommonMap) {
	if len(params) == 0 {
		return
	}
	taskId, ok := params["TaskId"]
	if ok && taskId.(int) > 0 {
		session.And("task_id = ?", taskId)
	}
	protocol, ok := params["Protocol"]
	if ok && protocol.(int) > 0 {
		session.And("protocol = ?", protocol)
	}
	status, ok := params["Status"]
	if ok && status.(int) > -1 {
		session.And("status = ?", status)
	}
}
