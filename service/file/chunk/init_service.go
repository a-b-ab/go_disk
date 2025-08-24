package chunk

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"time"

	"go-cloud-disk/cache"
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"

	"github.com/google/uuid"
)

const (
	ChunkSize = 5 * 1024 * 1024 // 5MB 分片大小
)

type ChunkInitService struct {
	FolderId string `form:"filefolder" json:"filefolder" binding:"required"` // 文件夹ID
}

type ChunkUploadInfo struct {
	UploadId       string    `json:"upload_id"`       // 上传任务ID
	FileName       string    `json:"file_name"`       // 文件名
	FileSize       int64     `json:"file_size"`       // 文件总大小
	ChunkSize      int64     `json:"chunk_size"`      // 分片大小
	TotalChunks    int       `json:"total_chunks"`    // 总分片数
	FolderId       string    `json:"folder_id"`       // 文件夹ID
	UserId         string    `json:"user_id"`         // 用户ID
	CreatedAt      time.Time `json:"created_at"`      // 创建时间
	UploadedChunks []int     `json:"uploaded_chunks"` // 已上传的分片列表
}

func (service *ChunkInitService) InitChunkUpload(userId string, file *multipart.FileHeader, dst string) serializer.Response {
	// 获取用户上传文件并保存到本地
	var userStore model.FileStore
	var err error

	// 检查添加文件大小后当前大小是否超过最大限制
	var isExceed bool
	if isExceed, err = checkIfFileSizeExceedsVolum(&userStore, userId, file.Size); err != nil {
		logger.Log().Error("[FileUploadService.UploadFile] 检查用户容量失败: ", err)
		return serializer.DBErr("", err)
	}
	if isExceed {
		return serializer.ParamsErr("ExceedStoreLimit", nil)
	}

	// 计算分片数
	totalChunks := int((file.Size + ChunkSize - 1) / ChunkSize)

	// 生成上传任务ID
	uploadId := uuid.New().String()

	// 创建分片上传信息
	uploadInfo := ChunkUploadInfo{
		UploadId:       uploadId,
		FileName:       file.Filename,
		FileSize:       file.Size,
		ChunkSize:      ChunkSize,
		TotalChunks:    totalChunks,
		FolderId:       service.FolderId,
		UserId:         userId,
		CreatedAt:      time.Now(),
		UploadedChunks: make([]int, 0),
	}

	// 保存到Redis，设置过期时间为24小时
	if err := saveChunkUploadInfoToRedis(uploadId, uploadInfo); err != nil {
		logger.Log().Error("[FileChunkInitService.InitChunkUpload] 保存上传信息到Redis失败: ", err)
		return serializer.InternalErr("", err)
	}

	response := map[string]interface{}{
		"upload_id":    uploadId,
		"chunk_size":   ChunkSize,
		"total_chunks": totalChunks,
		"file_size":    file.Size,
	}

	return serializer.Success(response)
}

// checkIfFileSizeExceedsVolum 检查上传文件大小是否超过用户存储空间限制
func checkIfFileSizeExceedsVolum(userStore *model.FileStore, userId string, size int64) (bool, error) {
	if err := model.DB.Where("owner_id = ?", userId).Find(userStore).Error; err != nil {
		return false, err
	}
	ans := userStore.CurrentSize+size > userStore.MaxSize
	return ans, nil
}

// saveChunkUploadInfoToRedis 哈希存储分片上传信息
func saveChunkUploadInfoToRedis(uploadId string, uploadInfo ChunkUploadInfo) error {
	key := cache.ChunkUploadInfoKey(uploadId)

	// 将 UploadedChunks 序列化为 JSON 字符串
	uploadedChunks, err := json.Marshal(uploadInfo.UploadedChunks)
	if err != nil {
		return err
	}

	// Redis 哈希字段
	fields := map[string]interface{}{
		"UploadId":       uploadInfo.UploadId,
		"FileName":       uploadInfo.FileName,
		"FileSize":       uploadInfo.FileSize,
		"ChunkSize":      uploadInfo.ChunkSize,
		"TotalChunks":    uploadInfo.TotalChunks,
		"FolderId":       uploadInfo.FolderId,
		"UserId":         uploadInfo.UserId,
		"CreatedAt":      uploadInfo.CreatedAt.Unix(), // 时间戳存储
		"UploadedChunks": string(uploadedChunks),
	}

	// 写入哈希
	if err := cache.RedisClient.HSet(context.Background(), key, fields).Err(); err != nil {
		return err
	}

	// 设置过期时间
	return cache.RedisClient.Expire(context.Background(), key, 24*time.Hour).Err()
}
