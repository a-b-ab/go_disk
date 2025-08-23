package disk

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// TencentCloudDisk 腾讯云对象存储(COS)实现
type TencentCloudDisk struct {
	bucketname string // 存储桶名称
	secretId   string // 腾讯云SecretId
	secretKey  string // 腾讯云SecretKey
}

// getDefaultClient 获取默认的腾讯云COS客户端
func (cloud *TencentCloudDisk) getDefaultClient() *cos.Client {
	u, _ := url.Parse(cloud.bucketname)
	b := &cos.BaseURL{BucketURL: u}
	c := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cloud.secretId,
			SecretKey: cloud.secretKey,
		},
	})
	return c
}

// getUploadPresignedURLPresigned 使用文件键生成预签名上传URL
// 用户可以使用预签名URL来上传文件
func (cloud *TencentCloudDisk) getUploadPresignedURLPresigned(key string) (string, error) {
	u, _ := url.Parse(cloud.bucketname)
	b := &cos.BaseURL{BucketURL: u}
	c := cos.NewClient(b, &http.Client{})
	ctx := context.Background()

	opt := &cos.PresignedURLOptions{
		Query:  &url.Values{},
		Header: &http.Header{},
	}
	opt.Query.Add("x-cos-security-token", "<token>")
	presignedURL, err := c.Object.GetPresignedURL(ctx, http.MethodPut, key, cloud.secretId, cloud.secretKey, time.Minute*15, opt)
	if err != nil {
		return "", fmt.Errorf("创建上传预签名URL错误：%v", err)
	}
	return presignedURL.String(), nil
}

// GetUploadPresignedURL 使用用户ID、文件路径、文件名生成云盘键并获取上传预签名URL
func (cloud *TencentCloudDisk) GetUploadPresignedURL(userId string, filePath string, fileName string) (string, error) {
	key := fastBuildKey(userId, filePath, fileName)
	presignedURL, err := cloud.getUploadPresignedURLPresigned(key)
	if err != nil {
		return "", err
	}
	return presignedURL, nil
}

// getDownloadPresignedURL 根据文件键生成下载预签名URL
// 废弃⚠️
func (cloud *TencentCloudDisk) getDownloadPresignedURL(key string) (string, error) {
	client := cloud.getDefaultClient()
	ctx := context.Background()
	presignedURL, err := client.Object.GetPresignedURL(ctx, http.MethodGet, key, cloud.secretId, cloud.secretKey, time.Hour, nil)
	if err != nil {
		return "", fmt.Errorf("创建下载预签名URL错误：%v", err)
	}
	return presignedURL.String(), nil
}

// GetDownloadPresignedURL 使用用户ID、文件路径、文件名生成云盘键并获取下载预签名URL
func (cloud *TencentCloudDisk) GetDownloadPresignedURL(userId string, filePath string, fileName string) (string, error) {
	key := fastBuildKey(userId, filePath, fileName)
	presignedURL, err := cloud.getDownloadPresignedURL(key)
	if err != nil {
		return "", err
	}
	return presignedURL, nil
}

// getObjectUrl 使用文件键生成对象URL，用户可以使用此URL下载文件或查看图片
func (cloud *TencentCloudDisk) getObjectUrl(key string) (str string, err error) {
	var ok bool
	if ok, err = cloud.checkObjectIsExist(key); err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("此对象在云端不存在")
	}
	client := cloud.getDefaultClient()
	ourl := client.Object.GetObjectURL(key)
	return ourl.String(), nil
}

// GetObjectURL 使用用户ID、文件路径、文件名生成云盘键并获取对象URL
func (cloud *TencentCloudDisk) GetObjectURL(userId string, filePath string, fileName string) (string, error) {
	key := fastBuildKey(userId, filePath, fileName)
	objectURL, err := cloud.getObjectUrl(key)
	if err != nil {
		return "", err
	}
	return objectURL, nil
}

// deleteObject 在云端删除多个对象
func (cloud *TencentCloudDisk) deleteObject(keys []string) error {
	client := cloud.getDefaultClient()
	obs := []cos.Object{}
	for _, v := range keys {
		obs = append(obs, cos.Object{Key: v})
	}
	opt := &cos.ObjectDeleteMultiOptions{
		Objects: obs,
	}

	_, _, err := client.Object.DeleteMulti(context.Background(), opt)
	if err != nil {
		return fmt.Errorf("删除对象错误：%v", err)
	}
	return nil
}

// DeleteObject 使用文件列表构建文件键并删除对象
func (cloud *TencentCloudDisk) DeleteObject(userId string, filePath string, items []string) error {
	var keys []string
	for _, file := range items {
		key := fastBuildKey(userId, filePath, file)
		keys = append(keys, key)
	}
	err := cloud.deleteObject(keys)
	return err
}

