package service

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/infrastructure/slack"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/optional"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// --- モック定義 ---

type mockUserRepositoryForAuth struct {
	mock.Mock
}

func (m *mockUserRepositoryForAuth) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepositoryForAuth) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepositoryForAuth) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepositoryForAuth) FindByIDWithAvatar(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepositoryForAuth) FindByUsernameWithAvatar(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepositoryForAuth) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepositoryForAuth) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *mockUserRepositoryForAuth) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *mockUserRepositoryForAuth) Search(ctx context.Context, filter repository.SearchUserFilter) ([]model.User, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.User), args.Get(1).(int64), args.Error(2)
}

func (m *mockUserRepositoryForAuth) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockCredentialRepository struct {
	mock.Mock
}

func (m *mockCredentialRepository) Create(ctx context.Context, credential *model.Credential) error {
	args := m.Called(ctx, credential)
	return args.Error(0)
}

func (m *mockCredentialRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*model.Credential, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Credential), args.Error(1)
}

func (m *mockCredentialRepository) Update(ctx context.Context, credential *model.Credential) error {
	args := m.Called(ctx, credential)
	return args.Error(0)
}

type mockPasswordHasher struct {
	mock.Mock
}

func (m *mockPasswordHasher) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *mockPasswordHasher) Compare(hashedPassword, password string) error {
	args := m.Called(hashedPassword, password)
	return args.Error(0)
}

type mockOAuthAccountRepository struct {
	mock.Mock
}

func (m *mockOAuthAccountRepository) Create(ctx context.Context, account *model.OAuthAccount) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *mockOAuthAccountRepository) Update(ctx context.Context, account *model.OAuthAccount) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *mockOAuthAccountRepository) FindByProviderAndProviderUserID(ctx context.Context, provider model.OAuthProvider, providerUserID string) (*model.OAuthAccount, error) {
	args := m.Called(ctx, provider, providerUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OAuthAccount), args.Error(1)
}

func (m *mockOAuthAccountRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.OAuthAccount, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.OAuthAccount), args.Error(1)
}

type mockImageRepositoryForAuth struct {
	mock.Mock
}

func (m *mockImageRepositoryForAuth) Create(ctx context.Context, image *model.Image) error {
	args := m.Called(ctx, image)
	return args.Error(0)
}

func (m *mockImageRepositoryForAuth) FindByID(ctx context.Context, id uuid.UUID) (*model.Image, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Image), args.Error(1)
}

func (m *mockImageRepositoryForAuth) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockImageRepositoryForAuth) FindOrphaned(ctx context.Context) ([]model.Image, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Image), args.Error(1)
}

type mockAudioJobRepositoryForAuth struct {
	mock.Mock
}

func (m *mockAudioJobRepositoryForAuth) FindByID(ctx context.Context, id uuid.UUID) (*model.AudioJob, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AudioJob), args.Error(1)
}

func (m *mockAudioJobRepositoryForAuth) FindByUserID(ctx context.Context, userID uuid.UUID, filter repository.AudioJobFilter) ([]model.AudioJob, error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).([]model.AudioJob), args.Error(1)
}

func (m *mockAudioJobRepositoryForAuth) FindByEpisodeID(ctx context.Context, episodeID uuid.UUID) ([]model.AudioJob, error) {
	args := m.Called(ctx, episodeID)
	return args.Get(0).([]model.AudioJob), args.Error(1)
}

func (m *mockAudioJobRepositoryForAuth) FindPendingByEpisodeID(ctx context.Context, episodeID uuid.UUID) (*model.AudioJob, error) {
	args := m.Called(ctx, episodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AudioJob), args.Error(1)
}

func (m *mockAudioJobRepositoryForAuth) Create(ctx context.Context, job *model.AudioJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *mockAudioJobRepositoryForAuth) Update(ctx context.Context, job *model.AudioJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *mockAudioJobRepositoryForAuth) UpdateProgress(ctx context.Context, id uuid.UUID, progress int) error {
	args := m.Called(ctx, id, progress)
	return args.Error(0)
}

func (m *mockAudioJobRepositoryForAuth) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockAudioJobRepositoryForAuth) CancelActiveByUserID(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type mockScriptJobRepositoryForAuth struct {
	mock.Mock
}

func (m *mockScriptJobRepositoryForAuth) FindByID(ctx context.Context, id uuid.UUID) (*model.ScriptJob, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ScriptJob), args.Error(1)
}

