package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
)

// ユーザーデータへのアクセスインターフェース
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}

type userRepository struct {
	db *gorm.DB
}

// UserRepository の実装を返す
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// ユーザーを作成する
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return apperror.ErrInternal.WithMessage("Failed to create user").WithError(err)
	}
	return nil
}

// 指定された ID のユーザーを取得する
func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("User not found")
		}
		return nil, apperror.ErrInternal.WithMessage("Failed to fetch user").WithError(err)
	}
	return &user, nil
}

// 指定されたメールアドレスのユーザーを取得する
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("User not found")
		}
		return nil, apperror.ErrInternal.WithMessage("Failed to fetch user").WithError(err)
	}
	return &user, nil
}

// 指定されたメールアドレスのユーザーが存在するか確認する
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, apperror.ErrInternal.WithMessage("Failed to check email existence").WithError(err)
	}
	return count > 0, nil
}

// 指定されたユーザー名のユーザーが存在するか確認する
func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, apperror.ErrInternal.WithMessage("Failed to check username existence").WithError(err)
	}
	return count > 0, nil
}
