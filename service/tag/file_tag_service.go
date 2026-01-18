package tag

import (
	"strings"

	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
)

// FileTagBindService 绑定文件标签
type FileTagBindService struct {
	TagID string `json:"tag_id" binding:"required"`
}

func isImagePostfix(postfix string) bool {
	switch strings.ToLower(postfix) {
	case "png", "jpg", "jpeg", "gif", "webp", "bmp":
		return true
	default:
		return false
	}
}

// GetTagsByFile 获取图片文件的标签列表
func GetTagsByFile(userID string, fileID string) serializer.Response {
	var file model.File
	if err := model.DB.Where("uuid = ? AND owner = ?", fileID, userID).First(&file).Error; err != nil {
		return serializer.ParamsErr("文件不存在", err)
	}
	if !isImagePostfix(file.FilePostfix) {
		return serializer.ParamsErr("仅图片支持标签", nil)
	}

	// file_tags 目前以 file_uuid（md5）作为 file_id 进行关联
	tags, err := model.GetTagsByFile(file.FileUuid)
	if err != nil {
		return serializer.DBErr("获取标签失败", err)
	}
	return serializer.Success(serializer.BuildTags(tags))
}

// BindTagToFile 为图片文件绑定标签
func (service *FileTagBindService) BindTagToFile(userID string, fileID string) serializer.Response {
	var file model.File
	if err := model.DB.Where("uuid = ? AND owner = ?", fileID, userID).First(&file).Error; err != nil {
		return serializer.ParamsErr("文件不存在", err)
	}
	if !isImagePostfix(file.FilePostfix) {
		return serializer.ParamsErr("仅图片支持标签", nil)
	}

	if err := model.AddTagToFile(file.FileUuid, service.TagID); err != nil {
		return serializer.DBErr("绑定标签失败", err)
	}
	return serializer.Success(nil)
}

// UnbindTagFromFile 为图片文件解绑标签
func UnbindTagFromFile(userID string, fileID string, tagID string) serializer.Response {
	var file model.File
	if err := model.DB.Where("uuid = ? AND owner = ?", fileID, userID).First(&file).Error; err != nil {
		return serializer.ParamsErr("文件不存在", err)
	}
	if !isImagePostfix(file.FilePostfix) {
		return serializer.ParamsErr("仅图片支持标签", nil)
	}

	// file_tags 目前以 file_uuid（md5）作为 file_id 进行关联
	if err := model.DB.Where("file_id = ? AND tag_id = ?", file.FileUuid, tagID).Delete(&model.FileTag{}).Error; err != nil {
		return serializer.DBErr("解绑标签失败", err)
	}
	return serializer.Success(nil)
}
