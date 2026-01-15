package serializer

import "go-cloud-disk/model"

// User 用户序列化器
type User struct {
	ID                   string `json:"id"`
	UserName             string `json:"username"`
	NickName             string `json:"nickname"`
	UserMainFileFolderID string `json:"filefolder"`
	// FileStore 的 ID。当前实现中 FileStore 以 OwnerID（即用户ID）作为主键，因此这里直接返回用户ID。
	UserFileStoreID string `json:"filestore"`
	Status          string `json:"status"`
}

// BuildUser 返回用户序列化器
func BuildUser(user model.User) User {
	return User{
		ID:                   user.Uuid,
		UserName:             user.UserName,
		NickName:             user.NickName,
		Status:               user.Status,
		UserMainFileFolderID: user.UserMainFileFolderID,
		UserFileStoreID:      user.Uuid,
	}
}

// BuildUsers 返回用户序列化器列表
func BuildUsers(users []model.User) (usersSerializer []User) {
	for _, user := range users {
		usersSerializer = append(usersSerializer, BuildUser(user))
	}
	return
}
