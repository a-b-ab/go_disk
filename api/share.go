package api

import (
	"go-cloud-disk/serializer"
	"go-cloud-disk/service/share"
	"github.com/gin-gonic/gin"
)

// CreateShare 使用文件ID和用户ID创建分享
func CreateShare(c *gin.Context) {
	var service share.ShareCreateService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	userId := c.MustGet("UserId").(string)
	res := service.CreateShare(userId)
	c.JSON(200, res)
}

// GetShareInfo 根据分享ID获取分享信息，增加分享查看次数
func GetShareInfo(c *gin.Context) {
	var service share.ShareGetInfoService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	shareId := c.Param("shareId")
	res := service.GetShareInfo(shareId)
	c.JSON(200, res)
}

// GetUserAllShare 获取用户的所有分享信息
func GetUserAllShare(c *gin.Context) {
	var service share.ShareGetAllService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	userId := c.MustGet("UserId").(string)
	res := service.GetAllShare(userId)
	c.JSON(200, res)
}

// DeleteShare 根据分享ID删除分享
func DeleteShare(c *gin.Context) {
	var service share.ShareDeleteService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	shareId := c.Param("shareId")
	userId := c.MustGet("UserId").(string)
	res := service.DeleteShare(shareId, userId)
	c.JSON(200, res)
}

// ShareSaveFile 将分享的文件保存到用户文件夹
func ShareSaveFile(c *gin.Context) {
	var service share.ShareSaveFileService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	userId := c.MustGet("UserId").(string)
	res := service.ShareSaveFile(userId)
	c.JSON(200, res)
}
