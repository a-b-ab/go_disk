package share

import (
	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

// ShareGetInfoService 获取分享信息服务结构体
type ShareGetInfoService struct{}

// GetShareInfo 获取分享信息
func (service *ShareGetInfoService) GetShareInfo(shareid string) serializer.Response {
	share := model.Share{
		Uuid: shareid,
	}
	// 先从数据库拿审核状态（避免未审核内容被 Redis 缓存绕过）
	if err := model.DB.Where("uuid = ?", shareid).First(&share).Error; err != nil {
		logger.Log().Error("[ShareGetInfoService.GetShareInfo] 获取分享信息失败: ", err)
		return serializer.DBErr("", err)
	}
	// 未审核通过：仅返回基础信息，不返回 downloadurl，不计 view，不写 Redis
	if share.AuditStatus != 1 {
		return serializer.Success(serializer.BuildShare(share))
	}

	// 尝试从Redis获取分享信息
	if share.CheckRedisExistsShare() {
		_ = share.GetShareInfoFromRedis()
		// 动态生成预签名下载URL（不要使用缓存）
		downloadUrl, err := share.DownloadURL()
		if err != nil {
			share.SetEmptyShare()
			return serializer.Success(serializer.BuildShareWithDownloadUrl(share, ""))
		}
		// 检查是否为空分享
		if downloadUrl != "" {
			share.AddViewCount()
		}
		return serializer.Success(serializer.BuildShareWithDownloadUrl(share, downloadUrl))
	}

	// 获取下载URL，如果无法获取下载URL说明分享已被删除
	downloadUrl, err := share.DownloadURL()
	if err != nil {
		share.SetEmptyShare()
	}

	// 如果日查看次数超过20次，将其添加到Redis中
	// 以提高搜索速度
	if share.DailyViewCount() > 20 {
		// 如果是空分享则从日排行榜中移除
		err := share.SaveShareInfoToRedis(downloadUrl)
		if err != nil {
			logger.Log().Error(err.Error())
		}
	}

	// 增加分享查看次数
	if downloadUrl != "" {
		share.AddViewCount()
	}
	return serializer.Success(serializer.BuildShareWithDownloadUrl(share, downloadUrl))
}
