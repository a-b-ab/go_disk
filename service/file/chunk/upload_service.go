package chunk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"go-cloud-disk/cache"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils"
	"go-cloud-disk/utils/logger"
)

type FileChunkUploadService struct {
	UploadId    string `form:"upload_id" json:"upload_id" binding:"required"`       // 上传任务ID
	ChunkNumber int    `form:"chunk_number" json:"chunk_number" binding:"required"` // 分片序号
}

// UploadChunk 上传单个分片
func (service *FileChunkUploadService) UploadChunk(userId string, chunkFile *multipart.FileHeader) serializer.Response {
	// 1. 获取上传任务信息
	uploadInfo, err := getChunkUploadInfoFromRedis(service.UploadId)
	if err != nil {
		logger.Log().Error("[FileChunkUploadService.UploadChunk] 获取上传信息失败: ", err)
		return serializer.ParamsErr("UploadIdNotFound", err)
	}

	// 2. 验证用户
	if uploadInfo.UserId != userId {
		return serializer.NotAuthErr("没有权限")
	}

	// 3. 验证分片序号
	if service.ChunkNumber < 0 || service.ChunkNumber > uploadInfo.TotalChunks {
		return serializer.ParamsErr("InvalidChunkNumber", nil)
	}

	// 4. 验证分片大小
	expectedSize := int64(ChunkSize)
	if service.ChunkNumber == uploadInfo.TotalChunks {
		expectedSize = uploadInfo.FileSize - int64(ChunkSize*(uploadInfo.TotalChunks-1))
	}

	if chunkFile.Size != int64(expectedSize) {
		return serializer.ParamsErr("InvalidChunkSize", nil)
	}

	// 5. 保存分片到本地临时目录
	tmDir, err := saveChunkToTempDir(service.UploadId, service.ChunkNumber, chunkFile)
	if err != nil {
		logger.Log().Error("[FileChunkUploadService.UploadChunk] 保存分片失败: ", err)
		return serializer.InternalErr("SaveChunkFailed", err)
	}

	dstPath := filepath.Join(tmDir, fmt.Sprintf("chunk_%d", service.ChunkNumber))
	if err := saveMultipartFile(chunkFile, dstPath); err != nil {
		return serializer.InternalErr("保存分片失败", err)
	}

	// 6. 计算分片MD5
	md5Str, err := utils.GetFileMD5(dstPath)
	if err != nil {
		return serializer.InternalErr("计算分片MD5失败", err)
	}

	// 7. 更新上传信息（Redis）
	uploadInfo.UploadedChunks = append(uploadInfo.UploadedChunks, service.ChunkNumber)
	uploadInfo.UploadedChunks = utils.UniqueAndSortInts(uploadInfo.UploadedChunks)
	if err := saveChunkUploadInfoToRedis(service.UploadId, *uploadInfo); err != nil {
		return serializer.InternalErr("更新上传信息失败", err)
	}

	response := map[string]interface{}{
		"chunk_number":   service.ChunkNumber,
		"chunk_md5":      md5Str,
		"uploaded_count": len(uploadInfo.UploadedChunks),
		"total_chunks":   uploadInfo.TotalChunks,
	}

	return serializer.Success(response)
}

// 从 Redis 中获取分片上传信息
func getChunkUploadInfoFromRedis(uploadId string) (*ChunkUploadInfo, error) {
	key := cache.ChunkUploadInfoKey(uploadId)
	result, err := cache.RedisClient.HGetAll(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("上传信息未找到")
	}

	// 解析 UploadedChunks
	var uploadedChunks []int
	if ucStr, ok := result["UploadedChunks"]; ok && ucStr != "" {
		if err := json.Unmarshal([]byte(ucStr), &uploadedChunks); err != nil {
			return nil, err
		}
	}

	// 解析 CreatedAt
	var createdAt time.Time
	if tsStr, ok := result["CreatedAt"]; ok && tsStr != "" {
		tsInt, err := strconv.ParseInt(tsStr, 10, 64)
		if err != nil {
			return nil, err
		}
		createdAt = time.Unix(tsInt, 0)
	}

	fileSize, _ := strconv.ParseInt(result["FileSize"], 10, 64)
	chunkSize, _ := strconv.ParseInt(result["ChunkSize"], 10, 64)
	totalChunks, _ := strconv.Atoi(result["TotalChunks"])

	info := &ChunkUploadInfo{
		UploadId:       result["UploadId"],
		FileName:       result["FileName"],
		FileSize:       fileSize,
		ChunkSize:      chunkSize,
		TotalChunks:    totalChunks,
		FolderId:       result["FolderId"],
		UserId:         result["UserId"],
		CreatedAt:      createdAt,
		UploadedChunks: uploadedChunks,
	}

	return info, nil
}

// saveChunkToTempDir 为每个上传任务和分片创建临时目录
func saveChunkToTempDir(uploadId string, chunkNumber int, fileHeader *multipart.FileHeader) (string, error) {
	// 临时目录按uploadId创建，保证不同任务隔离
	tmpRoot := "./tmp_chunks"
	taskDir := filepath.Join(tmpRoot, uploadId)
	if err := os.MkdirAll(taskDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("创建任务临时目录失败: %v", err)
	}
	return taskDir, nil
}

// 将multipart.FileHeader 保存到指定路径
func saveMultipartFile(fileHeader *multipart.FileHeader, dstPath string) error {
	// 打开 multipart 文件
	srcFile, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("打开分片文件失败: %v", err)
	}
	defer srcFile.Close()

	// 创建目标文件
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer dstFile.Close()

	// 拷贝数据
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("保存分片文件失败: %v", err)
	}
	return nil
}
