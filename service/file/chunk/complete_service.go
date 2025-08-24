package chunk

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"go-cloud-disk/cache"
	"go-cloud-disk/disk"
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils"
	"go-cloud-disk/utils/logger"

	"gorm.io/gorm"
)

// FileChunkCompleteService 完成分片上传服务
type FileChunkCompleteService struct {
	UploadId string `json:"upload_id" binding:"required"` // 上传任务ID
}

// CompleteChunkUpload 完成分片上传，合并分片并上传到COS
func (service *FileChunkCompleteService) CompleteChunkUpload(userId string) serializer.Response {
	// 1. 获取上传任务信息
	uploadInfo, err := getChunkUploadInfoFromRedis(service.UploadId)
	if err != nil {
		logger.Log().Error("[FileChunkCompleteService.CompleteChunkUpload] 获取上传信息失败: ", err)
		return serializer.ParamsErr("UploadIdNotFound", err)
	}

	// 2. 验证用户权限
	if uploadInfo.UserId != userId {
		return serializer.NotAuthErr("没有权限")
	}

	// 3. 检查所有分片是否都已上传
	if len(uploadInfo.UploadedChunks) != uploadInfo.TotalChunks {
		return serializer.ParamsErr("分片未完全上传", fmt.Errorf("已上传%d个，需要%d个", len(uploadInfo.UploadedChunks), uploadInfo.TotalChunks))
	}

	// 4. 验证分片完整性
	expectedChunks := make([]int, uploadInfo.TotalChunks)
	for i := 0; i < uploadInfo.TotalChunks; i++ {
		expectedChunks[i] = i + 1
	}

	sort.Ints(uploadInfo.UploadedChunks)
	if !isSliceEqual(expectedChunks, uploadInfo.UploadedChunks) {
		return serializer.ParamsErr("分片不完整", nil)
	}

	// 5. 合并分片文件
	mergedFilePath, fileMD5, err := service.mergeChunkFiles(uploadInfo)
	if err != nil {
		logger.Log().Error("[FileChunkCompleteService.CompleteChunkUpload] 合并分片失败: ", err)
		return serializer.InternalErr("合并分片失败", err)
	}
	defer os.Remove(mergedFilePath) // 确保清理临时文件

	// 6. 检查用户存储空间
	var userStore model.FileStore
	isExceed, err := checkIfFileSizeExceedsVolum(&userStore, userId, uploadInfo.FileSize)
	if err != nil {
		logger.Log().Error("[FileChunkCompleteService.CompleteChunkUpload] 检查用户容量失败: ", err)
		return serializer.DBErr("", err)
	}
	if isExceed {
		return serializer.ParamsErr("ExceedStoreLimit", nil)
	}

	// 7. 检查文件是否已存在（去重）
	existingFilePath := model.GetFileInfoFromRedis(fileMD5)
	needUploadToCOS := existingFilePath == ""
	if needUploadToCOS {
		existingFilePath = userId
	}

	// 8. 创建文件记录并入库（先完成数据库操作）
	fileModel, err := service.createFileRecord(uploadInfo, fileMD5, existingFilePath, userId, userStore)
	if err != nil {
		logger.Log().Error("[FileChunkCompleteService.CompleteChunkUpload] 创建文件记录失败: ", err)
		return serializer.DBErr("创建文件记录失败", err)
	}

	// 7.1 如果需要上传到COS，使用协程异步处理
	if needUploadToCOS {
		go func() {
			// 复制合并文件路径，因为主协程会删除原文件
			tempCopyPath := mergedFilePath + ".png"
			if err := service.copyFile(mergedFilePath, tempCopyPath); err != nil {
				logger.Log().Error("[FileChunkCompleteService.CompleteChunkUpload] 复制临时文件失败: ", err)
				return
			}
			defer os.Remove(tempCopyPath) // 确保清理复制的临时文件

			// 异步上传到COS
			err := disk.BaseCloudDisk.UploadSimpleFile(tempCopyPath, userId, fileMD5, uploadInfo.FileSize)
			if err != nil {
				logger.Log().Error("[FileChunkCompleteService.CompleteChunkUpload] 异步上传文件到COS失败: ", err)
				// 这里可以考虑重试机制或者更新文件状态
				return
			}

			logger.Log().Info("[FileChunkCompleteService.CompleteChunkUpload] 文件异步上传到COS成功: ", fileMD5)
		}()
	}

	// 9. 清理Redis分片信息
	if err := service.cleanupChunkInfo(service.UploadId, uploadInfo.TotalChunks); err != nil {
		logger.Log().Error("[FileChunkCompleteService.CompleteChunkUpload] 清理Redis分片信息失败: ", err)
		// 不返回错误，因为文件已经成功创建
	}

	// 10. 清理本地分片文件
	service.cleanupLocalChunkFiles(service.UploadId)

	// 11. 保存文件信息到Redis
	fileModel.SaveFileUploadInfoToRedis()

	return serializer.Success(serializer.BuildFile(*fileModel))
}

