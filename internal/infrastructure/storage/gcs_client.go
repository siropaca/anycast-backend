package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

// ストレージクライアントのインターフェース
type Client interface {
	Upload(ctx context.Context, data []byte, path, contentType string) (string, error)
	GenerateSignedURL(ctx context.Context, path string, expiration time.Duration) (string, error)
	Delete(ctx context.Context, path string) error
}

// 音声ファイルの GCS パスを生成する
func GenerateAudioPath(audioID string) string {
	return fmt.Sprintf("audios/%s.mp3", audioID)
}

// 画像ファイルの GCS パスを生成する
// ext は拡張子（例: ".png", ".jpg"）
func GenerateImagePath(imageID, ext string) string {
	return fmt.Sprintf("images/%s%s", imageID, ext)
}

type gcsClient struct {
	client     *storage.Client
	bucketName string
}

// GCS クライアントを作成する
func NewGCSClient(ctx context.Context, bucketName, credentialsJSON string) (Client, error) {
	var opts []option.ClientOption
	if credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(credentialsJSON)))
	}

	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	return &gcsClient{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// ファイルをアップロードする
func (c *gcsClient) Upload(ctx context.Context, data []byte, path, contentType string) (string, error) {
	log := logger.FromContext(ctx)
	log.Debug("uploading file to GCS", "path", path, "size", len(data))

	bucket := c.client.Bucket(c.bucketName)
	obj := bucket.Object(path)

	writer := obj.NewWriter(ctx)
	writer.ContentType = contentType

	if _, err := writer.Write(data); err != nil {
		log.Error("failed to write to GCS", "error", err)
		return "", apperror.ErrMediaUploadFailed.WithMessage("Failed to upload file").WithError(err)
	}

	if err := writer.Close(); err != nil {
		log.Error("failed to close GCS writer", "error", err)
		return "", apperror.ErrMediaUploadFailed.WithMessage("Failed to upload file").WithError(err)
	}

	log.Debug("file uploaded successfully", "path", path)

	return path, nil
}

// 署名付き URL を生成する
func (c *gcsClient) GenerateSignedURL(ctx context.Context, path string, expiration time.Duration) (string, error) {
	log := logger.FromContext(ctx)
	log.Debug("generating signed URL", "path", path, "expiration", expiration)

	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(expiration),
	}

	url, err := c.client.Bucket(c.bucketName).SignedURL(path, opts)
	if err != nil {
		log.Error("failed to generate signed URL", "error", err)
		return "", apperror.ErrInternal.WithMessage("Failed to generate signed URL").WithError(err)
	}

	log.Debug("signed URL generated successfully", "url", url)
	return url, nil
}

// ファイルを削除する
func (c *gcsClient) Delete(ctx context.Context, path string) error {
	log := logger.FromContext(ctx)
	log.Debug("deleting file from GCS", "path", path)

	bucket := c.client.Bucket(c.bucketName)
	obj := bucket.Object(path)

	if err := obj.Delete(ctx); err != nil {
		// ファイルが存在しない場合はエラーにしない
		if err == storage.ErrObjectNotExist {
			log.Debug("file does not exist, skipping delete", "path", path)
			return nil
		}
		log.Error("failed to delete from GCS", "error", err)
		return apperror.ErrInternal.WithMessage("Failed to delete file").WithError(err)
	}

	log.Debug("file deleted successfully", "path", path)
	return nil
}

// ファイルをダウンロードする
func (c *gcsClient) Download(ctx context.Context, path string) ([]byte, error) {
	log := logger.FromContext(ctx)
	log.Debug("downloading file from GCS", "path", path)

	bucket := c.client.Bucket(c.bucketName)
	obj := bucket.Object(path)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, apperror.ErrNotFound.WithMessage("File not found")
		}
		log.Error("failed to create GCS reader", "error", err)
		return nil, apperror.ErrInternal.WithMessage("Failed to download file").WithError(err)
	}

	defer func() {
		if err := reader.Close(); err != nil {
			log.Warn("failed to close GCS reader", "error", err)
		}
	}()

	data, err := io.ReadAll(reader)
	if err != nil {
		log.Error("failed to read from GCS", "error", err)
		return nil, apperror.ErrInternal.WithMessage("Failed to download file").WithError(err)
	}

	log.Debug("file downloaded successfully", "path", path, "size", len(data))
	return data, nil
}

// クライアントを閉じる
func (c *gcsClient) Close() error {
	return c.client.Close()
}
