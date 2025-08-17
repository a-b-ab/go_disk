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
	// 尝试从Redis获取分享信息
	if share.CheckRedisExistsShare() {
		downloadUrl := share.GetShareInfoFromRedis()
		// 检查是否为空分享
		if downloadUrl != "" {
			share.AddViewCount()
		}
		return serializer.Success(serializer.BuildShareWithDownloadUrl(share, downloadUrl))
	}

	// 无法从Redis获取分享信息时搜索数据库
	if err := model.DB.Where("uuid = ?", shareid).Find(&share).Error; err != nil {
		logger.Log().Error("[ShareGetInfoService.GetShareInfo] 获取分享信息失败: ", err)
		return serializer.DBErr("", err)
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
