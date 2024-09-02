package routers

import (
	"LittlePudding/modules/routers/executor"
	"LittlePudding/modules/routers/log"
	"LittlePudding/modules/routers/task"
	"LittlePudding/modules/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Register(r *gin.Engine) {
	r.Static("/static", "views/static")
	r.LoadHTMLGlob("views/templates/*")
	r.GET("/", func(c *gin.Context) {
		//读取首页文件
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// 定时任务
	r.GET("/tasks", func(c *gin.Context) {
		c.HTML(http.StatusOK, "tasks.html", nil)
	})

	r.GET("/executors", func(c *gin.Context) {
		c.HTML(http.StatusOK, "executors.html", nil)
	})
	taskGroup := r.Group("/task")
	{
		taskGroup.POST("/store", task.Store)
		taskGroup.GET("/:id", task.Detail)
		taskGroup.GET("", task.Index)
		taskGroup.POST("/update", task.Update)
		taskGroup.GET("/log", log.Index)
		taskGroup.POST("/log/clear", log.Clear)
		taskGroup.POST("/remove/:id", task.Remove)
		taskGroup.POST("/enable/:id", task.Enable)
		taskGroup.POST("/disable/:id", task.Disable)
		taskGroup.GET("/run/:id", task.Run)
	}

	// 执行器
	executorGroup := r.Group("/executor")
	{
		executorGroup.GET("/:id", executor.Detail)
		executorGroup.POST("/store", executor.Store)
		executorGroup.GET("", executor.Index)
		executorGroup.GET("/all", executor.All)
		executorGroup.POST("/remove/:id", executor.Remove)
	}

	// 404 错误处理
	r.NoRoute(func(c *gin.Context) {
		jsonResp := utils.JsonResponse{}
		c.String(http.StatusNotFound, jsonResp.Failure(utils.NotFound, "您访问的页面不存在"))
	})

	// 500 错误处理
	r.NoMethod(func(c *gin.Context) {
		jsonResp := utils.JsonResponse{}
		c.String(http.StatusInternalServerError, jsonResp.Failure(utils.ServerError, "服务器内部错误, 请稍后再试"))
	})
}