func (m *mockScriptJobRepositoryForAuth) FindByUserID(ctx context.Context, userID uuid.UUID, filter repository.ScriptJobFilter) ([]model.ScriptJob, error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).([]model.ScriptJob), args.Error(1)
}

func (m *mockScriptJobRepositoryForAuth) FindByEpisodeID(ctx context.Context, episodeID uuid.UUID) ([]model.ScriptJob, error) {
	args := m.Called(ctx, episodeID)
	return args.Get(0).([]model.ScriptJob), args.Error(1)
}

func (m *mockScriptJobRepositoryForAuth) FindPendingByEpisodeID(ctx context.Context, episodeID uuid.UUID) (*model.ScriptJob, error) {
	args := m.Called(ctx, episodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ScriptJob), args.Error(1)
}

func (m *mockScriptJobRepositoryForAuth) FindLatestCompletedByEpisodeID(ctx context.Context, episodeID uuid.UUID) (*model.ScriptJob, error) {
	args := m.Called(ctx, episodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ScriptJob), args.Error(1)
}

func (m *mockScriptJobRepositoryForAuth) Create(ctx context.Context, job *model.ScriptJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *mockScriptJobRepositoryForAuth) Update(ctx context.Context, job *model.ScriptJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *mockScriptJobRepositoryForAuth) UpdateProgress(ctx context.Context, id uuid.UUID, progress int) error {
	args := m.Called(ctx, id, progress)
	return args.Error(0)
}

func (m *mockScriptJobRepositoryForAuth) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockScriptJobRepositoryForAuth) CancelActiveByUserID(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type mockStorageClientForAuth struct {
	mock.Mock
}

func (m *mockStorageClientForAuth) Upload(ctx context.Context, data []byte, path, contentType string) (string, error) {
	args := m.Called(ctx, data, path, contentType)
	return args.String(0), args.Error(1)
}

func (m *mockStorageClientForAuth) GenerateSignedURL(ctx context.Context, path string, expiration time.Duration) (string, error) {
	args := m.Called(ctx, path, expiration)
	return args.String(0), args.Error(1)
}

func (m *mockStorageClientForAuth) Delete(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

type mockSlackClientForAuth struct {
	mock.Mock
}

func (m *mockSlackClientForAuth) SendFeedback(ctx context.Context, feedback slack.FeedbackNotification) error {
	args := m.Called(ctx, feedback)
	return args.Error(0)
}

func (m *mockSlackClientForAuth) SendContact(ctx context.Context, contact slack.ContactNotification) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}

func (m *mockSlackClientForAuth) SendAlert(ctx context.Context, alert slack.AlertNotification) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *mockSlackClientForAuth) SendRegistration(ctx context.Context, registration slack.RegistrationNotification) error {
	args := m.Called(ctx, registration)
	return args.Error(0)
}

func (m *mockSlackClientForAuth) IsFeedbackEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockSlackClientForAuth) IsContactEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockSlackClientForAuth) IsAlertEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockSlackClientForAuth) IsRegistrationEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

// --- テストヘルパー ---

// newAuthServiceForTest はテスト用の authService を組み立てるヘルパー
func newAuthServiceForTest(
	userRepo *mockUserRepositoryForAuth,
	credentialRepo *mockCredentialRepository,
	oauthAccountRepo *mockOAuthAccountRepository,
	imageRepo *mockImageRepositoryForAuth,
	storageClient *mockStorageClientForAuth,
) *authService {
	sc := new(mockSlackClientForAuth)
	sc.On("IsRegistrationEnabled").Return(false)
	return &authService{
		userRepo:         userRepo,
		credentialRepo:   credentialRepo,
		oauthAccountRepo: oauthAccountRepo,
		imageRepo:        imageRepo,
		storageClient:    storageClient,
		slackClient:      sc,
	}
}

// newAuthServiceForTestWithJobs はジョブ Repository 付きのテスト用 authService を組み立てるヘルパー
func newAuthServiceForTestWithJobs(
	userRepo *mockUserRepositoryForAuth,
	audioJobRepo *mockAudioJobRepositoryForAuth,
	scriptJobRepo *mockScriptJobRepositoryForAuth,
) *authService {
	sc := new(mockSlackClientForAuth)
	sc.On("IsRegistrationEnabled").Return(false)
	return &authService{
		userRepo:      userRepo,
		audioJobRepo:  audioJobRepo,
		scriptJobRepo: scriptJobRepo,
		slackClient:   sc,
	}
}

