package tag

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
)

// TagFilesService 获取某个标签下的文件列表
type TagFilesService struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}

func (service *TagFilesService) GetTagFiles(userID string, tagID string) serializer.Response {
	limit := service.Limit
	offset := service.Offset
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	var tag model.Tag
	if err := model.DB.Where("id = ?", tagID).First(&tag).Error; err != nil {
		return serializer.ParamsErr("标签不存在", err)
	}

	files, err := model.GetFilesByTagID(tagID, userID, limit, offset)
	if err != nil {
		return serializer.DBErr("获取标签文件失败", err)
	}
	return serializer.Success(serializer.BuildFiles(files))
}

