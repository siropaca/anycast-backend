package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// ScriptLineRepository は台本行データへのアクセスインターフェース
type ScriptLineRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.ScriptLine, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]model.ScriptLine, error)
	FindByEpisodeID(ctx context.Context, episodeID uuid.UUID) ([]model.ScriptLine, error)
	FindByEpisodeIDWithVoice(ctx context.Context, episodeID uuid.UUID) ([]model.ScriptLine, error)
	Create(ctx context.Context, scriptLine *model.ScriptLine) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByEpisodeID(ctx context.Context, episodeID uuid.UUID) error
	CreateBatch(ctx context.Context, scriptLines []model.ScriptLine) ([]model.ScriptLine, error)
	Update(ctx context.Context, scriptLine *model.ScriptLine) error
	IncrementLineOrderFrom(ctx context.Context, episodeID uuid.UUID, fromLineOrder int) error
	UpdateLineOrders(ctx context.Context, lineOrders map[uuid.UUID]int) error
	ExistsBySpeakerIDAndChannelID(ctx context.Context, speakerID, channelID uuid.UUID) (bool, error)
	UpdateSpeakerIDByChannelID(ctx context.Context, channelID, oldSpeakerID, newSpeakerID uuid.UUID) error
}

type scriptLineRepository struct {
	db *gorm.DB
}

// NewScriptLineRepository は ScriptLineRepository の実装を返す
func NewScriptLineRepository(db *gorm.DB) ScriptLineRepository {
	return &scriptLineRepository{db: db}
}

// FindByID は ID で台本行を取得する
func (r *scriptLineRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ScriptLine, error) {
	var scriptLine model.ScriptLine

	if err := r.db.WithContext(ctx).
		Preload("Speaker").
		Preload("Speaker.Voice").
		First(&scriptLine, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.ErrNotFound.WithMessage("台本行が見つかりません")
		}
		logger.FromContext(ctx).Error("failed to fetch script line", "error", err, "id", id)
		return nil, apperror.ErrInternal.WithMessage("台本行の取得に失敗しました").WithError(err)
	}

	return &scriptLine, nil
}

// FindByEpisodeID は指定されたエピソードの台本行一覧を取得する
func (r *scriptLineRepository) FindByEpisodeID(ctx context.Context, episodeID uuid.UUID) ([]model.ScriptLine, error) {
	var scriptLines []model.ScriptLine

	if err := r.db.WithContext(ctx).
		Preload("Speaker").
		Preload("Speaker.Voice").
		Where("episode_id = ?", episodeID).
		Order("line_order ASC").
		Find(&scriptLines).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch script lines", "error", err, "episode_id", episodeID)
		return nil, apperror.ErrInternal.WithMessage("台本行一覧の取得に失敗しました").WithError(err)
	}

	return scriptLines, nil
}

// FindByEpisodeIDWithVoice は指定されたエピソードの台本行一覧を取得する（Voice 情報を含む）
func (r *scriptLineRepository) FindByEpisodeIDWithVoice(ctx context.Context, episodeID uuid.UUID) ([]model.ScriptLine, error) {
	var scriptLines []model.ScriptLine

	if err := r.db.WithContext(ctx).
		Preload("Speaker").
		Preload("Speaker.Voice").
		Where("episode_id = ?", episodeID).
		Order("line_order ASC").
		Find(&scriptLines).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch script lines with voice", "error", err, "episode_id", episodeID)
		return nil, apperror.ErrInternal.WithMessage("台本行一覧の取得に失敗しました").WithError(err)
	}

	return scriptLines, nil
}

// Delete は指定された台本行を削除する
func (r *scriptLineRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.ScriptLine{}, "id = ?", id)
	if result.Error != nil {
		logger.FromContext(ctx).Error("failed to delete script line", "error", result.Error, "id", id)
		return apperror.ErrInternal.WithMessage("台本行の削除に失敗しました").WithError(result.Error)
	}

	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("台本行が見つかりません")
	}

	return nil
}

// DeleteByEpisodeID は指定されたエピソードの台本行を全て削除する
func (r *scriptLineRepository) DeleteByEpisodeID(ctx context.Context, episodeID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Where("episode_id = ?", episodeID).
		Delete(&model.ScriptLine{}).Error; err != nil {
		logger.FromContext(ctx).Error("failed to delete script lines", "error", err, "episode_id", episodeID)
		return apperror.ErrInternal.WithMessage("台本行の一括削除に失敗しました").WithError(err)
	}

	return nil
}

// CreateBatch は複数の台本行を一括作成する
func (r *scriptLineRepository) CreateBatch(ctx context.Context, scriptLines []model.ScriptLine) ([]model.ScriptLine, error) {
	if len(scriptLines) == 0 {
		return []model.ScriptLine{}, nil
	}

	if err := r.db.WithContext(ctx).Create(&scriptLines).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create script lines", "error", err)
		return nil, apperror.ErrInternal.WithMessage("台本行の一括作成に失敗しました").WithError(err)
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
		return nil, apperror.ErrInternal.WithMessage("作成した台本行の取得に失敗しました").WithError(err)
	}

	return created, nil
}

