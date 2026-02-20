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
	"github.com/siropaca/anycast-backend/internal/infrastructure/slack"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/crypto"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/token"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// リフレッシュトークンの有効期限
const refreshTokenExpiration = 30 * 24 * time.Hour

// AuthResult は認証結果を表す
type AuthResult struct {
	User         response.UserResponse
	RefreshToken string
	IsCreated    bool // 新規作成されたかどうか（OAuth 用）
}

// RefreshResult はトークンリフレッシュの結果を表す
type RefreshResult struct {
	UserID       uuid.UUID
	RefreshToken string
}

// AuthService は認証関連のビジネスロジックインターフェースを表す
type AuthService interface {
	Register(ctx context.Context, req request.RegisterRequest) (*AuthResult, error)
	Login(ctx context.Context, req request.LoginRequest) (*AuthResult, error)
	OAuthGoogle(ctx context.Context, req request.OAuthGoogleRequest) (*AuthResult, error)
	RefreshToken(ctx context.Context, req request.RefreshTokenRequest) (*RefreshResult, error)
	ChangePassword(ctx context.Context, userID string, req request.ChangePasswordRequest) error
	Logout(ctx context.Context, userID string, req request.LogoutRequest) error
	GetMe(ctx context.Context, userID string) (*response.MeResponse, error)
	UpdateMe(ctx context.Context, userID string, req request.UpdateMeRequest) (*response.MeResponse, error)
	UpdatePrompt(ctx context.Context, userID string, req request.UpdateUserPromptRequest) (*response.MeResponse, error)
	UpdateUsername(ctx context.Context, userID string, req request.UpdateUsernameRequest) (*response.MeResponse, error)
	CheckUsernameAvailability(ctx context.Context, userID string, req request.CheckUsernameRequest) (*response.UsernameCheckResponse, error)
	DeleteMe(ctx context.Context, userID string) error
}

type authService struct {
	userRepo         repository.UserRepository
	credentialRepo   repository.CredentialRepository
	oauthAccountRepo repository.OAuthAccountRepository
	refreshTokenRepo repository.RefreshTokenRepository
	imageRepo        repository.ImageRepository
	playlistRepo     repository.PlaylistRepository
	audioJobRepo     repository.AudioJobRepository
	scriptJobRepo    repository.ScriptJobRepository
	passwordHasher   crypto.PasswordHasher
	storageClient    storage.Client
	slackClient      slack.Client
}

// NewAuthService は authService を生成して AuthService として返す
func NewAuthService(
	userRepo repository.UserRepository,
	credentialRepo repository.CredentialRepository,
	oauthAccountRepo repository.OAuthAccountRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	imageRepo repository.ImageRepository,
	playlistRepo repository.PlaylistRepository,
	audioJobRepo repository.AudioJobRepository,
	scriptJobRepo repository.ScriptJobRepository,
	passwordHasher crypto.PasswordHasher,
	storageClient storage.Client,
	slackClient slack.Client,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		credentialRepo:   credentialRepo,
		oauthAccountRepo: oauthAccountRepo,
		refreshTokenRepo: refreshTokenRepo,
		imageRepo:        imageRepo,
		playlistRepo:     playlistRepo,
		audioJobRepo:     audioJobRepo,
		scriptJobRepo:    scriptJobRepo,
		passwordHasher:   passwordHasher,
		storageClient:    storageClient,
		slackClient:      slackClient,
	}
}

// Register は新規ユーザーを登録する
func (s *authService) Register(ctx context.Context, req request.RegisterRequest) (*AuthResult, error) {
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
		return nil, apperror.ErrInternal.WithMessage("パスワードのハッシュ化に失敗しました").WithError(err)
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

	// デフォルト再生リストを作成
	if err := s.createDefaultPlaylist(ctx, user.ID); err != nil {
		logger.FromContext(ctx).Error("failed to create default playlist", "error", err, "user_id", user.ID)
		// エラーでもユーザー登録は成功させる（ログだけ残す）
	}

	// リフレッシュトークンを生成
	refreshToken, err := s.createRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Slack 新規登録通知（非同期で実行し、エラーは無視）
	if s.slackClient.IsRegistrationEnabled() {
		go func() {
			notification := slack.RegistrationNotification{
				UserID:      user.ID.String(),
				DisplayName: user.DisplayName,
				Email:       user.Email,
				Method:      "email",
				CreatedAt:   user.CreatedAt,
			}
			if err := s.slackClient.SendRegistration(context.Background(), notification); err != nil {
				logger.Default().Warn("Slack新規登録通知の送信に失敗しました", "error", err)
			}
		}()
	}

	return &AuthResult{
		User:         *s.toUserResponse(ctx, user),
		RefreshToken: refreshToken,
	}, nil
}

