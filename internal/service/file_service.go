package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cccvno1/nova/internal/model"
	"github.com/cccvno1/nova/internal/repository"
	"github.com/cccvno1/nova/pkg/config"
	"github.com/cccvno1/nova/pkg/database"
	"github.com/cccvno1/nova/pkg/errors"
	"github.com/cccvno1/nova/pkg/storage"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
)

// FileService 文件服务接口
type FileService interface {
	Upload(ctx context.Context, fileHeader *multipart.FileHeader, category string, userID uint) (*FileResponse, error)
	Download(ctx context.Context, id uint, userID uint) (io.ReadCloser, *model.File, error)
	Delete(ctx context.Context, id uint, userID uint) error
	GetByID(ctx context.Context, id uint) (*FileResponse, error)
	List(ctx context.Context, userID uint, category string, pagination *database.Pagination) ([]FileResponse, error)
	Search(ctx context.Context, keyword string, pagination *database.Pagination) ([]FileResponse, error)
	GetUserStorageInfo(ctx context.Context, userID uint) (*StorageInfo, error)
}

type fileService struct {
	fileRepo repository.FileRepository
	storage  storage.Storage
	config   *config.UploadConfig
}

// NewFileService 创建文件服务
func NewFileService(fileRepo repository.FileRepository, storage storage.Storage, cfg *config.UploadConfig) FileService {
	return &fileService{
		fileRepo: fileRepo,
		storage:  storage,
		config:   cfg,
	}
}

// FileResponse 文件响应
type FileResponse struct {
	ID           uint      `json:"id"`
	OriginalName string    `json:"original_name"`
	SavedName    string    `json:"saved_name"`
	URL          string    `json:"url"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
	Size         int64     `json:"size"`
	MimeType     string    `json:"mime_type"`
	Extension    string    `json:"extension"`
	Category     string    `json:"category"`
	UploadedBy   uint      `json:"uploaded_by"`
	Width        int       `json:"width,omitempty"`
	Height       int       `json:"height,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// StorageInfo 存储信息
type StorageInfo struct {
	FileCount int64 `json:"file_count"` // 文件数量
	TotalSize int64 `json:"total_size"` // 总大小（字节）
	UsedMB    int64 `json:"used_mb"`    // 已使用空间（MB）
}

// Upload 上传文件
func (s *fileService) Upload(ctx context.Context, fileHeader *multipart.FileHeader, category string, userID uint) (*FileResponse, error) {
	// 1. 验证文件
	if err := s.validateFile(fileHeader); err != nil {
		return nil, err
	}

	// 2. 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, errors.Wrap(errors.ErrInvalidParams, fmt.Errorf("failed to open file: %w", err))
	}
	defer file.Close()

	// 3. 计算文件 Hash（用于秒传）
	hash, err := s.calculateHash(file)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternalServer, err)
	}

	// 4. 检查是否已存在相同文件（秒传功能）
	existingFile, err := s.fileRepo.FindByHash(ctx, hash)
	if err == nil && existingFile != nil {
		// 文件已存在，创建新的元数据记录（引用相同的物理文件）
		newFile := &model.File{
			OriginalName:  fileHeader.Filename,
			SavedName:     existingFile.SavedName,
			Path:          existingFile.Path,
			URL:           existingFile.URL,
			Size:          existingFile.Size,
			MimeType:      existingFile.MimeType,
			Extension:     existingFile.Extension,
			Hash:          hash,
			StorageType:   existingFile.StorageType,
			Category:      category,
			UploadedBy:    userID,
			Status:        model.FileStatusNormal,
			ThumbnailPath: existingFile.ThumbnailPath,
			ThumbnailURL:  existingFile.ThumbnailURL,
			Width:         existingFile.Width,
			Height:        existingFile.Height,
		}

		if err := s.fileRepo.Create(ctx, newFile); err != nil {
			return nil, errors.Wrap(errors.ErrDatabase, err)
		}

		return s.toResponse(newFile), nil
	}

	// 5. 生成保存文件名
	ext := filepath.Ext(fileHeader.Filename)
	savedName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// 6. 构建存储路径（按日期分目录）
	now := time.Now()
	relativePath := filepath.Join(
		category,
		now.Format("2006"),
		now.Format("01"),
		now.Format("02"),
		savedName,
	)

	// 7. 重置文件指针
	if _, err := file.Seek(0, 0); err != nil {
		return nil, errors.Wrap(errors.ErrInternalServer, fmt.Errorf("failed to seek file: %w", err))
	}

	// 8. 上传到存储
	url, err := s.storage.Upload(ctx, file, savedName, relativePath)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternalServer, fmt.Errorf("failed to upload file: %w", err))
	}

	// 9. 创建文件记录
	fileModel := &model.File{
		OriginalName: fileHeader.Filename,
		SavedName:    savedName,
		Path:         relativePath,
		URL:          url,
		Size:         fileHeader.Size,
		MimeType:     fileHeader.Header.Get("Content-Type"),
		Extension:    ext,
		Hash:         hash,
		StorageType:  s.config.StorageType,
		Category:     category,
		UploadedBy:   userID,
		Status:       model.FileStatusNormal,
	}

	// 10. 如果是图片，处理缩略图和获取尺寸
	if s.isImage(fileModel.MimeType) {
		if err := s.processImage(ctx, file, fileModel, relativePath); err != nil {
			// 图片处理失败不影响上传，只记录错误
			fmt.Printf("failed to process image: %v\n", err)
		}
	}

	// 11. 保存到数据库
	if err := s.fileRepo.Create(ctx, fileModel); err != nil {
		// 数据库保存失败，删除已上传的文件
		_ = s.storage.Delete(ctx, relativePath)
		return nil, errors.Wrap(errors.ErrDatabase, err)
	}

	return s.toResponse(fileModel), nil
}

