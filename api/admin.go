package api

import (
	"go-cloud-disk/serializer"
	"go-cloud-disk/service/admin"

	"github.com/gin-gonic/gin"
)

// UpdateUserAuth 更改用户权限
func UpdateUserAuth(c *gin.Context) {
	var service admin.UserChangeAuthService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	userStatus := c.MustGet("Status").(string)
	res := service.UserChangeAuth(userStatus)
	c.JSON(200, res)
}

// SearchUser 根据uuid、用户名或状态搜索用户
func SearchUser(c *gin.Context) {
	var service admin.UserSearchService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	res := service.UserSearch()
	c.JSON(200, res)
}

// UserFileStoreUpdate 更新用户文件存储
func UserFileStoreUpdate(c *gin.Context) {
	var service admin.UserFilestoreUpdateService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	res := service.UserFilestoreUpdate()
	c.JSON(200, res)
}

// SearchShare 根据uuid、标题或拥有者搜索分享
func SearchShare(c *gin.Context) {
	var service admin.ShareSearchService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	res := service.ShareSearch()
	c.JSON(200, res)
}

// AdminDeleteShare 根据uuid删除分享
func AdminDeleteShare(c *gin.Context) {
	var service admin.ShareDeleteService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	shareId := c.Param("shareId")
	res := service.ShareDelete(shareId)
	c.JSON(200, res)
}

// AdminDeleteFile 删除数据库中相同md5码的所有文件，不删除云端文件
func AdminDeleteFile(c *gin.Context) {
	var service admin.FileDeleteService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	userStatus := c.MustGet("Status").(string)
	fileId := c.Param("fileId")
	res := service.FileDelete(userStatus, fileId)
	c.JSON(200, res)
}

// AdminGetFileStoreInfo 根据用户ID获取文件存储信息
func AdminGetFileStoreInfo(c *gin.Context) {
	var service admin.FileStoreGetInfoService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	userId := c.Param("userId")
	res := service.FileStoreGetInfo(userId)
	c.JSON(200, res)
}