// Login はメールアドレスとパスワードで認証する
func (s *authService) Login(ctx context.Context, req request.LoginRequest) (*AuthResult, error) {
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

	// リフレッシュトークンを生成
	refreshToken, err := s.createRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         *s.toUserResponse(ctx, user),
		RefreshToken: refreshToken,
	}, nil
}

// OAuthGoogle は Google OAuth で認証する
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

		// リフレッシュトークンを生成
		refreshToken, err := s.createRefreshToken(ctx, existingAccount.UserID)
		if err != nil {
			return nil, err
		}

		return &AuthResult{
			User:         *s.toUserResponse(ctx, user),
			RefreshToken: refreshToken,
			IsCreated:    false,
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

	// デフォルト再生リストを作成
	if err := s.createDefaultPlaylist(ctx, user.ID); err != nil {
		logger.FromContext(ctx).Error("failed to create default playlist", "error", err, "user_id", user.ID)
		// エラーでもユーザー登録は成功させる（ログだけ残す）
	}

	// リフレッシュトークンを生成
	refreshToken, err := s.createRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Slack 新規登録通知（非同期で実行し、エラーは無視）
	if s.slackClient.IsRegistrationEnabled() {
		go func() {
			notification := slack.RegistrationNotification{
				UserID:      user.ID.String(),
				DisplayName: user.DisplayName,
				Email:       user.Email,
				Method:      "google",
				CreatedAt:   user.CreatedAt,
			}
			if err := s.slackClient.SendRegistration(context.Background(), notification); err != nil {
				logger.Default().Warn("Slack新規登録通知の送信に失敗しました", "error", err)
			}
		}()
	}

	return &AuthResult{
		User:         *s.toUserResponse(ctx, user),
		RefreshToken: refreshToken,
		IsCreated:    true,
	}, nil
}

// RefreshToken はリフレッシュトークンを使って新しいトークンを発行する
func (s *authService) RefreshToken(ctx context.Context, req request.RefreshTokenRequest) (*RefreshResult, error) {
	// リフレッシュトークンを検索
	existing, err := s.refreshTokenRepo.FindByToken(ctx, req.RefreshToken)
	if err != nil {
		if apperror.IsCode(err, apperror.CodeNotFound) {
			return nil, apperror.ErrInvalidRefreshToken.WithMessage("リフレッシュトークンが無効です")
		}
		return nil, err
	}

	// 有効期限をチェック
	if time.Now().After(existing.ExpiresAt) {
		// 期限切れのトークンを削除（エラーはログのみ）
		if delErr := s.refreshTokenRepo.DeleteByToken(ctx, req.RefreshToken); delErr != nil {
			logger.FromContext(ctx).Error("failed to delete expired refresh token", "error", delErr)
		}
		return nil, apperror.ErrInvalidRefreshToken.WithMessage("リフレッシュトークンの有効期限が切れています")
	}

	// トークンローテーション: 古いトークンを削除して新しいトークンを生成
	if err := s.refreshTokenRepo.DeleteByToken(ctx, req.RefreshToken); err != nil {
		return nil, err
	}

	newToken, err := s.createRefreshToken(ctx, existing.UserID)
	if err != nil {
		return nil, err
	}

	return &RefreshResult{
		UserID:       existing.UserID,
		RefreshToken: newToken,
	}, nil
}

// Logout はリフレッシュトークンを無効化する
func (s *authService) Logout(ctx context.Context, userID string, req request.LogoutRequest) error {
	// リフレッシュトークンを検索して所有者を確認
	existing, err := s.refreshTokenRepo.FindByToken(ctx, req.RefreshToken)
	if err != nil {
		if apperror.IsCode(err, apperror.CodeNotFound) {
			return apperror.ErrInvalidRefreshToken.WithMessage("リフレッシュトークンが無効です")
		}
		return err
	}

	// トークンの所有者を確認
	if existing.UserID.String() != userID {
		return apperror.ErrInvalidRefreshToken.WithMessage("リフレッシュトークンが無効です")
	}

	return s.refreshTokenRepo.DeleteByToken(ctx, req.RefreshToken)
}

