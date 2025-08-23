package middleware

import (
	"log"
	"strings"
	"time"

	"go-cloud-disk/auth"
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT认证中间件，检查JWT认证并保存JWT信息
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// token格式 Authorization: "Bearer token"
		authorization := c.Request.Header.Get("Authorization")
		if authorization == "" {
			c.JSON(200, serializer.NotLogin("Need Token"))
			c.Abort()
			return
		}

		parts := strings.Split(authorization, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(200, serializer.NotLogin("Token format error"))
			c.Abort()
			return
		}

		// 解析token
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			c.JSON(200, serializer.NotLogin("Token error"))
			c.Abort()
			return
		}

		// 检查token是否已过期
		if time.Now().Unix() > claims.ExpiresAt.Unix() {
			c.JSON(200, serializer.NotLogin("Token expiration"))
			c.Abort()
			return
		}

		c.Set("UserId", claims.UserId)
		c.Set("UserName", claims.UserName)
		c.Set("Status", claims.Status)

		c.Next()
	}
}

// CasbinAuth Casbin权限认证中间件
func CasbinAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取对象和操作
		userStatus := c.MustGet("Status").(string)
		method := c.Request.Method
		path := c.Request.URL.Path
		object := strings.TrimPrefix(path, "/api/v1/")

		// 把它们作为 主体（sub）、客体（obj）、动作（act） 传给 Casbin：
		if ok, _ := auth.Casbin.Enforce(userStatus, object, method); !ok {
			c.JSON(200, serializer.NotAuthErr("not auth"))
			c.Abort()
		}
	}
}

// AdminAuth 管理员权限认证中间件
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := c.MustGet("UserId").(string)
		var user model.User
		if err := model.DB.Where("uuid = ?", uuid).Find(&user).Error; err != nil {
			log.Println("检查管理员权限时获取用户信息失败", err)
			c.Abort()
			return
		}
		if user.Status != c.MustGet("Status").(string) {
			c.JSON(200, serializer.NotAuthErr("change jwt!!!"))
			c.Abort()
			return
		}
	}
}
