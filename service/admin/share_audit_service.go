package admin

import (
	"time"

	"go-cloud-disk/model"
	"go-cloud-disk/serializer"
	"go-cloud-disk/utils/logger"
)

// ShareAuditService 管理员审核分享
type ShareAuditService struct {
	AuditStatus  int    `json:"audit_status" binding:"oneof=1 2"` // 1=通过 2=驳回
	RejectReason string `json:"reject_reason"`
}

func (service *ShareAuditService) AuditShare(shareID string, reviewerID string) serializer.Response {
	var share model.Share
	if err := model.DB.Where("uuid = ?", shareID).First(&share).Error; err != nil {
		return serializer.DBErr("分享不存在", err)
	}

	now := time.Now()
	share.AuditStatus = service.AuditStatus
	share.Reviewer = reviewerID
	share.ReviewedAt = &now
	if service.AuditStatus == 2 {
		share.RejectReason = service.RejectReason
	} else {
		share.RejectReason = ""
	}

	if err := model.DB.Save(&share).Error; err != nil {
		logger.Log().Error("[ShareAuditService.AuditShare] 保存审核结果失败: ", err)
		return serializer.DBErr("保存审核结果失败", err)
	}

	// 审核状态变化，清理 Redis 缓存，避免旧下载链接/旧内容
	share.DeleteShareInfoInRedis()

	return serializer.Success(serializer.BuildShare(share))
}

