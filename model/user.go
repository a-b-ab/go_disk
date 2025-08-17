package model

import (
	"fmt"

	"go-cloud-disk/conf"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Uuid                 string `gorm:"primarykey"`
	UserName             string
	PasswordDigest       string
	NickName             string
	Status               string
	Avatar               string `gorm:"size:1000"`
	UserFileStoreID      string
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
	user.Uuid = uuid.New().String()
	fileStoreId, err := CreateFileStore(user.Uuid)
	if err != nil {
		return fmt.Errorf("创建文件存储错误 %v", err)
	}
	mainFileFolderId, err := CreateBaseFileFolder(user.Uuid, fileStoreId)
	if err != nil {
		return fmt.Errorf("创建基础文件夹错误 %v", err)
	}

	user.UserFileStoreID = fileStoreId
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
