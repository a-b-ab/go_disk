package file

import (
	"mime/multipart"

	"go-cloud-disk/disk"
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils"
	"go-cloud-disk/utils/logger"
	"gorm.io/gorm"
)

// FileUploadService 文件上传服务结构体
type FileUploadService struct {
	FolderId string `form:"filefolder" json:"filefolder" binding:"required"` // 文件夹ID
}

// checkIfFileSizeExceedsVolum 检查上传文件大小是否超过用户存储空间限制
func checkIfFileSizeExceedsVolum(userStore *model.FileStore, userId string, size int64) (bool, error) {
	if err := model.DB.Where("owner_id = ?", userId).Find(userStore).Error; err != nil {
		return false, err
	}
	ans := userStore.CurrentSize+size > userStore.MaxSize
	return ans, nil
}

// createFile 使用事务保存用户文件信息，确保用户存储空间安全
func createFile(t *gorm.DB, file model.File, userStore model.FileStore) error {
	// 保存文件信息到数据库
	var err error
	if err = t.Save(&file).Error; err != nil {
		return err
	}
	// 增加用户文件存储容量
	userStore.AddCurrentSize(file.Size)
	if err = t.Save(&userStore).Error; err != nil {
		return err
	}
	return nil
}

// UploadFile 上传文件到云端并创建文件记录
func (service *FileUploadService) UploadFile(userId string, file *multipart.FileHeader, dst string) serializer.Response {
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

	// 上传文件到云端
	md5String, err := utils.GetFileMD5(dst)
	if err != nil {
		logger.Log().Error("[FileUploadService.UploadFile] 获取文件MD5失败: ", err)
		return serializer.ParamsErr("", err)
	}
	// 如果文件最近已经上传过，不重复上传到云端
	// 从Redis获取文件信息
	filePath := model.GetFileInfoFromRedis(md5String)
	if filePath == "" {
		err = disk.BaseCloudDisk.UploadSimpleFile(dst, userId, md5String, file.Size)
		if err != nil {
			logger.Log().Error("[FileUploadService.UploadFile] 上传文件到云端失败: ", err)
			return serializer.InternalErr("", err)
		}
		filePath = userId
	}

	// 插入文件到数据库
	filename, extend := utils.SplitFilename(file.Filename)
	fileModel := model.File{
		Owner:          userId,
		FileName:       filename,
		FilePostfix:    extend,
		FileUuid:       md5String,
		FilePath:       filePath,
		ParentFolderId: service.FolderId,
		Size:           file.Size,
	}

	t := model.DB.Begin()
	// 插入用户文件信息到数据库
	if err := createFile(t, fileModel, userStore); err != nil {
		logger.Log().Error("[FileUploadService.UploadFile] 创建文件信息失败: ", err)
		t.Rollback()
		return serializer.DBErr("", err)
	}

	// 将文件大小添加到文件夹和父文件夹
	var userFileFolder model.FileFolder
	if err := t.Where("uuid = ?", service.FolderId).Find(&userFileFolder).Error; err != nil {
		logger.Log().Error("[FileUploadService.UploadFile] 获取文件夹信息失败: ", err)
		t.Rollback()
		return serializer.DBErr("", err)
	}
	if err := userFileFolder.AddFileFolderSize(t, file.Size); err != nil {
		logger.Log().Error("[FileUploadService.UploadFile] 更新文件夹容量失败: ", err)
		t.Rollback()
		return serializer.DBErr("", err)
	}

	t.Commit()

	// 保存文件信息到Redis
	fileModel.SaveFileUploadInfoToRedis()
	return serializer.Success(serializer.BuildFile(fileModel))
}
