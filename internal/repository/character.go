package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

// キャラクターデータへのアクセスインターフェース
type CharacterRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Character, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, filter CharacterFilter) ([]model.Character, int64, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Character, error)
	Create(ctx context.Context, character *model.Character) error
	Update(ctx context.Context, character *model.Character) error
	Delete(ctx context.Context, id uuid.UUID) error
	IsUsedInAnyChannel(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsByUserIDAndName(ctx context.Context, userID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error)
}

// キャラクター検索のフィルタ条件
type CharacterFilter struct {
	Limit  int
	Offset int
}

type characterRepository struct {
	db *gorm.DB
}

// CharacterRepository の実装を返す
func NewCharacterRepository(db *gorm.DB) CharacterRepository {
	return &characterRepository{db: db}
}

// 指定されたユーザーのキャラクター一覧を取得する
func (r *characterRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filter CharacterFilter) ([]model.Character, int64, error) {
	var characters []model.Character
	var total int64

	tx := r.db.WithContext(ctx).Model(&model.Character{}).Where("user_id = ?", userID)

	// 総件数を取得
	if err := tx.Count(&total).Error; err != nil {
		logger.FromContext(ctx).Error("failed to count characters", "error", err, "user_id", userID)
		return nil, 0, apperror.ErrInternal.WithMessage("Failed to count characters").WithError(err)
	}

	// ページネーションとリレーションのプリロード
	if err := tx.
		Preload("Voice").
		Order("created_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&characters).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch characters", "error", err, "user_id", userID)
		return nil, 0, apperror.ErrInternal.WithMessage("Failed to fetch characters").WithError(err)
	}

	return characters, total, nil
}

// 指定された ID のキャラクターを取得する
func (r *characterRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Character, error) {
	var character model.Character

	if err := r.db.WithContext(ctx).
		Preload("Voice").
		First(&character, "id = ?", id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("Character not found")
		}

		logger.FromContext(ctx).Error("failed to fetch character", "error", err, "character_id", id)
		return nil, apperror.ErrInternal.WithMessage("Failed to fetch character").WithError(err)
	}

	return &character, nil
}

// 複数の ID でキャラクターを取得する
func (r *characterRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Character, error) {
	var characters []model.Character

	if len(ids) == 0 {
		return characters, nil
	}

	if err := r.db.WithContext(ctx).
		Preload("Voice").
		Where("id IN ?", ids).
		Find(&characters).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch characters by ids", "error", err, "ids", ids)
		return nil, apperror.ErrInternal.WithMessage("Failed to fetch characters").WithError(err)
	}

	return characters, nil
}

// キャラクターを作成する
func (r *characterRepository) Create(ctx context.Context, character *model.Character) error {
	if err := r.db.WithContext(ctx).Create(character).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create character", "error", err)
		return apperror.ErrInternal.WithMessage("Failed to create character").WithError(err)
	}

	return nil
}

// キャラクターを更新する
func (r *characterRepository) Update(ctx context.Context, character *model.Character) error {
	if err := r.db.WithContext(ctx).Save(character).Error; err != nil {
		logger.FromContext(ctx).Error("failed to update character", "error", err, "character_id", character.ID)
		return apperror.ErrInternal.WithMessage("Failed to update character").WithError(err)
	}

	return nil
}

// キャラクターを削除する
func (r *characterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Character{}, "id = ?", id)
	if result.Error != nil {
		logger.FromContext(ctx).Error("failed to delete character", "error", result.Error, "character_id", id)
		return apperror.ErrInternal.WithMessage("Failed to delete character").WithError(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("Character not found")
	}

	return nil
}

// キャラクターがいずれかのチャンネルで使用中かどうかを確認する
func (r *characterRepository) IsUsedInAnyChannel(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Model(&model.ChannelCharacter{}).
		Where("character_id = ?", id).
		Count(&count).Error; err != nil {
		logger.FromContext(ctx).Error("failed to check character usage", "error", err, "character_id", id)
		return false, apperror.ErrInternal.WithMessage("Failed to check character usage").WithError(err)
	}

	return count > 0, nil
}

// 同一ユーザー内で同じ名前のキャラクターが存在するかどうかを確認する
func (r *characterRepository) ExistsByUserIDAndName(ctx context.Context, userID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	var count int64

	tx := r.db.WithContext(ctx).
		Model(&model.Character{}).
		Where("user_id = ? AND name = ?", userID, name)

	if excludeID != nil {
		tx = tx.Where("id != ?", *excludeID)
	}

	if err := tx.Count(&count).Error; err != nil {
		logger.FromContext(ctx).Error("failed to check character name", "error", err, "user_id", userID, "name", name)
		return false, apperror.ErrInternal.WithMessage("Failed to check character name").WithError(err)
	}

	return count > 0, nil
}
