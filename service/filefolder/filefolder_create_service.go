package filefolder

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

// FileFolderCreateService 创建文件夹服务结构体
type FileFolderCreateService struct {
	ParentFolderID string `json:"parent" form:"parent" binding:"required"` // 父文件夹ID
	FileFolderName string `json:"name" form:"name" binding:"required"`     // 文件夹名称
}

// CreateFileFolder 在用户数据库中创建文件夹
func (service *FileFolderCreateService) CreateFileFolder(userId string) serializer.Response {
	// 检查用户是否匹配
	var fileFolder model.FileFolder
	var err error
	if err = model.DB.Where("uuid = ?", service.ParentFolderID).Find(&fileFolder).Error; err != nil {
		logger.Log().Error("[FileFolderCreateService.CreateFileFolder] 获取文件夹信息失败: ", err)
		return serializer.DBErr("", err)
	}
	if fileFolder.OwnerID != userId {
		return serializer.NotAuthErr("")
	}

	// 插入文件夹到数据库
	createFilerFolder := model.FileFolder{
		FileFolderName: service.FileFolderName,
		ParentFolderID: service.ParentFolderID,
		FileStoreID:    fileFolder.FileStoreID, // 继承父文件夹的存储空间 ID
		OwnerID:        userId,
		Size:           0, // 新建文件夹默认大小为0
	}

	if err = model.DB.Create(&createFilerFolder).Error; err != nil {
		logger.Log().Error("[FileFolderCreateService.CreateFileFolder] 创建文件夹失败: ", err)
		return serializer.DBErr("", err)
	}
	return serializer.Success(serializer.BuildFileFolder(createFilerFolder))
}
