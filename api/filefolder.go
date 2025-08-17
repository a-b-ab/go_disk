package api

import (
	"go-cloud-disk/serializer"
	"go-cloud-disk/service/filefolder"
	"github.com/gin-gonic/gin"
)

// GetFilefolderAllFile 返回文件夹中的所有文件
func GetFilefolderAllFile(c *gin.Context) {
	var service filefolder.FileFolderGetAllFileService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	fileFolderId := c.Param("filefolderid")
	jwtUser := c.MustGet("UserId").(string)
	res := service.GetAllFile(jwtUser, fileFolderId)
	c.JSON(200, res)
}

// GetFilefolderAllFilefolder 返回文件夹中的所有子文件夹
func GetFilefolderAllFilefolder(c *gin.Context) {
	var service filefolder.FileFolderGetAllFileFolderService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	fileFolderId := c.Param("filefolderid")
	jwtUser := c.MustGet("UserId").(string)
	res := service.GetAllFileFolder(jwtUser, fileFolderId)
	c.JSON(200, res)
}

// CreateFileFolder 在用户指定的父文件夹中创建文件夹
// 并返回文件夹信息
func CreateFileFolder(c *gin.Context) {
	var service filefolder.FileFolderCreateService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	jwtUser := c.MustGet("UserId").(string)
	res := service.CreateFileFolder(jwtUser)
	c.JSON(200, res)
}

// DeleteFileFolder 根据文件夹ID删除文件夹
func DeleteFileFolder(c *gin.Context) {
	var service filefolder.DeleteFileFolderService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	jwtUser := c.MustGet("UserId").(string)
	fileFolderId := c.Param("filefolderid")
	res := service.DeleteFileFolder(jwtUser, fileFolderId)
	c.JSON(200, res)
}

// UpdateFileFolder 更新文件夹名称或文件夹位置
func UpdateFileFolder(c *gin.Context) {
	var service filefolder.FileFolderUpdateService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}
	jwtUser := c.MustGet("UserId").(string)
	res := service.UpdateFileFolderInfo(jwtUser)
	c.JSON(200, res)
}
