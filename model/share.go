package model

import (
	"context"
	"fmt"
	"strconv"

	"go-cloud-disk/cache"
	"go-cloud-disk/disk"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Share struct {
	Uuid        string `gorm:"primarykey"`
	Owner       string
	FileId      string // 分享文件的文件uuid
	FileName    string
	Title       string
	Size        int64
	SharingTime string
}

// SetEmptyShare 设置空分享,表示分享链接已失效
func (share *Share) SetEmptyShare() {
	// 从日排行榜中移除分享，并将分享添加到空分享集合中
	if share.DailyViewCount() > 10 {
		cache.RedisClient.ZRem(context.Background(), cache.DailyRankKey, share.Uuid)
		cache.RedisClient.SAdd(context.Background(), cache.EmptyShare, share.Uuid)
	}

	share.Owner = ""
	share.FileId = ""
	share.FileName = ""
	share.Title = "来晚了,分享的文件已被删除"
	share.Size = 0
	share.SharingTime = ""
}

// BeforeCreate 在插入数据库前创建uuid
func (file *Share) BeforeCreate(tx *gorm.DB) (err error) {
	if file.Uuid == "" {
		file.Uuid = uuid.New().String()
	}
	return
}

// DownloadURL 获取分享下载链接
func (share *Share) DownloadURL() (string, error) {
	var file File
	if err := DB.Where("uuid = ?", share.FileId).Find(&file).Error; err != nil {
		return "", fmt.Errorf("构建下载链接时查找用户文件失败 %v", err)
	}

	url, err := disk.BaseCloudDisk.GetObjectURL(file.FilePath, "", file.FileUuid+"."+file.FilePostfix)
	if err != nil {
		return "", fmt.Errorf("获取分享下载链接时获取对象URL失败，%v", err)
	}
	return url, nil
}

// ViewCount 从Redis获取分享查看次数
func (share *Share) ViewCount() (num int64) {
	countStr, _ := cache.RedisClient.Get(context.Background(), cache.ShareKey(share.Uuid)).Result()
	if countStr == "" {
		return 0
	}
	num, _ = strconv.ParseInt(countStr, 10, 64)
	return
}

// DailyViewCount 根据分享uuid获取日查看次数
func (share *Share) DailyViewCount() float64 {
	countStr := cache.RedisClient.ZScore(context.Background(), cache.DailyRankKey, share.Uuid).Val()
	return countStr
}

// AddViewCount 在Redis中增加分享查看次数
func (share *Share) AddViewCount() {
	// 1. Redis 中单独的访问计数器自增
	cache.RedisClient.Incr(context.Background(), cache.ShareKey(share.Uuid))
	// 2. Redis 中每日排行榜（有序集合）对应的分数自增 1
	cache.RedisClient.ZIncrBy(context.Background(), cache.DailyRankKey, 1, share.Uuid)
}

// SaveShareInfoToRedis 保存分享信息到Redis
func (share *Share) SaveShareInfoToRedis(downloadUrl string) error {
	ctx := context.Background()
	// 如果Owner不为空，说明函数已经写入到Redis中
	if s := cache.RedisClient.HGet(ctx, cache.ShareInfoKey(share.Uuid), "Owner").Val(); s != "" {
		return nil
	}

	// 使用管道保存分享信息到Redis以确保
	// 分享信息全部写入Redis
	saveShare := cache.RedisClient.Pipeline()
	// 向 Redis 中的 哈希（Hash）类型存入分享信息。
	saveShare.HSet(ctx, cache.ShareInfoKey(share.Uuid), "Owner", share.Owner)
	saveShare.HSet(ctx, cache.ShareInfoKey(share.Uuid), "FileId", share.FileId)
	saveShare.HSet(ctx, cache.ShareInfoKey(share.Uuid), "FileName", share.FileName)
	saveShare.HSet(ctx, cache.ShareInfoKey(share.Uuid), "Title", share.Title)
	saveShare.HSet(ctx, cache.ShareInfoKey(share.Uuid), "Size", share.Size)
	saveShare.HSet(ctx, cache.ShareInfoKey(share.Uuid), "SharingTime", share.SharingTime)
	saveShare.HSet(ctx, cache.ShareInfoKey(share.Uuid), "downloadUrl", downloadUrl)
	_, err := saveShare.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// GetShareInfoFromRedis 从Redis获取分享信息并返回下载链接
func (share *Share) GetShareInfoFromRedis() string {
	// 如果是空分享则填充空消息
	if cache.RedisClient.SIsMember(context.Background(), cache.EmptyShare, share.Uuid).Val() {
		share.Owner = ""
		share.FileId = ""
		share.FileName = ""
		share.Title = "来晚了,分享的文件已被删除"
		share.Size = 0
		share.SharingTime = ""
		return ""
	}
	// 从 Redis 的 哈希（Hash） 中获取该分享的所有字段
	shareInfo := cache.RedisClient.HGetAll(context.Background(), cache.ShareInfoKey(share.Uuid)).Val()
	share.Owner = shareInfo["Owner"]
	share.FileId = shareInfo["FileId"]
	share.FileName = shareInfo["FileName"]
	share.Title = shareInfo["Title"]
	share.Size, _ = strconv.ParseInt(shareInfo["Size"], 10, 64)
	share.SharingTime = shareInfo["SharingTime"]

	return shareInfo["downloadUrl"]
}

// CheckRedisExistsShare 使用标题信息检查，因为当分享信息存储到Redis时标题肯定存在
func (share *Share) CheckRedisExistsShare() bool {
	share.FileId, _ = cache.RedisClient.HGet(context.Background(), cache.ShareInfoKey(share.Uuid), "Title").Result()
	return share.FileId != "" || cache.RedisClient.SIsMember(context.Background(), cache.EmptyShare, share.Uuid).Val()
}

// DeleteShareInfoInRedis 删除Redis中的分享信息
func (share *Share) DeleteShareInfoInRedis() {
	_ = cache.RedisClient.ZRem(context.Background(), cache.DailyRankKey, share.Uuid)
	_ = cache.RedisClient.Del(context.Background(), cache.ShareInfoKey(share.Uuid)).Val()
}
