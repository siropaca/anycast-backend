package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// AudioRepository は音声データへのアクセスインターフェース
type AudioRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Audio, error)
	Create(ctx context.Context, audio *model.Audio) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindOrphaned(ctx context.Context) ([]model.Audio, error)
}

type audioRepository struct {
	db *gorm.DB
}

// NewAudioRepository は AudioRepository の実装を返す
func NewAudioRepository(db *gorm.DB) AudioRepository {
	return &audioRepository{db: db}
}

// FindByID は ID で音声を取得する
func (r *audioRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Audio, error) {
	var audio model.Audio

	if err := r.db.WithContext(ctx).First(&audio, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.ErrNotFound.WithMessage("音声が見つかりません")
		}

		logger.FromContext(ctx).Error("failed to fetch audio", "error", err, "id", id)
		return nil, apperror.ErrInternal.WithMessage("音声の取得に失敗しました").WithError(err)
	}

	return &audio, nil
}

// Create は音声を作成する
func (r *audioRepository) Create(ctx context.Context, audio *model.Audio) error {
	if err := r.db.WithContext(ctx).Create(audio).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create audio", "error", err)
		return apperror.ErrInternal.WithMessage("音声の作成に失敗しました").WithError(err)
	}

	return nil
}

// Delete は音声を削除する
func (r *audioRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Audio{}, "id = ?", id)
	if result.Error != nil {
		logger.FromContext(ctx).Error("failed to delete audio", "error", result.Error, "id", id)
		return apperror.ErrInternal.WithMessage("音声の削除に失敗しました").WithError(result.Error)
	}

	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("音声が見つかりません")
	}

	return nil
}

// FindOrphaned はどのテーブルからも参照されていない孤児レコードを取得する
//
// 対象: episodes.full_audio_id, bgms.audio_id, system_bgms.audio_id, audio_jobs.result_audio_id
// 条件: created_at から 1 時間以上経過したレコードのみ
func (r *audioRepository) FindOrphaned(ctx context.Context) ([]model.Audio, error) {
	var audios []model.Audio

	query := `
		SELECT a.* FROM audios a
		WHERE a.created_at < NOW() - INTERVAL '1 hour'
		AND NOT EXISTS (SELECT 1 FROM episodes e WHERE e.full_audio_id = a.id)
		AND NOT EXISTS (SELECT 1 FROM bgms b WHERE b.audio_id = a.id)
		AND NOT EXISTS (SELECT 1 FROM system_bgms sb WHERE sb.audio_id = a.id)
		AND NOT EXISTS (SELECT 1 FROM audio_jobs aj WHERE aj.result_audio_id = a.id)
		ORDER BY a.created_at DESC
	`

	if err := r.db.WithContext(ctx).Raw(query).Scan(&audios).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch orphaned audios", "error", err)
		return nil, apperror.ErrInternal.WithMessage("孤立した音声の取得に失敗しました").WithError(err)
	}

	return audios, nil
}
