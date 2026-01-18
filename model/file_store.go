package model

import (
	"fmt"

	"gorm.io/gorm"
)

type FileStore struct {
	// 兼容现有数据库：历史表结构可能存在 uuid 且为 NOT NULL。
	// 为避免插入/更新时报 “Field 'uuid' doesn't have a default value”，这里显式建模 uuid。
	// 约定：Uuid 与 OwnerID 都使用用户ID（6位 Base62），保持兼容与可读性。
	Uuid        string `gorm:"column:uuid;primarykey"`
	OwnerID     string `gorm:"column:owner_id;uniqueIndex"`
	CurrentSize int64  `gorm:"column:current_size"`
	MaxSize     int64  `gorm:"column:max_size"`
}

// BeforeCreate 在插入数据库前初始化 uuid（兼容旧库要求 uuid NOT NULL）
func (fileStore *FileStore) BeforeCreate(tx *gorm.DB) (err error) {
	if fileStore.Uuid == "" {
		// 优先用 OwnerID，保证稳定映射；否则退化为直接使用 owner_id
		if fileStore.OwnerID != "" {
			fileStore.Uuid = fileStore.OwnerID
		}
	}
	return nil
}

// AddCurrentSize 增加当前存储大小
func (fileStore *FileStore) AddCurrentSize(size int64) (err error) {
	if fileStore.CurrentSize+size > fileStore.MaxSize {
		return fmt.Errorf("添加大小超过最大存储容量")
	}
	fileStore.CurrentSize += size
	return nil
}

// SubCurrentSize 减少当前存储大小
func (fileStore *FileStore) SubCurrentSize(size int64) (err error) {
	fileStore.CurrentSize = max(fileStore.CurrentSize-size, 0)
	return nil
}

// CreateFileStore 根据用户ID创建新的文件存储
func CreateFileStore(userId string) (string, error) {
	fileStore := FileStore{
		Uuid:        userId,
		OwnerID:     userId,
		CurrentSize: 0,
		MaxSize:     1024 * 1024,
	}
	if err := DB.Create(&fileStore).Error; err != nil {
		return "", err
	}
	return fileStore.Uuid, nil
}
