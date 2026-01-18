package chunk

import (
	"context"
	"encoding/json"
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
	FolderId  string `form:"filefolder" json:"filefolder" binding:"required"`   // 文件夹ID
	FileName  string `form:"file_name" json:"file_name" binding:"required"`     // 文件名（含后缀）
	FileSize  int64  `form:"file_size" json:"file_size" binding:"required"`     // 文件总大小（bytes）
	// todo: 文件md5，前端传过来即可，支持秒传
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

func (service *ChunkInitService) InitChunkUpload(userId string) serializer.Response {
	var userStore model.FileStore
	var err error

	// 前端只需要传文件名/大小/目录，不需要传文件内容（真实内容会在分片上传接口传输）
	// 这里只是为了检查用户存储空间是否足够
	// 真实的文件内容会在分片上传时传输
	// 检查用户存储空间是否足够

	// 检查添加文件大小后当前大小是否超过最大限制
	var isExceed bool
	if isExceed, err = checkIfFileSizeExceedsVolum(&userStore, userId, service.FileSize); err != nil {
		logger.Log().Error("[FileUploadService.UploadFile] 检查用户容量失败: ", err)
		return serializer.DBErr("", err)
	}
	if isExceed {
		return serializer.ParamsErr("ExceedStoreLimit", nil)
	}

	// // todo:秒传判断+md5存储到file文件
	// existFile,err := model.GetFileByMd5(service.FileMd5)
	// if err == nil && existFile != nil {

	// 	// 建立用户文件关系（逻辑秒传）
	// 	if err := model.CreateUserFile(
	// 		userId,
	// 		existFile.Id,
	// 		service.FolderId,
	// 	); err != nil {
	// 		return serializer.DBErr("", err)
	// 	}

	// 	// 直接返回，不进入分片流程
	// 	return serializer.Success(map[string]interface{}{
	// 		"instant": true,
	// 		"file_id": existFile.Id,
	// 	})
	// }

	// 计算分片数（chunk_number 使用 1..totalChunks）
	totalChunks := int((service.FileSize + ChunkSize - 1) / ChunkSize)

	// 生成上传任务ID
	uploadId := uuid.New().String()

	// 创建分片上传信息
	uploadInfo := ChunkUploadInfo{
		UploadId:       uploadId,
		FileName:       service.FileName,
		FileSize:       service.FileSize,
		ChunkSize:      ChunkSize,
		// 文件md5
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
		"file_size":    service.FileSize,
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
