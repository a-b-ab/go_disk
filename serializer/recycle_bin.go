package serializer

import (
	"time"

	"go-cloud-disk/model"
)

// RecycleBinItem 用于回收站列表
type RecycleBinItem struct {
	FileID    string    `json:"file_id"`
	FileName  string    `json:"file_name"`
	FileType  string    `json:"file_type"`
	Size      int64     `json:"size"`
	DeletedAt time.Time `json:"deleted_at"`
	ExpireAt  time.Time `json:"expire_at"`
}

// RecycleBinConfig 序列化回收站配置
type RecycleBinConfig struct {
	AutoCleanDays   int64 `json:"auto_clean_days"`
	EnableAutoClean int   `json:"enable_auto_clean"`
}

// RecycleBinListResponse 回收站列表响应
type RecycleBinListResponse struct {
	Files    []RecycleBinItem `json:"files"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	Total    int64            `json:"total"`
	Config   RecycleBinConfig `json:"config"`
}

// BuildRecycleBinItem 将文件转为回收站记录
func BuildRecycleBinItem(file model.File, expireDays int64) RecycleBinItem {
	expireAt := time.Time{}
	if file.DeletedAt != nil && expireDays > 0 {
		expireAt = file.DeletedAt.Add(time.Hour * 24 * time.Duration(expireDays))
	}
	return RecycleBinItem{
		FileID:    file.ID,
		FileName:  file.FileName,
		FileType:  file.FilePostfix,
		Size:      file.Size,
		DeletedAt: getTime(file.DeletedAt),
		ExpireAt:  expireAt,
	}
}

func getTime(ptr *time.Time) time.Time {
	if ptr == nil {
		return time.Time{}
	}
	return *ptr
}

// BuildRecycleBinItems 构建回收站文件列表
func BuildRecycleBinItems(files []model.File, expireDays int64) []RecycleBinItem {
	items := make([]RecycleBinItem, 0, len(files))
	for _, f := range files {
		items = append(items, BuildRecycleBinItem(f, expireDays))
	}
	return items
}

// BuildRecycleBinConfig 构建配置序列化
func BuildRecycleBinConfig(cfg model.RecycleBinConfig) RecycleBinConfig {
	return RecycleBinConfig{
		AutoCleanDays:   cfg.AutoCleanDays,
		EnableAutoClean: cfg.EnableAutoClean,
	}
}