// Download 下载文件
func (s *fileService) Download(ctx context.Context, id uint, userID uint) (io.ReadCloser, *model.File, error) {
	// 查询文件信息
	file, err := s.fileRepo.FindByID(ctx, id)
	if err != nil {
		return nil, nil, errors.New(errors.ErrRecordNotFound, "file not found")
	}

	// 验证权限（只能下载自己的文件，或者有管理员权限）
	if file.UploadedBy != userID {
		// TODO: 这里可以添加管理员权限检查
		return nil, nil, errors.New(errors.ErrForbidden, "no permission to download this file")
	}

	// 从存储下载
	reader, err := s.storage.Download(ctx, file.Path)
	if err != nil {
		return nil, nil, errors.Wrap(errors.ErrInternalServer, err)
	}

	return reader, file, nil
}

// Delete 删除文件
func (s *fileService) Delete(ctx context.Context, id uint, userID uint) error {
	// 查询文件信息
	file, err := s.fileRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New(errors.ErrRecordNotFound, "file not found")
	}

	// 验证权限
	if file.UploadedBy != userID {
		// TODO: 管理员权限检查
		return errors.New(errors.ErrForbidden, "no permission to delete this file")
	}

	// 软删除数据库记录
	if err := s.fileRepo.Delete(ctx, id); err != nil {
		return errors.Wrap(errors.ErrDatabase, err)
	}

	// 注意：这里不立即删除物理文件，因为可能有其他记录引用（秒传）
	// 可以通过定时任务清理没有引用的文件

	return nil
}

// GetByID 根据 ID 获取文件信息
func (s *fileService) GetByID(ctx context.Context, id uint) (*FileResponse, error) {
	file, err := s.fileRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New(errors.ErrRecordNotFound, "file not found")
	}

	return s.toResponse(file), nil
}

// List 获取文件列表
func (s *fileService) List(ctx context.Context, userID uint, category string, pagination *database.Pagination) ([]FileResponse, error) {
	var files []model.File
	var err error

	if category != "" {
		files, err = s.fileRepo.ListByUserAndCategory(ctx, userID, category, pagination)
	} else {
		files, err = s.fileRepo.ListByUser(ctx, userID, pagination)
	}

	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, err)
	}

	result := make([]FileResponse, len(files))
	for i, file := range files {
		result[i] = *s.toResponse(&file)
	}

	return result, nil
}

// Search 搜索文件
func (s *fileService) Search(ctx context.Context, keyword string, pagination *database.Pagination) ([]FileResponse, error) {
	files, err := s.fileRepo.Search(ctx, keyword, pagination)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, err)
	}

	result := make([]FileResponse, len(files))
	for i, file := range files {
		result[i] = *s.toResponse(&file)
	}

	return result, nil
}

// GetUserStorageInfo 获取用户存储信息
func (s *fileService) GetUserStorageInfo(ctx context.Context, userID uint) (*StorageInfo, error) {
	count, err := s.fileRepo.CountByUser(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, err)
	}

	totalSize, err := s.fileRepo.GetUserStorageUsage(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, err)
	}

	return &StorageInfo{
		FileCount: count,
		TotalSize: totalSize,
		UsedMB:    totalSize / 1024 / 1024,
	}, nil
}

