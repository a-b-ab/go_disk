package file

import (
	"fmt"
	"time"

	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"

	"gorm.io/gorm"
)

// FileDeleteService 文件删除服务结构体
type FileDeleteService struct{}

// logicalDeleteFile 通过事务把文件标记为删除（写入 deleted_at），不物理删除记录
func logicalDeleteFile(t *gorm.DB, userFile model.File, userFileFolder model.FileFolder) error {
	// 1) 标记 deleted_at
	// 只保留到“分钟”（秒/纳秒归零）
	now := time.Now().Truncate(time.Minute)
	if err := t.Model(&userFile).Updates(map[string]any{
		"deleted_at": &now,
	}).Error; err != nil {
		return fmt.Errorf("删除文件时标记deleted_at失败：%v", err)
	}

	// 2) 从文件夹与父文件夹中减去文件大小（让列表/统计不再包含已删除文件）
	if err := userFileFolder.SubFileFolderSize(t, userFile.Size); err != nil {
		return fmt.Errorf("删除文件时减少文件夹大小失败：%v", err)
	}

	// 注意：采用“回收站占用容量”的策略时，软删除不释放容量；
	// 用户容量（current_size）只会在“清空回收站/过期清理”时扣减释放。
	return nil
}

// FileDelete 删除文件并更新用户存储空间
func (service *FileDeleteService) FileDelete(userId string, fileid string) serializer.Response {
	var userFile model.File
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
	// 已经在回收站的文件，避免重复扣减容量/文件夹大小
	if userFile.DeletedAt != nil {
		return serializer.Success(nil)
	}
	var userFileFolder model.FileFolder
	if err = t.Where("uuid = ?", userFile.ParentFolderId).Find(&userFileFolder).Error; err != nil {
		logger.Log().Error("[FileDeleteService.FileDelete] 查找用户文件夹失败")
		return serializer.DBErr("", err)
	}

	// 使用事务“软删除”文件（写 deleted_at）
	if err = logicalDeleteFile(t, userFile, userFileFolder); err != nil {
		logger.Log().Error("[FileDeleteService.FileDelete] 软删除失败: ", err)
		return serializer.DBErr("", err)
	}

	return serializer.Success(nil)
}
