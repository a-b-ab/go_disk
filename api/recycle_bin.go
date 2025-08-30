package api

import (
	"go-cloud-disk/serializer"
	"go-cloud-disk/service/file"

	"github.com/gin-gonic/gin"
)

// 逻辑删除文件（移到回收站）
func LogicalDeleteFile(c *gin.Context) {
	fileID := c.Param("fileid")
	userID := c.MustGet("userId").(string)

	var service file.FileRefCountService
	res := service.LogicalDeleteFile(userID, fileID)
	c.JSON(200, res)
}

// RestoreFile 从回收站恢复文件
func RestoreFile(c *gin.Context) {
	recycleBinID := c.Param("recycleBinId")
	userID := c.MustGet("userId").(string)

	var service file.FileRefCountService
	res := service.RestoreFile(userID, recycleBinID)
	c.JSON(200, res)
}

// GetRecycleBinList 获取回收站列表
func GetRecycleBinList(c *gin.Context) {
	var service file.GetRecycleBinListService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	userID := c.MustGet("UserId").(string)
	res := service.GetRecycleBinList(userID)
	c.JSON(200, res)
}

// EmptyRecycleBin 清空回收站
func EmptyRecycleBin(c *gin.Context) {
	userID := c.MustGet("UserId").(string)

	var service file.RecycleBinService
	res := service.EmptyRecycleBin(userID)
	c.JSON(200, res)
}

// GetRecycleBinConfig 获取回收站配置
func GetRecycleBinConfig(c *gin.Context) {
	userID := c.MustGet("UserId").(string)

	var service file.RecycleBinService
	res := service.GetRecycleBinConfig(userID)
	c.JSON(200, res)
}

// UpdateRecycleBinConfig 更新回收站配置
func UpdateRecycleBinConfig(c *gin.Context) {
	var service file.RecycleBinConfigService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	userID := c.MustGet("UserId").(string)
	res := service.UpdateRecycleBinConfig(userID)
	c.JSON(200, res)
}
