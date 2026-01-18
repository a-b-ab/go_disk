package api

import (
	"go-cloud-disk/serializer"
	tagService "go-cloud-disk/service/tag"

	"github.com/gin-gonic/gin"
)

// GetFileTags 获取文件标签（仅图片）
func GetFileTags(c *gin.Context) {
	userID := c.MustGet("UserId").(string)
	fileID := c.Param("fileid")
	c.JSON(200, tagService.GetTagsByFile(userID, fileID))
}

// BindFileTag 绑定文件标签（仅图片）
func BindFileTag(c *gin.Context) {
	userID := c.MustGet("UserId").(string)
	fileID := c.Param("fileid")

	var service tagService.FileTagBindService
	if err := c.ShouldBindJSON(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}
	c.JSON(200, service.BindTagToFile(userID, fileID))
}

// UnbindFileTag 解绑文件标签（仅图片）
func UnbindFileTag(c *gin.Context) {
	userID := c.MustGet("UserId").(string)
	fileID := c.Param("fileid")
	tagID := c.Param("tagId")
	c.JSON(200, tagService.UnbindTagFromFile(userID, fileID, tagID))
}
