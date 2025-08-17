package filefolder

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

// DeleteFileFolderService 删除文件夹服务结构体
type DeleteFileFolderService struct{}

// DeleteFileFolder 临时删除文件夹，当添加大小模型时此函数将被更新
func (service *DeleteFileFolderService) DeleteFileFolder(userId string, fileFolderId string) serializer.Response {
	// 检查用户权限是否匹配此文件夹
	var fileFolder model.FileFolder
	var err error
	if err := model.DB.Where("uuid = ?", fileFolderId).Find(&fileFolder).Error; err != nil {
		logger.Log().Error("[DeleteFileFolderService.DeleteFileFolder] 查找文件夹信息失败: ", err)
		return serializer.DBErr("", err)
	}
	if fileFolder.OwnerID != userId {
		return serializer.NotAuthErr("")
	}

	// 从列表中删除文件夹并防止文件夹重复删除
	if fileFolder.ParentFolderID == "root" || fileFolder.ParentFolderID == "" {
		return serializer.ParamsErr("CanDeleteRoot", nil)
	}
	t := model.DB.Begin()
	defer func() {
		if err != nil {
			t.Rollback()
		} else {
			t.Commit()
		}
	}()

	// 删除文件夹和其中的文件
	deleteFileFolderIDs := []string{}
	deleteFileFolderIDs = append(deleteFileFolderIDs, fileFolderId)
	for len(deleteFileFolderIDs) > 0 {
		deleteFileFolders := []model.FileFolder{}
		deleteIDs := []string{}
		// 获取要删除文件夹中的子文件夹
		if err := t.Select("uuid").Where("parent_folder_id in (?)", deleteFileFolderIDs).Find(&deleteFileFolders).Error; err != nil {
			logger.Log().Error("[DeleteFileFolderService.DeleteFileFolder] 查找要删除的文件夹信息失败: ", err)
			return serializer.DBErr("", err)
		}
		// 获取将要删除的文件夹ID
		for _, filefolder := range deleteFileFolders {
			deleteIDs = append(deleteIDs, filefolder.Uuid)
		}
		// 删除当前批次的文件夹
		if err := t.Where("uuid in (?)", deleteFileFolderIDs).Delete(&model.FileFolder{}).Error; err != nil {
			logger.Log().Error("[DeleteFileFolderService.DeleteFileFolder] 删除文件夹失败: ", err)
			return serializer.DBErr("", err)
		}
		// 删除当前批次文件夹内的文件
		if err := t.Where("parent_folder_id in (?)", deleteFileFolderIDs).Delete(&model.File{}).Error; err != nil {
			logger.Log().Error("[DeleteFileFolderService.DeleteFileFolder] 删除文件失败: ", err)
			return serializer.DBErr("", err)
		}
		// 下一轮处理子文件夹
		deleteFileFolderIDs = deleteIDs
	}

	// 从父文件夹中减去删除文件夹的大小
	if fileFolder.ParentFolderID != "root" {
		var parentFileFolder model.FileFolder
		if err := t.Where("uuid = ?", fileFolder.ParentFolderID).Find(&parentFileFolder).Error; err != nil {
			logger.Log().Error("[DeleteFileFolderService.DeleteFileFolder] 查找文件夹信息失败: ", err)
			return serializer.DBErr("", err)
		}
		if err := parentFileFolder.SubFileFolderSize(t, fileFolder.Size); err != nil {
			logger.Log().Error("[DeleteFileFolderService.DeleteFileFolder] 更新父文件夹信息失败: ", err)
			return serializer.DBErr("", err)
		}
	}

	// 从用户存储空间中减去文件夹大小
	var userStore model.FileStore
	if err := t.Where("uuid = ? and owner_id = ?", fileFolder.FileStoreID, userId).Find(&userStore).Error; err != nil {
		logger.Log().Error("[DeleteFileFolderService.DeleteFileFolder] 查找文件存储信息失败: ", err)
		return serializer.DBErr("", err)
	}
	userStore.SubCurrentSize(fileFolder.Size)
	if err = t.Save(&userStore).Error; err != nil {
		logger.Log().Error("[DeleteFileFolderService.DeleteFileFolder] 更新文件存储信息失败: ", err)
		return serializer.DBErr("", err)
	}

	return serializer.Success(nil)
}