// mergeChunkFiles 合并分片文件
func (service *FileChunkCompleteService) mergeChunkFiles(uploadInfo *ChunkUploadInfo) (string, string, error) {
	// 创建临时合并文件
	tempDir := "./tmp_merge"
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return "", "", fmt.Errorf("创建临时目录失败: %v", err)
	}

	mergedFilePath := filepath.Join(tempDir, fmt.Sprintf("merged_%s_%s", service.UploadId, uploadInfo.FileName))
	mergedFile, err := os.Create(mergedFilePath)
	if err != nil {
		return "", "", fmt.Errorf("创建合并文件失败: %v", err)
	}
	defer mergedFile.Close()

	// 按顺序合并所有分片
	chunkDir := filepath.Join("./tmp_chunks", service.UploadId)
	for i := 0; i < uploadInfo.TotalChunks; i++ {
		chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%d", i+1))

		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return "", "", fmt.Errorf("打开分片文件 %s 失败: %v", chunkPath, err)
		}

		_, err = io.Copy(mergedFile, chunkFile)
		chunkFile.Close()

		if err != nil {
			return "", "", fmt.Errorf("复制分片 %d 失败: %v", i, err)
		}
	}

	// 计算合并后文件的MD5
	fileMD5, err := utils.GetFileMD5(mergedFilePath)
	if err != nil {
		return "", "", fmt.Errorf("计算文件MD5失败: %v", err)
	}

	return mergedFilePath, fileMD5, nil
}

// createFileRecord 创建文件记录并入库
func (service *FileChunkCompleteService) createFileRecord(uploadInfo *ChunkUploadInfo, fileMD5, filePath, userId string, userStore model.FileStore) (*model.File, error) {
	// 分离文件名和扩展名
	filename, extend := utils.SplitFilename(uploadInfo.FileName)

	// 创建文件模型
	fileModel := model.File{
		Owner:          userId,
		FileName:       filename,
		FilePostfix:    extend,
		FileUuid:       fileMD5,
		FilePath:       filePath,
		ParentFolderId: uploadInfo.FolderId,
		Size:           uploadInfo.FileSize,
	}

	// 开始数据库事务
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建文件记录
	if err := createFile(tx, fileModel, userStore); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("保存文件信息失败: %v", err)
	}

	// 更新文件夹大小
	var userFileFolder model.FileFolder
	if err := tx.Where("uuid = ?", uploadInfo.FolderId).First(&userFileFolder).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("获取文件夹信息失败: %v", err)
	}

	if err := userFileFolder.AddFileFolderSize(tx, uploadInfo.FileSize); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新文件夹容量失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %v", err)
	}

	return &fileModel, nil
}

// cleanupChunkInfo 清理Redis中的分片信息
func (service *FileChunkCompleteService) cleanupChunkInfo(uploadId string, totalChunks int) error {
	// 构建要删除的Redis键列表
	keys := []string{
		cache.ChunkUploadInfoKey(uploadId),
	}

	// 添加所有分片信息的键
	for i := 0; i < totalChunks; i++ {
		chunkKey := fmt.Sprintf("chunk:%s:%d", uploadId, i)
		keys = append(keys, chunkKey+":info", chunkKey+":data")
	}

	// 批量删除Redis键
	return cache.RedisClient.Del(context.Background(), keys...).Err()
}

// cleanupLocalChunkFiles 清理本地分片文件
func (service *FileChunkCompleteService) cleanupLocalChunkFiles(uploadId string) {
	chunkDir := filepath.Join("./tmp_chunks", uploadId)
	if err := os.RemoveAll(chunkDir); err != nil {
		logger.Log().Error("[FileChunkCompleteService.cleanupLocalChunkFiles] 清理本地分片文件失败: ", err)
	}
}

// createFile 使用事务保存用户文件信息
func createFile(tx *gorm.DB, file model.File, userStore model.FileStore) error {
	// 保存文件信息到数据库
	if err := tx.Save(&file).Error; err != nil {
		return err
	}

	// 增加用户文件存储容量
	userStore.AddCurrentSize(file.Size)
	if err := tx.Save(&userStore).Error; err != nil {
		return err
	}

	return nil
}

// copyFile 复制文件
func (service *FileChunkCompleteService) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// isSliceEqual 比较两个int切片是否相等
func isSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
