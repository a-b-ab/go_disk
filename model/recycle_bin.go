package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RecycleBin 回收站模型
type RecycleBin struct {
	ID               string    `gorm:"primarykey" json:"id"`
	UserID           string    `gorm:"not null;index" json:"user_id"`      // 用户ID
	FileID           string    `gorm:"not null;index" json:"file_id"`      // 文件ID
	OriginalFileName string    `gorm:"not null" json:"original_file_name"` // 原始文件名
	OriginalPath     string    `gorm:"not null" json:"original_path"`      // 原始路径
	Size             int64     `gorm:"not null" json:"size"`               // 文件大小
	DeletedAt        time.Time `gorm:"not null;index" json:"deleted_at"`   // 删除时间
	ExpireAt         time.Time `gorm:"not null;index" json:"expire_at"`    // 过期时间
	IsRestored       int       `gorm:"default:0" json:"is_restored"`       // 是否已恢复
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// RecycleBinConfig 回收站配置模型
type RecycleBinConfig struct {
	ID                  string    `gorm:"primarykey" json:"id"`
	UserID              string    `gorm:"unique;not null" json:"user_id"`         // 用户ID
	AutoCleanDays       int64     `gorm:"default:30" json:"auto_clean_days"`      // 用户自动清理天数
	MaxCapacityMB       int64     `gorm:"default:1024" json:"max_capacity_mb"`    // 最大容量（MB）
	EnableAutoClean     int       `gorm:"default:1" json:"enable_auto_clean"`     // 是否启用自动清理
	EnableCapacityClean int       `gorm:"default:1" json:"enable_capacity_clean"` // 是否启用超容量清理
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// BeforeCreate 在插入数据库前创建uuid
func (rb *RecycleBin) BeforeCreate(tx *gorm.DB) (err error) {
	if rb.ID == "" {
		rb.ID = uuid.New().String()
	}
	return
}

// BeforeCreate 在插入数据库前创建uuid
func (rbc *RecycleBinConfig) BeforeCreate(tx *gorm.DB) (err error) {
	if rbc.ID == "" {
		rbc.ID = uuid.New().String()
	}
	return
}
