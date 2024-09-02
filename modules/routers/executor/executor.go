package executor

import (
	"LittlePudding/models"
	"LittlePudding/modules/logger"
	"LittlePudding/modules/routers/base"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

type ExecutorForm struct {
	Id            int    `binding:"Required;Min(1)"`
	ExecutorName  string `binding:"Required;MaxSize(64)"`
	ExecutorTitle string `binding:"Required;MaxSize(128)"`
	Address       string `binding:"Required;MaxSize(64)"`
}

// All 获取所有执行器
func All(c *gin.Context) {
	execModel := models.Executor{}
	executors, err := execModel.AllList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "获取执行器列表失败",
		})
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("获取成功"),
		"data":    executors,
	})
}

func Index(c *gin.Context) {
	executorModel := new(models.Executor)
	queryParams := parseQueryParams(c)
	total, err := executorModel.Total(queryParams)
	if err != nil {
		logger.Error(err)
	}
	executors, err := executorModel.List(queryParams)
	if err != nil {
		logger.Error(err)
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"total": total,
		"data":  executors,
	})
}

// Detail 获取执行器详情
func Detail(c *gin.Context) {
	id := queryInt(c, "id")

	execModel := models.Executor{}
	err := execModel.Search(id)
	if err != nil {
		c.JSON(http.StatusNotFound, map[string]interface{}{
			"message": "执行器不存在",
		})
		return
	}

	c.JSON(http.StatusOK, execModel)
}

// Store 保存或更新执行器信息
func Store(c *gin.Context) {
	var form ExecutorForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "表单验证失败",
		})
		return
	}

	execModel := models.Executor{
		Id:            form.Id,
		ExecutorName:  form.ExecutorName,
		ExecutorTitle: form.ExecutorTitle,
		Address:       form.Address,
		UpdateTime:    time.Now(),
	}

	var err error
	if form.Id > 0 {
		_, err = execModel.UpdateBean(form.Id)
	} else {
		_, err = execModel.Create()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "保存失败",
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"message": "保存成功",
	})
}

// Remove 删除执行器
func Remove(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "无效的ID",
		})
		return
	}

	execModel := models.Executor{}
	_, err = execModel.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "删除失败",
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"message": "删除成功",
	})
}

func parseQueryParams(ctx *gin.Context) models.CommonMap {
	var params = models.CommonMap{}
	params["ExecutorId"], _ = strconv.Atoi(ctx.Query("executor_id"))
	params["ExecutorName"] = ctx.Query("executor_name")
	base.ParsePageAndPageSize(ctx, params)

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
