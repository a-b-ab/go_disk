package file

import (
	"fmt"
	"time"

	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"

	"gorm.io/gorm"
)

// RecycleBinService 回收站服务
type RecycleBinService struct{}

// GetRecycleBinListService 获取回收站列表服务
type GetRecycleBinListService struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

// RecycleBinConfigService 回收站配置服务
type RecycleBinConfigService struct {
	AutoCleanDays       int64 `json:"auto_clean_days" binding:"min=1,max=365"`
	MaxCapacityMB       int64 `json:"max_capacity_mb" binding:"min=100"`
	EnableAutoClean     int   `json:"enable_auto_clean"`
	EnableCapacityClean int   `json:"enable_capacity_clean"`
}

// GetRecycleBinList 获取用户回收站列表
func (service *GetRecycleBinListService) GetRecycleBinList(userID string) serializer.Response {
	if service.Page <= 0 {
		service.Page = 1
	}
	if service.PageSize <= 0 || service.PageSize > 100 {
		service.PageSize = 10
	}

	var recycleBins []model.RecycleBin
	var total int64

	// 查询总数
	if err := model.DB.Model(&model.RecycleBin{}).Where("user_id = ? AND is_restored = 1", userID).Count(&total).Error; err != nil {
		logger.Log().Error("[GetRecycleBinList] 查询回收站总数失败: ", err)
		return serializer.DBErr("查询回收站失败", err)
	}

	// 分页查询
	offset := (service.Page - 1) * service.PageSize
	if err := model.DB.Where("user_id = ? AND is_restored = false", userID).
		Order("deleted_at DESC").
		Offset(offset).
		Limit(service.PageSize).
		Find(&recycleBins).Error; err != nil {
		logger.Log().Error("[GetRecycleBinList] 查询回收站列表失败: ", err)
		return serializer.DBErr("查询回收站失败", err)
	}

	// 计算总容量
	var totalSize int64
	model.DB.Model(&model.RecycleBin{}).
		Where("user_id = ? AND is_restored = 0", userID).
		Select("COALESCE(SUM(size),0)").Row().Scan(&totalSize)

	return serializer.Response{
		Code: 200,
		Data: map[string]interface{}{
			"list":       recycleBins,
			"total":      total,
			"page":       service.Page,
			"page_size":  service.PageSize,
			"total_size": totalSize,
		},
		Msg: "查询回收站成功",
	}
}

// EmptyRecycleBin 清空回收站
func (service *RecycleBinService) EmptyRecycleBin(userID string) serializer.Response {
	// 开始事务
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取用户回收站中的所有文件
	var recycleBins []model.RecycleBin
	if err := tx.Where("user_id = ? AND is_restored = 0", userID).Find(&recycleBins).Error; err != nil {
		tx.Rollback()
		logger.Log().Error("[EmptyRecycleBin] 查询回收站失败: ", err)
		return serializer.DBErr("查询回收站失败", err)
	}

	// 处理每个文件
	for _, rb := range recycleBins {
		// 检查文件引用计数
		var file model.File
		if err := tx.Select("ref_count, file_uuid").Where("uuid = ?", rb.FileID).First(&file).Error; err != nil {
			logger.Log().Error(fmt.Sprintf("[EmptyRecycleBin] 查询文件失败: FileID=%s, Error=%v", rb.FileID, err))
			continue
		}

		// 如果引用计数<=1，可以物理删除
		if file.RefCount <= 1 {
			if err := service.schedulePhysicalDeletion(file.FileUuid); err != nil {
				logger.Log().Error(fmt.Sprintf("[processExpiredFile] 调度物理删除失败: FileUuid=%s, Error=%v", file.FileUuid, err))
			}
		}
	}

	// 删除回收站记录
	if err := tx.Where("user_id = ? AND is_restored = 0", userID).Delete(&model.RecycleBin{}).Error; err != nil {
		tx.Rollback()
		logger.Log().Error("[EmptyRecycleBin] 删除回收站记录失败: ", err)
		return serializer.DBErr("清空回收站失败", err)
	}

	if err := tx.Commit().Error; err != nil {
		logger.Log().Error("[EmptyRecycleBin] 提交事务失败: ", err)
		return serializer.DBErr("清空回收站失败", err)
	}

	logger.Log().Info(fmt.Sprintf("[EmptyRecycleBin] 用户回收站已清空: UserID=%s", userID))
	return serializer.Response{Code: 0, Msg: "回收站已清空"}
}

