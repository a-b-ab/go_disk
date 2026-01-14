package file

import (
	"errors"
	"time"

	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"

	"gorm.io/gorm"
)

// RecycleBinService 回收站服务
type RecycleBinService struct{}

// AutoCleanExpiredFiles 自动清理过期回收站文件（全局任务：按用户配置执行）
// 说明：
// - 当前回收站是通过 model.File.deleted_at 标记实现的
// - LogicalDeleteFile 时已扣减用户空间，因此这里永久删除记录无需再扣减空间
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

		tx := model.DB.Where("owner = ? AND deleted_at IS NOT NULL AND deleted_at < ?", cfg.UserID, expiredBefore).
			Delete(&model.File{})
		if tx.Error != nil {
			logger.Log().Error("[AutoCleanExpiredFiles] 清理失败 user=%s days=%d err=%v", cfg.UserID, cfg.AutoCleanDays, tx.Error)
			lastErr = tx.Error
			continue
		}
		if tx.RowsAffected > 0 {
			logger.Log().Info("[AutoCleanExpiredFiles] 清理完成 user=%s days=%d deleted=%d", cfg.UserID, cfg.AutoCleanDays, tx.RowsAffected)
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
	if err := model.DB.Where("owner = ? AND deleted_at IS NOT NULL", userID).
		Delete(&model.File{}).Error; err != nil {
		logger.Log().Error("[EmptyRecycleBin] 清空回收站失败: ", err)
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
