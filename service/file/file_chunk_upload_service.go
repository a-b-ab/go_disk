package file

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-cloud-disk/cache"
	"go-cloud-disk/disk"
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils"
	"go-cloud-disk/utils/logger"

	"github.com/google/uuid"
)

const (
	ChunkSize = 5 * 1024 * 1024 // 5MB 分片大小
)

// ChunkUploadInfo 分片上传信息
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

// ChunkInfo 分片信息
type ChunkInfo struct {
	ChunkNumber int    `json:"chunk_number"` // 分片序号
	ChunkSize   int64  `json:"chunk_size"`   // 分片大小
	MD5         string `json:"md5"`          // 分片MD5
}

// FileChunkInitService 初始化分片上传服务
type FileChunkInitService struct {
	FileName string `form:"filename" json:"filename" binding:"required"`     // 文件名
	FileSize int64  `form:"filesize" json:"filesize" binding:"required"`     // 文件大小
	FolderId string `form:"filefolder" json:"filefolder" binding:"required"` // 文件夹ID
}

// FileChunkUploadService 分片上传服务
type FileChunkUploadService struct {
	UploadId    string `form:"upload_id" json:"upload_id" binding:"required"`       // 上传任务ID
	ChunkNumber int    `form:"chunk_number" json:"chunk_number" binding:"required"` // 分片序号
}

// FileChunkCheckService 检查分片服务
type FileChunkCheckService struct {
	UploadId string `form:"upload_id" json:"upload_id" binding:"required"` // 上传任务ID
}

// FileChunkCompleteService 完成分片上传服务
type FileChunkCompleteService struct {
	UploadId string `form:"upload_id" json:"upload_id" binding:"required"` // 上传任务ID
}

// InitChunkUpload 初始化分片上传
func (service *FileChunkInitService) InitChunkUploads(userId string) serializer.Response {
	// 检查用户存储空间
	var userStore model.FileStore
	isExceed, err := checkIfFileSizeExceedsVolum(&userStore, userId, service.FileSize)
	if err != nil {
		logger.Log().Error("[FileChunkInitService.InitChunkUpload] 检查用户容量失败: ", err)
		return serializer.DBErr("", err)
	}
	if isExceed {
		return serializer.ParamsErr("ExceedStoreLimit", nil)
	}

	// 检查文件夹权限
	var fileFolder model.FileFolder
	if err := model.DB.Where("uuid = ?", service.FolderId).First(&fileFolder).Error; err != nil {
		logger.Log().Error("[FileChunkInitService.InitChunkUpload] 查找文件夹失败: ", err)
		return serializer.DBErr("", err)
	}
	if fileFolder.OwnerID != userId {
		return serializer.NotAuthErr("")
	}

	// 计算分片数
	totalChunks := int((service.FileSize + ChunkSize - 1) / ChunkSize)

	// 生成上传任务ID
	uploadId := uuid.New().String()

	// 创建分片上传信息
	uploadInfo := ChunkUploadInfo{
		UploadId:       uploadId,
		FileName:       service.FileName,
		FileSize:       service.FileSize,
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
		"file_size":    service.FileSize,
	}

	return serializer.Success(response)
}

// UploadChunk 上传分片
func (service *FileChunkUploadService) UploadChunks(userId string, chunkData []byte) serializer.Response {
	// 从Redis获取上传信息
	uploadInfo, err := getChunkUploadInfoFromRedis(service.UploadId)
	if err != nil {
		logger.Log().Error("[FileChunkUploadService.UploadChunk] 获取上传信息失败: ", err)
		return serializer.ParamsErr("UploadIdNotFound", err)
	}

	// 验证用户权限
	if uploadInfo.UserId != userId {
		return serializer.NotAuthErr("")
	}

	// 验证分片序号
	if service.ChunkNumber < 1 || service.ChunkNumber > uploadInfo.TotalChunks {
		return serializer.ParamsErr("InvalidChunkNumber", nil)
	}

	// 验证分片大小
	var expectedSize int64 = ChunkSize
	if service.ChunkNumber == uploadInfo.TotalChunks {
		// 最后一个分片可能小于标准分片大小
		expectedSize = uploadInfo.FileSize - int64(uploadInfo.TotalChunks-1)*ChunkSize
	}
	if int64(len(chunkData)) != expectedSize {
		return serializer.ParamsErr("InvalidChunkSize", nil)
	}

	// 计算分片MD5
	chunkMD5 := utils.GetBytesMD5(chunkData)

	// 保存分片到临时存储
	chunkKey := fmt.Sprintf("chunk:%s:%d", service.UploadId, service.ChunkNumber)
	chunkInfo := ChunkInfo{
		ChunkNumber: service.ChunkNumber,
		ChunkSize:   int64(len(chunkData)),
		MD5:         chunkMD5,
	}

	// 保存分片信息到Redis
	if err := saveChunkInfoToRedis(chunkKey, chunkInfo, chunkData); err != nil {
		logger.Log().Error("[FileChunkUploadService.UploadChunk] 保存分片信息失败: ", err)
		return serializer.InternalErr("", err)
	}

	// 更新已上传分片列表
	uploadInfo.UploadedChunks = append(uploadInfo.UploadedChunks, service.ChunkNumber)
	// 去重并排序
	uploadInfo.UploadedChunks = utils.UniqueAndSortInts(uploadInfo.UploadedChunks)

	// 更新Redis中的上传信息
	if err := saveChunkUploadInfoToRedis(service.UploadId, *uploadInfo); err != nil {
		logger.Log().Error("[FileChunkUploadService.UploadChunk] 更新上传信息失败: ", err)
		return serializer.InternalErr("", err)
	}

	response := map[string]interface{}{
		"chunk_number":    service.ChunkNumber,
		"chunk_md5":       chunkMD5,
		"uploaded_chunks": len(uploadInfo.UploadedChunks),
		"total_chunks":    uploadInfo.TotalChunks,
	}

	return serializer.Success(response)
}