// GetRecycleBinConfig 获取回收站配置
func (service *RecycleBinService) GetRecycleBinConfig(userID string) serializer.Response {
	var config model.RecycleBinConfig
	err := model.DB.Where("user_id = ?", userID).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建默认配置
			config = model.RecycleBinConfig{
				UserID:              userID,
				AutoCleanDays:       30,
				MaxCapacityMB:       102400,
				EnableAutoClean:     1,
				EnableCapacityClean: 1,
			}
			if err := model.DB.Create(&config).Error; err != nil {
				logger.Log().Error("[GetRecycleBinConfig] 创建默认配置失败: ", err)
				return serializer.DBErr("获取配置失败", err)
			}
		} else {
			logger.Log().Error("[GetRecycleBinConfig] 查询配置失败: ", err)
			return serializer.DBErr("获取配置失败", err)
		}
	}
	return serializer.Response{
		Code: 0,
		Data: config,
		Msg:  "获取配置成功",
	}
}

// UpdateRecycleBinConfig 更新回收站配置
func (service *RecycleBinConfigService) UpdateRecycleBinConfig(userID string) serializer.Response {
	var config model.RecycleBinConfig
	err := model.DB.Where("user_id = ?", userID).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新配置
			config = model.RecycleBinConfig{
				UserID:              userID,
				AutoCleanDays:       service.AutoCleanDays,
				MaxCapacityMB:       service.MaxCapacityMB,
				EnableAutoClean:     service.EnableAutoClean,
				EnableCapacityClean: service.EnableCapacityClean,
			}
			if err := model.DB.Create(&config).Error; err != nil {
				logger.Log().Error("[UpdateRecycleBinConfig] 创建配置失败: ", err)
				return serializer.DBErr("更新配置失败", err)
			}
		} else {
			logger.Log().Error("[UpdateRecycleBinConfig] 查询配置失败: ", err)
			return serializer.DBErr("更新配置失败", err)
		}
	} else {
		// 更新配置
		config.AutoCleanDays = service.AutoCleanDays
		config.MaxCapacityMB = service.MaxCapacityMB
		config.EnableAutoClean = service.EnableAutoClean
		config.EnableCapacityClean = service.EnableCapacityClean
		if err := model.DB.Save(&config).Error; err != nil {
			logger.Log().Error("[UpdateRecycleBinConfig] 更新配置失败: ", err)
			return serializer.DBErr("更新配置失败", err)
		}
	}

	logger.Log().Info(fmt.Sprintf("[UpdateRecycleBinConfig] 回收站配置已更新: UserID=%s", userID))
	return serializer.Response{Code: 0, Msg: "配置更新成功"}
}

// AutoCleanExpiredFiles 自动清理过期文件
func (service *RecycleBinService) AutoCleanExpiredFiles() error {
	// 查询所有启用自动清理的用户配置
	var configs []model.RecycleBinConfig
	if err := model.DB.Where("enable_auto_clean = 1").Find(&configs).Error; err != nil {
		logger.Log().Error("[AutoCleanExpiredFiles] 查询用户配置失败: ", err)
		return err
	}

	for _, config := range configs {
		// 查询过期的回收站文件
		expireTime := time.Now().AddDate(0, 0, -int(config.AutoCleanDays))
		var expiredFiles []model.RecycleBin
		if err := model.DB.Where("user_id = ? AND is_restored = 0 AND deleted_at <= ?", config.UserID, expireTime).Find(&expiredFiles).Error; err != nil {
			logger.Log().Error(fmt.Sprintf("[AutoCleanExpiredFiles] 查询过期文件失败: UserID=%s, Error=%v", config.UserID, err))
			continue
		}

		// 处理过期文件
		for _, file := range expiredFiles {
			if err := model.DB.Delete(&file).Error; err != nil {
				logger.Log().Error(fmt.Sprintf("[AutoCleanExpiredFiles] 删除过期文件失败: UserID=%s, FileID=%s, Error=%v", config.UserID, file.FileID, err))
			}
		}
		logger.Log().Info(fmt.Sprintf("[AutoCleanExpiredFiles] 用户过期文件清理完成: UserID=%s, Count=%d", config.UserID, len(expiredFiles)))
	}
	return nil
}