// validateFile 验证文件
func (s *fileService) validateFile(fileHeader *multipart.FileHeader) error {
	// 1. 检查文件大小
	maxSize := s.config.MaxSize * 1024 * 1024 // 转换为字节
	if fileHeader.Size > maxSize {
		return errors.New(errors.ErrInvalidParams, fmt.Sprintf("file size exceeds limit: %dMB", s.config.MaxSize))
	}

	// 2. 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if len(s.config.AllowedExts) > 0 {
		allowed := false
		for _, allowedExt := range s.config.AllowedExts {
			if ext == strings.ToLower(allowedExt) {
				allowed = true
				break
			}
		}
		if !allowed {
			return errors.New(errors.ErrInvalidParams, fmt.Sprintf("file extension not allowed: %s", ext))
		}
	}

	// 3. 检查 MIME 类型
	mimeType := fileHeader.Header.Get("Content-Type")
	if len(s.config.AllowedTypes) > 0 {
		allowed := false
		for _, allowedType := range s.config.AllowedTypes {
			if mimeType == allowedType {
				allowed = true
				break
			}
		}
		if !allowed {
			return errors.New(errors.ErrInvalidParams, fmt.Sprintf("file type not allowed: %s", mimeType))
		}
	}

	return nil
}

// calculateHash 计算文件 Hash（SHA256）
func (s *fileService) calculateHash(file multipart.File) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// isImage 判断是否为图片
func (s *fileService) isImage(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}

// processImage 处理图片（获取尺寸、生成缩略图）
func (s *fileService) processImage(ctx context.Context, file multipart.File, fileModel *model.File, originalPath string) error {
	// 重置文件指针
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	// 解码图片
	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// 获取图片尺寸
	bounds := img.Bounds()
	fileModel.Width = bounds.Dx()
	fileModel.Height = bounds.Dy()

	// 生成缩略图
	if s.config.EnableThumbnail {
		thumbnail := resize.Thumbnail(
			uint(s.config.ThumbnailWidth),
			uint(s.config.ThumbnailHeight),
			img,
			resize.Lanczos3,
		)

		// 保存缩略图
		thumbnailPath := s.getThumbnailPath(originalPath)
		if err := s.saveThumbnail(ctx, thumbnail, thumbnailPath, format); err != nil {
			return fmt.Errorf("failed to save thumbnail: %w", err)
		}

		fileModel.ThumbnailPath = thumbnailPath
		fileModel.ThumbnailURL = s.storage.GetURL(thumbnailPath)
	}

	return nil
}

// getThumbnailPath 获取缩略图路径
func (s *fileService) getThumbnailPath(originalPath string) string {
	dir := filepath.Dir(originalPath)
	ext := filepath.Ext(originalPath)
	name := strings.TrimSuffix(filepath.Base(originalPath), ext)
	return filepath.Join(dir, fmt.Sprintf("%s_thumb%s", name, ext))
}

// saveThumbnail 保存缩略图
func (s *fileService) saveThumbnail(ctx context.Context, img image.Image, path string, format string) error {
	// 创建临时文件
	tmpFile := fmt.Sprintf("/tmp/thumb_%s_%d", uuid.New().String(), time.Now().Unix())
	f, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile)
	defer f.Close()

	// 编码图片
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(f, img, &jpeg.Options{Quality: s.config.ThumbnailQuality})
	case "png":
		err = png.Encode(f, img)
	default:
		err = jpeg.Encode(f, img, &jpeg.Options{Quality: s.config.ThumbnailQuality})
	}

	if err != nil {
		return err
	}

	// 重置文件指针
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}

	// 上传缩略图
	_, err = s.storage.Upload(ctx, f, filepath.Base(path), path)
	return err
}

// toResponse 转换为响应对象
func (s *fileService) toResponse(file *model.File) *FileResponse {
	return &FileResponse{
		ID:           file.ID,
		OriginalName: file.OriginalName,
		SavedName:    file.SavedName,
		URL:          file.URL,
		ThumbnailURL: file.ThumbnailURL,
		Size:         file.Size,
		MimeType:     file.MimeType,
		Extension:    file.Extension,
		Category:     file.Category,
		UploadedBy:   file.UploadedBy,
		Width:        file.Width,
		Height:       file.Height,
		CreatedAt:    file.CreatedAt,
	}
}
