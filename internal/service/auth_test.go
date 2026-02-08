package service

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
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

// --- テストヘルパー ---

// newAuthServiceForTest はテスト用の authService を組み立てるヘルパー
func newAuthServiceForTest(
	userRepo *mockUserRepositoryForAuth,
	credentialRepo *mockCredentialRepository,
	oauthAccountRepo *mockOAuthAccountRepository,
	imageRepo *mockImageRepositoryForAuth,
	storageClient *mockStorageClientForAuth,
) *authService {
	return &authService{
		userRepo:         userRepo,
		credentialRepo:   credentialRepo,
		oauthAccountRepo: oauthAccountRepo,
		imageRepo:        imageRepo,
		storageClient:    storageClient,
	}
}

// newAuthServiceForTestWithJobs はジョブ Repository 付きのテスト用 authService を組み立てるヘルパー
func newAuthServiceForTestWithJobs(
	userRepo *mockUserRepositoryForAuth,
	audioJobRepo *mockAudioJobRepositoryForAuth,
	scriptJobRepo *mockScriptJobRepositoryForAuth,
) *authService {
	return &authService{
		userRepo:      userRepo,
		audioJobRepo:  audioJobRepo,
		scriptJobRepo: scriptJobRepo,
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
			AvatarImageID: &avatarIDStr,
		}

		result, err := svc.UpdateMe(ctx, userID.String(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result.Avatar)
		userRepo.AssertExpectations(t)
		imageRepo.AssertExpectations(t)
	})

	t.Run("avatarImageId を空文字でクリアできる", func(t *testing.T) {
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

		emptyStr := ""
		req := request.UpdateMeRequest{
			DisplayName:   "Test User",
			AvatarImageID: &emptyStr,
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
			HeaderImageID: &headerIDStr,
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
			AvatarImageID: &avatarIDStr,
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

		// AvatarImageID を nil にして送信
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
