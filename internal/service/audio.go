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
	"github.com/siropaca/anycast-backend/internal/pkg/audio"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// allowedAudioMimeTypes は許可される音声の MIME タイプ
var allowedAudioMimeTypes = map[string]string{
	"audio/mpeg":  ".mp3",
	"audio/mp3":   ".mp3",
	"audio/wav":   ".wav",
	"audio/wave":  ".wav",
	"audio/ogg":   ".ogg",
	"audio/aac":   ".aac",
	"audio/mp4":   ".m4a",
	"audio/x-m4a": ".m4a",
}

// UploadAudioInput は音声アップロード用の入力データを表す
type UploadAudioInput struct {
	File        io.Reader
	Filename    string
	ContentType string
	FileSize    int
}

// AudioService は音声サービスのインターフェースを表す
type AudioService interface {
	UploadAudio(ctx context.Context, input UploadAudioInput) (*response.AudioUploadDataResponse, error)
}

type audioService struct {
	audioRepo     repository.AudioRepository
	storageClient storage.Client
}

// NewAudioService は audioService を生成して AudioService として返す
func NewAudioService(audioRepo repository.AudioRepository, storageClient storage.Client) AudioService {
	return &audioService{
		audioRepo:     audioRepo,
		storageClient: storageClient,
	}
}

// UploadAudio は音声ファイルをアップロードする
func (s *audioService) UploadAudio(ctx context.Context, input UploadAudioInput) (*response.AudioUploadDataResponse, error) {
	log := logger.FromContext(ctx)

	// MIME タイプのバリデーション
	ext, ok := allowedAudioMimeTypes[input.ContentType]
	if !ok {
		return nil, apperror.ErrValidation.WithMessage("無効な音声形式です。使用可能な形式: mp3, wav, ogg, aac, m4a")
	}

	// ファイルデータの読み込み
	data, err := io.ReadAll(input.File)
	if err != nil {
		log.Error("音声データの読み込みに失敗しました", "error", err)
		return nil, apperror.ErrInternal.WithMessage("音声データの読み込みに失敗しました").WithError(err)
	}

	// 音声 ID の生成
	audioID := uuid.New()

	// 再生時間を取得
	durationMs := audio.GetDurationMs(data)

	// GCS へアップロード
	path := storage.GenerateAudioPathWithExt(audioID.String(), ext)
	if _, err := s.storageClient.Upload(ctx, data, path, input.ContentType); err != nil {
		return nil, err
	}

	// DB に保存
	audioModel := &model.Audio{
		ID:         audioID,
		MimeType:   input.ContentType,
		Path:       path,
		Filename:   input.Filename,
		FileSize:   input.FileSize,
		DurationMs: durationMs,
	}

	if err := s.audioRepo.Create(ctx, audioModel); err != nil {
		// DB 保存に失敗した場合は GCS のファイルを削除
		if deleteErr := s.storageClient.Delete(ctx, path); deleteErr != nil {
			log.Warn("アップロード済み音声のクリーンアップに失敗しました", "error", deleteErr, "path", path)
		}
		return nil, err
	}

	// 署名付き URL を生成
	signedURL, err := s.storageClient.GenerateSignedURL(ctx, path, 1*time.Hour)
	if err != nil {
		return nil, err
	}

	return &response.AudioUploadDataResponse{
		Data: response.AudioUploadResponse{
			ID:         audioModel.ID,
			MimeType:   audioModel.MimeType,
			URL:        signedURL,
			Filename:   filepath.Base(audioModel.Filename),
			FileSize:   audioModel.FileSize,
			DurationMs: audioModel.DurationMs,
		},
	}, nil
}
