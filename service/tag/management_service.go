package tag

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

// TagCreateService 负责新增标签
type TagCreateService struct {
	Name string `json:"name" form:"name" binding:"required"`
}

// TagUpdateService 负责更新标签名称
// CreateTag 创建新标签
func (service *TagCreateService) CreateTag() serializer.Response {
	tag := model.Tag{
		Name: service.Name,
	}
	if err := model.DB.Create(&tag).Error; err != nil {
		logger.Log().Error("[TagCreateService.CreateTag] 创建标签失败: ", err)
		return serializer.DBErr("", err)
	}
	return serializer.Success(serializer.BuildTag(tag))
}

// TagUpdateService 负责更新标签名称
type TagUpdateService struct {
	Name string `json:"name" form:"name" binding:"required"`
}

// TagDeleteService 负责删除标签
type TagDeleteService struct{}

// TagListService 负责获取标签列表和详情
type TagListService struct{}

// UpdateTag 修改标签名称
func (service *TagUpdateService) UpdateTag(tagID string) serializer.Response {
	var tag model.Tag
	if err := model.DB.Where("id = ?", tagID).First(&tag).Error; err != nil {
		logger.Log().Error("[TagUpdateService.UpdateTag] 查找标签失败: ", err)
		return serializer.DBErr("", err)
	}
	tag.Name = service.Name
	if err := model.DB.Save(&tag).Error; err != nil {
		logger.Log().Error("[TagUpdateService.UpdateTag] 保存标签失败: ", err)
		return serializer.DBErr("", err)
	}
	return serializer.Success(serializer.BuildTag(tag))
}

// DeleteTag 删除标签
func (service TagDeleteService) DeleteTag(tagID string) serializer.Response {
	if err := model.DB.Where("id = ?", tagID).Delete(&model.Tag{}).Error; err != nil {
		logger.Log().Error("[TagListService.DeleteTag] 删除标签失败: ", err)
		return serializer.DBErr("", err)
	}
	return serializer.Success(nil)
}

// ListTags 返回所有标签
func (service TagListService) ListTags() serializer.Response {
	var tags []model.Tag
	if err := model.DB.Find(&tags).Error; err != nil {
		logger.Log().Error("[TagListService.ListTags] 查询标签失败: ", err)
		return serializer.DBErr("", err)
	}
	return serializer.Success(serializer.BuildTags(tags))
}

// GetTag 获取单个标签信息
func (service TagListService) GetTag(tagID string) serializer.Response {
	var tag model.Tag
	if err := model.DB.Where("id = ?", tagID).First(&tag).Error; err != nil {
		logger.Log().Error("[TagListService.GetTag] 查找标签失败: ", err)
		return serializer.DBErr("", err)
	}
	return serializer.Success(serializer.BuildTag(tag))
}
