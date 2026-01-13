package model

import (
	"context"
	"math/rand"
	"time"

	"go-cloud-disk/cache"
	"go-cloud-disk/utils"

	"gorm.io/gorm"
)

type File struct {
	// 兼容现有代码/SQL：仍映射到数据库列 uuid，但代码层不再叫 Uuid（实现“废弃 uuid 字段”）
	ID             string `gorm:"primarykey;column:uuid;size:12"`
	Owner          string // 文件所有者，如果文件被删除则所有者为空
	FileName       string // 真实文件名
	FilePostfix    string
	FileUuid       string `gorm:"unique;not null"` // 云端文件使用md5作为名称
	FilePath       string // 云端文件的文件夹路径，用于保存分享文件
	ParentFolderId string
	Size           int64 // 文件大小
	IsDeleted      int   `gorm:"default:0"` // 逻辑删除标记 // 这个字段也废弃
}

// BeforeCreate 在插入数据库前创建文件ID（12位 Base62）
func (file *File) BeforeCreate(tx *gorm.DB) (err error) {
	if file.ID == "" {
		// 碰撞概率很低，但仍做一次轻量规避
		for i := 0; i < 20; i++ {
			id, e := utils.RandomBase62(12)
			if e != nil {
				return e
			}
			var exist File
			e = tx.Select("uuid").Where("uuid = ?", id).First(&exist).Error
			if e == nil {
				continue
			}
			if e != nil && e != gorm.ErrRecordNotFound {
				return e
			}
			file.ID = id
			break
		}
		if file.ID == "" {
			return gorm.ErrInvalidData
		}
	}
	return
}

// GetFileInfoFromRedis 从Redis获取文件上传路径
func GetFileInfoFromRedis(md5 string) string {
	filePath := cache.RedisClient.Get(context.Background(), cache.FileInfoStoreKey(md5)).Val()
	return filePath
}

// SaveFileUploadInfoToRedis 保存文件路径到Redis
func (file *File) SaveFileUploadInfoToRedis() {
	randTime := time.Hour*12 + time.Minute*time.Duration(rand.Intn(60))
	cache.RedisClient.Set(context.Background(), cache.FileInfoStoreKey(file.FileUuid), file.FilePath, randTime)
}
