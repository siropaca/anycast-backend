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
	FindByEpisodeID(ctx context.Context, episodeID uuid.UUID) ([]model.ScriptLine, error)
}

type scriptLineRepository struct {
	db *gorm.DB
}

// ScriptLineRepository の実装を返す
func NewScriptLineRepository(db *gorm.DB) ScriptLineRepository {
	return &scriptLineRepository{db: db}
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
