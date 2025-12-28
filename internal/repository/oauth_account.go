package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
)

// OAuth 認証情報へのアクセスインターフェース
type OAuthAccountRepository interface {
	Create(ctx context.Context, account *model.OAuthAccount) error
	Update(ctx context.Context, account *model.OAuthAccount) error
	FindByProviderAndProviderUserID(ctx context.Context, provider, providerUserID string) (*model.OAuthAccount, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.OAuthAccount, error)
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

// 指定されたユーザー ID の OAuth 認証情報一覧を取得する
func (r *oauthAccountRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.OAuthAccount, error) {
	var accounts []model.OAuthAccount
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&accounts).Error; err != nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to fetch OAuth accounts").WithError(err)
	}
	return accounts, nil
}
