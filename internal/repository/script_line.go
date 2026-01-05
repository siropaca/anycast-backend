package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/logger"
	"github.com/siropaca/anycast-backend/internal/model"
)

// 台本行データへのアクセスインターフェース
type ScriptLineRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.ScriptLine, error)
	FindByEpisodeID(ctx context.Context, episodeID uuid.UUID) ([]model.ScriptLine, error)
	DeleteByEpisodeID(ctx context.Context, episodeID uuid.UUID) error
	CreateBatch(ctx context.Context, scriptLines []model.ScriptLine) ([]model.ScriptLine, error)
	Update(ctx context.Context, scriptLine *model.ScriptLine) error
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
		Preload("Sfx").
		Preload("Audio").
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
		Preload("Sfx").
		Preload("Audio").
		Where("episode_id = ?", episodeID).
		Order("line_order ASC").
		Find(&scriptLines).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch script lines", "error", err, "episode_id", episodeID)
		return nil, apperror.ErrInternal.WithMessage("Failed to fetch script lines").WithError(err)
	}

	return scriptLines, nil
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
		Preload("Sfx").
		Preload("Audio").
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
