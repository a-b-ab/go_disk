package file

import (
	"go-cloud-disk/disk"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
	"github.com/google/uuid"
)

// GetUploadURLService 获取上传URL服务结构体
type GetUploadURLService struct {
	FileType string `form:"filetype" json:"filetype" binding:"required,min=1"` // 文件类型/扩展名
}

// getUploadURLResponse 获取上传URL响应结构体
type getUploadURLResponse struct {
	Url      string `json:"url"`       // 上传URL
	FileUuid string `json:"file_uuid"` // 文件UUID
}

// GetUploadURL 获取文件上传预签名URL
func (service *GetUploadURLService) GetUploadURL(fileowner string) serializer.Response {
	fileID := uuid.New().String()
	fileName := fileID + "." + service.FileType
	url, err := disk.BaseCloudDisk.GetUploadPresignedURL(fileowner, "", fileName)
	if err != nil {
		logger.Log().Error("[GetUploadURLService.GetUploadURL] 获取上传URL失败: ", err)
		return serializer.InternalErr("", err)
	}

	return serializer.Success(getUploadURLResponse{
		Url:      url,
		FileUuid: fileID,
	})
}
