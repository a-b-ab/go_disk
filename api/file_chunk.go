package api

import (
	"go-cloud-disk/serializer"
	"go-cloud-disk/service/file"
	"go-cloud-disk/service/file/chunk"

	"github.com/gin-gonic/gin"
)

// InitChunkUpload 初始化分片上传
func InitChunkUpload(c *gin.Context) {
	var service chunk.ChunkInitService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}
	userId, file, dst, err := getUploadFileParam(c)
	if err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}
	res := service.InitChunkUpload(userId, file, dst)
	c.JSON(200, res)
}

// UploadChunk 上传分片
func UploadChunk(c *gin.Context) {
	var service chunk.FileChunkUploadService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	// 读取分片数据
	chunkFile, err := c.FormFile("chunk_data")
	if err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	userId := c.MustGet("UserId").(string)
	res := service.UploadChunk(userId, chunkFile)
	c.JSON(200, res)
}

// CheckChunks 检查已上传的分片
func CheckChunks(c *gin.Context) {
	uploadId := c.Query("upload_id")
	if uploadId == "" {
		c.JSON(200, serializer.ParamsErr("upload_id is required", nil))
		return
	}

	service := file.FileChunkCheckService{
		UploadId: uploadId,
	}

	userId := c.MustGet("UserId").(string)
	res := service.CheckChunks(userId)
	c.JSON(200, res)
}

// CompleteChunkUpload 完成分片上传
func CompleteChunkUpload(c *gin.Context) {
	var service file.FileChunkCompleteService
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(200, serializer.ErrorResponse(err))
		return
	}

	userId := c.MustGet("UserId").(string)
	res := service.CompleteChunkUpload(userId)
	c.JSON(200, res)
}
