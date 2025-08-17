package share

import (
	"time"

	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils"
	"go-cloud-disk/utils/logger"
)

// ShareCreateService 创建分享服务结构体
type ShareCreateService struct {
	FileId string `json:"fileid" form:"fileid" binding:"required"` // 文件ID
	Title  string `json:"title" form:"title" binding:"required"`   // 分享标题
}

// createShareSuccessResponse 创建分享成功响应结构体
type createShareSuccessResponse struct {
	ShareId string `json:"shareid"` // 分享ID
}

// CreateShare 创建文件分享
func (service *ShareCreateService) CreateShare(userId string) serializer.Response {
	// 检查文件所有者
	var shareFile model.File
	if err := model.DB.Where("uuid = ? and owner = ?", service.FileId, userId).Find(&shareFile).Error; err != nil {
		logger.Log().Error("[ShareCreateService.CreateShare] 查找文件信息失败: ", err)
		return serializer.DBErr("", err)
	}

	// 创建分享并保存到数据库
	newShare := model.Share{
		Owner:       userId,
		FileId:      service.FileId,
		Title:       service.Title,
		Size:        shareFile.Size,
		FileName:    shareFile.FileName + "." + shareFile.FilePostfix,
		SharingTime: time.Unix(time.Now().Unix(), 0).Format(utils.DefaultTimeTemplate),
	}
	if err := model.DB.Create(&newShare).Error; err != nil {
		logger.Log().Error("[ShareCreateService.CreateShare] 创建分享失败: ", err)
		return serializer.DBErr("", err)
	}

	return serializer.Success(createShareSuccessResponse{
		ShareId: newShare.Uuid,
	})
}
