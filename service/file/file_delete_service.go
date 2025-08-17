package file

import (
	"fmt"

	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
	"gorm.io/gorm"
)

// FileDeleteService 文件删除服务结构体
type FileDeleteService struct{}

// deleteFile 通过GORM提供的事务删除文件
func deleteFile(t *gorm.DB, userFile model.File, userStore model.FileStore, userFileFolder model.FileFolder) error {
	// 从文件夹和父文件夹中减去删除文件的大小
	if err := userFileFolder.SubFileFolderSize(t, userFile.Size); err != nil {
		return fmt.Errorf("删除文件时减少文件夹大小失败：%v", err)
	}

	// 从用户存储空间中减去删除文件的大小
	userStore.SubCurrentSize(userFile.Size)
	if err := model.DB.Delete(&userFile).Error; err != nil {
		return fmt.Errorf("删除文件时删除文件记录失败：%v", err)
	}
	if err := model.DB.Save(&userStore).Error; err != nil {
		return fmt.Errorf("删除文件时更新用户存储空间失败：%v", err)
	}
	return nil
}

// FileDelete 删除文件并更新用户存储空间
func (service *FileDeleteService) FileDelete(userId string, fileid string) serializer.Response {
	var userFile model.File
	var userStore model.FileStore
	var err error
	t := model.DB.Begin()
	defer func() {
		if err != nil {
			t.Rollback()
		} else {
			t.Commit()
		}
	}()

	// 检查文件所有者
	if err = t.Where("uuid = ?", fileid).Find(&userFile).Error; err != nil {
		logger.Log().Error("[FileDeleteService.FileDelete] 查找用户文件失败")
		return serializer.DBErr("", err)
	}
	if userFile.Owner != userId {
		return serializer.NotAuthErr("")
	}
	if err = t.Where("owner_id = ?", userId).First(&userStore).Error; err != nil {
		logger.Log().Error("[FileDeleteService.FileDelete] 查找用户文件存储信息失败")
		return serializer.DBErr("", err)
	}
	var userFileFolder model.FileFolder
	if err = t.Where("uuid = ?", userFile.ParentFolderId).Find(&userFileFolder).Error; err != nil {
		logger.Log().Error("[FileDeleteService.FileDelete] 查找用户文件夹失败")
		return serializer.DBErr("", err)
	}

	// 使用事务删除文件
	if err = deleteFile(t, userFile, userStore, userFileFolder); err != nil {
		logger.Log().Error("[FileDeleteService.FileDelete] 更新用户文件存储容量失败: ", err)
		return serializer.DBErr("", err)
	}

	return serializer.Success(nil)
}
