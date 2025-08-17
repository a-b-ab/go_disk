package file

import (
	"go-cloud-disk/disk"
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils"
	"go-cloud-disk/utils/logger"
)

// FileCreateService 文件创建服务结构体
type FileCreateService struct {
	FileName       string `json:"filename" form:"filename" binding:"required"`         // 文件名
	FilePostfix    string `json:"file_postfix" form:"file_postfix" binding:"required"` // 文件后缀
	FileUuid       string `json:"file_uuid" form:"file_uuid" binding:"required"`       // 文件UUID
	ParentFolderId string `json:"folder" form:"folder" binding:"required"`             // 父文件夹ID
	Size           int64  `json:"size" form:"size" binding:"required"`                 // 文件大小
}

// CreateFile 通过使用上传URL上传文件来创建文件记录
func (service *FileCreateService) CreateFile(owner string) serializer.Response {
	// 检查文件是否已成功上传到云端
	uploadFileNameInCloud := utils.FastBuildFileName(service.FileUuid, service.FilePostfix)
	successUpload, err := disk.BaseCloudDisk.IsObjectExist(owner, "", uploadFileNameInCloud)
	if err != nil {
		return serializer.ErrorResponse(err)
	}
	if !successUpload {
		return serializer.DBErr("", nil)
	}

	// 检查文件夹权限
	var fileFolder model.FileFolder
	if err = model.DB.Where("uuid = ?", service.FileUuid).Find(&fileFolder).Error; err != nil {
		logger.Log().Error("[FileCreateService.CreateFile] 查找文件夹失败: ", err)
		return serializer.DBErr("", err)
	}

	if fileFolder.OwnerID != owner {
		return serializer.NotAuthErr("")
	}

	// 在数据库中创建文件记录
	file := model.File{
		Owner:          owner,
		FileName:       service.FileName,
		FilePostfix:    service.FilePostfix,
		FileUuid:       service.FileUuid,
		ParentFolderId: service.ParentFolderId,
		Size:           service.Size,
		FilePath:       owner,
	}

	if err = model.DB.Create(&file).Error; err != nil {
		logger.Log().Error("[FileCreateService.CreateFile] 创建文件失败: ", err)
		return serializer.DBErr("", err)
	}
	return serializer.Success(nil)
}
