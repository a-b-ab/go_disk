package utils

import (
	"errors"
	"time"

	"go-cloud-disk/conf"
	"go-cloud-disk/model"
	"github.com/golang-jwt/jwt/v5"
)

type MyClaims struct {
	UserId               string `json:"user_id"`
	UserName             string `json:"user_name"`
	Status               string `json:"status"`
	jwt.RegisteredClaims        // 嵌入JWT标准声明
}

// GenToken 生成JWT令牌
func GenToken(issuer string, expireHour int, user *model.User) (string, error) {
	mySigningKey := []byte(conf.JwtKey)
	claims := MyClaims{
		UserId:   user.Uuid,
		UserName: user.UserName,
		Status:   user.Status,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHour) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	// 使用库创建JWT令牌对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 签名并生成最终令牌
	return token.SignedString(mySigningKey)
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string) (*MyClaims, error) {
	mySigningKey := []byte(conf.JwtKey)
	// 解析JWT令牌
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(t *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	})
	if err != nil {
		return nil, err
	}

	// 类型断言成功 && 令牌有效
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token 错误")
}
