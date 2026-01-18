package file

import (
	"strings"

	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
)

// ListAllFilesService 列出用户全部文件（用于标签页“添加文件”）
type ListAllFilesService struct {
	Limit     int  `form:"limit"`
	Offset    int  `form:"offset"`
	OnlyImage bool `form:"only_image"`
}

func isImagePostfix(postfix string) bool {
	switch strings.ToLower(postfix) {
	case "png", "jpg", "jpeg", "gif", "webp", "bmp", "svg":
		return true
	default:
		return false
	}
}

func (service *ListAllFilesService) ListAllFiles(userID string) serializer.Response {
	limit := service.Limit
	offset := service.Offset
	if limit <= 0 || limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}

	var files []model.File
	tx := model.DB.Where("owner = ? AND deleted_at IS NULL", userID).
		Limit(limit).
		Offset(offset)
	if err := tx.Find(&files).Error; err != nil {
		return serializer.DBErr("获取文件失败", err)
	}

	if service.OnlyImage {
		filtered := make([]model.File, 0, len(files))
		for _, f := range files {
			if isImagePostfix(f.FilePostfix) {
				filtered = append(filtered, f)
			}
		}
		files = filtered
	}

	return serializer.Success(serializer.BuildFiles(files))
}

