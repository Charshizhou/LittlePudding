package task

import (
	"LittlePudding/models"
	"LittlePudding/modules/logger"
	"LittlePudding/modules/routers/base"
	"LittlePudding/modules/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type TaskForm struct {
	Id                    int                 `json:"Id,id,omitempty"`
	TaskName              string              `json:"TaskName,task_name"`
	TaskCron              string              `json:"TaskCron" json:"TaskCron,task_corm,omitempty"`            //任务定时
	TaskDesc              string              `json:"TaskDesc" json:"TaskDesc,task_desc,omitempty"`            //任务描述
	Author                string              `json:"Author,author,omitempty"`                                 //任务创建者
	ScheduleType          models.ScheduleType `json:"ScheduleType,schedule_type,omitempty"`                    //调度类型
	ScheduleConf          string              `json:"ScheduleConf,schedule_conf,omitempty"`                    //调度配置
	MisfireStrategy       string              `json:"MisfireStrategy,misfire_strategy,omitempty"`              //调度过期策略
	ExecutorRouteStrategy string              `json:"ExecutorRouteStrategy,executor_route_strategy,omitempty"` //执行器路由策略
	ExecutorId            int                 `json:"ExecutorId,executor_id,omitempty"`                        //执行器组id
	TaskParam             string              `json:"TaskParam,task_param,omitempty"`                          //任务参数
	Priority              int                 `json:"Priority,priority,omitempty"`                             //任务优先级
	ExecuteTimeout        int                 `json:"ExecuteTimeout,execute_timeout,omitempty"`
	ExecuteFailRetryCount int                 `json:"ExecuteFailRetryCount,execute_fail_retry_count,omitempty"` //失败重试次数
	TaskType              string              `json:"TaskType,task_type,omitempty"`                             //任务代码方式
	TaskRemark            string              `json:"TaskRemark,task_remark,omitempty"`                         //任务代码备注
	Status                models.Status       `json:"Status,status,omitempty"`                                  //状态 1:运行 0:停止
	NextRunTime           string              `json:"NextRunTime,omitempty"`                                    //下次执行时间
}

func reverse(tasks []models.Task) {
	for i, j := 0, len(tasks)-1; i < j; i, j = i+1, j-1 {
		tasks[i], tasks[j] = tasks[j], tasks[i]
	}
}

func GetTasks(c *gin.Context) {
	// 获取任务列表
	var tasks []models.Task
	models.Db.Find(&tasks)
	c.JSON(200, gin.H{"tasks": tasks})
}

func (f TaskForm) Error(c *gin.Context, errs []error) {
	if len(errs) == 0 {
		return
	}
	json := utils.JsonResponse{}
	content := json.CommonFailure("表单验证失败, 请检测输入")

	c.String(http.StatusBadRequest, content)
}

// Index 首页

func Index(c *gin.Context) {
	taskModel := new(models.Task)
	queryParams := parseQueryParams(c)
	total, err := taskModel.Total(queryParams)
	if err != nil {
		logger.Error(err)
	}
	tasks, err := taskModel.List(queryParams)
	if err != nil {
		logger.Error(err)
	}
	for i, item := range tasks {
		tasks[i].NextRunTime, _ = item.GetNextRunTime()
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"total": total,
		"data":  tasks,
	})
}

func Detail(c *gin.Context) {
	id := queryInt(c, "id")
	taskModel := new(models.Task)
	task, err := taskModel.Detail(id)
	if err != nil || task.Id == 0 {
		logger.Errorf("编辑任务#获取任务详情失败#任务ID-%d", id)
		c.JSON(http.StatusInternalServerError, "获取任务详情失败")
		return
	}

	c.JSON(http.StatusOK, task)
}

func Store(c *gin.Context) {
	var form TaskForm
	// 调试表单数据
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, "表单验证失败, 请检测输入")
		return
	}

	taskModel := models.Task{}
	var id = form.Id
	nameExists, err := taskModel.NameExist(form.TaskName, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	if nameExists {
		c.JSON(http.StatusBadRequest, "任务名称已存在")
		return
	}

	// 填充任务模型
	taskModel.Name = form.TaskName
	taskModel.TaskCron = strings.TrimSpace(form.TaskCron)
	taskModel.TaskDesc = form.TaskDesc
	taskModel.Author = form.Author
	taskModel.ScheduleType = form.ScheduleType
	taskModel.ScheduleConf = form.ScheduleConf
	taskModel.MisfireStrategy = form.MisfireStrategy
	taskModel.ExecutorRouteStrategy = form.ExecutorRouteStrategy
	taskModel.ExecutorId = form.ExecutorId
	taskModel.TaskParam = form.TaskParam
	taskModel.Priority = form.Priority
	taskModel.ExecuteTimeout = form.ExecuteTimeout
	taskModel.ExecuteFailRetryCount = form.ExecuteFailRetryCount
	taskModel.TaskType = form.TaskType
	taskModel.TaskRemark = form.TaskRemark

	if taskModel.ExecuteFailRetryCount > 10 || taskModel.ExecuteFailRetryCount < 0 {
		c.JSON(http.StatusBadRequest, "任务重试次数取值0-10")
		return
	}

	if id == 0 {
		// 任务添加后开始调度执行
		taskModel.Status = models.Enabled
		id, err = taskModel.Create()
	} else {
		_, err = taskModel.UpdateBean(id)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"保存失败": err,
		})
		return
	}

	status, _ := taskModel.GetStatus(id)
	if status == models.Enabled {
		nextRunTime, err := taskModel.CalculateNextRunTime()
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"计算下次执行时间失败": err,
			})
			return
		}
		_, err = taskModel.UpdateNextRunTime(id, nextRunTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"更新下次执行时间失败": err,
			})
			return
		}
	}

	c.JSON(http.StatusOK, "保存成功")
}

