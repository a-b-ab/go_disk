package api

import (
	"go-cloud-disk/serializer"
	"go-cloud-disk/service/file"

	"github.com/gin-gonic/gin"
)

// ListAllFiles 列出用户全部文件（用于标签联动/添加文件）
func ListAllFiles(c *gin.Context) {
	userID := c.MustGet("UserId").(string)
	var service file.ListAllFilesService
	if err := c.ShouldBindQuery(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}
	c.JSON(200, service.ListAllFiles(userID))
}

