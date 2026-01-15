package model

import (
	"errors"
	"fmt"

	"go-cloud-disk/conf"
	"go-cloud-disk/idgen"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	Uuid                 string `gorm:"primarykey"`
	UserName             string
	PasswordDigest       string
	NickName             string
	Status               string
	UserMainFileFolderID string
}

const (
	// PasswordCount 密码加密难度
	PasswordCount = 12
	// StatusSuperAdmin 超级管理员
	StatusSuperAdmin = "super_admin"
	// StatusAdmin 普通管理员
	StatusAdmin = "common_admin"
	// StatusActiveUser 激活用户
	StatusActiveUser = "active"
	// StatusInactiveUser 未激活用户
	StatusInactiveUser = "inactive"
	// StatusSuspendUser 暂停用户
	StatusSuspendUser = "suspend"
)

// SetPassword 加密用户密码以保存数据
func (user *User) SetPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), PasswordCount)
	if err != nil {
		return err
	}
	user.PasswordDigest = string(bytes)
	return nil
}

// CheckPassword 检查用户密码
func (user *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordDigest), []byte(password))
	return err == nil
}

// CreateUser 在数据库中创建用户，并为用户绑定一个文件存储
func (user *User) CreateUser() error {
	// 使用 6 位 Base62 短ID 作为用户ID（通过数据库唯一性兜底，并在创建前做一次碰撞规避）
	for i := 0; i < 20; i++ {
		id, err := idgen.RandomBase62(6)
		if err != nil {
			return fmt.Errorf("生成用户ID失败 %v", err)
		}
		var exist User
		err = DB.Select("uuid").Where("uuid = ?", id).First(&exist).Error
		if err == nil {
			continue // 碰撞，重试
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user.Uuid = id
			break
		}
		return fmt.Errorf("检查用户ID唯一性失败 %v", err)
	}
	if user.Uuid == "" {
		return fmt.Errorf("生成用户ID失败：重试次数过多")
	}
	fileStoreId, err := CreateFileStore(user.Uuid)
	if err != nil {
		return fmt.Errorf("创建文件存储错误 %v", err)
	}
	mainFileFolderId, err := CreateBaseFileFolder(user.Uuid, fileStoreId)
	if err != nil {
		return fmt.Errorf("创建基础文件夹错误 %v", err)
	}

	// 为用户初始化回收站配置（确保定时任务能覆盖到“从未打开回收站页面”的用户）
	// 说明：主键 ID 使用 userID，避免空字符串主键导致冲突。
	var cfg RecycleBinConfig
	if err := DB.Where("user_id = ?", user.Uuid).First(&cfg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			cfg = RecycleBinConfig{
				ID:              user.Uuid,
				UserID:          user.Uuid,
				AutoCleanDays:   30,
				EnableAutoClean: 1,
			}
			if err := DB.Create(&cfg).Error; err != nil {
				return fmt.Errorf("创建回收站配置错误 %v", err)
			}
		} else {
			return fmt.Errorf("检查回收站配置错误 %v", err)
		}
	}

	user.UserMainFileFolderID = mainFileFolderId
	if err := DB.Create(user).Error; err != nil {
		return fmt.Errorf("创建用户错误 %v", err)
	}

	return nil
}

func createSuperAdmin() error {
	admin := User{
		UserName: conf.AdminUserName,
		NickName: conf.AdminUserName,
		Status:   StatusSuperAdmin,
	}

	if err := admin.SetPassword(conf.AdminPassword); err != nil {
		return fmt.Errorf("设置超级管理员密码错误 %v", err)
	}
	if err := admin.CreateUser(); err != nil {
		return fmt.Errorf("创建超级管理员用户错误 %v", err)
	}

	return nil
}
