package chunk

import (
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

type FileChunkCheckService struct {
	UploadId string `json:"upload_id" binding:"required"` // 上传任务ID
}

// CheckChunks 检查已上传的分片状态
func (service *FileChunkCheckService) CheckChunk(userId string) serializer.Response {
	// 1. 获取上传任务信息
	uploadInfo, err := getChunkUploadInfoFromRedis(service.UploadId)
	if err != nil {
		logger.Log().Error("[FileChunkCheckService.CheckChunks] 获取上传信息失败: ", err)
		return serializer.ParamsErr("UploadIdNotFound", err)
	}

	// 2. 验证用户权限
	if uploadInfo.UserId != userId {
		return serializer.NotAuthErr("没有权限")
	}

	// 3. 计算上传进度
	uploadedCount := len(uploadInfo.UploadedChunks)
	uploadProgress := float64(uploadedCount) / float64(uploadInfo.TotalChunks) * 100

	// 4. 查找缺失的分片
	missingChunks := findMissingChunks(uploadInfo.UploadedChunks, uploadInfo.TotalChunks)

	// 5. 判断上传状态
	status := "uploading"
	if uploadedCount == uploadInfo.TotalChunks {
		status = "ready_to_complete"
	} else if uploadedCount == 0 {
		status = "not_started"
	}

	// 6. 构建响应数据
	response := map[string]interface{}{
		"upload_id":       service.UploadId,
		"file_name":       uploadInfo.FileName,
		"file_size":       uploadInfo.FileSize,
		"chunk_size":      uploadInfo.ChunkSize,
		"total_chunks":    uploadInfo.TotalChunks,
		"uploaded_chunks": uploadInfo.UploadedChunks,
		"uploaded_count":  uploadedCount,
		"missing_chunks":  missingChunks,
		"missing_count":   len(missingChunks),
		"upload_progress": uploadProgress,
		"status":          status,
		"created_at":      uploadInfo.CreatedAt,
		"folder_id":       uploadInfo.FolderId,
	}

	return serializer.Success(response)
}

// findMissingChunks 查找缺失的分片序号
func findMissingChunks(uploadedChunks []int, totalChunks int) []int {
	var missingChunks []int

	// 创建已上传分片的映射
	uploadedMap := make(map[int]bool)
	for _, chunk := range uploadedChunks {
		uploadedMap[chunk] = true
	}

	// 找出缺失的分片
	for i := 0; i < totalChunks; i++ {
		if !uploadedMap[i] {
			missingChunks = append(missingChunks, i)
		}
	}

	return missingChunks
}
