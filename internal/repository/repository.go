package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/siropaca/anycast-backend/internal/model"
)

// ボイスデータへのアクセスインターフェース
type VoiceRepository interface {
	FindAll(ctx context.Context, filter VoiceFilter) ([]model.Voice, error)
	FindByID(ctx context.Context, id string) (*model.Voice, error)
}

// ボイス検索のフィルタ条件
type VoiceFilter struct {
	Provider *string
	Gender   *string
}

// ユーザーデータへのアクセスインターフェース
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}

// パスワード認証情報へのアクセスインターフェース
type CredentialRepository interface {
	Create(ctx context.Context, credential *model.Credential) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*model.Credential, error)
}

// OAuth 認証情報へのアクセスインターフェース
type OAuthAccountRepository interface {
	Create(ctx context.Context, account *model.OAuthAccount) error
	Update(ctx context.Context, account *model.OAuthAccount) error
	FindByProviderAndProviderUserID(ctx context.Context, provider, providerUserID string) (*model.OAuthAccount, error)
}