// newAuthServiceForTestWithPassword はパスワードハッシャー付きのテスト用 authService を組み立てるヘルパー
func newAuthServiceForTestWithPassword(
	credentialRepo *mockCredentialRepository,
	passwordHasher *mockPasswordHasher,
) *authService {
	sc := new(mockSlackClientForAuth)
	sc.On("IsRegistrationEnabled").Return(false)
	return &authService{
		credentialRepo: credentialRepo,
		passwordHasher: passwordHasher,
		slackClient:    sc,
	}
}

// --- テスト ---

func TestDisplayNameToUsername(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
		want        string
	}{
		{
			name:        "半角スペースをアンダースコアに変換",
			displayName: "John Doe",
			want:        "John_Doe",
		},
		{
			name:        "全角スペースをアンダースコアに変換",
			displayName: "田中　太郎",
			want:        "田中_太郎",
		},
		{
			name:        "連続する半角スペースを1つのアンダースコアに圧縮",
			displayName: "John  Doe",
			want:        "John_Doe",
		},
		{
			name:        "連続する全角スペースを1つのアンダースコアに圧縮",
			displayName: "田中　　太郎",
			want:        "田中_太郎",
		},
		{
			name:        "混合スペースを1つのアンダースコアに圧縮",
			displayName: "John 　Doe",
			want:        "John_Doe",
		},
		{
			name:        "前後の空白を削除",
			displayName: "  John Doe  ",
			want:        "John_Doe",
		},
		{
			name:        "スペースがない場合はそのまま",
			displayName: "JohnDoe",
			want:        "JohnDoe",
		},
		{
			name:        "複数の単語",
			displayName: "John Middle Doe",
			want:        "John_Middle_Doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := displayNameToUsername(tt.displayName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAppendRandomSuffix(t *testing.T) {
	t.Run("ユーザー名にランダムなサフィックスを付与する", func(t *testing.T) {
		username := "testuser"
		result := appendRandomSuffix(username)

		// 形式が "testuser_数字" であることを確認
		pattern := regexp.MustCompile(`^testuser_\d+$`)
		assert.True(t, pattern.MatchString(result), "結果は 'testuser_数字' の形式であるべき: %s", result)
	})

	t.Run("サフィックスは0から9999の範囲", func(t *testing.T) {
		username := "user"
		pattern := regexp.MustCompile(`^user_(\d{1,4})$`)

		// 複数回実行して形式を確認
		for i := 0; i < 100; i++ {
			result := appendRandomSuffix(username)
			assert.True(t, pattern.MatchString(result), "サフィックスは1〜4桁の数字であるべき: %s", result)
		}
	})
}

func TestUpdateMe(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	avatarImageID := uuid.New()
	headerImageID := uuid.New()

	baseUser := &model.User{
		ID:          userID,
		Email:       "test@example.com",
		Username:    "testuser",
		DisplayName: "Test User",
		Bio:         "",
		Role:        model.RoleUser,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	t.Run("displayName と bio を更新できる", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		credentialRepo := new(mockCredentialRepository)
		oauthAccountRepo := new(mockOAuthAccountRepository)
		imageRepo := new(mockImageRepositoryForAuth)
		storageClient := new(mockStorageClientForAuth)
		svc := newAuthServiceForTest(userRepo, credentialRepo, oauthAccountRepo, imageRepo, storageClient)

		user := *baseUser
		userRepo.On("FindByID", ctx, userID).Return(&user, nil).Twice()
		userRepo.On("Update", ctx, mock.Anything).Return(nil)
		credentialRepo.On("FindByUserID", ctx, userID).Return(nil, apperror.ErrNotFound)
		oauthAccountRepo.On("FindByUserID", ctx, userID).Return([]model.OAuthAccount{}, nil)

		req := request.UpdateMeRequest{
			DisplayName: "新しい名前",
			Bio:         "よろしくお願いします",
		}

		result, err := svc.UpdateMe(ctx, userID.String(), req)

		assert.NoError(t, err)
		assert.Equal(t, "新しい名前", result.DisplayName)
		assert.Equal(t, "よろしくお願いします", result.Bio)
		userRepo.AssertExpectations(t)
	})

	t.Run("avatarImageId で画像を設定できる", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		credentialRepo := new(mockCredentialRepository)
		oauthAccountRepo := new(mockOAuthAccountRepository)
		imageRepo := new(mockImageRepositoryForAuth)
		storageClient := new(mockStorageClientForAuth)
		svc := newAuthServiceForTest(userRepo, credentialRepo, oauthAccountRepo, imageRepo, storageClient)

		user := *baseUser
		userRepo.On("FindByID", ctx, userID).Return(&user, nil).Twice()
		userRepo.On("Update", ctx, mock.MatchedBy(func(u *model.User) bool {
			return u.AvatarID != nil && *u.AvatarID == avatarImageID
		})).Return(nil)
		imageRepo.On("FindByID", ctx, avatarImageID).Return(&model.Image{
			ID:   avatarImageID,
			Path: "images/avatar.png",
		}, nil)
		storageClient.On("GenerateSignedURL", ctx, "images/avatar.png", storage.SignedURLExpirationImage).Return("https://signed.example.com/avatar.png", nil)
		credentialRepo.On("FindByUserID", ctx, userID).Return(nil, apperror.ErrNotFound)
		oauthAccountRepo.On("FindByUserID", ctx, userID).Return([]model.OAuthAccount{}, nil)

		avatarIDStr := avatarImageID.String()
		req := request.UpdateMeRequest{
			DisplayName:   "Test User",
			AvatarImageID: optional.Field[string]{Value: &avatarIDStr, IsSet: true},
		}

		result, err := svc.UpdateMe(ctx, userID.String(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result.Avatar)
		userRepo.AssertExpectations(t)
		imageRepo.AssertExpectations(t)
	})

	t.Run("avatarImageId を null でクリアできる", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		credentialRepo := new(mockCredentialRepository)
		oauthAccountRepo := new(mockOAuthAccountRepository)
		imageRepo := new(mockImageRepositoryForAuth)
		storageClient := new(mockStorageClientForAuth)
		svc := newAuthServiceForTest(userRepo, credentialRepo, oauthAccountRepo, imageRepo, storageClient)

		user := *baseUser
		user.AvatarID = &avatarImageID
		userRepo.On("FindByID", ctx, userID).Return(&user, nil).Twice()
		userRepo.On("Update", ctx, mock.MatchedBy(func(u *model.User) bool {
			return u.AvatarID == nil
		})).Return(nil)
		credentialRepo.On("FindByUserID", ctx, userID).Return(nil, apperror.ErrNotFound)
		oauthAccountRepo.On("FindByUserID", ctx, userID).Return([]model.OAuthAccount{}, nil)

		req := request.UpdateMeRequest{
			DisplayName:   "Test User",
			AvatarImageID: optional.Field[string]{Value: nil, IsSet: true},
		}

		result, err := svc.UpdateMe(ctx, userID.String(), req)

		assert.NoError(t, err)
		assert.Nil(t, result.Avatar)
		userRepo.AssertExpectations(t)
	})

	t.Run("headerImageId で画像を設定できる", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		credentialRepo := new(mockCredentialRepository)
		oauthAccountRepo := new(mockOAuthAccountRepository)
		imageRepo := new(mockImageRepositoryForAuth)
		storageClient := new(mockStorageClientForAuth)
		svc := newAuthServiceForTest(userRepo, credentialRepo, oauthAccountRepo, imageRepo, storageClient)

		user := *baseUser
		userRepo.On("FindByID", ctx, userID).Return(&user, nil).Twice()
		userRepo.On("Update", ctx, mock.MatchedBy(func(u *model.User) bool {
			return u.HeaderImageID != nil && *u.HeaderImageID == headerImageID
		})).Return(nil)
		imageRepo.On("FindByID", ctx, headerImageID).Return(&model.Image{
			ID:   headerImageID,
			Path: "images/header.png",
		}, nil)
		storageClient.On("GenerateSignedURL", ctx, "images/header.png", storage.SignedURLExpirationImage).Return("https://signed.example.com/header.png", nil)
		credentialRepo.On("FindByUserID", ctx, userID).Return(nil, apperror.ErrNotFound)
		oauthAccountRepo.On("FindByUserID", ctx, userID).Return([]model.OAuthAccount{}, nil)

		headerIDStr := headerImageID.String()
		req := request.UpdateMeRequest{
			DisplayName:   "Test User",
			HeaderImageID: optional.Field[string]{Value: &headerIDStr, IsSet: true},
		}

		result, err := svc.UpdateMe(ctx, userID.String(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result.HeaderImage)
		userRepo.AssertExpectations(t)
		imageRepo.AssertExpectations(t)
	})

	t.Run("存在しない画像 ID を指定するとエラーになる", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		credentialRepo := new(mockCredentialRepository)
		oauthAccountRepo := new(mockOAuthAccountRepository)
		imageRepo := new(mockImageRepositoryForAuth)
		storageClient := new(mockStorageClientForAuth)
		svc := newAuthServiceForTest(userRepo, credentialRepo, oauthAccountRepo, imageRepo, storageClient)

		user := *baseUser
		userRepo.On("FindByID", ctx, userID).Return(&user, nil)
		imageRepo.On("FindByID", ctx, avatarImageID).Return(nil, apperror.ErrNotFound.WithMessage("画像が見つかりません"))

		avatarIDStr := avatarImageID.String()
		req := request.UpdateMeRequest{
			DisplayName:   "Test User",
			AvatarImageID: optional.Field[string]{Value: &avatarIDStr, IsSet: true},
		}

		result, err := svc.UpdateMe(ctx, userID.String(), req)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeNotFound))
	})

	t.Run("avatarImageId が nil の場合は変更しない", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		credentialRepo := new(mockCredentialRepository)
		oauthAccountRepo := new(mockOAuthAccountRepository)
		imageRepo := new(mockImageRepositoryForAuth)
		storageClient := new(mockStorageClientForAuth)
		svc := newAuthServiceForTest(userRepo, credentialRepo, oauthAccountRepo, imageRepo, storageClient)

		user := *baseUser
		user.AvatarID = &avatarImageID
		userRepo.On("FindByID", ctx, userID).Return(&user, nil).Twice()
		userRepo.On("Update", ctx, mock.MatchedBy(func(u *model.User) bool {
			return u.AvatarID != nil && *u.AvatarID == avatarImageID
		})).Return(nil)
		imageRepo.On("FindByID", ctx, avatarImageID).Return(&model.Image{
			ID:   avatarImageID,
			Path: "images/avatar.png",
		}, nil)
		storageClient.On("GenerateSignedURL", ctx, "images/avatar.png", storage.SignedURLExpirationImage).Return("https://signed.example.com/avatar.png", nil)
		credentialRepo.On("FindByUserID", ctx, userID).Return(nil, apperror.ErrNotFound)
		oauthAccountRepo.On("FindByUserID", ctx, userID).Return([]model.OAuthAccount{}, nil)

		// AvatarImageID を未送信（IsSet: false）にして送信
		req := request.UpdateMeRequest{
			DisplayName: "Test User",
		}

		result, err := svc.UpdateMe(ctx, userID.String(), req)

		assert.NoError(t, err)
		// アバターは変更されていないのでそのまま残る
		assert.NotNil(t, result.Avatar)
		assert.Equal(t, avatarImageID, result.Avatar.ID)
		userRepo.AssertExpectations(t)
	})
}