func Update(c *gin.Context) {
	var form TaskForm
	// 调试表单数据
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, "表单验证失败, 请检测输入")
		return
	}

	taskModel := new(models.Task)
	id := form.Id

	// 检查任务是否存在
	existingTask, err := taskModel.Detail(id)
	if err != nil || existingTask.Id == 0 {
		c.JSON(http.StatusInternalServerError, "获取任务详情失败")
		return
	}

	// 检查任务名称是否已经存在（除了当前任务）
	nameExists, err := taskModel.NameExist(form.TaskName, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	if nameExists {
		c.JSON(http.StatusBadRequest, "任务名称已存在")
		return
	}

	// 更新任务模型字段
	if form.TaskName != "" {
		taskModel.Name = form.TaskName
	}
	if form.TaskCron != "" {
		taskModel.TaskCron = strings.TrimSpace(form.TaskCron)
	}
	if form.TaskDesc != "" {
		taskModel.TaskDesc = form.TaskDesc
	}
	if form.Author != "" {
		taskModel.Author = form.Author
	}
	if form.ScheduleType != 0 {
		taskModel.ScheduleType = form.ScheduleType
	}
	if form.ScheduleConf != "" {
		taskModel.ScheduleConf = form.ScheduleConf
	}
	if form.MisfireStrategy != "" {
		taskModel.MisfireStrategy = form.MisfireStrategy
	}
	if form.ExecutorRouteStrategy != "" {
		taskModel.ExecutorRouteStrategy = form.ExecutorRouteStrategy
	}
	if form.ExecutorId != 0 {
		taskModel.ExecutorId = form.ExecutorId
	}
	if form.TaskParam != "" {
		taskModel.TaskParam = form.TaskParam
	}
	if form.Priority != 0 {
		taskModel.Priority = form.Priority
	}
	if form.ExecuteTimeout != 0 {
		taskModel.ExecuteTimeout = form.ExecuteTimeout
	}
	if form.ExecuteFailRetryCount != 0 {
		taskModel.ExecuteFailRetryCount = form.ExecuteFailRetryCount
	}
	if form.TaskType != "" {
		taskModel.TaskType = form.TaskType
	}
	if form.TaskRemark != "" {
		taskModel.TaskRemark = form.TaskRemark
	}

	// 检查任务重试次数范围
	if taskModel.ExecuteFailRetryCount > 10 || taskModel.ExecuteFailRetryCount < 0 {
		c.JSON(http.StatusBadRequest, "任务重试次数取值0-10")
		return
	}

	_, err = taskModel.UpdateBean(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"保存失败": err,
		})
		return
	}

	status, _ := taskModel.GetStatus(id)
	if status == models.Enabled {
		nextRunTime, err := taskModel.CalculateNextRunTime()
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"计算下次执行时间失败": err,
			})
			return
		}
		_, err = taskModel.UpdateNextRunTime(id, nextRunTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"更新下次执行时间失败": err,
			})
			return
		}
	}

	c.JSON(http.StatusOK, "保存成功")
}

func Remove(c *gin.Context) {
	id := queryInt(c, "id")
	taskModel := new(models.Task)
	_, err := taskModel.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	deleteId, err := taskModel.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("task deleted: %d", deleteId),
	})
}

func Enable(c *gin.Context) {
	changeStatus(c, models.Enabled)
}

func Disable(c *gin.Context) {
	changeStatus(c, models.Disabled)
}

func changeStatus(c *gin.Context, status models.Status) {
	id := queryInt(c, "id")
	taskModel := new(models.Task)
	task, _ := taskModel.Detail(id)
	_, err := task.Update(id, models.CommonMap{
		"status": status,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	if status == models.Enabled {
		location, _ := time.LoadLocation("Asia/Shanghai")
		nextRunTime, err := task.CalculateNextRunTimeWithLocation(location)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"message": fmt.Sprintf("计算下次执行时间失败 %v", err),
			})
			return
		}
		_, err = task.UpdateNextRunTime(id, nextRunTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"message": fmt.Sprintf("更新下次执行时间失败 %v", err),
			})
			return
		}
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"message": "保存成功",
	})
}

func parseQueryParams(c *gin.Context) models.CommonMap {
	var params models.CommonMap = models.CommonMap{}
	params["Id"] = queryInt(c, "id")
	params["executor_id"] = queryInt(c, "executor_id")
	params["Name"] = c.Query("name")
	status := queryInt(c, "status")
	if status >= 0 {
		status -= 1
	}
	params["Status"] = status
	base.ParsePageAndPageSize(c, params)

	return params
}

func queryInt(c *gin.Context, key string) int {
	valueStr := c.Query(key)
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0
	}
	return value
}

func Run(c *gin.Context) {
	id := queryInt(c, "id")
	taskModel := new(models.Task)
	task, err := taskModel.Detail(id)
	if err != nil || task.Id <= 0 {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": fmt.Sprintf("获取任务详情失败 %v", err),
		})
		return
	}

	err = task.Run(task.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": fmt.Sprintf("任务运行失败 %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"message": "任务已开始运行, 请到任务日志中查看结果",
	})
}
