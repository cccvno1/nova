package storage

import (
	"context"
	"io"
	"mime/multipart"
)

// Storage 文件存储接口
type Storage interface {
	// Upload 上传文件
	// ctx: 上下文
	// file: 文件
	// path: 存储路径（相对路径）
	// 返回: 访问 URL, 错误
	Upload(ctx context.Context, file multipart.File, filename string, path string) (string, error)

	// Delete 删除文件
	Delete(ctx context.Context, path string) error

	// Download 下载文件
	// 返回: 文件内容读取器, 错误
	Download(ctx context.Context, path string) (io.ReadCloser, error)

	// GetURL 获取文件访问 URL
	GetURL(path string) string

	// Exists 检查文件是否存在
	Exists(ctx context.Context, path string) (bool, error)

	// GetSize 获取文件大小
	GetSize(ctx context.Context, path string) (int64, error)
}

// UploadOptions 上传选项
type UploadOptions struct {
	ContentType string            // MIME 类型
	Metadata    map[string]string // 元数据
	ACL         string            // 访问控制（public-read, private 等）
}