// Update は台本行を更新する
func (r *scriptLineRepository) Update(ctx context.Context, scriptLine *model.ScriptLine) error {
	if err := r.db.WithContext(ctx).Save(scriptLine).Error; err != nil {
		logger.FromContext(ctx).Error("failed to update script line", "error", err, "id", scriptLine.ID)
		return apperror.ErrInternal.WithMessage("台本行の更新に失敗しました").WithError(err)
	}

	return nil
}

// Create は台本行を作成する
func (r *scriptLineRepository) Create(ctx context.Context, scriptLine *model.ScriptLine) error {
	if err := r.db.WithContext(ctx).Create(scriptLine).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create script line", "error", err)
		return apperror.ErrInternal.WithMessage("台本行の作成に失敗しました").WithError(err)
	}

	return nil
}

// IncrementLineOrderFrom は指定した lineOrder 以上の行の lineOrder を +1 する
func (r *scriptLineRepository) IncrementLineOrderFrom(ctx context.Context, episodeID uuid.UUID, fromLineOrder int) error {
	if err := r.db.WithContext(ctx).
		Model(&model.ScriptLine{}).
		Where("episode_id = ? AND line_order >= ?", episodeID, fromLineOrder).
		UpdateColumn("line_order", gorm.Expr("line_order + 1")).Error; err != nil {
		logger.FromContext(ctx).Error("failed to increment line order", "error", err, "episode_id", episodeID)
		return apperror.ErrInternal.WithMessage("行順序の更新に失敗しました").WithError(err)
	}

	return nil
}

// FindByIDs は複数の ID で台本行を取得する
func (r *scriptLineRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]model.ScriptLine, error) {
	var scriptLines []model.ScriptLine

	if len(ids) == 0 {
		return scriptLines, nil
	}

	if err := r.db.WithContext(ctx).
		Preload("Speaker").
		Preload("Speaker.Voice").
		Where("id IN ?", ids).
		Find(&scriptLines).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch script lines by ids", "error", err)
		return nil, apperror.ErrInternal.WithMessage("台本行一覧の取得に失敗しました").WithError(err)
	}

	return scriptLines, nil
}

// ExistsBySpeakerIDAndChannelID はチャンネル配下のエピソードの台本行にそのキャラクターが使われているかを返す
func (r *scriptLineRepository) ExistsBySpeakerIDAndChannelID(ctx context.Context, speakerID, channelID uuid.UUID) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Model(&model.ScriptLine{}).
		Joins("JOIN episodes ON episodes.id = script_lines.episode_id").
		Where("episodes.channel_id = ? AND script_lines.speaker_id = ?", channelID, speakerID).
		Limit(1).
		Count(&count).Error; err != nil {
		logger.FromContext(ctx).Error("failed to check script line speaker existence", "error", err, "speaker_id", speakerID, "channel_id", channelID)
		return false, apperror.ErrInternal.WithMessage("台本行の話者チェックに失敗しました").WithError(err)
	}

	return count > 0, nil
}

// UpdateSpeakerIDByChannelID はチャンネル配下の全エピソードの台本行で旧 speakerID を新 speakerID に一括更新する
func (r *scriptLineRepository) UpdateSpeakerIDByChannelID(ctx context.Context, channelID, oldSpeakerID, newSpeakerID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Model(&model.ScriptLine{}).
		Where("speaker_id = ? AND episode_id IN (?)",
			oldSpeakerID,
			r.db.Model(&model.Episode{}).Select("id").Where("channel_id = ?", channelID),
		).
		UpdateColumn("speaker_id", newSpeakerID).Error; err != nil {
		logger.FromContext(ctx).Error("failed to update speaker_id in script lines", "error", err, "channel_id", channelID, "old_speaker_id", oldSpeakerID, "new_speaker_id", newSpeakerID)
		return apperror.ErrInternal.WithMessage("台本行の話者更新に失敗しました").WithError(err)
	}

	return nil
}

// UpdateLineOrders は複数の台本行の lineOrder を一括更新する
func (r *scriptLineRepository) UpdateLineOrders(ctx context.Context, lineOrders map[uuid.UUID]int) error {
	for id, order := range lineOrders {
		if err := r.db.WithContext(ctx).
			Model(&model.ScriptLine{}).
			Where("id = ?", id).
			UpdateColumn("line_order", order).Error; err != nil {
			logger.FromContext(ctx).Error("failed to update line order", "error", err, "id", id)
			return apperror.ErrInternal.WithMessage("行順序の更新に失敗しました").WithError(err)
		}
	}

	return nil
}
