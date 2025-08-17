package file

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

// FileUpdateService 文件更新服务结构体
type FileUpdateService struct {
	FileId      string `json:"file" form:"file" binding:"required"`     // 文件ID
	FileName    string `json:"name" form:"name"`                        // 文件名
	NewParentId string `json:"parent" form:"parent" binding:"required"` // 新的父文件夹ID
}

// UpdateFileInfo 更新文件信息，包括文件名和所属文件夹
func (service *FileUpdateService) UpdateFileInfo(userId string) serializer.Response {
	var file model.File
	var err error
	if err := model.DB.Where("uuid = ?", service.FileId).Find(&file).Error; err != nil {
		logger.Log().Error("[FileUpdateService.UpdateFileInfo] 查找文件失败: ", err)
		return serializer.DBErr("", err)
	}
	if file.Owner != userId {
		return serializer.NotAuthErr("")
	}

	var nowFilefolder model.FileFolder
	if err := model.DB.Where("uuid = ?", file.ParentFolderId).Find(&nowFilefolder).Error; err != nil {
		logger.Log().Error("[FileUpdateService.UpdateFileInfo] 查找文件夹失败: ", err)
		return serializer.DBErr("", err)
	}
	// 检查目标文件夹所有者
	var parentFilefolder model.FileFolder
	if err := model.DB.Where("uuid = ?", service.NewParentId).Find(&parentFilefolder).Error; err != nil {
		logger.Log().Error("[FileUpdateService.UpdateFileInfo] 查找文件夹失败: ", err)
		return serializer.DBErr("更新文件信息时查找父文件夹失败", err)
	}
	if userId != parentFilefolder.OwnerID {
		return serializer.NotAuthErr("")
	}

	// 构建新的文件信息
	file.ParentFolderId = service.NewParentId
	newFilename := file.FileName
	if service.FileName != "" {
		newFilename = service.FileName
	}
	file.FileName = newFilename
	// 更新文件信息到数据库
	t := model.DB.Begin()
	defer func() {
		if err != nil {
			t.Rollback()
		} else {
			t.Commit()
		}
	}()
	if err := t.Save(&file).Error; err != nil {
		logger.Log().Error("[FileUpdateService.UpdateFileInfo] 更新文件失败: ", err)
		return serializer.DBErr("", err)
	}

	// 更改文件夹大小
	if nowFilefolder.Uuid != parentFilefolder.Uuid {
		err = nowFilefolder.SubFileFolderSize(t, file.Size)
		if err != nil {
			logger.Log().Error("[FileUpdateService.UpdateFileInfo] 减少文件夹大小失败: ", err)
			return serializer.DBErr("", err)
		}
		err = parentFilefolder.AddFileFolderSize(t, file.Size)
		if err != nil {
			logger.Log().Error("[FileUpdateService.UpdateFileInfo] 增加文件夹大小失败: ", err)
			return serializer.DBErr("", err)
		}
	}

	return serializer.Success(serializer.BuildFile(file))
}
