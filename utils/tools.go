package utils

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ImageURLToBase64 从 COS 签名 URL 下载图片并转为 Base64 字符串
func ImageURLToBase64(imageURL string) (string, error) {
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("下载图片失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载图片失败: %v", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取图片数据失败: %v", err)
	}

	// 获取MIME类型，；例如:"image/jpeg"
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpg"
	}

	// 编码为 Base64 并拼接 data URI
	base64Str := base64.StdEncoding.EncodeToString(data)
	dataURI := fmt.Sprintf("data:%s;base64,%s", strings.ToLower(contentType), base64Str)

	return dataURI, nil
}
