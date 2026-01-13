package serializer

import "go-cloud-disk/model"

// User 用户序列化器
type User struct {
	ID                   string `json:"id"`
	UserName             string `json:"username"`
	NickName             string `json:"nickname"`
	UserMainFileFolderID string `json:"filefolder"`
	Status               string `json:"status"`
}

// BuildUser 返回用户序列化器
func BuildUser(user model.User) User {
	return User{
		ID:                   user.Uuid,
		UserName:             user.UserName,
		NickName:             user.NickName,
		Status:               user.Status,
		UserMainFileFolderID: user.UserMainFileFolderID,
	}
}

// BuildUsers 返回用户序列化器列表
func BuildUsers(users []model.User) (usersSerializer []User) {
	for _, user := range users {
		usersSerializer = append(usersSerializer, BuildUser(user))
	}
	return
}
