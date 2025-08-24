package api

import (
	"fmt"
	"mime/multipart"
	"time"

	"go-cloud-disk/serializer"
	"go-cloud-disk/service/file"
	"go-cloud-disk/utils"

	"github.com/gin-gonic/gin"
)

// GetUploadURL 返回文件上传URL
func GetUploadURL(c *gin.Context) {
	var service file.GetUploadURLService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	userIdInJWT := c.MustGet("UserId").(string)
	res := service.GetUploadURL(userIdInJWT)
	c.JSON(200, res)
}

// CreateFile 在数据库中创建文件记录
func CreateFile(c *gin.Context) {
	var service file.FileCreateService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	userIdInJWT := c.MustGet("UserId").(string)
	res := service.CreateFile(userIdInJWT)
	c.JSON(200, res)
}

// GetDownloadURL 返回文件下载URL
func GetDownloadURL(c *gin.Context) {
	var service file.FileGetDownloadURLService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	fileId := c.Param("fileid")
	userIdInJWT := c.MustGet("UserId").(string)
	res := service.GetDownloadURL(userIdInJWT, fileId)
	c.JSON(200, res)
}

// getUploadFileParam 从请求中获取上传文件参数
func getUploadFileParam(c *gin.Context) (userId string, file *multipart.FileHeader, dst string, err error) {
	userId = c.MustGet("UserId").(string)
	file, err = c.FormFile("file")
	if err != nil {
		err = fmt.Errorf("获取上传文件错误：%v", err)
		return
	}
	// 保存文件到本地
	if file == nil {
		err = fmt.Errorf("参数中没有文件")
		return
	}

	// 简单检查文件大小是否可以上传。当允许用户上传更大文件时，应该使用用户存储空间来检查
	// 文件是否可以上传。
	// 例如，使用 file.checkIfFileSizeExceedsVolum() 来检查文件是否可以上传
	// 在这种情况下，使用简单检查来提高API速度
	if file.Size > 1024*1024*100 {
		err = fmt.Errorf("文件大小过大")
		return
	}
	// 将文件保存到指定文件夹，以便将来方便删除文件
	uploadDay := time.Now().Format("2006-01-02")
	dst = utils.FastBuildString("./user/", uploadDay, "/", userId, "/", file.Filename)
	c.SaveUploadedFile(file, dst)
	return
}

// UploadFile 上传文件到云端
func UploadFile(c *gin.Context) {
	var service file.FileUploadService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	userId, file, dst, err := getUploadFileParam(c)
	if err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}
	res := service.UploadFile(userId, file, dst)
	c.JSON(200, res)
}

// DeleteFile 删除数据库中的文件记录，不删除云端文件
func DeleteFile(c *gin.Context) {
	var service file.FileDeleteService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	fileid := c.Param("fileid")
	userId := c.MustGet("UserId").(string)
	res := service.FileDelete(userId, fileid)
	c.JSON(200, res)
}

// UpdateFile 更新文件信息，如移动文件、更新文件名
func UpdateFile(c *gin.Context) {
	var service file.FileUpdateService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	userId := c.MustGet("UserId").(string)
	res := service.UpdateFileInfo(userId)
	c.JSON(200, res)
}
