package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path"
	"strings"
)

// FastBuildFileName 快速构建文件名，将文件名和文件后缀拼接
func FastBuildFileName(fileName string, filePostfix string) string {
	var res strings.Builder
	res.Write([]byte(fileName))
	res.Write([]byte("."))
	res.Write([]byte(filePostfix))
	return res.String()
}

// FastBuildString 快速构建字符串，将多个字符串拼接
func FastBuildString(str ...string) string {
	var res strings.Builder
	for _, s := range str {
		res.Write([]byte(s))
	}
	return res.String()
}

// GetFileMD5 获取文件的MD5校验码
func GetFileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	defer func() {
		_ = file.Close()
	}()
	if err != nil {
		return "", err
	}
	hash := md5.New()
	// 将扩展名添加到MD5计算中，因为MD5计算会对内容相同但扩展名不同的文件产生相同的MD5码
	ext := path.Ext(file.Name())
	hash.Write([]byte(ext))
	_, _ = io.Copy(hash, file)
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// SplitFilename 分割文件名，将file.filename拆分为文件名和扩展名
func SplitFilename(str string) (filename string, extend string) {
	for i := len(str) - 1; i >= 0 && str[i] != '/'; i-- {
		if str[i] == '.' {
			if i != 0 {
				filename = str[:i]
			}
			if i != len(str)-1 {
				extend = str[i+1:]
			}
			return
		}
	}
	return str, ""
}

// GetBytesMD5 获取字节数组的MD5校验码
func GetBytesMD5(data []byte) string {
	hash := md5.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

// UniqueAndSortInts 对整数切片去重并排序
func UniqueAndSortInts(nums []int) []int {
	if len(nums) == 0 {
		return nums
	}

	// 使用map去重
	uniqueMap := make(map[int]bool)
	for _, num := range nums {
		uniqueMap[num] = true
	}

	// 转换回切片
	result := make([]int, 0, len(uniqueMap))
	for num := range uniqueMap {
		result = append(result, num)
	}

	// 排序
	for i := 0; i < len(result)-1; i++ {
		for j := 0; j < len(result)-1-i; j++ {
			if result[j] > result[j+1] {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}

	return result
}

// ContainsInt 检查切片是否包含指定整数
func ContainsInt(nums []int, target int) bool {
	for _, num := range nums {
		if num == target {
			return true
		}
	}
	return false
}

// EnsureDir 确保目录存在，如果不存在则创建
func EnsureDir(dirPath string) error {
	return os.MkdirAll(dirPath, 0o755)
}

// WriteBytesToFile 将字节数组写入文件
func WriteBytesToFile(filePath string, data []byte) error {
	return os.WriteFile(filePath, data, 0o644)
}

// RemoveFile 删除文件
func RemoveFile(filePath string) error {
	return os.Remove(filePath)
}
