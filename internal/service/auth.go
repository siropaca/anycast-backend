package service

import (
	"context"
	"fmt"
	"math/rand/v2"
	"regexp"
	"strings"
	"time"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/logger"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/crypto"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// 認証結果
type AuthResult struct {
	User      response.UserResponse
	IsCreated bool // 新規作成されたかどうか（OAuth 用）
}

// 認証関連のビジネスロジックインターフェース
type AuthService interface {
	Register(ctx context.Context, req request.RegisterRequest) (*response.UserResponse, error)
	Login(ctx context.Context, req request.LoginRequest) (*response.UserResponse, error)
	OAuthGoogle(ctx context.Context, req request.OAuthGoogleRequest) (*AuthResult, error)
	GetMe(ctx context.Context, userID string) (*response.MeResponse, error)
}

type authService struct {
	userRepo         repository.UserRepository
	credentialRepo   repository.CredentialRepository
	oauthAccountRepo repository.OAuthAccountRepository
	imageRepo        repository.ImageRepository
	passwordHasher   crypto.PasswordHasher
	storageClient    storage.Client
}

// AuthService の実装を返す
func NewAuthService(
	userRepo repository.UserRepository,
	credentialRepo repository.CredentialRepository,
	oauthAccountRepo repository.OAuthAccountRepository,
	imageRepo repository.ImageRepository,
	passwordHasher crypto.PasswordHasher,
	storageClient storage.Client,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		credentialRepo:   credentialRepo,
		oauthAccountRepo: oauthAccountRepo,
		imageRepo:        imageRepo,
		passwordHasher:   passwordHasher,
		storageClient:    storageClient,
	}
}

// ユーザーを登録する
func (s *authService) Register(ctx context.Context, req request.RegisterRequest) (*response.UserResponse, error) {
	// メールアドレスの重複チェック
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.ErrDuplicateEmail.WithMessage("このメールアドレスは既に使用されています")
	}

	// ユーザー名を自動生成
	username, err := s.generateUniqueUsername(ctx, req.DisplayName)
	if err != nil {
		return nil, err
	}

	// パスワードをハッシュ化
	hashedPassword, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		logger.FromContext(ctx).Error("failed to hash password", "error", err)
		return nil, apperror.ErrInternal.WithMessage("Failed to hash password").WithError(err)
	}

	// ユーザーを作成
	user := &model.User{
		Email:       req.Email,
		Username:    username,
		DisplayName: req.DisplayName,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 認証情報を作成
	credential := &model.Credential{
		UserID:       user.ID,
		PasswordHash: hashedPassword,
	}
	if err := s.credentialRepo.Create(ctx, credential); err != nil {
		return nil, err
	}

	return s.toUserResponse(ctx, user), nil
}

// メールアドレスとパスワードで認証する
func (s *authService) Login(ctx context.Context, req request.LoginRequest) (*response.UserResponse, error) {
	// ユーザーを取得
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		// NotFound エラーの場合も認証エラーとして扱う
		if apperror.IsCode(err, apperror.CodeNotFound) {
			return nil, apperror.ErrInvalidCredentials.WithMessage("メールアドレスまたはパスワードが正しくありません")
		}
		return nil, err
	}

	// 認証情報を取得
	credential, err := s.credentialRepo.FindByUserID(ctx, user.ID)
	if err != nil {
		// 認証情報がない場合（OAuth のみのユーザー）
		if apperror.IsCode(err, apperror.CodeNotFound) {
			return nil, apperror.ErrInvalidCredentials.WithMessage("メールアドレスまたはパスワードが正しくありません")
		}
		return nil, err
	}

	// パスワードを検証
	if err := s.passwordHasher.Compare(credential.PasswordHash, req.Password); err != nil {
		return nil, apperror.ErrInvalidCredentials.WithMessage("メールアドレスまたはパスワードが正しくありません")
	}

	return s.toUserResponse(ctx, user), nil
}

