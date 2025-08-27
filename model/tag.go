package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Tag 标签模型
type Tag struct {
	ID   string `gorm:"primarykey" json:"id"`
	Name string `gorm:"unique;not null;size:100" json:"name"` // 标签名称，唯一
}

// FileTag 文件标签关联模型
type FileTag struct {
	ID     string `gorm:"primarykey" json:"id"`
	FileID string `gorm:"index" json:"file_id"`         // 文件ID，关联files表 外键 -> files.file_uuid
	TagID  string `gorm:"not null;index" json:"tag_id"` // 标签ID，关联tags表
	File   File   `gorm:"foreignKey:FileID;references:FileUuid;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"file"`
	Tag    Tag    `gorm:"foreignKey:TagID;references:ID" json:"tag"`
}

// BeforeCreate 在插入数据库前创建uuid
func (tag *Tag) BeforeCreate(tx *gorm.DB) (err error) {
	if tag.ID == "" {
		tag.ID = uuid.New().String()
	}
	return
}

// BeforeCreate 在插入数据库前创建uuid
func (fileTag *FileTag) BeforeCreate(tx *gorm.DB) (err error) {
	if fileTag.ID == "" {
		fileTag.ID = uuid.New().String()
	}
	return
}

// GetOrCreateTag 获取或创建标签
func GetOrCreateTag(tagName string) (*Tag, error) {
	var tag Tag
	err := DB.Where("name = ?", tagName).First(&tag).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 标签不存在，创建新标签
			tag = Tag{Name: tagName}
			if err := DB.Create(&tag).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return &tag, nil
}

// AddTagToFile 为文件添加标签
func AddTagToFile(fileID, tagID string) error {
	// 检查关联是否已存在
	var existingFileTag FileTag
	err := DB.Where("file_id = ? AND tag_id = ?", fileID, tagID).First(&existingFileTag).Error
	if err == nil {
		// 关联已存在
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	// 创建新的文件标签关联
	fileTag := FileTag{
		FileID: fileID,
		TagID:  tagID,
	}
	return DB.Create(&fileTag).Error
}

// GetFilesByTag 根据标签获取文件列表
func GetFilesByTag(tagName string, userID string, limit, offset int) ([]File, error) {
	var files []File
	err := DB.Joins("JOIN file_tags ON files.uuid = file_tags.file_id").
		Joins("JOIN tags ON file_tags.tag_id = tags.id").
		Where("tags.name = ? AND files.owner = ?", tagName, userID).
		Limit(limit).
		Offset(offset).
		Find(&files).Error
	return files, err
}

// GetTagsByFile 获取文件的所有标签
func GetTagsByFile(fileID string) ([]Tag, error) {
	var tags []Tag
	err := DB.Joins("JOIN file_tags ON tags.id = file_tags.tag_id").
		Where("file_tags.file_id = ?", fileID).
		Find(&tags).Error
	return tags, err
}