func TestDeleteMe(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("アカウント削除が成功する", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		audioJobRepo := new(mockAudioJobRepositoryForAuth)
		scriptJobRepo := new(mockScriptJobRepositoryForAuth)
		svc := newAuthServiceForTestWithJobs(userRepo, audioJobRepo, scriptJobRepo)

		userRepo.On("FindByID", ctx, userID).Return(&model.User{ID: userID}, nil)
		audioJobRepo.On("CancelActiveByUserID", ctx, userID).Return(nil)
		scriptJobRepo.On("CancelActiveByUserID", ctx, userID).Return(nil)
		userRepo.On("Delete", ctx, userID).Return(nil)

		err := svc.DeleteMe(ctx, userID.String())

		assert.NoError(t, err)
		userRepo.AssertExpectations(t)
		audioJobRepo.AssertExpectations(t)
		scriptJobRepo.AssertExpectations(t)
	})

	t.Run("ユーザーが見つからない場合はエラーを返す", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		audioJobRepo := new(mockAudioJobRepositoryForAuth)
		scriptJobRepo := new(mockScriptJobRepositoryForAuth)
		svc := newAuthServiceForTestWithJobs(userRepo, audioJobRepo, scriptJobRepo)

		userRepo.On("FindByID", ctx, userID).Return(nil, apperror.ErrNotFound.WithMessage("ユーザーが見つかりません"))

		err := svc.DeleteMe(ctx, userID.String())

		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeNotFound))
		userRepo.AssertExpectations(t)
	})

	t.Run("音声ジョブキャンセルが失敗した場合はエラーを返す", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		audioJobRepo := new(mockAudioJobRepositoryForAuth)
		scriptJobRepo := new(mockScriptJobRepositoryForAuth)
		svc := newAuthServiceForTestWithJobs(userRepo, audioJobRepo, scriptJobRepo)

		userRepo.On("FindByID", ctx, userID).Return(&model.User{ID: userID}, nil)
		audioJobRepo.On("CancelActiveByUserID", ctx, userID).Return(apperror.ErrInternal.WithMessage("キャンセルに失敗しました"))

		err := svc.DeleteMe(ctx, userID.String())

		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeInternal))
		userRepo.AssertExpectations(t)
		audioJobRepo.AssertExpectations(t)
	})

	t.Run("台本ジョブキャンセルが失敗した場合はエラーを返す", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		audioJobRepo := new(mockAudioJobRepositoryForAuth)
		scriptJobRepo := new(mockScriptJobRepositoryForAuth)
		svc := newAuthServiceForTestWithJobs(userRepo, audioJobRepo, scriptJobRepo)

		userRepo.On("FindByID", ctx, userID).Return(&model.User{ID: userID}, nil)
		audioJobRepo.On("CancelActiveByUserID", ctx, userID).Return(nil)
		scriptJobRepo.On("CancelActiveByUserID", ctx, userID).Return(apperror.ErrInternal.WithMessage("キャンセルに失敗しました"))

		err := svc.DeleteMe(ctx, userID.String())

		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeInternal))
		userRepo.AssertExpectations(t)
		audioJobRepo.AssertExpectations(t)
		scriptJobRepo.AssertExpectations(t)
	})

	t.Run("ユーザー削除が失敗した場合はエラーを返す", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		audioJobRepo := new(mockAudioJobRepositoryForAuth)
		scriptJobRepo := new(mockScriptJobRepositoryForAuth)
		svc := newAuthServiceForTestWithJobs(userRepo, audioJobRepo, scriptJobRepo)

		userRepo.On("FindByID", ctx, userID).Return(&model.User{ID: userID}, nil)
		audioJobRepo.On("CancelActiveByUserID", ctx, userID).Return(nil)
		scriptJobRepo.On("CancelActiveByUserID", ctx, userID).Return(nil)
		userRepo.On("Delete", ctx, userID).Return(apperror.ErrInternal.WithMessage("削除に失敗しました"))

		err := svc.DeleteMe(ctx, userID.String())

		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeInternal))
		userRepo.AssertExpectations(t)
		audioJobRepo.AssertExpectations(t)
		scriptJobRepo.AssertExpectations(t)
	})

	t.Run("無効な UUID の場合はエラーを返す", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		audioJobRepo := new(mockAudioJobRepositoryForAuth)
		scriptJobRepo := new(mockScriptJobRepositoryForAuth)
		svc := newAuthServiceForTestWithJobs(userRepo, audioJobRepo, scriptJobRepo)

		err := svc.DeleteMe(ctx, "invalid-uuid")

		assert.Error(t, err)
	})
}

