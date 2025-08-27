package api

import (
	"go-cloud-disk/serializer"
	"go-cloud-disk/service/tag"

	"github.com/gin-gonic/gin"
)

// AutoTagFile 自动为文件添加标签
func AutoTagFile(c *gin.Context) {
	var service tag.AutoGetTag
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	res := service.GetAutoTags(c)
	c.JSON(200, res)
}
