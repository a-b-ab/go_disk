# 大文件分片上传功能使用说明

## 功能概述

新增的大文件分片上传功能支持：
- **5MB 分片大小**：将大文件分成 5MB 的小块进行上传
- **断点续传**：支持网络中断后继续上传
- **Redis 存储**：分片信息存储在 Redis 中，支持集群部署
- **完整性校验**：通过 MD5 校验确保文件完整性

## API 接口说明

### 1. 初始化上传任务
**接口**: `POST /api/v1/file/chunk/init`

**请求参数**:
```json
{
  "filename": "large_file.zip",
  "filesize": 52428800,
  "filefolder": "folder-uuid-here"
}
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "upload_id": "upload-task-uuid",
    "chunk_size": 5242880,
    "total_chunks": 10,
    "file_size": 52428800
  },
  "msg": "success"
}
```

### 2. 上传单个分片
**接口**: `POST /api/v1/file/chunk/upload`

**请求参数**:
- `upload_id`: 上传任务ID（form 参数）
- `chunk_number`: 分片序号，从1开始（form 参数）
- 请求 body 为二进制分片数据

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "chunk_number": 1,
    "chunk_md5": "d41d8cd98f00b204e9800998ecf8427e",
    "uploaded_chunks": 1,
    "total_chunks": 10
  },
  "msg": "success"
}
```

### 3. 查询已上传分片
**接口**: `GET /api/v1/file/chunk/check?upload_id=upload-task-uuid`

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "upload_id": "upload-task-uuid",
    "uploaded_chunks": [1, 2, 3, 5],
    "total_chunks": 10,
    "file_name": "large_file.zip",
    "file_size": 52428800,
    "chunk_size": 5242880
  },
  "msg": "success"
}
```

### 4. 完成上传任务
**接口**: `POST /api/v1/file/chunk/complete`

**请求参数**:
```json
{
  "upload_id": "upload-task-uuid"
}
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "uuid": "file-uuid",
    "filename": "large_file",
    "file_postfix": "zip",
    "size": 52428800,
    // ... 其他文件信息
  },
  "msg": "success"
}
```

## 前端实现示例

### JavaScript 分片上传实现

```javascript
class ChunkUploader {
  constructor(file, folderId, chunkSize = 5 * 1024 * 1024) {
    this.file = file;
    this.folderId = folderId;
    this.chunkSize = chunkSize;
    this.uploadId = null;
    this.totalChunks = 0;
    this.uploadedChunks = [];
  }

  // 初始化上传任务
  async initUpload() {
    const response = await fetch('/api/v1/file/chunk/init', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + getToken()
      },
      body: JSON.stringify({
        filename: this.file.name,
        filesize: this.file.size,
        filefolder: this.folderId
      })
    });
    
    const result = await response.json();
    if (result.code === 200) {
      this.uploadId = result.data.upload_id;
      this.totalChunks = result.data.total_chunks;
      return true;
    }
    return false;
  }

  // 检查已上传的分片
  async checkUploaded() {
    const response = await fetch(`/api/v1/file/chunk/check?upload_id=${this.uploadId}`, {
      headers: {
        'Authorization': 'Bearer ' + getToken()
      }
    });
    
    const result = await response.json();
    if (result.code === 200) {
      this.uploadedChunks = result.data.uploaded_chunks;
      return this.uploadedChunks;
    }
    return [];
  }

  // 上传单个分片
  async uploadChunk(chunkNumber) {
    const start = (chunkNumber - 1) * this.chunkSize;
    const end = Math.min(start + this.chunkSize, this.file.size);
    const chunk = this.file.slice(start, end);

    const formData = new FormData();
    formData.append('upload_id', this.uploadId);
    formData.append('chunk_number', chunkNumber);

    const response = await fetch('/api/v1/file/chunk/upload', {
      method: 'POST',
      headers: {
        'Authorization': 'Bearer ' + getToken()
      },
      body: chunk // 直接发送二进制数据
    });
    
    return await response.json();
  }

  // 完成上传
  async completeUpload() {
    const response = await fetch('/api/v1/file/chunk/complete', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + getToken()
      },
      body: JSON.stringify({
        upload_id: this.uploadId
      })
    });
    
    return await response.json();
  }

  // 完整的上传流程（支持断点续传）
  async upload(onProgress) {
    // 1. 初始化上传任务
    if (!await this.initUpload()) {
      throw new Error('初始化上传失败');
    }

    // 2. 检查已上传的分片（断点续传）
    await this.checkUploaded();

    // 3. 上传缺失的分片
    for (let i = 1; i <= this.totalChunks; i++) {
      if (!this.uploadedChunks.includes(i)) {
        await this.uploadChunk(i);
        this.uploadedChunks.push(i);
        
        // 更新进度
        if (onProgress) {
          onProgress({
            uploaded: this.uploadedChunks.length,
            total: this.totalChunks,
            percentage: (this.uploadedChunks.length / this.totalChunks * 100).toFixed(2)
          });
        }
      }
    }

    // 4. 完成上传
    return await this.completeUpload();
  }
}

// 使用示例
async function handleFileUpload(file, folderId) {
  const uploader = new ChunkUploader(file, folderId);
  
  try {
    const result = await uploader.upload((progress) => {
      console.log(`上传进度: ${progress.percentage}%`);
    });
    
    if (result.code === 200) {
      console.log('文件上传成功', result.data);
    }
  } catch (error) {
    console.error('上传失败:', error);
  }
}
```

## 技术特点

1. **内存友好**: 分片数据临时存储在 Redis 中，不占用服务器磁盘空间
2. **断点续传**: 网络中断后可以继续上传，无需重新开始
3. **并发控制**: 可以控制同时上传的分片数量
4. **完整性校验**: 每个分片都有 MD5 校验，确保数据完整性
5. **自动清理**: 上传完成后自动清理 Redis 中的临时数据
6. **存储优化**: 相同 MD5 的文件会复用，节省存储空间

## 配置说明

- **分片大小**: 默认 5MB，在 `ChunkSize` 常量中定义
- **Redis 过期时间**: 分片信息保存 24 小时
- **临时文件目录**: `./temp/` 目录用于临时文件合并
- **最大并发数**: 可在前端控制同时上传的分片数量
