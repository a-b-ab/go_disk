package rank

import (
	"context"
	"sort"

	"go-cloud-disk/cache"
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

// GetDailyRankService 获取日排行榜服务结构体
type GetDailyRankService struct{}

// GetDailyRank 获取分享文件的日排行榜
func (service *GetDailyRankService) GetDailyRank() serializer.Response {
	shares := make([]model.Share, 0, 16)

	// 从缓存中获取分享排行榜
	shareRank, err := cache.RedisClient.ZRevRange(context.Background(), cache.DailyRankKey, 0, 9).Result()
	if err != nil {
		logger.Log().Error("[GetDailyRankService.GetDailyRank] 从缓存获取排行榜失败: ", err)
		return serializer.DBErr("", err)
	}

	if len(shareRank) > 0 {
		err := model.DB.Model(&model.Share{}).Where("uuid in (?)", shareRank).Find(&shares).Error
		if err != nil {
			logger.Log().Error("[GetDailyRankService.GetDailyRank] 从数据库获取排行榜失败: ", err)
			return serializer.DBErr("", err)
		}
	}

	// 用空分享填充分享列表
	emptyShare := model.Share{
		Uuid:        "",
		Owner:       "",
		FileId:      "",
		Title:       "虚位以待",
		SharingTime: "",
	}
	for len(shares) < 10 {
		shares = append(shares, emptyShare)
	}

	// 对分享进行排序
	rspShare := serializer.BuildShares(shares)
	sort.Slice(rspShare, func(i, j int) bool {
		return rspShare[i].View > rspShare[j].View
	})

	return serializer.Success(rspShare)
}