// ChangePassword は認証済みユーザーのパスワードを更新する
func (s *authService) ChangePassword(ctx context.Context, userID string, req request.ChangePasswordRequest) error {
	// UUID をパース
	id, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	// 認証情報を取得
	credential, err := s.credentialRepo.FindByUserID(ctx, id)
	if err != nil {
		if apperror.IsCode(err, apperror.CodeNotFound) {
			return apperror.ErrInvalidCredentials.WithMessage("現在のパスワードが正しくありません")
		}
		return err
	}

	// 現在のパスワードを検証
	if err := s.passwordHasher.Compare(credential.PasswordHash, req.CurrentPassword); err != nil {
		return apperror.ErrInvalidCredentials.WithMessage("現在のパスワードが正しくありません")
	}

	// 新しいパスワードをハッシュ化
	hashedPassword, err := s.passwordHasher.Hash(req.NewPassword)
	if err != nil {
		logger.FromContext(ctx).Error("failed to hash password", "error", err)
		return apperror.ErrInternal.WithMessage("パスワードのハッシュ化に失敗しました").WithError(err)
	}

	// パスワードを更新
	credential.PasswordHash = hashedPassword
	return s.credentialRepo.Update(ctx, credential)
}

// createRefreshToken はリフレッシュトークンを生成して DB に保存する
func (s *authService) createRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	tokenStr, err := token.Generate()
	if err != nil {
		logger.FromContext(ctx).Error("failed to generate refresh token", "error", err)
		return "", apperror.ErrInternal.WithMessage("リフレッシュトークンの生成に失敗しました").WithError(err)
	}

	refreshToken := &model.RefreshToken{
		UserID:    userID,
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(refreshTokenExpiration),
	}
	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return "", err
	}

	return tokenStr, nil
}

// multiSpaceRegex は連続する半角スペースにマッチする正規表現
var multiSpaceRegex = regexp.MustCompile(`\s+`)

// displayNameToUsername は displayName をユーザー名形式に変換する
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

// appendRandomSuffix はユーザー名にランダムなサフィックスを付与する
func appendRandomSuffix(username string) string {
	suffix := rand.IntN(10000)
	return fmt.Sprintf("%s_%d", username, suffix)
}

// generateUniqueUsername は displayName からユニークなユーザー名を生成する
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
	return "", apperror.ErrInternal.WithMessage("ユニークなユーザー名の生成に失敗しました")
}

