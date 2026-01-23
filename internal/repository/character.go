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

// CharacterRepository はキャラクターデータへのアクセスインターフェース
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

// CharacterFilter はキャラクター検索のフィルタ条件を表す
type CharacterFilter struct {
	Limit  int
	Offset int
}

type characterRepository struct {
	db *gorm.DB
}

// NewCharacterRepository は CharacterRepository の実装を返す
func NewCharacterRepository(db *gorm.DB) CharacterRepository {
	return &characterRepository{db: db}
}

// FindByUserID は指定されたユーザーのキャラクター一覧を取得する
func (r *characterRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filter CharacterFilter) ([]model.Character, int64, error) {
	var characters []model.Character
	var total int64

	tx := r.db.WithContext(ctx).Model(&model.Character{}).Where("user_id = ?", userID)

	// 総件数を取得
	if err := tx.Count(&total).Error; err != nil {
		logger.FromContext(ctx).Error("failed to count characters", "error", err, "user_id", userID)
		return nil, 0, apperror.ErrInternal.WithMessage("キャラクター数の取得に失敗しました").WithError(err)
	}

	// ページネーションとリレーションのプリロード
	if err := tx.
		Preload("Avatar").
		Preload("Voice").
		Preload("ChannelCharacters.Channel").
		Order("created_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&characters).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch characters", "error", err, "user_id", userID)
		return nil, 0, apperror.ErrInternal.WithMessage("キャラクター一覧の取得に失敗しました").WithError(err)
	}

	return characters, total, nil
}

// FindByID は指定された ID のキャラクターを取得する
func (r *characterRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Character, error) {
	var character model.Character

	if err := r.db.WithContext(ctx).
		Preload("Avatar").
		Preload("Voice").
		Preload("ChannelCharacters.Channel").
		Preload("ChannelCharacters.Channel.Category").
		Preload("ChannelCharacters.Channel.Artwork").
		First(&character, "id = ?", id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("キャラクターが見つかりません")
		}

		logger.FromContext(ctx).Error("failed to fetch character", "error", err, "character_id", id)
		return nil, apperror.ErrInternal.WithMessage("キャラクターの取得に失敗しました").WithError(err)
	}

	return &character, nil
}

// FindByIDs は複数の ID でキャラクターを取得する
func (r *characterRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Character, error) {
	var characters []model.Character

	if len(ids) == 0 {
		return characters, nil
	}

	if err := r.db.WithContext(ctx).
		Preload("Avatar").
		Preload("Voice").
		Where("id IN ?", ids).
		Find(&characters).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch characters by ids", "error", err, "ids", ids)
		return nil, apperror.ErrInternal.WithMessage("キャラクター一覧の取得に失敗しました").WithError(err)
	}

	return characters, nil
}

// Create はキャラクターを作成する
func (r *characterRepository) Create(ctx context.Context, character *model.Character) error {
	if err := r.db.WithContext(ctx).Create(character).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create character", "error", err)
		return apperror.ErrInternal.WithMessage("キャラクターの作成に失敗しました").WithError(err)
	}

	return nil
}

// Update はキャラクターを更新する
func (r *characterRepository) Update(ctx context.Context, character *model.Character) error {
	if err := r.db.WithContext(ctx).Save(character).Error; err != nil {
		logger.FromContext(ctx).Error("failed to update character", "error", err, "character_id", character.ID)
		return apperror.ErrInternal.WithMessage("キャラクターの更新に失敗しました").WithError(err)
	}

	return nil
}

// Delete はキャラクターを削除する
func (r *characterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Character{}, "id = ?", id)
	if result.Error != nil {
		logger.FromContext(ctx).Error("failed to delete character", "error", result.Error, "character_id", id)
		return apperror.ErrInternal.WithMessage("キャラクターの削除に失敗しました").WithError(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("キャラクターが見つかりません")
	}

	return nil
}

// IsUsedInAnyChannel はキャラクターがいずれかのチャンネルで使用中かどうかを確認する
func (r *characterRepository) IsUsedInAnyChannel(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Model(&model.ChannelCharacter{}).
		Where("character_id = ?", id).
		Count(&count).Error; err != nil {
		logger.FromContext(ctx).Error("failed to check character usage", "error", err, "character_id", id)
		return false, apperror.ErrInternal.WithMessage("キャラクターの使用状況確認に失敗しました").WithError(err)
	}

	return count > 0, nil
}

// ExistsByUserIDAndName は同一ユーザー内で同じ名前のキャラクターが存在するかどうかを確認する
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
		return false, apperror.ErrInternal.WithMessage("キャラクター名の確認に失敗しました").WithError(err)
	}

	return count > 0, nil
}
