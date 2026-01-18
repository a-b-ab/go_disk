package filefolder

import (
	"errors"

	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"

	"gorm.io/gorm"
)

// DeleteFileFolderService 删除文件夹服务结构体
type DeleteFileFolderService struct{}

// DeleteFileFolder 临时删除文件夹，当添加大小模型时此函数将被更新
func (service *DeleteFileFolderService) DeleteFileFolder(userId string, fileFolderId string) serializer.Response {
	// 检查用户权限是否匹配此文件夹
	var fileFolder model.FileFolder
	var err error
	err = model.DB.Where("uuid = ?", fileFolderId).First(&fileFolder).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return serializer.ParamsErr("FolderNotFound", nil)
		}
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
	if t.Error != nil {
		return serializer.DBErr("", t.Error)
	}
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
		err = t.Select("uuid").Where("parent_folder_id in (?)", deleteFileFolderIDs).Find(&deleteFileFolders).Error
		if err != nil {
			logger.Log().Error("[DeleteFileFolderService.DeleteFileFolder] 查找要删除的文件夹信息失败: ", err)
			return serializer.DBErr("", err)
		}
		// 获取将要删除的文件夹ID
		for _, filefolder := range deleteFileFolders {
			deleteIDs = append(deleteIDs, filefolder.Uuid)
		}
		// 删除当前批次文件夹内的文件
		// 注意：如果数据库存在外键约束（file.parent_folder_id -> file_folder.uuid），必须先删文件再删文件夹
		err = t.Where("parent_folder_id in (?)", deleteFileFolderIDs).Delete(&model.File{}).Error
		if err != nil {
			logger.Log().Error("[DeleteFileFolderService.DeleteFileFolder] 删除文件失败: ", err)
			return serializer.DBErr("", err)
		}
		// 删除当前批次的文件夹
		err = t.Where("uuid in (?)", deleteFileFolderIDs).Delete(&model.FileFolder{}).Error
		if err != nil {
			logger.Log().Error("[DeleteFileFolderService.DeleteFileFolder] 删除文件夹失败: ", err)
			return serializer.DBErr("", err)
		}
		// 下一轮处理子文件夹
		deleteFileFolderIDs = deleteIDs
	}

	// 从父文件夹中减去删除文件夹的大小
	if fileFolder.ParentFolderID != "root" {
		var parentFileFolder model.FileFolder
		err = t.Where("uuid = ?", fileFolder.ParentFolderID).Find(&parentFileFolder).Error
		if err != nil {
			logger.Log().Error("[DeleteFileFolderService.DeleteFileFolder] 查找文件夹信息失败: ", err)
			return serializer.DBErr("", err)
		}
		err = parentFileFolder.SubFileFolderSize(t, fileFolder.Size)
		if err != nil {
			logger.Log().Error("[DeleteFileFolderService.DeleteFileFolder] 更新父文件夹信息失败: ", err)
			return serializer.DBErr("", err)
		}
	}

	// 从用户存储空间中减去文件夹大小
	var userStore model.FileStore
	// 兼容不同历史字段：FileStoreID 可能是 uuid 或 owner_id（当前实现通常等于 userId）
	err = t.Where("owner_id = ? OR uuid = ?", userId, fileFolder.FileStoreID).First(&userStore).Error
	if err != nil {
		logger.Log().Error("[DeleteFileFolderService.DeleteFileFolder] 查找文件存储信息失败: ", err)
		return serializer.DBErr("", err)
	}
	userStore.SubCurrentSize(fileFolder.Size)
	err = t.Save(&userStore).Error
	if err != nil {
		logger.Log().Error("[DeleteFileFolderService.DeleteFileFolder] 更新文件存储信息失败: ", err)
		return serializer.DBErr("", err)
	}

	return serializer.Success(nil)
}
