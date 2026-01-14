package file

import (
	"errors"
	"fmt"
	"time"

	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"

	"gorm.io/gorm"
)

// RecycleBinService 回收站服务
type RecycleBinService struct{}

// RestoreFile 从回收站恢复文件（将 deleted_at 置空，并回补文件夹大小）
func (service *RecycleBinService) RestoreFile(userID, fileID string) serializer.Response {
	// 读取回收站配置，用于判断是否已过期（如果启用自动清理）
	cfg, err := ensureRecycleBinConfig(userID)
	if err != nil {
		logger.Log().Error("[RestoreFile] 获取回收站配置失败: ", err)
		return serializer.DBErr("获取回收站配置失败", err)
	}

	var file model.File
	if err := model.DB.Where("uuid = ? AND owner = ? AND deleted_at IS NOT NULL", fileID, userID).First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return serializer.ParamsErr("文件不在回收站或不存在", err)
		}
		return serializer.DBErr("查找回收站文件失败", err)
	}

	// 若启用自动清理且设置了天数，则超过过期时间不允许恢复（避免“应被清理但尚未跑任务”的边界）
	if cfg.EnableAutoClean == 1 && cfg.AutoCleanDays > 0 && file.DeletedAt != nil {
		expireAt := file.DeletedAt.Add(time.Hour * 24 * time.Duration(cfg.AutoCleanDays))
		if time.Now().After(expireAt) {
			return serializer.ParamsErr("文件已过期，无法恢复", nil)
		}
	}

	// 开始事务
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 回收站占用容量策略：恢复不会改变用户容量（current_size），因此无需容量校验/回补容量。
	// 1) deleted_at 置空（恢复）
	if err := tx.Model(&model.File{}).Where("uuid = ? AND owner = ?", fileID, userID).
		Updates(map[string]any{"deleted_at": nil}).Error; err != nil {
		tx.Rollback()
		return serializer.DBErr("恢复文件失败", err)
	}

	// 2) 回补文件夹大小（包含父级累计）
	var folder model.FileFolder
	if err := tx.Where("uuid = ? AND owner_id = ?", file.ParentFolderId, userID).First(&folder).Error; err != nil {
		tx.Rollback()
		return serializer.DBErr("查找文件夹失败", err)
	}
	if err := folder.AddFileFolderSize(tx, file.Size); err != nil {
		tx.Rollback()
		return serializer.DBErr("恢复文件夹容量失败", err)
	}

	if err := tx.Commit().Error; err != nil {
		return serializer.DBErr("恢复失败", err)
	}
	logger.Log().Info(fmt.Sprintf("[RestoreFile] 恢复成功 fileID=%s userID=%s", fileID, userID))
	return serializer.Success(nil)
}

// AutoCleanExpiredFiles 自动清理过期回收站文件（全局任务：按用户配置执行）
// 说明：
// - 当前回收站是通过 model.File.deleted_at 标记实现的
// - 采用“回收站占用容量”策略：软删除不释放容量，这里物理删除时才释放容量（扣减 current_size）
// - 物理对象（COS）删除目前仍是 todo（与 EmptyRecycleBin 保持一致）
func (service *RecycleBinService) AutoCleanExpiredFiles() error {
	now := time.Now()

	var configs []model.RecycleBinConfig
	if err := model.DB.Where("enable_auto_clean = ?", 1).Find(&configs).Error; err != nil {
		logger.Log().Error("[AutoCleanExpiredFiles] 查询回收站配置失败: %v", err)
		return err
	}
	if len(configs) == 0 {
		return nil
	}

	var lastErr error
	for _, cfg := range configs {
		if cfg.AutoCleanDays <= 0 {
			continue
		}
		expiredBefore := now.Add(-time.Hour * 24 * time.Duration(cfg.AutoCleanDays))

		// 事务：先统计要删除的容量，再删除记录，最后释放容量
		t := model.DB.Begin()
		var totalSize int64
		if err := t.Model(&model.File{}).
			Where("owner = ? AND deleted_at IS NOT NULL AND deleted_at < ?", cfg.UserID, expiredBefore).
			Select("COALESCE(SUM(size),0)").Scan(&totalSize).Error; err != nil {
			t.Rollback()
			logger.Log().Error("[AutoCleanExpiredFiles] 统计失败 user=%s days=%d err=%v", cfg.UserID, cfg.AutoCleanDays, err)
			lastErr = err
			continue
		}

		del := t.Where("owner = ? AND deleted_at IS NOT NULL AND deleted_at < ?", cfg.UserID, expiredBefore).
			Delete(&model.File{})
		if del.Error != nil {
			t.Rollback()
			logger.Log().Error("[AutoCleanExpiredFiles] 清理失败 user=%s days=%d err=%v", cfg.UserID, cfg.AutoCleanDays, del.Error)
			lastErr = del.Error
			continue
		}

		if totalSize > 0 {
			var store model.FileStore
			if err := t.Where("owner_id = ?", cfg.UserID).First(&store).Error; err != nil {
				t.Rollback()
				logger.Log().Error("[AutoCleanExpiredFiles] 查容量失败 user=%s err=%v", cfg.UserID, err)
				lastErr = err
				continue
			}
			store.SubCurrentSize(totalSize)
			if err := t.Save(&store).Error; err != nil {
				t.Rollback()
				logger.Log().Error("[AutoCleanExpiredFiles] 更新容量失败 user=%s err=%v", cfg.UserID, err)
				lastErr = err
				continue
			}
		}

		if err := t.Commit().Error; err != nil {
			logger.Log().Error("[AutoCleanExpiredFiles] 提交失败 user=%s err=%v", cfg.UserID, err)
			lastErr = err
			continue
		}

		if del.RowsAffected > 0 {
			logger.Log().Info("[AutoCleanExpiredFiles] 清理完成 user=%s days=%d deleted=%d freed=%d", cfg.UserID, cfg.AutoCleanDays, del.RowsAffected, totalSize)
		}
	}
	return lastErr
}

