package disk

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tiia "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tiia/v20190529"
)

// TencentImageTag 腾讯云标签识别实现
type TencentImageTag struct {
	secretId  string
	secretKey string
	region    string
	client    *tiia.Client
}

// TagResult 标签识别结果
type TagResult struct {
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"` // 置信度
}

// NewTencentImageTag 创建新的图像标签实例
func NewTencentImageTag() *TencentImageTag {
	return &TencentImageTag{
		secretId:  os.Getenv("BUCKET_SECRET_ID"),
		secretKey: os.Getenv("BUCKET_SECRET_KEY"),
		region:    os.Getenv("BUCKET_REGION"),
	}
}

// getDefaultClient 获取默认的腾讯云TIIA客户端
func (tag *TencentImageTag) getDefaultClient() (*tiia.Client, error) {
	if tag.client != nil {
		return tag.client, nil
	}
	cred := common.NewCredential(tag.secretId, tag.secretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "tiia.tencentcloudapi.com"

	client, err := tiia.NewClient(cred, tag.region, cpf)
	if err != nil {
		return nil, fmt.Errorf("创建腾讯云图像标签识别客户端失败: %v", err)
	}
	tag.client = client
	return tag.client, nil
}

// DetectLabels 图像标签识别
func (tag *TencentImageTag) DetectLabels(imageURL string) ([]TagResult, error) {
	client, err := tag.getDefaultClient()
	if err != nil {
		return nil, err
	}

	request := tiia.NewDetectLabelProRequest()
	request.ImageUrl = common.StringPtr(imageURL)

	response, err := client.DetectLabelPro(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return nil, fmt.Errorf("腾讯云API错误: %v", err)
	}
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}

	var results []TagResult
	for _, label := range response.Response.Labels {
		if label.Name != nil && label.Confidence != nil {
			results = append(results, TagResult{
				Name:       *label.Name,
				Confidence: float64(*label.Confidence),
			})
		}
	}
	return results, nil
}

// DetectLabels 图像标签识别（Base64）
func (tag *TencentImageTag) DetectLabelBase64(imageBase64 string) ([]TagResult, error) {
	client, err := tag.getDefaultClient()
	if err != nil {
		return nil, err
	}

	request := tiia.NewDetectLabelProRequest()
	request.ImageBase64 = common.StringPtr(imageBase64)

	response, err := client.DetectLabelPro(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return nil, fmt.Errorf("腾讯云API错误: %v", err)
	}
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}

	var results []TagResult
	for _, label := range response.Response.Labels {
		if label.Name != nil && label.Confidence != nil {
			results = append(results, TagResult{
				Name:       *label.Name,
				Confidence: float64(*label.Confidence),
			})
		}
	}
	return results, nil
}

// FilterSensitiveWords 过滤敏感词
func (tag *TencentImageTag) FilterSensitiveWords(tags []TagResult) []TagResult {
	sensitiveWords := map[string]bool{
		"暴力": true,
		"血腥": true,
		"恐怖": true,
		"政治": true,
		"敏感": true,
	}
	var filtered []TagResult
	for _, t := range tags {
		if !sensitiveWords[t.Name] {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// TagResultsToJSON 序列化标签结果
func (tag *TencentImageTag) TagResultsToJSON(tags []TagResult) (string, error) {
	data, err := json.Marshal(tags)
	if err != nil {
		return "", fmt.Errorf("标签结果序列化失败: %v", err)
	}
	return string(data), nil
}

// TagResultsFromJSON 反序列化标签结果
func (tag *TencentImageTag) TagResultsFromJSON(jsonStr string) ([]TagResult, error) {
	var tags []TagResult
	err := json.Unmarshal([]byte(jsonStr), &tags)
	if err != nil {
		return nil, fmt.Errorf("标签结果反序列化失败: %v", err)
	}
	return tags, nil
}
