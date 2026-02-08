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