// toUserResponse は model.User を response.UserResponse に変換する
func (s *authService) toUserResponse(ctx context.Context, user *model.User) *response.UserResponse {
	var avatarURL *string
	if user.AvatarID != nil {
		image, err := s.imageRepo.FindByID(ctx, *user.AvatarID)
		if err == nil {
			if storage.IsExternalURL(image.Path) {
				avatarURL = &image.Path
			} else {
				signedURL, err := s.storageClient.GenerateSignedURL(ctx, image.Path, storage.SignedURLExpirationImage)
				if err == nil {
					avatarURL = &signedURL
				}
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

// GetMe は現在のユーザー情報を取得する
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
	avatar := s.resolveImageResponse(ctx, user.AvatarID)

	// ヘッダー画像を取得
	headerImage := s.resolveImageResponse(ctx, user.HeaderImageID)

	return &response.MeResponse{
		ID:             user.ID,
		Email:          user.Email,
		Username:       user.Username,
		DisplayName:    user.DisplayName,
		Bio:            user.Bio,
		Role:           string(user.Role),
		Avatar:         avatar,
		HeaderImage:    headerImage,
		UserPrompt:     user.UserPrompt,
		HasPassword:    hasPassword,
		OAuthProviders: providers,
		CreatedAt:      user.CreatedAt,
	}, nil
}

// UpdateMe はユーザーのプロフィール情報を更新する
func (s *authService) UpdateMe(ctx context.Context, userID string, req request.UpdateMeRequest) (*response.MeResponse, error) {
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

	// displayName, bio を更新
	user.DisplayName = req.DisplayName
	user.Bio = req.Bio

	// avatarImageId の処理
	if req.AvatarImageID.IsSet {
		if req.AvatarImageID.Value == nil {
			user.AvatarID = nil
		} else {
			avatarID, err := uuid.Parse(*req.AvatarImageID.Value)
			if err != nil {
				return nil, apperror.ErrValidation.WithMessage("avatarImageId は有効な UUID である必要があります")
			}
			if _, err := s.imageRepo.FindByID(ctx, avatarID); err != nil {
				return nil, err
			}
			user.AvatarID = &avatarID
		}
		user.Avatar = nil
	}

	// headerImageId の処理
	if req.HeaderImageID.IsSet {
		if req.HeaderImageID.Value == nil {
			user.HeaderImageID = nil
		} else {
			headerID, err := uuid.Parse(*req.HeaderImageID.Value)
			if err != nil {
				return nil, apperror.ErrValidation.WithMessage("headerImageId は有効な UUID である必要があります")
			}
			if _, err := s.imageRepo.FindByID(ctx, headerID); err != nil {
				return nil, err
			}
			user.HeaderImageID = &headerID
		}
		user.HeaderImage = nil
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// 更新後のユーザー情報を返す
	return s.GetMe(ctx, userID)
}

// resolveImageResponse は画像 ID から AvatarResponse を生成する
func (s *authService) resolveImageResponse(ctx context.Context, imageID *uuid.UUID) *response.AvatarResponse {
	if imageID == nil {
		return nil
	}

	image, err := s.imageRepo.FindByID(ctx, *imageID)
	if err != nil {
		return nil
	}

	if storage.IsExternalURL(image.Path) {
		return &response.AvatarResponse{
			ID:  image.ID,
			URL: image.Path,
		}
	}

	signedURL, err := s.storageClient.GenerateSignedURL(ctx, image.Path, storage.SignedURLExpirationImage)
	if err != nil {
		return nil
	}

	return &response.AvatarResponse{
		ID:  image.ID,
		URL: signedURL,
	}
}

// UpdatePrompt はユーザーの台本生成用プロンプトを更新する
func (s *authService) UpdatePrompt(ctx context.Context, userID string, req request.UpdateUserPromptRequest) (*response.MeResponse, error) {
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

	// プロンプトを更新（null の場合は空文字列）
	if req.UserPrompt != nil {
		user.UserPrompt = *req.UserPrompt
	} else {
		user.UserPrompt = ""
	}
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// 更新後のユーザー情報を返す
	return s.GetMe(ctx, userID)
}

// usernameRegex はユーザー名に使用可能な文字パターン（英数字・アンダースコア・日本語）
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\p{Hiragana}\p{Katakana}\p{Han}]+$`)

// validateUsernameFormat はユーザー名の形式を検証する
func validateUsernameFormat(username string) error {
	if !usernameRegex.MatchString(username) {
		return apperror.ErrValidation.WithMessage("ユーザー名に使用できない文字が含まれています")
	}
	if strings.HasPrefix(username, "__") {
		return apperror.ErrValidation.WithMessage("ユーザー名を __ で始めることはできません")
	}
	return nil
}

// UpdateUsername はユーザー名を変更する
func (s *authService) UpdateUsername(ctx context.Context, userID string, req request.UpdateUsernameRequest) (*response.MeResponse, error) {
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

	// ユーザー名の形式を検証
	if err := validateUsernameFormat(req.Username); err != nil {
		return nil, err
	}

	// 現在のユーザー名と同じなら変更せず返す
	if user.Username == req.Username {
		return s.GetMe(ctx, userID)
	}

	// 重複チェック
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.ErrDuplicateUsername.WithMessage("このユーザー名は既に使用されています")
	}

	// ユーザー名を更新
	user.Username = req.Username
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return s.GetMe(ctx, userID)
}

// CheckUsernameAvailability はユーザー名の利用可否をチェックする
func (s *authService) CheckUsernameAvailability(ctx context.Context, userID string, req request.CheckUsernameRequest) (*response.UsernameCheckResponse, error) {
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

	// ユーザー名の形式を検証
	if err := validateUsernameFormat(req.Username); err != nil {
		return nil, err
	}

	// 自分の現在のユーザー名なら利用可能
	if user.Username == req.Username {
		return &response.UsernameCheckResponse{
			Username:  req.Username,
			Available: true,
		}, nil
	}

	// 重複チェック
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	return &response.UsernameCheckResponse{
		Username:  req.Username,
		Available: !exists,
	}, nil
}

// DeleteMe はユーザーアカウントを削除する
func (s *authService) DeleteMe(ctx context.Context, userID string) error {
	// UUID をパース
	id, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	// ユーザーの存在確認
	if _, err := s.userRepo.FindByID(ctx, id); err != nil {
		return err
	}

	// 実行中ジョブをキャンセル
	if err := s.audioJobRepo.CancelActiveByUserID(ctx, id); err != nil {
		return err
	}
	if err := s.scriptJobRepo.CancelActiveByUserID(ctx, id); err != nil {
		return err
	}

	// ユーザーを削除（ON DELETE CASCADE で関連データも削除される）
	return s.userRepo.Delete(ctx, id)
}

// createDefaultPlaylist はユーザーのデフォルト再生リストを作成する
func (s *authService) createDefaultPlaylist(ctx context.Context, userID uuid.UUID) error {
	playlist := &model.Playlist{
		UserID:      userID,
		Name:        DefaultPlaylistName,
		Description: "",
		IsDefault:   true,
	}

	return s.playlistRepo.Create(ctx, playlist)
}
