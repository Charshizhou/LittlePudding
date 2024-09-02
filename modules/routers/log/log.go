package log

import (
	"LittlePudding/models"
	"LittlePudding/modules/logger"
	"LittlePudding/modules/routers/base"
	"LittlePudding/modules/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// 获取任务日志列表
func Index(c *gin.Context) {
	logModel := new(models.TaskLog)
	queryParams := parseQueryParams(c)
	total, err := logModel.Total(queryParams)
	if err != nil {
		logger.Error(err)
	}
	logs, err := logModel.List(queryParams)
	if err != nil {
		logger.Error(err)
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"total": total,
		"data":  logs,
	})
}

// 清空日志
func Clear(c *gin.Context) {
	taskLogModel := new(models.TaskLog)
	_, err := taskLogModel.Clear()
	json := utils.JsonResponse{}
	if err != nil {
		c.JSON(http.StatusInternalServerError, json.CommonFailure(utils.FailureContent))
		return
	}

	c.JSON(http.StatusOK, json.Success(utils.SuccessContent, nil))
}

// 删除N个月前的日志
func Remove(c *gin.Context) {
	jsonResp := utils.JsonResponse{}
	month := queryInt(c, "month")
	if month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, jsonResp.CommonFailure("参数取值范围1-12"))
		return
	}
	taskLogModel := new(models.TaskLog)
	_, _ = taskLogModel.DeleteByTaskId(month)

	c.JSON(http.StatusOK, jsonResp.Success("删除成功", nil))
}

func queryInt(c *gin.Context, key string) int {
	valueStr := c.Query(key)
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0
	}
	return value
}

// 解析查询参数
func parseQueryParams(c *gin.Context) models.CommonMap {
	var params models.CommonMap = models.CommonMap{}
	params["TaskId"] = queryInt(c, "task_id")
	status := c.Query("status")
	if status != "" {
		s, err := strconv.Atoi(status)
		if err == nil {
			params["Status"] = s - 1
		}
	}
	base.ParsePageAndPageSize(c, params)

	return params
}
