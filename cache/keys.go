package cache

import "fmt"

const (
	// DailyRankKey 每日查看排行榜
	DailyRankKey = "rank:daily"
	// EmptyShare 存储空分享键的集合
	EmptyShare = "share:empty"
)

// ShareKey 使用ID构建缓存中的分享键
func ShareKey(id string) string {
	return fmt.Sprintf("share:%s", id)
}

// ShareInfoKey 使用ID构建缓存中的分享信息键
func ShareInfoKey(id string) string {
	return fmt.Sprintf("info:share:%s", id)
}

// FileInfoStoreKey 使用ID构建缓存中的文件存储信息键
func FileInfoStoreKey(id string) string {
	return fmt.Sprintf("file:cloud:%s", id)
}

// EmailCodeKey 用于在缓存中存储确认码
func EmailCodeKey(email string) string {
	return fmt.Sprintf("email:%s", email)
}

// RecentSendUserKey 存储用户最近的请求
func RecentSendUserKey(email string) string {
	return fmt.Sprintf("user:confirm:%s", email)
}

// ChunkUploadInfoKey 分片上传信息键
func ChunkUploadInfoKey(uploadId string) string {
	return fmt.Sprintf("chunk:upload:%s", uploadId)
}