// Google OAuth で認証する
func (s *authService) OAuthGoogle(ctx context.Context, req request.OAuthGoogleRequest) (*AuthResult, error) {
	// 既存の OAuth アカウントを検索
	existingAccount, err := s.oauthAccountRepo.FindByProviderAndProviderUserID(ctx, model.OAuthProviderGoogle, req.ProviderUserID)
	if err != nil {
		// NotFound エラーの場合は新規作成へ進む
		if !apperror.IsCode(err, apperror.CodeNotFound) {
			return nil, err
		}
		existingAccount = nil
	}

	if existingAccount != nil {
		// 既存ユーザー: トークン情報を更新
		existingAccount.AccessToken = &req.AccessToken
		existingAccount.RefreshToken = req.RefreshToken
		if req.ExpiresAt != nil {
			expiresAt := time.Unix(*req.ExpiresAt, 0)
			existingAccount.ExpiresAt = &expiresAt
		}
		existingAccount.UpdatedAt = time.Now()

		if err := s.oauthAccountRepo.Update(ctx, existingAccount); err != nil {
			return nil, err
		}

		// ユーザー情報を取得
		user, err := s.userRepo.FindByID(ctx, existingAccount.UserID)
		if err != nil {
			return nil, err
		}

		return &AuthResult{
			User:      *s.toUserResponse(ctx, user),
			IsCreated: false,
		}, nil
	}

	// ユーザー名を自動生成
	username, err := s.generateUniqueUsername(ctx, req.DisplayName)
	if err != nil {
		return nil, err
	}

	// 新規ユーザー: ユーザーと OAuth アカウントを作成
	user := &model.User{
		Email:       req.Email,
		Username:    username,
		DisplayName: req.DisplayName,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	oauthAccount := &model.OAuthAccount{
		UserID:         user.ID,
		Provider:       model.OAuthProviderGoogle,
		ProviderUserID: req.ProviderUserID,
		AccessToken:    &req.AccessToken,
		RefreshToken:   req.RefreshToken,
	}
	if req.ExpiresAt != nil {
		expiresAt := time.Unix(*req.ExpiresAt, 0)
		oauthAccount.ExpiresAt = &expiresAt
	}

	if err := s.oauthAccountRepo.Create(ctx, oauthAccount); err != nil {
		return nil, err
	}

	return &AuthResult{
		User:      *s.toUserResponse(ctx, user),
		IsCreated: true,
	}, nil
}

// 連続する半角スペースにマッチする正規表現
var multiSpaceRegex = regexp.MustCompile(`\s+`)

// displayName をユーザー名形式に変換する
func displayNameToUsername(displayName string) string {
	// 全角スペースを半角スペースに変換
	s := strings.ReplaceAll(displayName, "\u3000", " ")
	// 連続する半角スペースを1個に圧縮
	s = multiSpaceRegex.ReplaceAllString(s, " ")
	// 前後の空白を削除
	s = strings.TrimSpace(s)
	// 半角スペースをアンダースコアに変換
	return strings.ReplaceAll(s, " ", "_")
}

// ユーザー名にランダムなサフィックスを付与する
func appendRandomSuffix(username string) string {
	suffix := rand.IntN(10000)
	return fmt.Sprintf("%s_%d", username, suffix)
}

// displayName からユニークなユーザー名を生成する
func (s *authService) generateUniqueUsername(ctx context.Context, displayName string) (string, error) {
	base := displayNameToUsername(displayName)

	// まずベース名で重複チェック
	exists, err := s.userRepo.ExistsByUsername(ctx, base)
	if err != nil {
		return "", err
	}
	if !exists {
		return base, nil
	}

	// 重複する場合はランダムな番号を付与してリトライ
	const maxRetries = 10
	for i := 0; i < maxRetries; i++ {
		candidate := appendRandomSuffix(base)

		exists, err := s.userRepo.ExistsByUsername(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}

	logger.FromContext(ctx).Error("failed to generate unique username after retries", "display_name", displayName)
	return "", apperror.ErrInternal.WithMessage("ユーザー名の生成に失敗しました")
}

// model.User を response.UserResponse に変換する
func (s *authService) toUserResponse(ctx context.Context, user *model.User) *response.UserResponse {
	var avatarURL *string
	if user.AvatarID != nil {
		image, err := s.imageRepo.FindByID(ctx, *user.AvatarID)
		if err == nil {
			signedURL, err := s.storageClient.GenerateSignedURL(ctx, image.Path, signedURLExpiration)
			if err == nil {
				avatarURL = &signedURL
			}
		}
		// アバターが見つからない場合やURL生成に失敗した場合はエラーにせず nil のまま
	}

	return &response.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Role:        string(user.Role),
		AvatarURL:   avatarURL,
	}
}

// 現在のユーザー情報を取得する
func (s *authService) GetMe(ctx context.Context, userID string) (*response.MeResponse, error) {
	// UUID をパース
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	// ユーザーを取得
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// パスワード設定の有無を確認
	_, credErr := s.credentialRepo.FindByUserID(ctx, id)
	hasPassword := credErr == nil

	// 連携済みの OAuth プロバイダを取得
	oauthAccounts, err := s.oauthAccountRepo.FindByUserID(ctx, id)
	if err != nil {
		return nil, err
	}
	providers := make([]string, 0, len(oauthAccounts))
	for _, account := range oauthAccounts {
		providers = append(providers, string(account.Provider))
	}

	// アバター画像を取得
	var avatar *response.AvatarResponse
	if user.AvatarID != nil {
		image, err := s.imageRepo.FindByID(ctx, *user.AvatarID)
		if err == nil {
			signedURL, err := s.storageClient.GenerateSignedURL(ctx, image.Path, signedURLExpiration)
			if err == nil {
				avatar = &response.AvatarResponse{
					ID:  image.ID,
					URL: signedURL,
				}
			}
		}
		// アバターが見つからない場合やURL生成に失敗した場合はエラーにせず nil のまま
	}

	return &response.MeResponse{
		ID:             user.ID,
		Email:          user.Email,
		Username:       user.Username,
		DisplayName:    user.DisplayName,
		Role:           string(user.Role),
		Avatar:         avatar,
		HasPassword:    hasPassword,
		OAuthProviders: providers,
		CreatedAt:      user.CreatedAt,
	}, nil
}
