package service

import (
	"context"
	"io"
	"path/filepath"
	"time"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// allowedImageMimeTypes は許可される画像の MIME タイプ
var allowedImageMimeTypes = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// UploadImageInput は画像アップロード用の入力データを表す
type UploadImageInput struct {
	File        io.Reader
	Filename    string
	ContentType string
	FileSize    int
}

// ImageService は画像サービスのインターフェースを表す
type ImageService interface {
	UploadImage(ctx context.Context, input UploadImageInput) (*response.ImageUploadDataResponse, error)
}

type imageService struct {
	imageRepo     repository.ImageRepository
	storageClient storage.Client
}

// NewImageService は imageService を生成して ImageService として返す
func NewImageService(imageRepo repository.ImageRepository, storageClient storage.Client) ImageService {
	return &imageService{
		imageRepo:     imageRepo,
		storageClient: storageClient,
	}
}

// UploadImage は画像ファイルをアップロードする
func (s *imageService) UploadImage(ctx context.Context, input UploadImageInput) (*response.ImageUploadDataResponse, error) {
	log := logger.FromContext(ctx)

	// MIME タイプのバリデーション
	ext, ok := allowedImageMimeTypes[input.ContentType]
	if !ok {
		return nil, apperror.ErrValidation.WithMessage("無効な画像形式です。使用可能な形式: png, jpeg, gif, webp")
	}

	// ファイルデータの読み込み
	data, err := io.ReadAll(input.File)
	if err != nil {
		log.Error("failed to read image data", "error", err)
		return nil, apperror.ErrInternal.WithMessage("画像データの読み込みに失敗しました").WithError(err)
	}

	// 画像 ID の生成
	imageID := uuid.New()

	// GCS へアップロード
	path := storage.GenerateImagePath(imageID.String(), ext)
	if _, err := s.storageClient.Upload(ctx, data, path, input.ContentType); err != nil {
		return nil, err
	}

	// DB に保存
	image := &model.Image{
		ID:       imageID,
		MimeType: input.ContentType,
		Path:     path,
		Filename: input.Filename,
		FileSize: input.FileSize,
	}

	if err := s.imageRepo.Create(ctx, image); err != nil {
		// DB 保存に失敗した場合は GCS のファイルを削除
		if deleteErr := s.storageClient.Delete(ctx, path); deleteErr != nil {
			log.Warn("failed to cleanup uploaded image", "error", deleteErr, "path", path)
		}
		return nil, err
	}

	// 署名付き URL を生成
	signedURL, err := s.storageClient.GenerateSignedURL(ctx, path, 1*time.Hour)
	if err != nil {
		return nil, err
	}

	return &response.ImageUploadDataResponse{
		Data: response.ImageUploadResponse{
			ID:       image.ID,
			MimeType: image.MimeType,
			URL:      signedURL,
			Filename: filepath.Base(image.Filename),
			FileSize: image.FileSize,
		},
	}, nil
}
