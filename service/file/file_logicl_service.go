package file

import (
	"fmt"
	"time"

	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"

	"gorm.io/gorm"
)

// FileRefCountService 文件逻辑删除服务
type FileRefCountService struct{}

// todo
// LogicalDeleteFile 逻辑删除文件（移到回收站）
func (service *FileRefCountService) LogicalDeleteFile(userID, fileID string) serializer.Response {
	var file model.File
	if err := model.DB.Where("uuid = ? AND owner = ?", fileID, userID).First(&file).Error; err != nil {
		logger.Log().Error("[LogicalDeleteFile] 查找文件失败: ", err)
		if err == gorm.ErrRecordNotFound {
			return serializer.ParamsErr("文件不存在", err)
		}
		return serializer.DBErr("查找文件失败", err)
	}

	// 开始事务
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 2. todo:直接在库删除
	// 只保留到“分钟”（秒/纳秒归零）
	now := time.Now().Truncate(time.Minute)
	if err := tx.Model(&file).Updates(map[string]interface{}{
		"deleted_at": &now,
	}).Error; err != nil {
		tx.Rollback()
		logger.Log().Error("[LogicalDeleteFile] 标记删除失败: ", err)
		return serializer.DBErr("逻辑删除文件失败", err)
	}

	// 4. 更新用户存储空间
	if err := tx.Model(&model.FileStore{}).Where("user_id = ?", userID).UpdateColumn("current_size", gorm.Expr("current_size - ?", file.Size)).Error; err != nil {
		tx.Rollback()
		logger.Log().Error("[LogicalDeleteFile] 更新用户存储空间失败: ", err)
		return serializer.DBErr("更新用户存储空间失败", err)
	}

	if err := tx.Commit().Error; err != nil {
		logger.Log().Error("[LogicalDeleteFile] 提交事务失败: ", err)
		return serializer.DBErr("删除文件失败", err)
	}

	logger.Log().Info(fmt.Sprintf("[LogicalDeleteFile] 文件已移到回收站: FileID=%s, UserID=%s", fileID, userID))
	return serializer.Response{Code: 0, Msg: "文件已移到回收站"}
}

// // sendDelayedCleanMessage 发送延迟清理消息
// func (service *FileRefCountService) sendDelayedCleanMessage(fileUuid string, delay time.Duration) error {
// 	// 这里应该发送到延迟队列，暂时用简单实现
// 	// 实际项目中可以使用 RabbitMQ 的延迟队列插件或者 Redis 延迟队列
// 	logger.Log().Info(fmt.Sprintf("[sendDelayedCleanMessage] 文件 %s 将在 %v 后物理删除", fileUuid, delay))
// 	return nil
// }

// RestoreFile 从回收站恢复文件
// todo
// func (service *FileRefCountService) RestoreFile(userID, recycleBinID string) serializer.Response {
// 	var recycleBin model.RecycleBin
// 	if err := model.DB.Where("id = ? AND user_id = ? AND is_restored = false", recycleBinID, userID).First(&recycleBin).Error; err != nil {
// 		logger.Log().Error("[RestoreFile] 查找回收站记录失败: ", err)
// 		if err == gorm.ErrRecordNotFound {
// 			return serializer.ParamsErr("回收站记录不存在", err)
// 		}
// 		return serializer.DBErr("查找回收站记录失败", err)
// 	}

// 	// 检查是否已过期
// 	if time.Now().After(recycleBin.ExpireAt) {
// 		return serializer.ParamsErr("文件已过期，无法恢复", nil)
// 	}

// 	// 开始事务
// 	tx := model.DB.Begin()
// 	defer func() {
// 		if r := recover(); r != nil {
// 			tx.Rollback()
// 		}
// 	}()

// 	// 1. TODO:新建文件记录
// 	now := time.Now()
// 	if err := tx.Model(&model.File{}).Where("uuid = ?", recycleBin.FileID).Updates(map[string]interface{}{
// 		"is_deleted": 0,
// 		"deleted_at": nil,
// 		"owner":      userID,
// 		"updated_at": now,
// 	}).Error; err != nil {
// 		tx.Rollback()
// 		logger.Log().Error("[RestoreFile] 恢复文件失败: ", err)
// 		return serializer.DBErr("恢复文件失败", err)
// 	}

// 	// 3. 标记回收站记录为已恢复
// 	if err := tx.Model(&recycleBin).Update("is_restored", 1).Error; err != nil {
// 		tx.Rollback()
// 		logger.Log().Error("[RestoreFile] 更新回收站记录失败: ", err)
// 		return serializer.DBErr("更新回收站记录失败", err)
// 	}

// 	// 4. 更新用户存储空间
// 	if err := tx.Model(&model.FileStore{}).Where("user_id = ?", userID).UpdateColumn("current_size", gorm.Expr("current + ?", recycleBin.Size)).Error; err != nil {
// 		tx.Rollback()
// 		logger.Log().Error("[RestoreFile] 更新用户存储空间失败: ", err)
// 		return serializer.DBErr("更新用户存储空间失败", err)
// 	}

// 	if err := tx.Commit().Error; err != nil {
// 		logger.Log().Error("[RestoreFile] 提交事务失败: ", err)
// 		return serializer.DBErr("恢复文件失败", err)
// 	}

// 	logger.Log().Info(fmt.Sprintf("[RestoreFile] 文件已从回收站恢复: FileID=%s, UserID=%s", recycleBin.FileID, userID))
// 	return serializer.Response{Code: 0, Msg: "文件已恢复"}
// }
