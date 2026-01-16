package serializer

import "go-cloud-disk/model"

type Share struct {
	Uuid        string `json:"shareid"`
	FileId      string `json:"sharefileid"`
	Owner       string `json:"owner"`
	Title       string `json:"title"`
	Filename    string `json:"filename"`
	SharingTime string `json:"sharetime"`
	View        int64  `json:"view"`
	DownloadURL string `json:"downloadurl,omitempty"`
	Size        int64  `json:"filesize"`
	AuditStatus int    `json:"audit_status"`
	// 驳回原因（仅当 audit_status=2 时有意义）
	RejectReason string `json:"reject_reason,omitempty"`
}

func BuildShare(share model.Share) Share {
	return Share{
		Uuid:        share.Uuid,
		FileId:      share.FileId,
		Owner:       share.Owner,
		Title:       share.Title,
		Filename:    share.FileName,
		View:        share.ViewCount(),
		SharingTime: share.SharingTime,
		Size:        share.Size,
		AuditStatus: share.AuditStatus,
		RejectReason: share.RejectReason,
	}
}

func BuildShareWithDownloadUrl(share model.Share, url string) Share {
	return Share{
		Uuid:        share.Uuid,
		FileId:      share.FileId,
		Owner:       share.Owner,
		Title:       share.Title,
		Filename:    share.FileName,
		View:        share.ViewCount(),
		SharingTime: share.SharingTime,
		DownloadURL: url,
		Size:        share.Size,
		AuditStatus: share.AuditStatus,
		RejectReason: share.RejectReason,
	}
}

func BuildShares(Shares []model.Share) (shareSerializer []Share) {
	for _, share := range Shares {
		shareSerializer = append(shareSerializer, BuildShare(share))
	}
	return
}
