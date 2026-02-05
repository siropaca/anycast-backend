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

// ScriptJobRepository は台本ジョブデータへのアクセスインターフェース
type ScriptJobRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.ScriptJob, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, filter ScriptJobFilter) ([]model.ScriptJob, error)
	FindByEpisodeID(ctx context.Context, episodeID uuid.UUID) ([]model.ScriptJob, error)
	FindPendingByEpisodeID(ctx context.Context, episodeID uuid.UUID) (*model.ScriptJob, error)
	FindLatestCompletedByEpisodeID(ctx context.Context, episodeID uuid.UUID) (*model.ScriptJob, error)
	Create(ctx context.Context, job *model.ScriptJob) error
	Update(ctx context.Context, job *model.ScriptJob) error
	UpdateProgress(ctx context.Context, id uuid.UUID, progress int) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ScriptJobFilter は台本ジョブ検索のフィルタ条件
type ScriptJobFilter struct {
	Status *model.ScriptJobStatus
}

type scriptJobRepository struct {
	db *gorm.DB
}

// NewScriptJobRepository は ScriptJobRepository の実装を返す
func NewScriptJobRepository(db *gorm.DB) ScriptJobRepository {
	return &scriptJobRepository{db: db}
}

// FindByID は指定された ID の台本ジョブを取得する
func (r *scriptJobRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ScriptJob, error) {
	var job model.ScriptJob

	if err := r.db.WithContext(ctx).
		Preload("Episode").
		Preload("Episode.Channel").
		First(&job, "id = ?", id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("台本生成ジョブが見つかりません")
		}

		logger.FromContext(ctx).Error("failed to fetch script job", "error", err, "job_id", id)
		return nil, apperror.ErrInternal.WithMessage("台本生成ジョブの取得に失敗しました").WithError(err)
	}

	return &job, nil
}

// FindByUserID はユーザーの台本ジョブ一覧を取得する
func (r *scriptJobRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filter ScriptJobFilter) ([]model.ScriptJob, error) {
	var jobs []model.ScriptJob

	tx := r.db.WithContext(ctx).Where("user_id = ?", userID)

	// ステータスフィルタ
	if filter.Status != nil {
		tx = tx.Where("status = ?", *filter.Status)
	}

	if err := tx.
		Preload("Episode").
		Preload("Episode.Channel").
		Order("created_at DESC").
		Find(&jobs).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch script jobs", "error", err, "user_id", userID)
		return nil, apperror.ErrInternal.WithMessage("台本生成ジョブ一覧の取得に失敗しました").WithError(err)
	}

	return jobs, nil
}

// FindByEpisodeID はエピソードの台本ジョブ一覧を取得する
func (r *scriptJobRepository) FindByEpisodeID(ctx context.Context, episodeID uuid.UUID) ([]model.ScriptJob, error) {
	var jobs []model.ScriptJob

	if err := r.db.WithContext(ctx).
		Where("episode_id = ?", episodeID).
		Order("created_at DESC").
		Find(&jobs).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch script jobs by episode", "error", err, "episode_id", episodeID)
		return nil, apperror.ErrInternal.WithMessage("台本生成ジョブの取得に失敗しました").WithError(err)
	}

	return jobs, nil
}

// FindPendingByEpisodeID はエピソードの処理待ちジョブを取得する
// 見つからない場合は nil, nil を返す（エラーではない）
func (r *scriptJobRepository) FindPendingByEpisodeID(ctx context.Context, episodeID uuid.UUID) (*model.ScriptJob, error) {
	var job model.ScriptJob

	err := r.db.WithContext(ctx).
		Where("episode_id = ?", episodeID).
		Where("status IN ?", []model.ScriptJobStatus{model.ScriptJobStatusPending, model.ScriptJobStatusProcessing}).
		First(&job).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil //nolint:nilnil // not found is not an error
		}
		logger.FromContext(ctx).Error("failed to find pending script job", "error", err, "episode_id", episodeID)
		return nil, apperror.ErrInternal.WithMessage("処理待ちジョブの確認に失敗しました").WithError(err)
	}

	return &job, nil
}

// FindLatestCompletedByEpisodeID はエピソードの最新の完了済みジョブを取得する
// 見つからない場合は nil, nil を返す（エラーではない）
func (r *scriptJobRepository) FindLatestCompletedByEpisodeID(ctx context.Context, episodeID uuid.UUID) (*model.ScriptJob, error) {
	var job model.ScriptJob

	err := r.db.WithContext(ctx).
		Where("episode_id = ?", episodeID).
		Where("status = ?", model.ScriptJobStatusCompleted).
		Preload("Episode").
		Preload("Episode.Channel").
		Order("created_at DESC").
		First(&job).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil //nolint:nilnil // not found is not an error
		}
		logger.FromContext(ctx).Error("failed to find latest completed script job", "error", err, "episode_id", episodeID)
		return nil, apperror.ErrInternal.WithMessage("完了済みジョブの取得に失敗しました").WithError(err)
	}

	return &job, nil
}

// Create は台本ジョブを作成する
func (r *scriptJobRepository) Create(ctx context.Context, job *model.ScriptJob) error {
	if err := r.db.WithContext(ctx).Create(job).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create script job", "error", err)
		return apperror.ErrInternal.WithMessage("台本生成ジョブの作成に失敗しました").WithError(err)
	}

	return nil
}

// Update は台本ジョブを更新する
func (r *scriptJobRepository) Update(ctx context.Context, job *model.ScriptJob) error {
	if err := r.db.WithContext(ctx).Save(job).Error; err != nil {
		logger.FromContext(ctx).Error("failed to update script job", "error", err, "job_id", job.ID)
		return apperror.ErrInternal.WithMessage("台本生成ジョブの更新に失敗しました").WithError(err)
	}

	return nil
}

// UpdateProgress は台本ジョブの進捗のみを更新する
//
// ステータスなど他のフィールドは変更しない
func (r *scriptJobRepository) UpdateProgress(ctx context.Context, id uuid.UUID, progress int) error {
	if err := r.db.WithContext(ctx).Model(&model.ScriptJob{}).Where("id = ?", id).Update("progress", progress).Error; err != nil {
		logger.FromContext(ctx).Error("failed to update script job progress", "error", err, "job_id", id)
		return apperror.ErrInternal.WithMessage("進捗の更新に失敗しました").WithError(err)
	}

	return nil
}

// Delete は台本ジョブを削除する
func (r *scriptJobRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.ScriptJob{}, "id = ?", id)
	if result.Error != nil {
		logger.FromContext(ctx).Error("failed to delete script job", "error", result.Error, "job_id", id)
		return apperror.ErrInternal.WithMessage("台本生成ジョブの削除に失敗しました").WithError(result.Error)
	}

	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("台本生成ジョブが見つかりません")
	}

	return nil
}
