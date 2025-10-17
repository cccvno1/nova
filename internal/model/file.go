package model

import "github.com/cccvno1/nova/pkg/database"

// File 文件模型
type File struct {
	database.Model
	OriginalName string `gorm:"not null;size:255" json:"original_name"`               // 原始文件名
	SavedName    string `gorm:"not null;size:255;index" json:"saved_name"`            // 保存的文件名（UUID）
	Path         string `gorm:"not null;size:500" json:"path"`                        // 文件存储路径
	URL          string `gorm:"size:500" json:"url"`                                  // 访问 URL
	Size         int64  `gorm:"not null" json:"size"`                                 // 文件大小（字节）
	MimeType     string `gorm:"not null;size:100" json:"mime_type"`                   // MIME 类型
	Extension    string `gorm:"size:20;index" json:"extension"`                       // 文件扩展名
	Hash         string `gorm:"size:64;index" json:"hash"`                            // 文件 Hash（MD5/SHA256）
	StorageType  string `gorm:"not null;size:20;default:'local'" json:"storage_type"` // 存储类型: local, oss, s3
	BucketName   string `gorm:"size:100" json:"bucket_name,omitempty"`                // 存储桶名称（OSS/S3）
	Category     string `gorm:"size:50;index" json:"category"`                        // 文件分类: avatar, document, image, video, other
	UploadedBy   uint   `gorm:"not null;index" json:"uploaded_by"`                    // 上传用户ID
	Status       int    `gorm:"default:1;not null;index" json:"status"`               // 状态: 1=正常, 2=已删除, 3=审核中

	// 可选：缩略图相关
	ThumbnailPath string `gorm:"size:500" json:"thumbnail_path,omitempty"` // 缩略图路径
	ThumbnailURL  string `gorm:"size:500" json:"thumbnail_url,omitempty"`  // 缩略图 URL

	// 可选：元数据
	Width  int `gorm:"default:0" json:"width,omitempty"`  // 图片宽度
	Height int `gorm:"default:0" json:"height,omitempty"` // 图片高度
}

func (File) TableName() string {
	return "files"
}

// FileCategory 文件分类常量
const (
	FileCategoryAvatar   = "avatar"
	FileCategoryDocument = "document"
	FileCategoryImage    = "image"
	FileCategoryVideo    = "video"
	FileCategoryAudio    = "audio"
	FileCategoryOther    = "other"
)

// FileStatus 文件状态常量
const (
	FileStatusNormal  = 1
	FileStatusDeleted = 2
	FileStatusPending = 3
)