// GetRecycleBinListService 获取回收站列表服务
type GetRecycleBinListService struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

// RecycleBinConfigService 回收站配置服务
type RecycleBinConfigService struct {
	AutoCleanDays   int64 `json:"auto_clean_days" binding:"min=1,max=365"`
	EnableAutoClean int   `json:"enable_auto_clean"`
}

func ensureRecycleBinConfig(userID string) (model.RecycleBinConfig, error) {
	var config model.RecycleBinConfig
	if err := model.DB.Where("user_id = ?", userID).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			config = model.RecycleBinConfig{
				UserID:          userID,
				AutoCleanDays:   30,
				EnableAutoClean: 1,
			}
			if err := model.DB.Create(&config).Error; err != nil {
				return config, err
			}
			return config, nil
		}
		return config, err
	}
	return config, nil
}

// GetRecycleBinList 查询回收站文件
func (service *GetRecycleBinListService) GetRecycleBinList(userID string) serializer.Response {
	page := service.Page
	if page <= 0 {
		page = 1
	}
	pageSize := service.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	query := model.DB.Model(&model.File{}).
		Where("owner = ? AND deleted_at IS NOT NULL", userID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Log().Error("[GetRecycleBinList] 统计回收站失败: ", err)
		return serializer.DBErr("获取回收站失败", err)
	}

	var files []model.File
	if err := query.Order("deleted_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&files).Error; err != nil {
		logger.Log().Error("[GetRecycleBinList] 获取回收站列表失败: ", err)
		return serializer.DBErr("获取回收站列表失败", err)
	}

	config, err := ensureRecycleBinConfig(userID)
	if err != nil {
		logger.Log().Error("[GetRecycleBinList] 获取回收站配置失败: ", err)
		return serializer.DBErr("获取回收站配置失败", err)
	}

	response := serializer.RecycleBinListResponse{
		Files:    serializer.BuildRecycleBinItems(files, config.AutoCleanDays),
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Config:   serializer.BuildRecycleBinConfig(config),
	}
	return serializer.Success(response)
}

// EmptyRecycleBin 永久清空回收站
func (service *RecycleBinService) EmptyRecycleBin(userID string) serializer.Response {
	// 回收站占用容量策略：清空回收站时才真正释放容量（扣减 current_size）
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1) 统计要清理的总大小（用于释放容量）
	var totalSize int64
	if err := tx.Model(&model.File{}).
		Where("owner = ? AND deleted_at IS NOT NULL", userID).
		Select("COALESCE(SUM(size),0)").Scan(&totalSize).Error; err != nil {
		tx.Rollback()
		logger.Log().Error("[EmptyRecycleBin] 统计回收站大小失败: ", err)
		return serializer.DBErr("清空回收站失败", err)
	}

	// 2) 删除记录
	if err := tx.Where("owner = ? AND deleted_at IS NOT NULL", userID).
		Delete(&model.File{}).Error; err != nil {
		tx.Rollback()
		logger.Log().Error("[EmptyRecycleBin] 清空回收站失败: ", err)
		return serializer.DBErr("清空回收站失败", err)
	}

	// 3) 释放容量（扣减 current_size，避免扣成负数）
	if totalSize > 0 {
		var store model.FileStore
		if err := tx.Where("owner_id = ?", userID).First(&store).Error; err != nil {
			tx.Rollback()
			return serializer.DBErr("清空回收站失败", err)
		}
		store.SubCurrentSize(totalSize)
		if err := tx.Save(&store).Error; err != nil {
			tx.Rollback()
			return serializer.DBErr("清空回收站失败", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return serializer.DBErr("清空回收站失败", err)
	}
	// todo: 发送异步清理任务删除 COS 中的文件
	return serializer.Success(nil)
}

// GetRecycleBinConfig 返回用户的回收站配置
func (service *RecycleBinService) GetRecycleBinConfig(userID string) serializer.Response {
	config, err := ensureRecycleBinConfig(userID)
	if err != nil {
		logger.Log().Error("[GetRecycleBinConfig] 获取配置失败: ", err)
		return serializer.DBErr("获取回收站配置失败", err)
	}
	return serializer.Success(serializer.BuildRecycleBinConfig(config))
}

// UpdateRecycleBinConfig 更新用户的清理策略
func (service *RecycleBinConfigService) UpdateRecycleBinConfig(userID string) serializer.Response {
	config, err := ensureRecycleBinConfig(userID)
	if err != nil {
		logger.Log().Error("[UpdateRecycleBinConfig] 获取配置失败: ", err)
		return serializer.DBErr("获取回收站配置失败", err)
	}
	config.AutoCleanDays = service.AutoCleanDays
	config.EnableAutoClean = service.EnableAutoClean
	if err := model.DB.Save(&config).Error; err != nil {
		logger.Log().Error("[UpdateRecycleBinConfig] 保存配置失败: ", err)
		return serializer.DBErr("更新回收站配置失败", err)
	}
	return serializer.Success(serializer.BuildRecycleBinConfig(config))
}
