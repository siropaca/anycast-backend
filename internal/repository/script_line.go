package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// 台本行データへのアクセスインターフェース
type ScriptLineRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.ScriptLine, error)
	FindByEpisodeID(ctx context.Context, episodeID uuid.UUID) ([]model.ScriptLine, error)
	FindByEpisodeIDWithVoice(ctx context.Context, episodeID uuid.UUID) ([]model.ScriptLine, error)
	Create(ctx context.Context, scriptLine *model.ScriptLine) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByEpisodeID(ctx context.Context, episodeID uuid.UUID) error
	CreateBatch(ctx context.Context, scriptLines []model.ScriptLine) ([]model.ScriptLine, error)
	Update(ctx context.Context, scriptLine *model.ScriptLine) error
	IncrementLineOrderFrom(ctx context.Context, episodeID uuid.UUID, fromLineOrder int) error
}

type scriptLineRepository struct {
	db *gorm.DB
}

// ScriptLineRepository の実装を返す
func NewScriptLineRepository(db *gorm.DB) ScriptLineRepository {
	return &scriptLineRepository{db: db}
}

// ID で台本行を取得する
func (r *scriptLineRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ScriptLine, error) {
	var scriptLine model.ScriptLine

	if err := r.db.WithContext(ctx).
		Preload("Speaker").
		Preload("Speaker.Voice").
		First(&scriptLine, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.ErrNotFound.WithMessage("Script line not found")
		}
		logger.FromContext(ctx).Error("failed to fetch script line", "error", err, "id", id)
		return nil, apperror.ErrInternal.WithMessage("Failed to fetch script line").WithError(err)
	}

	return &scriptLine, nil
}

// 指定されたエピソードの台本行一覧を取得する
func (r *scriptLineRepository) FindByEpisodeID(ctx context.Context, episodeID uuid.UUID) ([]model.ScriptLine, error) {
	var scriptLines []model.ScriptLine

	if err := r.db.WithContext(ctx).
		Preload("Speaker").
		Preload("Speaker.Voice").
		Where("episode_id = ?", episodeID).
		Order("line_order ASC").
		Find(&scriptLines).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch script lines", "error", err, "episode_id", episodeID)
		return nil, apperror.ErrInternal.WithMessage("Failed to fetch script lines").WithError(err)
	}

	return scriptLines, nil
}

// 指定されたエピソードの台本行一覧を取得する（Voice 情報を含む）
func (r *scriptLineRepository) FindByEpisodeIDWithVoice(ctx context.Context, episodeID uuid.UUID) ([]model.ScriptLine, error) {
	var scriptLines []model.ScriptLine

	if err := r.db.WithContext(ctx).
		Preload("Speaker").
		Preload("Speaker.Voice").
		Where("episode_id = ?", episodeID).
		Order("line_order ASC").
		Find(&scriptLines).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch script lines with voice", "error", err, "episode_id", episodeID)
		return nil, apperror.ErrInternal.WithMessage("Failed to fetch script lines").WithError(err)
	}

	return scriptLines, nil
}

// 指定された台本行を削除する
func (r *scriptLineRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.ScriptLine{}, "id = ?", id)
	if result.Error != nil {
		logger.FromContext(ctx).Error("failed to delete script line", "error", result.Error, "id", id)
		return apperror.ErrInternal.WithMessage("Failed to delete script line").WithError(result.Error)
	}

	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("Script line not found")
	}

	return nil
}

// 指定されたエピソードの台本行を全て削除する
func (r *scriptLineRepository) DeleteByEpisodeID(ctx context.Context, episodeID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Where("episode_id = ?", episodeID).
		Delete(&model.ScriptLine{}).Error; err != nil {
		logger.FromContext(ctx).Error("failed to delete script lines", "error", err, "episode_id", episodeID)
		return apperror.ErrInternal.WithMessage("Failed to delete script lines").WithError(err)
	}

	return nil
}

// 複数の台本行を一括作成する
func (r *scriptLineRepository) CreateBatch(ctx context.Context, scriptLines []model.ScriptLine) ([]model.ScriptLine, error) {
	if len(scriptLines) == 0 {
		return []model.ScriptLine{}, nil
	}

	if err := r.db.WithContext(ctx).Create(&scriptLines).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create script lines", "error", err)
		return nil, apperror.ErrInternal.WithMessage("Failed to create script lines").WithError(err)
	}

	// 作成した行を再取得して Speaker 情報を含める
	var created []model.ScriptLine
	if err := r.db.WithContext(ctx).
		Preload("Speaker").
		Preload("Speaker.Voice").
		Where("episode_id = ?", scriptLines[0].EpisodeID).
		Order("line_order ASC").
		Find(&created).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch created script lines", "error", err)
		return nil, apperror.ErrInternal.WithMessage("Failed to fetch created script lines").WithError(err)
	}

	return created, nil
}

// 台本行を更新する
func (r *scriptLineRepository) Update(ctx context.Context, scriptLine *model.ScriptLine) error {
	if err := r.db.WithContext(ctx).Save(scriptLine).Error; err != nil {
		logger.FromContext(ctx).Error("failed to update script line", "error", err, "id", scriptLine.ID)
		return apperror.ErrInternal.WithMessage("Failed to update script line").WithError(err)
	}

	return nil
}

// 台本行を作成する
func (r *scriptLineRepository) Create(ctx context.Context, scriptLine *model.ScriptLine) error {
	if err := r.db.WithContext(ctx).Create(scriptLine).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create script line", "error", err)
		return apperror.ErrInternal.WithMessage("Failed to create script line").WithError(err)
	}

	return nil
}

// 指定した lineOrder 以上の行の lineOrder を +1 する
func (r *scriptLineRepository) IncrementLineOrderFrom(ctx context.Context, episodeID uuid.UUID, fromLineOrder int) error {
	if err := r.db.WithContext(ctx).
		Model(&model.ScriptLine{}).
		Where("episode_id = ? AND line_order >= ?", episodeID, fromLineOrder).
		UpdateColumn("line_order", gorm.Expr("line_order + 1")).Error; err != nil {
		logger.FromContext(ctx).Error("failed to increment line order", "error", err, "episode_id", episodeID)
		return apperror.ErrInternal.WithMessage("Failed to increment line order").WithError(err)
	}

	return nil
}