// processExpiredFile 处理过期文件
func (service *RecycleBinService) processExpiredFile(recycleBin model.RecycleBin) error {
	// 开始事务
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 检查文件引用计数
	var file model.File
	if err := tx.Select("ref_count, file_uuid").Where("uuid = ?", recycleBin.FileID).First(&file).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("查询文件失败: %v", err)
	}

	// 如果引用计数<=1，可以物理删除
	if file.RefCount <= 1 {
		if err := service.schedulePhysicalDeletion(file.FileUuid); err != nil {
			logger.Log().Error(fmt.Sprintf("[processExpiredFile] 调度物理删除失败: FileUuid=%s, Error=%v", file.FileUuid, err))
		}
	}

	// 删除回收站记录
	if err := tx.Delete(&recycleBin).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除回收站记录失败: %v", err)
	}

	return tx.Commit().Error
}

// schedulePhysicalDeletion 调度物理删除
func (service *RecycleBinService) schedulePhysicalDeletion(fileUuid string) error {
	// 这里应该发送到延迟队列进行物理删除
	// 暂时用日志记录
	logger.Log().Info(fmt.Sprintf("[schedulePhysicalDeletion] 调度物理删除文件: FileUuid=%s", fileUuid))
	return nil
}

// AutoCleanByCapacity 按容量自动清理
func (service *RecycleBinService) AutoCleanByCapacity() error {
	// 查询所有启用容量清理的用户配置
	var configs []model.RecycleBinConfig
	if err := model.DB.Where("enable_capacity_clean = 1").Find(&configs).Error; err != nil {
		logger.Log().Error("[AutoCleanByCapacity] 查询配置失败: ", err)
		return err
	}

	for _, config := range configs {
		// 查询用户回收站总容量
		var totalSize int64
		model.DB.Model(&model.RecycleBin{}).
			Where("user_id = ? AND is_restored = 0", config.UserID).
			Select("COALESCE(SUM(size), 0)").Row().Scan(&totalSize)

		// 如果超过限制，删除最早的文件
		if totalSize > config.MaxCapacityMB*1024*1024 {
			if err := service.cleanOldestFiles(config.UserID, totalSize, config.MaxCapacityMB*1024*1024); err != nil {
				logger.Log().Error(fmt.Sprintf("[AutoCleanByCapacity] 清理最早文件失败: UserID=%s, Error=%v", config.UserID, err))
			}
		}
	}

	return nil
}

// cleanOldestFiles 清理最早的文件
func (service *RecycleBinService) cleanOldestFiles(userID string, currentSize, maxSize int64) error {
	needToFree := currentSize - maxSize

	var files []model.RecycleBin
	if err := model.DB.Where("user_id = ? AND is_restored = 0", userID).
		Order("deleted_at ASC").Find(&files).Error; err != nil {
		return fmt.Errorf("查询文件失败: %v", err)
	}

	var freedSize int64
	for _, file := range files {
		if freedSize >= needToFree {
			break
		}

		if err := service.processExpiredFile(file); err != nil {
			logger.Log().Error(fmt.Sprintf("[cleanOldestFiles] 处理文件失败: FileID=%s, Error=%v", file.FileID, err))
			continue
		}

		freedSize += file.Size
	}

	logger.Log().Info(fmt.Sprintf("[cleanOldestFiles] 容量清理完成: UserID=%s, FreedSize=%d", userID, freedSize))
	return nil
}
