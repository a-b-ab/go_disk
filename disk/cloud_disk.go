package disk

import "go-cloud-disk/conf"

// CloudDisk 云盘接口定义，封装了云存储服务的基本操作
// 支持文件上传、下载、删除和存在性检查等功能
type CloudDisk interface {
	// GetUploadPresignedURL 生成预签名URL。
	// 用户可以使用预签名URL通过PUT方法上传文件
	GetUploadPresignedURL(userId string, filePath string, fileName string) (string, error)
	// GetDownloadPresignedURL 生成预签名URL。
	// 用户可以使用预签名URL通过GET方法下载文件
	GetDownloadPresignedURL(userId string, filePath string, fileName string) (string, error)
	// GetObjectURL 生成对象URL。用户可以使用URL查看文件。
	GetObjectURL(userId string, filePath string, fileName string) (string, error)
	// DeleteObject 删除用户对象
	DeleteObject(userId string, filePath string, items []string) error
	// DeleteObjectFilefolder 删除用户对象文件夹
	DeleteObjectFilefolder(userId string, filePath string) error
	// IsObjectExist 检查文件是否存在
	IsObjectExist(userId string, filePath string, fileName string) (bool, error)
	// UploadSimpleFile 上传小于1GB的文件到云端
	UploadSimpleFile(localFilePath string, userId string, md5 string, fileSize int64) error
}

// 确保TencentCloudDisk实现了CloudDisk接口
var _ CloudDisk = (*TencentCloudDisk)(nil)

// NewCloudDisk 云盘构造函数类型定义
type NewCloudDisk func() CloudDisk

// BaseCloudDisk 基础云盘实例，提供全局访问的云存储服务
var BaseCloudDisk CloudDisk

// NewCloudDiskMap 云盘类型映射表，根据配置选择对应的云存储实现
var NewCloudDiskMap map[string]NewCloudDisk

// init 初始化云盘映射表，注册可用的云存储提供商
func init() {
	NewCloudDiskMap = make(map[string]NewCloudDisk)
	NewCloudDiskMap["TENCENT"] = NewTencentCloudDisk
}

// SetBaseCloudDisk 设置基础云盘实例，根据配置文件选择对应的云存储提供商
func SetBaseCloudDisk() {
	version := conf.CloudDiskVersion
	ver, ok := NewCloudDiskMap[version]
	if !ok {
		panic("不支持此云盘版本")
	}
	BaseCloudDisk = ver()
}
