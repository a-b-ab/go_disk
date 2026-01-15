package serializer

import "go-cloud-disk/model"

// Tag 标签序列化
type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// BuildTag 构建序列化 Tag
func BuildTag(tag model.Tag) Tag {
	return Tag{
		ID:   tag.ID,
		Name: tag.Name,
	}
}

// BuildTags 构建序列化 Tag 列表
func BuildTags(tags []model.Tag) (tagsSerializer []Tag) {
	for _, tag := range tags {
		tagsSerializer = append(tagsSerializer, BuildTag(tag))
	}
	return
}