// CheckChunks 检查已上传的分片
func (service *FileChunkCheckService) CheckChunks(userId string) serializer.Response {
	// 从Redis获取上传信息
	uploadInfo, err := getChunkUploadInfoFromRedis(service.UploadId)
	if err != nil {
		logger.Log().Error("[FileChunkCheckService.CheckChunks] 获取上传信息失败: ", err)
		return serializer.ParamsErr("UploadIdNotFound", err)
	}

	// 验证用户权限
	if uploadInfo.UserId != userId {
		return serializer.NotAuthErr("")
	}

	response := map[string]interface{}{
		"upload_id":       service.UploadId,
		"uploaded_chunks": uploadInfo.UploadedChunks,
		"total_chunks":    uploadInfo.TotalChunks,
		"file_name":       uploadInfo.FileName,
		"file_size":       uploadInfo.FileSize,
		"chunk_size":      uploadInfo.ChunkSize,
	}

	return serializer.Success(response)
}

// CompleteChunkUpload 完成分片上传
func (service *FileChunkCompleteService) CompleteChunkUpload(userId string) serializer.Response {
	// 从Redis获取上传信息
	uploadInfo, err := getChunkUploadInfoFromRedis(service.UploadId)
	if err != nil {
		logger.Log().Error("[FileChunkCompleteService.CompleteChunkUpload] 获取上传信息失败: ", err)
		return serializer.ParamsErr("UploadIdNotFound", err)
	}

	// 验证用户权限
	if uploadInfo.UserId != userId {
		return serializer.NotAuthErr("")
	}

	// 检查是否所有分片都已上传
	if len(uploadInfo.UploadedChunks) != uploadInfo.TotalChunks {
		return serializer.ParamsErr("IncompleteChunks", nil)
	}

	// 验证分片完整性
	for i := 1; i <= uploadInfo.TotalChunks; i++ {
		if !utils.ContainsInt(uploadInfo.UploadedChunks, i) {
			return serializer.ParamsErr("MissingChunks", nil)
		}
	}

	// 合并分片并计算文件MD5
	fileMD5, err := mergeChunksAndUpload(service.UploadId, *uploadInfo)
	if err != nil {
		logger.Log().Error("[FileChunkCompleteService.CompleteChunkUpload] 合并分片失败: ", err)
		return serializer.InternalErr("", err)
	}

	// 创建文件记录
	filename, extend := utils.SplitFilename(uploadInfo.FileName)
	fileModel := model.File{
		Owner:          userId,
		FileName:       filename,
		FilePostfix:    extend,
		FileUuid:       fileMD5,
		FilePath:       userId,
		ParentFolderId: uploadInfo.FolderId,
		Size:           uploadInfo.FileSize,
	}

	// 开始数据库事务
	tx := model.DB.Begin()

	// 获取用户存储信息
	var userStore model.FileStore
	if err := tx.Where("owner_id = ?", userId).First(&userStore).Error; err != nil {
		logger.Log().Error("[FileChunkCompleteService.CompleteChunkUpload] 获取用户存储信息失败: ", err)
		tx.Rollback()
		return serializer.DBErr("", err)
	}

	// 创建文件记录
	if err := createFile(tx, fileModel, userStore); err != nil {
		logger.Log().Error("[FileChunkCompleteService.CompleteChunkUpload] 创建文件记录失败: ", err)
		tx.Rollback()
		return serializer.DBErr("", err)
	}

	// 更新文件夹大小
	var userFileFolder model.FileFolder
	if err := tx.Where("uuid = ?", uploadInfo.FolderId).First(&userFileFolder).Error; err != nil {
		logger.Log().Error("[FileChunkCompleteService.CompleteChunkUpload] 获取文件夹信息失败: ", err)
		tx.Rollback()
		return serializer.DBErr("", err)
	}
	if err := userFileFolder.AddFileFolderSize(tx, uploadInfo.FileSize); err != nil {
		logger.Log().Error("[FileChunkCompleteService.CompleteChunkUpload] 更新文件夹容量失败: ", err)
		tx.Rollback()
		return serializer.DBErr("", err)
	}

	tx.Commit()

	// 保存文件信息到Redis
	fileModel.SaveFileUploadInfoToRedis()

	// 清理Redis中的分片信息
	if err := cleanupChunkUploadInfo(service.UploadId, uploadInfo.TotalChunks); err != nil {
		logger.Log().Error("[FileChunkCompleteService.CompleteChunkUpload] 清理分片信息失败: ", err)
		// 不返回错误，因为文件已经成功创建
	}

	return serializer.Success(serializer.BuildFile(fileModel))
}

