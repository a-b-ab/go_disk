package model

import (
	"go-cloud-disk/conf"
)

// migration 数据库迁移
func migration() {
	_ = DB.AutoMigrate(&User{})
	_ = DB.AutoMigrate(&File{})
	_ = DB.AutoMigrate(&FileFolder{})
	_ = DB.AutoMigrate(&FileStore{})
	_ = DB.AutoMigrate(&Share{})
	_ = DB.AutoMigrate(&Tag{})
	_ = DB.AutoMigrate(&FileTag{})
	_ = DB.AutoMigrate(&RecycleBin{})
	_ = DB.AutoMigrate(&RecycleBinConfig{})
	initSuperAdmin()
}

func initSuperAdmin() {
	// 创建超级管理员
	var count int64
	adminUserName := conf.AdminUserName
	if err := DB.Model(&User{}).Where("user_name = ?", adminUserName).Count(&count).Error; err != nil {
		panic("创建超级管理员失败 %v")
	}

	if count == 0 {
		if err := createSuperAdmin(); err != nil {
			panic("创建超级管理员失败 %v")
		}
	}
}
