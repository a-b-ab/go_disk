package serializer

import "go-cloud-disk/model"

// FileStore 文件存储序列化器
type FileStore struct {
	MaxSize     int64 `json:"maxsize"`     // 最大存储空间
	CurrentSize int64 `json:"currentsize"` // 当前已使用空间
}

// BuildFileStore 构建文件存储序列化器
func BuildFileStore(fileStore model.FileStore) FileStore {
	return FileStore{
		MaxSize:     fileStore.MaxSize,
		CurrentSize: fileStore.CurrentSize,
	}
}
