package models

import (
	"LittlePudding/modules/rpc/client"
	pb "LittlePudding/modules/rpc/proto"
	"LittlePudding/service"
	"errors"
	"github.com/robfig/cron/v3"
	"time"
	"xorm.io/xorm"
)

type Status int8
type ScheduleType int8
type AlarmStatus int8
type Executoraddr string
type TaskResult service.TaskResult

const (
	Enabled  Status = 1
	Disabled Status = 2
)

const (
	None ScheduleType = iota
	Cron
	FixRate
	FixDelay
)

type Task struct {
	Id                    int          `json:"id" xorm:"int(11) pk autoincr"`
	Name                  string       `json:"name" xorm:"varchar(32) notnull"`
	TaskCron              string       `json:"task_cron" xorm:"varchar(64) notnull"`                              // crontab
	TaskDesc              string       `json:"task_desc" xorm:"varchar(255) notnull"`                             // 任务描述
	Author                string       `json:"author" xorm:"varchar(64) default ('')"`                            // 任务创建者
	ScheduleType          ScheduleType `json:"schedule_type" xorm:"tinyint(4) notnull default (0)"`               // 调度类型 cron,FIX_RATE,FIX_DELAY
	ScheduleConf          string       `json:"schedule_conf" xorm:"varchar(128) default ('')"`                    // 调度配置
	MisfireStrategy       string       `json:"misfire_strategy xorm::varchar(50) notnull default ('DO_NOTHING')"` // 调度过期策略
	ExecutorRouteStrategy string       `json:"executor_route_strategy" xorm:"varchar(50) default ('')"`           // 执行器路由策略
	ExecutorId            int          `json:"executor_id" xorm:"int(11) not null"`                               // 执行器组id
	TaskParam             string       `json:"task_param" xorm:"varchar(512) default ('')"`                       // 任务参数
	Priority              int          `json:"priority" xorm:"int(11) notnull default (0)"`                       // 任务优先级
	ExecuteTimeout        int          `json:"execute_timeout" xorm:"int(11) notnull default (0)"`                // 执行器任务超时时间
	ExecuteFailRetryCount int          `json:"execute_fail_retry_count" xorm:"int(11) notnull default (0)"`       // 失败重试次数
	TaskType              string       `json:"task_type" xorm:"varchar(50) notnull"`                              // 任务代码方式
	TaskRemark            string       `json:"task_remark" xorm:"varchar(256) default ('')"`                      // 任务代码备注
	Status                Status       `json:"status" xorm:"tinyint(4) notnull index default (0)"`                // 状态 1:运行 0:停止
	AddTime               time.Time    `json:"add_time" xorm:"datetime created"`                                  // 任务添加时间
	UpdateTime            time.Time    `json:"update_time" xorm:"datetime default null"`                          // 任务更新时间
	NextRunTime           time.Time    `json:"next_run_time" xorm:"datetime default null"`                        // 下次执行时间
	IsUpdated             bool         `json:"is_updated" xorm:"tinyint(1) notnull default (0)"`                  // 是否更新过
	BaseModel             `xorm:"-"`
	ExecutorAddr          string      `json:"executor_addr" xorm:"-"`
	TaskResult            *TaskResult `json:"task_result" xorm:"-"`
}

func executorTableName() string {
	return "executor"
}

// Create 新增
func (task *Task) Create() (insertId int, err error) {
	_, err = Db.Insert(task)
	if err == nil {
		insertId = task.Id
	}
	return
}

// UpdateBean 更新
func (task *Task) UpdateBean(id int) (int64, error) {
	return Db.ID(id).Cols("name,task_cron,task_desc,author,schedule_type,schedule_conf,misfire_strategy,executor_route_strategy,executor_id,task_param,priority,execute_timeout,execute_fail_retry_count,task_type,task_remark,status").Update(task)
}

// Update 更新
func (task *Task) Update(id int, data CommonMap) (int64, error) {
	return Db.Table(task).ID(id).Update(data)
}

func (task *Task) UpdateNextRunTime(id int, nextRunTime time.Time) (int64, error) {
	return Db.ID(id).Cols("next_run_time").Update(&Task{NextRunTime: nextRunTime})
}

// Delete 删除
func (task *Task) Delete(id int) (int64, error) {
	return Db.ID(id).Delete(&Task{})
}

// Disable 禁用
func (task *Task) Disable(id int) (int64, error) {
	return task.Update(id, CommonMap{"status": Disabled})
}

// Enable 激活
func (task *Task) Enable(id int) (int64, error) {
	return task.Update(id, CommonMap{"status": Enabled})
}

// ActiveList 获取所有激活任务
func (task *Task) ActiveList(page, pageSize int) ([]Task, error) {
	params := CommonMap{"Page": page, "PageSize": pageSize}
	task.parsePageAndPageSize(params)
	list := make([]Task, 0)
	err := Db.Where("status = ?", Enabled).Limit(task.PageSize, task.pageLimitOffset()).
		Find(&list)

	if err != nil {
		return list, err
	}

	return task.setExecutorsForTasks(list)
}

// ActiveListByExecutorId 获取某个主机下的所有激活任务
func (task *Task) ActiveListByExecutorId(id int) ([]Task, error) {
	list := make([]Task, 0)
	err := Db.Where("status = ? AND executor_id = ?", Enabled, id).Find(&list)

	if err != nil {
		return list, err
	}

	return task.setExecutorsForTasks(list)
}

