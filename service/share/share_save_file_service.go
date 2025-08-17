package share

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

// ShareSaveFileService 保存分享文件服务结构体
type ShareSaveFileService struct {
	FileId         string `json:"fileid" form:"fileid" binding:"required"`         // 文件ID
	SaveFilefolder string `json:"filefolder" form:"filefolder" binding:"required"` // 保存目标文件夹
}

// ShareSaveFile 将分享的文件保存到用户的文件夹中
func (service *ShareSaveFileService) ShareSaveFile(userId string) serializer.Response {
	// 从数据库获取要保存的文件信息
	var saveFile model.File
	var err error
	if err = model.DB.Where("uuid = ?", service.FileId).Find(&saveFile).Error; err != nil {
		logger.Log().Error("[ShareSaveFileService.ShareSaveFile] 查找文件信息失败: ", err)
		return serializer.DBErr("", err)
	}

	// 从数据库获取保存目标文件夹并检查所有者
	var targetFilefolder model.FileFolder
	if err = model.DB.Where("uuid = ? and owner_id = ?", service.SaveFilefolder, userId).Find(&targetFilefolder).Error; err != nil {
		logger.Log().Error("[ShareSaveFileService.ShareSaveFile] 查找文件夹失败: ", err)
		return serializer.DBErr("", err)
	}

	// 从数据库获取用户文件存储信息
	var targetFileStore model.FileStore
	if err := model.DB.Where("uuid = ?", targetFilefolder.FileStoreID).Find(&targetFileStore).Error; err != nil {
		logger.Log().Error("[ShareSaveFileService.ShareSaveFile] 查找文件存储信息失败: ", err)
		return serializer.DBErr("", err)
	}

	// 检查添加文件大小后是否超过当前大小限制
	if targetFileStore.CurrentSize+saveFile.Size > targetFileStore.MaxSize {
		return serializer.ParamsErr("ExceedStoreLimit", nil)
	}
	// 更改文件夹大小
	targetFileStore.AddCurrentSize(saveFile.Size)
	t := model.DB.Begin()
	defer func() {
		if err != nil {
			t.Rollback()
		} else {
			t.Commit()
		}
	}()

	if err := t.Save(&targetFileStore).Error; err != nil {
		logger.Log().Error("[ShareSaveFileService.ShareSaveFile] 更新用户存储信息失败: ", err)
		return serializer.DBErr("", err)
	}
	if err := targetFilefolder.AddFileFolderSize(t, saveFile.Size); err != nil {
		logger.Log().Error("[ShareSaveFileService.ShareSaveFile] 增加文件夹大小失败: ", err)
		return serializer.DBErr("", err)
	}

	// 保存文件到文件夹
	newFile := model.File{
		Owner:          targetFileStore.OwnerID,
		FileName:       saveFile.FileName,
		FilePostfix:    saveFile.FilePostfix,
		FileUuid:       saveFile.FileUuid,
		FilePath:       saveFile.FilePath,
		Size:           saveFile.Size,
		ParentFolderId: service.SaveFilefolder,
	}
	if err := model.DB.Create(&newFile).Error; err != nil {
		logger.Log().Error("[ShareSaveFileService.ShareSaveFile] 创建文件失败: ", err)
		return serializer.DBErr("", err)
	}

	return serializer.Success(nil)
}
