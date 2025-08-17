package model

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileStore struct {
	Uuid        string `gorm:"primarykey"`
	OwnerID     string `gorm:"column:owner_id"`
	CurrentSize int64
	MaxSize     int64
}

// BeforeCreate 在插入数据库前创建uuid
func (fileStore *FileStore) BeforeCreate(tx *gorm.DB) (err error) {
	if fileStore.Uuid == "" {
		fileStore.Uuid = uuid.NewString()
	}
	return
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

// CreateFileStore 根据用户ID创建新的文件存储，并返回其uuid或错误
func CreateFileStore(userId string) (string, error) {
	fileStore := FileStore{
		OwnerID:     userId,
		CurrentSize: 0,
		MaxSize:     1024 * 1024,
	}
	if err := DB.Create(&fileStore).Error; err != nil {
		return "", err
	}
	return fileStore.Uuid, nil
}