func TestChangePassword(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("パスワード更新が成功する", func(t *testing.T) {
		credentialRepo := new(mockCredentialRepository)
		passwordHasher := new(mockPasswordHasher)
		svc := newAuthServiceForTestWithPassword(credentialRepo, passwordHasher)

		credential := &model.Credential{
			UserID:       userID,
			PasswordHash: "hashed_current_password",
		}
		credentialRepo.On("FindByUserID", ctx, userID).Return(credential, nil)
		passwordHasher.On("Compare", "hashed_current_password", "current_password").Return(nil)
		passwordHasher.On("Hash", "new_password123").Return("hashed_new_password", nil)
		credentialRepo.On("Update", ctx, mock.MatchedBy(func(c *model.Credential) bool {
			return c.PasswordHash == "hashed_new_password"
		})).Return(nil)

		err := svc.ChangePassword(ctx, userID.String(), request.ChangePasswordRequest{
			CurrentPassword: "current_password",
			NewPassword:     "new_password123",
		})

		assert.NoError(t, err)
		credentialRepo.AssertExpectations(t)
		passwordHasher.AssertExpectations(t)
	})

	t.Run("現在のパスワードが間違っている場合はエラーを返す", func(t *testing.T) {
		credentialRepo := new(mockCredentialRepository)
		passwordHasher := new(mockPasswordHasher)
		svc := newAuthServiceForTestWithPassword(credentialRepo, passwordHasher)

		credential := &model.Credential{
			UserID:       userID,
			PasswordHash: "hashed_current_password",
		}
		credentialRepo.On("FindByUserID", ctx, userID).Return(credential, nil)
		passwordHasher.On("Compare", "hashed_current_password", "wrong_password").Return(errors.New("mismatch"))

		err := svc.ChangePassword(ctx, userID.String(), request.ChangePasswordRequest{
			CurrentPassword: "wrong_password",
			NewPassword:     "new_password123",
		})

		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeInvalidCredentials))
		credentialRepo.AssertExpectations(t)
		passwordHasher.AssertExpectations(t)
	})

	t.Run("認証情報が存在しない場合はエラーを返す", func(t *testing.T) {
		credentialRepo := new(mockCredentialRepository)
		passwordHasher := new(mockPasswordHasher)
		svc := newAuthServiceForTestWithPassword(credentialRepo, passwordHasher)

		credentialRepo.On("FindByUserID", ctx, userID).Return(nil, apperror.ErrNotFound.WithMessage("認証情報が見つかりません"))

		err := svc.ChangePassword(ctx, userID.String(), request.ChangePasswordRequest{
			CurrentPassword: "current_password",
			NewPassword:     "new_password123",
		})

		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeInvalidCredentials))
		credentialRepo.AssertExpectations(t)
	})

	t.Run("無効な UUID の場合はエラーを返す", func(t *testing.T) {
		credentialRepo := new(mockCredentialRepository)
		passwordHasher := new(mockPasswordHasher)
		svc := newAuthServiceForTestWithPassword(credentialRepo, passwordHasher)

		err := svc.ChangePassword(ctx, "invalid-uuid", request.ChangePasswordRequest{
			CurrentPassword: "current_password",
			NewPassword:     "new_password123",
		})

		assert.Error(t, err)
	})
}