// 辅助函数

// saveChunkUploadInfoToRedis 保存分片上传信息到Redis
// 后面用哈希存储
func saveChunkUploadInfoToRedis(uploadId string, uploadInfo ChunkUploadInfo) error {
	key := cache.ChunkUploadInfoKey(uploadId)
	data, err := json.Marshal(uploadInfo)
	if err != nil {
		return err
	}
	return cache.RedisClient.Set(context.Background(), key, data, 24*time.Hour).Err()
}

// getChunkUploadInfoFromRedis 从Redis获取分片上传信息
func getChunkUploadInfoFromRedis(uploadId string) (*ChunkUploadInfo, error) {
	key := cache.ChunkUploadInfoKey(uploadId)
	data, err := cache.RedisClient.Get(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}

	var uploadInfo ChunkUploadInfo
	if err := json.Unmarshal([]byte(data), &uploadInfo); err != nil {
		return nil, err
	}

	return &uploadInfo, nil
}

// saveChunkInfoToRedis 保存分片信息到Redis
func saveChunkInfoToRedis(chunkKey string, chunkInfo ChunkInfo, chunkData []byte) error {
	// 保存分片元信息
	infoData, err := json.Marshal(chunkInfo)
	if err != nil {
		return err
	}
	if err := cache.RedisClient.Set(context.Background(), chunkKey+":info", infoData, 24*time.Hour).Err(); err != nil {
		return err
	}

	// 保存分片数据
	return cache.RedisClient.Set(context.Background(), chunkKey+":data", chunkData, 24*time.Hour).Err()
}

// mergeChunksAndUpload 合并分片并上传到云端
func mergeChunksAndUpload(uploadId string, uploadInfo ChunkUploadInfo) (string, error) {
	// 创建临时文件来合并分片
	tempFilePath := fmt.Sprintf("./temp/%s_%s", uploadId, uploadInfo.FileName)

	// 确保临时目录存在
	if err := utils.EnsureDir("./temp"); err != nil {
		return "", err
	}

	// 按顺序读取分片并合并
	var allData []byte
	for i := 1; i <= uploadInfo.TotalChunks; i++ {
		chunkKey := fmt.Sprintf("chunk:%s:%d:data", uploadId, i)
		chunkData, err := cache.RedisClient.Get(context.Background(), chunkKey).Bytes()
		if err != nil {
			return "", fmt.Errorf("获取分片 %d 失败: %v", i, err)
		}
		allData = append(allData, chunkData...)
	}

	// 计算完整文件的MD5
	fileMD5 := utils.GetBytesMD5(allData)

	// 检查文件是否已经存在于云端
	existingFilePath := model.GetFileInfoFromRedis(fileMD5)
	if existingFilePath != "" {
		// 文件已存在，无需重复上传
		return fileMD5, nil
	}

	// 写入临时文件
	if err := utils.WriteBytesToFile(tempFilePath, allData); err != nil {
		return "", err
	}

	// 上传到云端
	err := disk.BaseCloudDisk.UploadSimpleFile(tempFilePath, uploadInfo.UserId, fileMD5, uploadInfo.FileSize)
	if err != nil {
		return "", err
	}

	// 删除临时文件
	utils.RemoveFile(tempFilePath)

	return fileMD5, nil
}

// cleanupChunkUploadInfo 清理Redis中的分片信息
func cleanupChunkUploadInfo(uploadId string, totalChunks int) error {
	keys := []string{cache.ChunkUploadInfoKey(uploadId)}

	// 添加所有分片相关的key
	for i := 1; i <= totalChunks; i++ {
		chunkKey := fmt.Sprintf("chunk:%s:%d", uploadId, i)
		keys = append(keys, chunkKey+":info", chunkKey+":data")
	}

	return cache.RedisClient.Del(context.Background(), keys...).Err()
}
