package middleware

import (
	"regexp"

	"go-cloud-disk/conf"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Cors set cors
func Cors() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	// 允许客户端请求时携带的请求头
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Cookie", "Authorization"}
	if gin.Mode() == gin.ReleaseMode {
		// 生产环境，严格指定允许跨域的域名（前端域名）
		config.AllowOrigins = []string{conf.FrontWeb}
	} else {
		// 测试环境，模糊匹配来自本地的请求
		config.AllowOriginFunc = func(origin string) bool {
			if regexp.MustCompile(`^http://127\.0\.0\.1:\d+$`).MatchString(origin) {
				return true
			}
			if regexp.MustCompile(`^http://localhost:\d+$`).MatchString(origin) {
				return true
			}
			return false
		}
	}
	// 允许携带cookie
	config.AllowCredentials = true
	// 返回基于上述配置的CORS中间件
	return cors.New(config)
}
