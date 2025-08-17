package filefolder

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

// FileFolderUpdateService 文件夹更新服务结构体
type FileFolderUpdateService struct {
	FileFolderId      string `json:"filefolder" form:"filefolder" binding:"required"` // 文件夹ID
	NewFileFolderName string `json:"name" form:"name"`                                // 新文件夹名称
	NewParentId       string `json:"parent" form:"parent" binding:"required"`         // 新父文件夹ID
}

// UpdateFileFolderInfo 更新文件夹信息，包括文件夹名称和所属位置
func (service *FileFolderUpdateService) UpdateFileFolderInfo(userid string) serializer.Response {
	var filefolder model.FileFolder
	var err error
	// 找当前用户要修改的文件夹
	if err := model.DB.Where("uuid = ? and owner_id = ?", service.FileFolderId, userid).Find(&filefolder).Error; err != nil {
		logger.Log().Error("[FileFolderUpdateService.UpdateFileFolderInfo] 查找文件夹信息失败: ", err)
		return serializer.DBErr("", err)
	}

	// 检查目标文件夹所有者
	var targetFilefolder model.FileFolder
	if err := model.DB.Where("uuid = ? and owner_id = ?", service.NewParentId, userid).Find(&targetFilefolder).Error; err != nil {
		logger.Log().Error("[FileFolderUpdateService.UpdateFileFolderInfo] 查找新父文件夹信息失败: ", err)
		return serializer.DBErr("", err)
	}

	// 找到旧父文件夹
	var parentFilefolder model.FileFolder
	if err := model.DB.Where("uuid = ?", filefolder.ParentFolderID).Find(&parentFilefolder).Error; err != nil {
		logger.Log().Error("[FileFolderUpdateService.UpdateFileFolderInfo] 查找旧父文件夹信息失败: ", err)
		return serializer.DBErr("", err)
	}

	// 获取新的文件夹信息
	newFileFolderName := filefolder.FileFolderName
	if service.NewFileFolderName != "" {
		newFileFolderName = service.NewFileFolderName
	}
	filefolder.FileFolderName = newFileFolderName
	filefolder.ParentFolderID = service.NewParentId

	// 更新文件夹信息到数据库
	t := model.DB.Begin()
	defer func() {
		if err != nil {
			t.Rollback()
		} else {
			t.Commit()
		}
	}()

	if err := t.Save(&filefolder).Error; err != nil {
		logger.Log().Error("[FileFolderUpdateService.UpdateFileFolderInfo] 更新文件夹信息失败: ", err)
		return serializer.DBErr("", err)
	}

	// 更改文件夹大小
	// 移动一个文件夹，就像搬箱子 → 旧箱子里少一份体积，新箱子里多一份体积
	if targetFilefolder.Uuid != parentFilefolder.Uuid {
		err = parentFilefolder.SubFileFolderSize(t, filefolder.Size)
		if err != nil {
			logger.Log().Error("[FileFolderUpdateService.UpdateFileFolderInfo] 更新旧文件夹信息失败: ", err)
			return serializer.DBErr("", err)
		}
		targetFilefolder.AddFileFolderSize(t, filefolder.Size)
		if err != nil {
			logger.Log().Error("[FileFolderUpdateService.UpdateFileFolderInfo] 更新新文件夹信息失败: ", err)
			return serializer.DBErr("", err)
		}
	}

	return serializer.Success(nil)
}
