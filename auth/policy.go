package auth

import "go-cloud-disk/model"

func initPolicy() {
	// 添加基础权限策略
	Casbin.AddPolicies(
		[][]string{
			// 被封禁用户无法执行任何操作
			{model.StatusSuspendUser, "*", "*", "deny"},
			// 未激活用户无法创建文件和文件夹
			{model.StatusInactiveUser, "user*", "*", "allow"},
			{model.StatusInactiveUser, "file*", "GET", "allow"},
			{model.StatusInactiveUser, "file*", "DELETE", "allow"},
			{model.StatusInactiveUser, "filefolder*", "GET", "allow"},
			{model.StatusInactiveUser, "filefolder*", "DELETE", "allow"},
			{model.StatusInactiveUser, "filestore*", "GET", "allow"},
			{model.StatusInactiveUser, "share*", "GET", "allow"},
			{model.StatusInactiveUser, "share*", "DELETE", "allow"},
			// 激活用户可以创建文件、文件夹和分享
			{model.StatusActiveUser, "share*", "*", "allow"},
			{model.StatusActiveUser, "file*", "*", "allow"},
			{model.StatusActiveUser, "filefolder*", "*", "allow"},
			{model.StatusActiveUser, "rank*", "GET", "allow"},
			// 管理员用户可以修改用户状态
			{model.StatusAdmin, "admin/user*", "*", "allow"},
			{model.StatusAdmin, "admin/login*", "*", "allow"},
			{model.StatusAdmin, "admin/filestore*", "*", "allow"},
			{model.StatusAdmin, "admin/share*", "*", "allow"},
			{model.StatusAdmin, "admin/file*", "*", "allow"},
			// 超级管理员可以执行任何操作
			{model.StatusSuperAdmin, "*", "*", "allow"},
		},
	)

	// 添加角色分组策略（角色继承）
	Casbin.AddGroupingPolicies(
		[][]string{
			// 激活用户继承未激活用户的权限
			{model.StatusActiveUser, model.StatusInactiveUser},
			// 管理员继承激活用户的权限
			{model.StatusAdmin, model.StatusActiveUser},
		},
	)
	// 保存策略到持久化存储
	Casbin.SavePolicy()
}
