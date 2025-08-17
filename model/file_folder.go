package model

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileFolder struct {
	Uuid           string `gorm:"primarykey"` // 主键，自动生成
	FileFolderName string // 文件夹名称
	ParentFolderID string // 父文件ID，支持层级结构
	FileStoreID    string // 关联的存储空间
	OwnerID        string // 所有者
	Size           int64  // 文件夹大小
}

// BeforeCreate 在插入数据库前创建uuid
func (fileFolder *FileFolder) BeforeCreate(tx *gorm.DB) (err error) {
	if fileFolder.Uuid == "" {
		fileFolder.Uuid = uuid.New().String()
	}
	return
}

// CreateBaseFileFolder 为用户创建文件夹，使用fileStoreId和ownerId，
// 并返回其uuid或错误
func CreateBaseFileFolder(ownerId string, fileStoreId string) (string, error) {
	fileStore := FileFolder{
		FileFolderName: "main",      // 文件夹名称：main（主目录）
		ParentFolderID: "root",      // 父目录ID：root（表示顶级目录）
		FileStoreID:    fileStoreId, // 文件存储ID
		OwnerID:        ownerId,     // 所有者用户ID
	}
	if err := DB.Create(&fileStore).Error; err != nil {
		return "", err
	}
	return fileStore.Uuid, nil
}

// SubSize 减少文件夹大小
func (fileFolder *FileFolder) SubSize(size int64) error {
	fileFolder.Size = max(fileFolder.Size-size, 0)
	return nil
}

// AddFileFolderSize 增加文件夹大小并为父文件夹增加大小（事务保证）
func (fileFolder *FileFolder) AddFileFolderSize(t *gorm.DB, appendSize int64) (err error) {
	// 为文件夹增加大小
	fileFolder.Size += appendSize
	parentId := fileFolder.ParentFolderID
	if err := t.Save(fileFolder).Error; err != nil {
		return fmt.Errorf("增加文件大小时保存文件夹出错 %v", err)
	}

	// 为父文件夹增加大小（循环处理）
	for parentId != "root" && parentId != "" {
		var nowFileFolder FileFolder
		if err = t.Where("uuid = ?", parentId).Find(&nowFileFolder).Error; err != nil {
			return fmt.Errorf("增加文件大小时查找文件夹出错 %v", err)
		}
		// 没有找到父文件夹
		if nowFileFolder.Uuid == "" {
			break
		}
		nowFileFolder.Size += appendSize
		if err = t.Save(&nowFileFolder).Error; err != nil {
			return fmt.Errorf("增加文件大小时保存文件夹出错 %v", err)
		}
		parentId = nowFileFolder.ParentFolderID
	}

	return
}

// SubFileFolderSize 减少文件夹大小并为父文件夹减少大小（事务保证）
func (fileFolder *FileFolder) SubFileFolderSize(t *gorm.DB, size int64) (err error) {
	// 为文件夹减少大小
	fileFolder.SubSize(size)
	parentId := fileFolder.ParentFolderID

	if err = t.Save(fileFolder).Error; err != nil {
		return fmt.Errorf("减少文件大小时保存文件夹出错 %v", err)
	}

	// 为父文件夹减少大小
	for parentId != "root" && parentId != "" {
		var nowFileFolder FileFolder
		if err = t.Where("uuid = ?", parentId).Find(&nowFileFolder).Error; err != nil {
			return fmt.Errorf("减少文件大小时查找文件夹出错 %v", err)
		}
		if nowFileFolder.Uuid == "" {
			break
		}
		nowFileFolder.SubSize(size)
		if err = t.Save(&nowFileFolder).Error; err != nil {
			return fmt.Errorf("减少文件大小时保存文件夹出错 %v", err)
		}
		parentId = nowFileFolder.ParentFolderID
	}

	return
}