func (task *Task) setExecutorsForTasks(tasks []Task) ([]Task, error) {
	TaskExecutor := new(Executor)
	for i, value := range tasks {
		ExecutorAddr, err := TaskExecutor.GetAddress(value.ExecutorId)
		if err != nil {
			return nil, err
		}
		tasks[i].ExecutorAddr = ExecutorAddr
	}

	return tasks, nil
}

// NameExist 判断任务名称是否存在
func (task *Task) NameExist(name string, id int) (bool, error) {
	count, err := Db.Where("id = ? AND name = ? AND status = ?", id, name, Enabled).Count(task)

	return count > 0, err
}

// GetStatus 获取状态
func (task *Task) GetStatus(id int) (Status, error) {
	exist, err := Db.ID(id).Get(task)
	if err != nil {
		return 0, err
	}
	if !exist {
		return 0, errors.New("not exist")
	}

	return task.Status, nil
}

func (task *Task) Detail(id int) (Task, error) {
	t := Task{}
	_, err := Db.Where("id=?", id).Get(&t)

	if err != nil {
		return t, err
	}
	return t, err
}

func (task *Task) List(params CommonMap) ([]Task, error) {
	task.parsePageAndPageSize(params)
	list := make([]Task, 0)
	session := Db.Alias("t").Join("LEFT", executorTableName(), "t.executor_id = executor.id")
	task.parseWhere(session, params)
	err := session.GroupBy("t.id").Asc("t.id").Cols("t.*").Limit(task.PageSize, task.pageLimitOffset()).Find(&list)

	if err != nil {
		return nil, err
	}

	return task.setExecutorsForTasks(list)
}

// Total 获取总数
func (task *Task) Total(params CommonMap) (int64, error) {
	session := Db.Alias("t").Join("LEFT", executorTableName(), "t.executor_id = executor.id")
	task.parseWhere(session, params)
	list := make([]Task, 0)

	err := session.GroupBy("t.id").Find(&list)

	return int64(len(list)), err
}

// 解析where
func (task *Task) parseWhere(session *xorm.Session, params CommonMap) {
	if len(params) == 0 {
		return
	}
	id, ok := params["Id"]
	if ok && id.(int) > 0 {
		session.And("t.id = ?", id)
	}
	hostId, ok := params["executor_id"]
	if ok && hostId.(int) > 0 {
		session.And("t.executor_id = ?", hostId)
	}
	name, ok := params["Name"]
	if ok && name.(string) != "" {
		session.And("t.name LIKE ?", "%"+name.(string)+"%")
	}
	status, ok := params["Status"]
	if ok && status.(int) > -1 {
		session.And("status = ?", status)
	}
}

func (task *Task) CalculateNextRunTime() (time.Time, error) {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(task.TaskCron)
	if err != nil {
		return time.Time{}, err
	}

	return schedule.Next(time.Now()).In(time.Local), nil
}

func (task *Task) CalculateNextRunTimeWithLocation(loc *time.Location) (time.Time, error) {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	schedule, err := parser.Parse(task.TaskCron)
	if err != nil {
		return time.Time{}, err
	}
	now := time.Now().In(loc)
	nextRunTime := schedule.Next(now)
	return nextRunTime, nil
}

func (task *Task) GetNextRunTime() (time.Time, error) {
	t := Task{}
	_, err := Db.Where("id=?", task.Id).Get(&t)

	if err != nil {
		return time.Time{}, err
	}
	return t.NextRunTime, nil
}

// CreateTaskLog 创建任务日志
func CreateTaskLog(taskModel Task) (int64, error) {
	taskLogModel := new(TaskLog)
	taskLogModel.TaskId = taskModel.Id
	taskLogModel.TaskExecutor = taskModel.ExecutorId
	taskLogModel.TaskId = taskModel.Id
	taskLogModel.ExecutorAddress = taskModel.ExecutorAddr
	taskLogModel.ExecutorParam = taskModel.TaskParam
	taskLogModel.ExecutorFailRetryCount = taskModel.ExecuteFailRetryCount
	taskLogModel.DispatchTime = taskModel.TaskResult.DispatchTime
	taskLogModel.ExecTime = taskModel.TaskResult.ExecTime
	taskLogModel.ExecResult = Result(taskModel.TaskResult.Result)
	insertId, err := taskLogModel.Create()

	return insertId, err
}

func (task *Task) Run(id int) error {
	t := Task{}
	_, err := Db.Where("id=?", id).Get(&t)

	if err != nil {
		return err
	}

	req := &pb.TaskRequest{
		Id:            int32(t.Id),
		Priority:      int32(t.Priority),
		ExecTime:      time.Now().Unix(),
		RouteStrategy: t.ExecutorRouteStrategy,
		TaskType:      t.TaskType,
		TaskParam:     t.TaskParam,
		TaskTimeout:   int64(t.ExecuteTimeout),
	}

	// 发送请求
	result, err := client.Exec("127.0.0.1:50051", req)
	if result == nil {
		result = &service.TaskResult{
			Result: service.Failure,
		}
	}
	if task.ExecuteFailRetryCount > 0 && result.Result == service.Failure {
		for i := 0; i < task.ExecuteFailRetryCount; i++ {
			result, err = client.Exec(task.ExecutorAddr, req)
			if result == nil {
				result = &service.TaskResult{
					Result: service.Failure,
				}
			}
			if err == nil && result.Result == service.Success {
				task.ExecuteFailRetryCount = i + 1
				break
			}
		}
	} else {
		task.ExecuteFailRetryCount = 0
	}
	task.TaskResult = (*TaskResult)(result)
	if _, err := CreateTaskLog(*task); err != nil {
		return err
	}
	return nil
}
