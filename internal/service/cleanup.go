package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// CleanupResult は孤児メディアクリーンアップの結果を表す
type CleanupResult struct {
	OrphanedAudios    []model.Audio
	OrphanedImages    []model.Image
	DeletedAudioCount int
	DeletedImageCount int
	FailedAudioCount  int
	FailedImageCount  int
}

// CleanupService はクリーンアップサービスのインターフェースを表す
type CleanupService interface {
	CleanupOrphanedMedia(ctx context.Context, dryRun bool) (*CleanupResult, error)
}

type cleanupService struct {
	audioRepo     repository.AudioRepository
	imageRepo     repository.ImageRepository
	storageClient storage.Client
}

// NewCleanupService は cleanupService を生成して CleanupService として返す
func NewCleanupService(
	audioRepo repository.AudioRepository,
	imageRepo repository.ImageRepository,
	storageClient storage.Client,
) CleanupService {
	return &cleanupService{
		audioRepo:     audioRepo,
		imageRepo:     imageRepo,
		storageClient: storageClient,
	}
}

// CleanupOrphanedMedia は孤児メディアファイルを検出し削除する
func (s *cleanupService) CleanupOrphanedMedia(ctx context.Context, dryRun bool) (*CleanupResult, error) {
	log := logger.FromContext(ctx)
	result := &CleanupResult{}

	// 孤児 Audio を取得
	orphanedAudios, err := s.audioRepo.FindOrphaned(ctx)
	if err != nil {
		return nil, err
	}
	result.OrphanedAudios = orphanedAudios

	// 孤児 Image を取得
	orphanedImages, err := s.imageRepo.FindOrphaned(ctx)
	if err != nil {
		return nil, err
	}
	result.OrphanedImages = orphanedImages

	log.Info("orphaned media detected",
		"orphaned_audios", len(orphanedAudios),
		"orphaned_images", len(orphanedImages),
		"dry_run", dryRun,
	)

	// dry-run の場合は削除せずに終了
	if dryRun {
		return result, nil
	}

	// Audio を削除
	for _, audio := range orphanedAudios {
		// GCS から削除
		if s.storageClient != nil {
			if err := s.storageClient.Delete(ctx, audio.Path); err != nil {
				log.Warn("failed to delete audio from GCS", "audio_id", audio.ID, "path", audio.Path, "error", err)
				result.FailedAudioCount++
				continue
			}
		}

		// DB から削除
		if err := s.audioRepo.Delete(ctx, audio.ID); err != nil {
			log.Warn("failed to delete audio from DB", "audio_id", audio.ID, "error", err)
			result.FailedAudioCount++
			continue
		}

		result.DeletedAudioCount++
		log.Debug("orphaned audio deleted", "audio_id", audio.ID, "path", audio.Path)
	}

	// Image を削除
	for _, image := range orphanedImages {
		// GCS から削除（外部 URL の場合はスキップ）
		if s.storageClient != nil && !storage.IsExternalURL(image.Path) {
			if err := s.storageClient.Delete(ctx, image.Path); err != nil {
				log.Warn("failed to delete image from GCS", "image_id", image.ID, "path", image.Path, "error", err)
				result.FailedImageCount++
				continue
			}
		}

		// DB から削除
		if err := s.imageRepo.Delete(ctx, image.ID); err != nil {
			log.Warn("failed to delete image from DB", "image_id", image.ID, "error", err)
			result.FailedImageCount++
			continue
		}

		result.DeletedImageCount++
		log.Debug("orphaned image deleted", "image_id", image.ID, "path", image.Path)
	}

	log.Info("orphaned media cleanup completed",
		"deleted_audios", result.DeletedAudioCount,
		"deleted_images", result.DeletedImageCount,
		"failed_audios", result.FailedAudioCount,
		"failed_images", result.FailedImageCount,
	)

	return result, nil
}