func TestValidateUsernameFormat(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{name: "英数字のみ", username: "user123", wantErr: false},
		{name: "アンダースコアを含む", username: "user_name", wantErr: false},
		{name: "日本語（ひらがな）", username: "たろう", wantErr: false},
		{name: "日本語（カタカナ）", username: "タロウ", wantErr: false},
		{name: "日本語（漢字）", username: "太郎", wantErr: false},
		{name: "混合（英数字と日本語）", username: "user太郎123", wantErr: false},
		{name: "アンダースコア1つで始まる", username: "_user", wantErr: false},
		{name: "__ で始まる", username: "__system", wantErr: true},
		{name: "ハイフンを含む", username: "user-name", wantErr: true},
		{name: "スペースを含む", username: "user name", wantErr: true},
		{name: "記号を含む", username: "user@name", wantErr: true},
		{name: "ドットを含む", username: "user.name", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUsernameFormat(tt.username)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateUsername(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	baseUser := &model.User{
		ID:          userID,
		Email:       "test@example.com",
		Username:    "current_user",
		DisplayName: "Test User",
		Role:        model.RoleUser,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	t.Run("ユーザー名変更が成功する", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		credentialRepo := new(mockCredentialRepository)
		oauthAccountRepo := new(mockOAuthAccountRepository)
		imageRepo := new(mockImageRepositoryForAuth)
		storageClient := new(mockStorageClientForAuth)
		svc := newAuthServiceForTest(userRepo, credentialRepo, oauthAccountRepo, imageRepo, storageClient)

		user := *baseUser
		userRepo.On("FindByID", ctx, userID).Return(&user, nil).Twice()
		userRepo.On("ExistsByUsername", ctx, "new_user").Return(false, nil)
		userRepo.On("Update", ctx, mock.MatchedBy(func(u *model.User) bool {
			return u.Username == "new_user"
		})).Return(nil)
		credentialRepo.On("FindByUserID", ctx, userID).Return(nil, apperror.ErrNotFound)
		oauthAccountRepo.On("FindByUserID", ctx, userID).Return([]model.OAuthAccount{}, nil)

		result, err := svc.UpdateUsername(ctx, userID.String(), request.UpdateUsernameRequest{
			Username: "new_user",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		userRepo.AssertExpectations(t)
	})

	t.Run("同じユーザー名なら変更せず返す", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		credentialRepo := new(mockCredentialRepository)
		oauthAccountRepo := new(mockOAuthAccountRepository)
		imageRepo := new(mockImageRepositoryForAuth)
		storageClient := new(mockStorageClientForAuth)
		svc := newAuthServiceForTest(userRepo, credentialRepo, oauthAccountRepo, imageRepo, storageClient)

		user := *baseUser
		userRepo.On("FindByID", ctx, userID).Return(&user, nil).Twice()
		credentialRepo.On("FindByUserID", ctx, userID).Return(nil, apperror.ErrNotFound)
		oauthAccountRepo.On("FindByUserID", ctx, userID).Return([]model.OAuthAccount{}, nil)

		result, err := svc.UpdateUsername(ctx, userID.String(), request.UpdateUsernameRequest{
			Username: "current_user",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "current_user", result.Username)
		userRepo.AssertExpectations(t)
	})

	t.Run("形式不正の場合はエラーを返す", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		svc := newAuthServiceForTest(userRepo, nil, nil, nil, nil)

		user := *baseUser
		userRepo.On("FindByID", ctx, userID).Return(&user, nil)

		result, err := svc.UpdateUsername(ctx, userID.String(), request.UpdateUsernameRequest{
			Username: "user-name",
		})

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeValidation))
	})

	t.Run("__ 始まりの場合はエラーを返す", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		svc := newAuthServiceForTest(userRepo, nil, nil, nil, nil)

		user := *baseUser
		userRepo.On("FindByID", ctx, userID).Return(&user, nil)

		result, err := svc.UpdateUsername(ctx, userID.String(), request.UpdateUsernameRequest{
			Username: "__reserved",
		})

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeValidation))
	})

	t.Run("重複する場合は 409 エラーを返す", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		svc := newAuthServiceForTest(userRepo, nil, nil, nil, nil)

		user := *baseUser
		userRepo.On("FindByID", ctx, userID).Return(&user, nil)
		userRepo.On("ExistsByUsername", ctx, "taken_user").Return(true, nil)

		result, err := svc.UpdateUsername(ctx, userID.String(), request.UpdateUsernameRequest{
			Username: "taken_user",
		})

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeDuplicateUsername))
		userRepo.AssertExpectations(t)
	})
}

