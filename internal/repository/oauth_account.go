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

// OAuth 認証情報へのアクセスインターフェース
type OAuthAccountRepository interface {
	Create(ctx context.Context, account *model.OAuthAccount) error
	Update(ctx context.Context, account *model.OAuthAccount) error
	FindByProviderAndProviderUserID(ctx context.Context, provider model.OAuthProvider, providerUserID string) (*model.OAuthAccount, error)
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
		logger.FromContext(ctx).Error("failed to create oauth account", "error", err, "provider", account.Provider)
		return apperror.ErrInternal.WithMessage("OAuth アカウントの作成に失敗しました").WithError(err)
	}

	return nil
}

// OAuth 認証情報を更新する
func (r *oauthAccountRepository) Update(ctx context.Context, account *model.OAuthAccount) error {
	if err := r.db.WithContext(ctx).Save(account).Error; err != nil {
		logger.FromContext(ctx).Error("failed to update oauth account", "error", err, "account_id", account.ID)
		return apperror.ErrInternal.WithMessage("OAuth アカウントの更新に失敗しました").WithError(err)
	}

	return nil
}

// 指定されたプロバイダとプロバイダユーザー ID の OAuth 認証情報を取得する
func (r *oauthAccountRepository) FindByProviderAndProviderUserID(ctx context.Context, provider model.OAuthProvider, providerUserID string) (*model.OAuthAccount, error) {
	var account model.OAuthAccount

	if err := r.db.WithContext(ctx).First(&account, "provider = ? AND provider_user_id = ?", provider, providerUserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("OAuth アカウントが見つかりません")
		}

		logger.FromContext(ctx).Error("failed to fetch oauth account", "error", err, "provider", provider)
		return nil, apperror.ErrInternal.WithMessage("OAuth アカウントの取得に失敗しました").WithError(err)
	}

	return &account, nil
}

// 指定されたユーザー ID の OAuth 認証情報一覧を取得する
func (r *oauthAccountRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.OAuthAccount, error) {
	var accounts []model.OAuthAccount

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&accounts).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch oauth accounts", "error", err, "user_id", userID)
		return nil, apperror.ErrInternal.WithMessage("OAuth アカウント一覧の取得に失敗しました").WithError(err)
	}

	return accounts, nil
}
