package auth

import (
	"go-cloud-disk/conf"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
)

// Casbin 全局权限执行器实例
var Casbin *casbin.Enforcer

// InitCasbin 初始化Casbin权限控制系统
func InitCasbin() {
	// 创建GORM适配器，用于将权限策略持久化到MySQL数据库
	// 参数: 数据库类型(mysql), 数据源名称(DSN), 是否自动创建表(true)
	a, err := gormadapter.NewAdapter("mysql", conf.MysqlDSN, true)
	if err != nil {
		panic(err)
	}

	// 从字符串创建RBAC模型配置
	// 定义了请求格式、策略格式、角色继承、策略效果和匹配规则
	m, err := model.NewModelFromString(`
	[request_definition]
	r = sub, obj, act
	
	[policy_definition]
	p = sub, obj, act, eft
	
	[role_definition]
	g = _, _
	
	[policy_effect]
	e = some(where (p.eft == allow)) && !some(where (p.eft == deny))
	
	[matchers]
	m = g(r.sub, p.sub) && keyMatch(r.act, p.act) && keyMatch(r.obj, p.obj)
	`)
	if err != nil {
		panic(err)
	}

	// 使用模型和适配器创建权限执行器
	e, err := casbin.NewEnforcer(m, a)
	if err != nil {
		panic(err)
	}
	Casbin = e

	// 从数据库加载已存在的权限策略
	Casbin.LoadPolicy()

	// 检查是否已有基础权限策略，如果没有则初始化默认策略
	// 测试管理员是否有访问用户管理接口的权限
	if ok, _ := Casbin.Enforce("common_admin", "admin/user", "POST"); !ok {
		initPolicy() // 初始化基础权限策略
	}
}
