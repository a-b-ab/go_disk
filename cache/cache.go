package cache

import (
	"context"
	"strconv"

	"go-cloud-disk/conf"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// Redis 初始化Redis客户端
func Redis() {
	db, _ := strconv.ParseUint(conf.RedisDB, 10, 64)
	client := redis.NewClient(&redis.Options{
		Addr:       conf.RedisAddr,
		Password:   conf.RedisPassword,
		DB:         int(db),
		MaxRetries: 1,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		panic("无法连接到Redis")
	}

	RedisClient = client
}
