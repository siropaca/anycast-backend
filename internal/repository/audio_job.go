package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// AudioJobRepository は音声ジョブデータへのアクセスインターフェース
type AudioJobRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.AudioJob, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, filter AudioJobFilter) ([]model.AudioJob, error)
	FindByEpisodeID(ctx context.Context, episodeID uuid.UUID) ([]model.AudioJob, error)
	FindPendingByEpisodeID(ctx context.Context, episodeID uuid.UUID) (*model.AudioJob, error)
	Create(ctx context.Context, job *model.AudioJob) error
	Update(ctx context.Context, job *model.AudioJob) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// AudioJobFilter は音声ジョブ検索のフィルタ条件
type AudioJobFilter struct {
	Status *model.AudioJobStatus
}

type audioJobRepository struct {
	db *gorm.DB
}

// NewAudioJobRepository は AudioJobRepository の実装を返す
func NewAudioJobRepository(db *gorm.DB) AudioJobRepository {
	return &audioJobRepository{db: db}
}

// FindByID は指定された ID の音声ジョブを取得する
func (r *audioJobRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.AudioJob, error) {
	var job model.AudioJob

	if err := r.db.WithContext(ctx).
		Preload("Episode").
		Preload("Episode.Channel").
		Preload("ResultAudio").
		First(&job, "id = ?", id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("音声生成ジョブが見つかりません")
		}

		logger.FromContext(ctx).Error("failed to fetch audio job", "error", err, "job_id", id)
		return nil, apperror.ErrInternal.WithMessage("音声生成ジョブの取得に失敗しました").WithError(err)
	}

	return &job, nil
}

// FindByUserID はユーザーの音声ジョブ一覧を取得する
func (r *audioJobRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filter AudioJobFilter) ([]model.AudioJob, error) {
	var jobs []model.AudioJob

	tx := r.db.WithContext(ctx).Where("user_id = ?", userID)

	// ステータスフィルタ
	if filter.Status != nil {
		tx = tx.Where("status = ?", *filter.Status)
	}

	if err := tx.
		Preload("Episode").
		Preload("Episode.Channel").
		Preload("ResultAudio").
		Order("created_at DESC").
		Find(&jobs).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch audio jobs", "error", err, "user_id", userID)
		return nil, apperror.ErrInternal.WithMessage("音声生成ジョブ一覧の取得に失敗しました").WithError(err)
	}

	return jobs, nil
}

// FindByEpisodeID はエピソードの音声ジョブ一覧を取得する
func (r *audioJobRepository) FindByEpisodeID(ctx context.Context, episodeID uuid.UUID) ([]model.AudioJob, error) {
	var jobs []model.AudioJob

	if err := r.db.WithContext(ctx).
		Preload("ResultAudio").
		Where("episode_id = ?", episodeID).
		Order("created_at DESC").
		Find(&jobs).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch audio jobs by episode", "error", err, "episode_id", episodeID)
		return nil, apperror.ErrInternal.WithMessage("音声生成ジョブの取得に失敗しました").WithError(err)
	}

	return jobs, nil
}

// FindPendingByEpisodeID はエピソードの処理待ちジョブを取得する
// 見つからない場合は nil, nil を返す（エラーではない）
func (r *audioJobRepository) FindPendingByEpisodeID(ctx context.Context, episodeID uuid.UUID) (*model.AudioJob, error) {
	var job model.AudioJob

	err := r.db.WithContext(ctx).
		Where("episode_id = ?", episodeID).
		Where("status IN ?", []model.AudioJobStatus{model.AudioJobStatusPending, model.AudioJobStatusProcessing}).
		First(&job).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil //nolint:nilnil // not found is not an error
		}
		logger.FromContext(ctx).Error("failed to find pending job", "error", err, "episode_id", episodeID)
		return nil, apperror.ErrInternal.WithMessage("処理待ちジョブの確認に失敗しました").WithError(err)
	}

	return &job, nil
}

// Create は音声ジョブを作成する
func (r *audioJobRepository) Create(ctx context.Context, job *model.AudioJob) error {
	if err := r.db.WithContext(ctx).Create(job).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create audio job", "error", err)
		return apperror.ErrInternal.WithMessage("音声生成ジョブの作成に失敗しました").WithError(err)
	}

	return nil
}

// Update は音声ジョブを更新する
func (r *audioJobRepository) Update(ctx context.Context, job *model.AudioJob) error {
	if err := r.db.WithContext(ctx).Save(job).Error; err != nil {
		logger.FromContext(ctx).Error("failed to update audio job", "error", err, "job_id", job.ID)
		return apperror.ErrInternal.WithMessage("音声生成ジョブの更新に失敗しました").WithError(err)
	}

	return nil
}

// Delete は音声ジョブを削除する
func (r *audioJobRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.AudioJob{}, "id = ?", id)
	if result.Error != nil {
		logger.FromContext(ctx).Error("failed to delete audio job", "error", result.Error, "job_id", id)
		return apperror.ErrInternal.WithMessage("音声生成ジョブの削除に失敗しました").WithError(result.Error)
	}

	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("音声生成ジョブが見つかりません")
	}

	return nil
}
