package model

import (
	"context"
	"math/rand"
	"time"

	"go-cloud-disk/cache"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type File struct {
	Uuid           string `gorm:"primarykey"`
	Owner          string // 文件所有者，如果文件被删除则所有者为空
	FileName       string // 真实文件名
	FilePostfix    string
	FileUuid       string `gorm:"unique;not null"` // 云端文件使用md5作为名称
	FilePath       string // 云端文件的文件夹路径，用于保存分享文件
	ParentFolderId string
	Size           int64 // 文件大小
	RefCount       int64 `gorm:"default:1"` // 文件引用计数,默认为1 // 这个字段废弃
	IsDeleted      int   `gorm:"default:0"` // 逻辑删除标记 // 这个字段也废弃
}

// BeforeCreate 在插入数据库前创建uuid
func (file *File) BeforeCreate(tx *gorm.DB) (err error) {
	if file.Uuid == "" {
		file.Uuid = uuid.New().String()
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
