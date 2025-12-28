package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
)

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

type credentialRepository struct {
	db *gorm.DB
}

// CredentialRepository の実装を返す
func NewCredentialRepository(db *gorm.DB) CredentialRepository {
	return &credentialRepository{db: db}
}

// パスワード認証情報を作成する
func (r *credentialRepository) Create(ctx context.Context, credential *model.Credential) error {
	if err := r.db.WithContext(ctx).Create(credential).Error; err != nil {
		return apperror.ErrInternal.WithMessage("Failed to create credential").WithError(err)
	}
	return nil
}

// 指定されたユーザー ID のパスワード認証情報を取得する
func (r *credentialRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*model.Credential, error) {
	var credential model.Credential
	if err := r.db.WithContext(ctx).First(&credential, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("Credential not found")
		}
		return nil, apperror.ErrInternal.WithMessage("Failed to fetch credential").WithError(err)
	}
	return &credential, nil
}

type oauthAccountRepository struct {
	db *gorm.DB
}

// OAuthAccountRepository の実装を返す
func NewOAuthAccountRepository(db *gorm.DB) OAuthAccountRepository {
	return &oauthAccountRepository{db: db}
}

// OAuth 認証情報を作成する
func (r *oauthAccountRepository) Create(ctx context.Context, account *model.OAuthAccount) error {
	if err := r.db.WithContext(ctx).Create(account).Error; err != nil {
		return apperror.ErrInternal.WithMessage("Failed to create OAuth account").WithError(err)
	}
	return nil
}

// OAuth 認証情報を更新する
func (r *oauthAccountRepository) Update(ctx context.Context, account *model.OAuthAccount) error {
	if err := r.db.WithContext(ctx).Save(account).Error; err != nil {
		return apperror.ErrInternal.WithMessage("Failed to update OAuth account").WithError(err)
	}
	return nil
}

// 指定されたプロバイダとプロバイダユーザー ID の OAuth 認証情報を取得する
func (r *oauthAccountRepository) FindByProviderAndProviderUserID(ctx context.Context, provider, providerUserID string) (*model.OAuthAccount, error) {
	var account model.OAuthAccount
	if err := r.db.WithContext(ctx).First(&account, "provider = ? AND provider_user_id = ?", provider, providerUserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("OAuth account not found")
		}
		return nil, apperror.ErrInternal.WithMessage("Failed to fetch OAuth account").WithError(err)
	}
	return &account, nil
}
