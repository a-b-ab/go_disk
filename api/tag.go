package api

import (
	"go-cloud-disk/serializer"
	tagService "go-cloud-disk/service/tag"

	"github.com/gin-gonic/gin"
)

// AutoTagFile 自动为文件添加标签
func AutoTagFile(c *gin.Context) {
	var service tagService.AutoGetTag
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	res := service.GetAutoTags(c)
	c.JSON(200, res)
}

// CreateTag 新建标签
func CreateTag(c *gin.Context) {
	var service tagService.TagCreateService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}
	c.JSON(200, service.CreateTag())
}

// ListTag 获取标签列表
func ListTag(c *gin.Context) {
	c.JSON(200, tagService.TagListService{}.ListTags())
}

// GetTag 查询单个标签
func GetTag(c *gin.Context) {
	tagID := c.Param("tagId")
	c.JSON(200, tagService.TagListService{}.GetTag(tagID))
}

// UpdateTag 修改标签
func UpdateTag(c *gin.Context) {
	var service tagService.TagUpdateService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	tagID := c.Param("tagId")
	c.JSON(200, service.UpdateTag(tagID))
}

// DeleteTag 删除标签
func DeleteTag(c *gin.Context) {
	tagID := c.Param("tagId")
	c.JSON(200, tagService.TagDeleteService{}.DeleteTag(tagID))
}
