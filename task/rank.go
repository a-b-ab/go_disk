package task

import (
	"context"

	"go-cloud-disk/cache"
)

// RestartDailyRank 重新计算日排行榜
func RestartDailyRank() error {
	// 日排行榜很可能是一个大key，使用
	// unlink删除以提高执行速度
	return cache.RedisClient.Unlink(context.Background(), cache.DailyRankKey).Err()
}
