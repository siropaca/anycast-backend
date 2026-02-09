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

// SearchUserFilter はユーザー検索のフィルタ条件を表す
type SearchUserFilter struct {
	Query  string
	Limit  int
	Offset int
}

// UserRepository はユーザーデータへのアクセスインターフェース
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByIDWithAvatar(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByUsernameWithAvatar(ctx context.Context, username string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	Search(ctx context.Context, filter SearchUserFilter) ([]model.User, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository は UserRepository の実装を返す
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create はユーザーを作成する
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create user", "error", err)
		return apperror.ErrInternal.WithMessage("ユーザーの作成に失敗しました").WithError(err)
	}

	return nil
}

// Update はユーザーを更新する
func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		logger.FromContext(ctx).Error("failed to update user", "error", err, "user_id", user.ID)
		return apperror.ErrInternal.WithMessage("ユーザーの更新に失敗しました").WithError(err)
	}

	return nil
}

// FindByID は指定された ID のユーザーを取得する
func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User

	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("ユーザーが見つかりません")
		}
		logger.FromContext(ctx).Error("failed to fetch user by id", "error", err, "user_id", id)
		return nil, apperror.ErrInternal.WithMessage("ユーザーの取得に失敗しました").WithError(err)
	}

	return &user, nil
}

// FindByIDWithAvatar は指定された ID のユーザーを Avatar リレーション付きで取得する
func (r *userRepository) FindByIDWithAvatar(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User

	if err := r.db.WithContext(ctx).
		Preload("Avatar").
		First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("ユーザーが見つかりません")
		}
		logger.FromContext(ctx).Error("failed to fetch user by id with avatar", "error", err, "user_id", id)
		return nil, apperror.ErrInternal.WithMessage("ユーザーの取得に失敗しました").WithError(err)
	}

	return &user, nil
}

// FindByUsernameWithAvatar は指定されたユーザー名のユーザーを Avatar リレーション付きで取得する
func (r *userRepository) FindByUsernameWithAvatar(ctx context.Context, username string) (*model.User, error) {
	var user model.User

	if err := r.db.WithContext(ctx).
		Preload("Avatar").
		Preload("HeaderImage").
		First(&user, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("ユーザーが見つかりません")
		}
		logger.FromContext(ctx).Error("failed to fetch user by username with avatar", "error", err, "username", username)
		return nil, apperror.ErrInternal.WithMessage("ユーザーの取得に失敗しました").WithError(err)
	}

	return &user, nil
}

// FindByEmail は指定されたメールアドレスのユーザーを取得する
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User

	if err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("ユーザーが見つかりません")
		}
		logger.FromContext(ctx).Error("failed to fetch user by email", "error", err)
		return nil, apperror.ErrInternal.WithMessage("ユーザーの取得に失敗しました").WithError(err)
	}

	return &user, nil
}

// ExistsByEmail は指定されたメールアドレスのユーザーが存在するか確認する
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		logger.FromContext(ctx).Error("failed to check email existence", "error", err)
		return false, apperror.ErrInternal.WithMessage("メールアドレスの重複確認に失敗しました").WithError(err)
	}

	return count > 0, nil
}

// ExistsByUsername は指定されたユーザー名のユーザーが存在するか確認する
func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		logger.FromContext(ctx).Error("failed to check username existence", "error", err)
		return false, apperror.ErrInternal.WithMessage("ユーザー名の重複確認に失敗しました").WithError(err)
	}

	return count > 0, nil
}

// Search はキーワードでユーザーを検索する
func (r *userRepository) Search(ctx context.Context, filter SearchUserFilter) ([]model.User, int64, error) {
	keyword := "%" + filter.Query + "%"

	tx := r.db.WithContext(ctx).Model(&model.User{}).
		Where("(username ILIKE ? OR display_name ILIKE ?)", keyword, keyword)

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		logger.FromContext(ctx).Error("failed to count search users", "error", err)
		return nil, 0, apperror.ErrInternal.WithMessage("ユーザーの検索に失敗しました").WithError(err)
	}

	var users []model.User
	if err := tx.Preload("Avatar").
		Order("created_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&users).Error; err != nil {
		logger.FromContext(ctx).Error("failed to search users", "error", err)
		return nil, 0, apperror.ErrInternal.WithMessage("ユーザーの検索に失敗しました").WithError(err)
	}

	return users, total, nil
}

// Delete は指定された ID のユーザーを削除する
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.User{}, "id = ?", id)
	if result.Error != nil {
		logger.FromContext(ctx).Error("failed to delete user", "error", result.Error, "user_id", id)
		return apperror.ErrInternal.WithMessage("ユーザーの削除に失敗しました").WithError(result.Error)
	}

	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("ユーザーが見つかりません")
	}

	return nil
}