// deleteFilefold 删除文件夹及其所有内容
func (cloud *TencentCloudDisk) deleteFilefold(dir string) error {
	client := cloud.getDefaultClient()
	var marker string // 分页查询的游标，标记下一页起点
	opt := &cos.BucketGetOptions{
		Prefix:  dir,
		MaxKeys: 1000,
	}

	isTruncated := true // 是否有更多内容需要处理
	var errInTruncated error
	for isTruncated {
		opt.Marker = marker
		v, _, err := client.Bucket.Get(context.Background(), opt) // 分页获取对象列表
		if err != nil {
			errInTruncated = err
			break
		}
		for _, content := range v.Contents {
			_, err = client.Object.Delete(context.Background(), content.Key)
			if err != nil {
				errInTruncated = err
				break
			}
		}
		if errInTruncated != nil {
			break
		}
		isTruncated = v.IsTruncated
		marker = v.NextMarker
	}
	if errInTruncated != nil {
		return errInTruncated
	}
	return nil
}

// DeleteObjectFilefolder 删除用户在云端的文件夹
func (cloud *TencentCloudDisk) DeleteObjectFilefolder(userId string, filePath string) error {
	key := fastBuildKey(userId, filePath, "")
	err := cloud.deleteFilefold(key)
	return err
}

// checkObjectIsExist 检查对象是否存在
func (cloud *TencentCloudDisk) checkObjectIsExist(key string) (bool, error) {
	client := cloud.getDefaultClient()
	ok, err := client.Object.IsExist(context.Background(), key)
	if err != nil {
		return false, err
	}
	return ok, nil
}

// IsObjectExist 检查对象是否存在
func (cloud *TencentCloudDisk) IsObjectExist(userId string, filePath string, fileName string) (bool, error) {
	key := fastBuildKey(userId, filePath, fileName)
	ok, err := cloud.checkObjectIsExist(key)
	return ok, err
}

// uploadSimpleFile 使用PutFromFile将本地文件上传到云端
func (cloud *TencentCloudDisk) uploadSimpleFile(localFilePath string, key string) error {
	client := cloud.getDefaultClient()
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentDisposition: "attachment",
		},
	}
	_, err := client.Object.PutFromFile(context.Background(), key, localFilePath, opt)
	if err != nil {
		return err
	}
	return nil
}

// todo: 分片上传大文件（支持断点续传）

// UploadSimpleFile 上传小于1GB的文件到云端
func (cloud *TencentCloudDisk) UploadSimpleFile(localFilePath string, userId string, md5 string, fileSize int64) error {
	if fileSize/1024/1024/1024 > 1 {
		return fmt.Errorf("文件过大，请使用uploadfile方法")
	}

	// 检查文件是否已存在于云端
	extend := path.Ext(localFilePath)
	ok, err := cloud.IsObjectExist(userId, "", md5+extend)
	if err != nil {
		return err
	}

	// 如果云端不存在，则上传文件
	if !ok {
		key := fastBuildKey(userId, "", md5+extend)
		if err = cloud.uploadSimpleFile(localFilePath, key); err != nil {
			return err
		}
	}

	return nil
}

// GetDownloadURL 根据文件路径生成预签名下载URL
func (cloud *TencentCloudDisk) GetDownloadURL(filePath string, fileUUID string) (string, error) {
	// 确保bucket url包含协议
	bucketURL := cloud.bucketname
	if !strings.HasPrefix(bucketURL, "http://") && !strings.HasPrefix(bucketURL, "https://") {
		bucketURL = "https://" + bucketURL
	}
	// 拼接对象key
	key := fmt.Sprintf("user/%s/%s.png", filePath, fileUUID)

	u, err := url.Parse(bucketURL)
	if err != nil {
		return "", nil
	}

	client := &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cloud.secretId,
			SecretKey: cloud.secretKey,
		},
	}

	cosClient := cos.NewClient(&cos.BaseURL{BucketURL: u}, client)

	// 生成预签名下载URL，有效期24小时
	presignedURL, err := cosClient.Object.GetPresignedURL(
		context.Background(),
		http.MethodGet,
		key,
		cloud.secretId,
		cloud.secretKey,
		24*time.Hour,
		nil,
	)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

// fastBuildKey 使用用户ID、文件路径、文件名通过Builder生成文件键
func fastBuildKey(userId string, filePath string, file string) string {
	var key strings.Builder
	key.Write([]byte("user/"))
	if userId != "" {
		key.Write([]byte(userId))
		key.Write([]byte("/"))
	}
	if filePath != "" {
		key.Write([]byte(filePath))
		key.Write([]byte("/"))
	}
	key.Write([]byte(file))
	return key.String()
}

// NewTencentCloudDisk 创建新的腾讯云盘实例
func NewTencentCloudDisk() CloudDisk {
	return &TencentCloudDisk{
		bucketname: os.Getenv("BUCKET_NAME"),
		secretId:   os.Getenv("BUCKET_SECRET_ID"),
		secretKey:  os.Getenv("BUCKET_SECRET_KEY"),
	}
}
