package test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const ChunkSize = 5 * 1024 * 1024 // 5MB

// SplitFileToChunks 将文件按指定分片大小切片，保存到临时目录
func SplitFileToChunks(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 创建临时目录存放分片
	tempDir := filepath.Join("./tmp_test_chunks")
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %v", err)
	}

	var chunkFiles []string
	totalSize := fileInfo.Size()
	totalChunks := int((totalSize + ChunkSize - 1) / ChunkSize)

	for i := 0; i < totalChunks; i++ {
		chunkFilePath := filepath.Join(tempDir, fmt.Sprintf("chunk_%d", i))
		chunkFiles = append(chunkFiles, chunkFilePath)

		chunkFile, err := os.Create(chunkFilePath)
		if err != nil {
			return nil, fmt.Errorf("创建分片文件失败: %v", err)
		}

		// 计算当前分片大小（最后一片可能小于 ChunkSize）
		currentChunkSize := int64(ChunkSize)
		if int64(i+1)*currentChunkSize > totalSize {
			currentChunkSize = totalSize - int64(i)*currentChunkSize
		}

		written, err := io.CopyN(chunkFile, file, currentChunkSize)
		if err != nil && err != io.EOF {
			chunkFile.Close()
			return nil, fmt.Errorf("写入分片失败: %v", err)
		}

		if written != currentChunkSize {
			chunkFile.Close()
			return nil, fmt.Errorf("写入分片大小不匹配: %d != %d", written, currentChunkSize)
		}

		chunkFile.Close()
	}

	return chunkFiles, nil
}

// PrintChunkInfo 打印分片信息（可选的辅助函数）
func PrintChunkInfo(filePath string, chunkFiles []string) {
	fileInfo, _ := os.Stat(filePath)

	fmt.Printf("🖼️  处理文件: %s\n", filepath.Base(filePath))
	fmt.Printf("📏 文件大小: %d bytes (%.2f MB)\n", fileInfo.Size(), float64(fileInfo.Size())/(1024*1024))
	fmt.Printf("✅ 切片完成! 共生成 %d 个分片文件:\n", len(chunkFiles))

	var totalChunkSize int64
	for i, chunkPath := range chunkFiles {
		chunkInfo, _ := os.Stat(chunkPath)
		totalChunkSize += chunkInfo.Size()
		fmt.Printf("   📦 chunk_%d: %s (%d bytes)\n", i, filepath.Base(chunkPath), chunkInfo.Size())
	}
}