func TestCheckUsernameAvailability(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	baseUser := &model.User{
		ID:          userID,
		Email:       "test@example.com",
		Username:    "current_user",
		DisplayName: "Test User",
		Role:        model.RoleUser,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	t.Run("利用可能なユーザー名の場合 available: true を返す", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		svc := newAuthServiceForTest(userRepo, nil, nil, nil, nil)

		user := *baseUser
		userRepo.On("FindByID", ctx, userID).Return(&user, nil)
		userRepo.On("ExistsByUsername", ctx, "available_name").Return(false, nil)

		result, err := svc.CheckUsernameAvailability(ctx, userID.String(), request.CheckUsernameRequest{
			Username: "available_name",
		})

		assert.NoError(t, err)
		assert.Equal(t, "available_name", result.Username)
		assert.True(t, result.Available)
		userRepo.AssertExpectations(t)
	})

	t.Run("既に使用されているユーザー名の場合 available: false を返す", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		svc := newAuthServiceForTest(userRepo, nil, nil, nil, nil)

		user := *baseUser
		userRepo.On("FindByID", ctx, userID).Return(&user, nil)
		userRepo.On("ExistsByUsername", ctx, "taken_name").Return(true, nil)

		result, err := svc.CheckUsernameAvailability(ctx, userID.String(), request.CheckUsernameRequest{
			Username: "taken_name",
		})

		assert.NoError(t, err)
		assert.Equal(t, "taken_name", result.Username)
		assert.False(t, result.Available)
		userRepo.AssertExpectations(t)
	})

	t.Run("自分の現在のユーザー名の場合 available: true を返す", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		svc := newAuthServiceForTest(userRepo, nil, nil, nil, nil)

		user := *baseUser
		userRepo.On("FindByID", ctx, userID).Return(&user, nil)

		result, err := svc.CheckUsernameAvailability(ctx, userID.String(), request.CheckUsernameRequest{
			Username: "current_user",
		})

		assert.NoError(t, err)
		assert.Equal(t, "current_user", result.Username)
		assert.True(t, result.Available)
		userRepo.AssertExpectations(t)
	})

	t.Run("形式不正の場合はエラーを返す", func(t *testing.T) {
		userRepo := new(mockUserRepositoryForAuth)
		svc := newAuthServiceForTest(userRepo, nil, nil, nil, nil)

		user := *baseUser
		userRepo.On("FindByID", ctx, userID).Return(&user, nil)

		result, err := svc.CheckUsernameAvailability(ctx, userID.String(), request.CheckUsernameRequest{
			Username: "invalid-name",
		})

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeValidation))
	})
}
