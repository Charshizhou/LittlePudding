package base

import (
	"LittlePudding/models"
	"github.com/gin-gonic/gin"
	"strconv"
)

// ParsePageAndPageSize 解析查询参数中的页数和每页数量
func ParsePageAndPageSize(ctx *gin.Context, params models.CommonMap) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil {
		page = 1
	}
	pageSize, err := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	if err != nil {
		pageSize = models.PageSize
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = models.PageSize
	}

	params["Page"] = page
	params["PageSize"] = pageSize
}
