package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// モックリポジトリ

type mockChannelRepository struct {
	mock.Mock
}

func (m *mockChannelRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Channel, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Channel), args.Error(1)
}

func (m *mockChannelRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filter repository.ChannelFilter) ([]model.Channel, int64, error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).([]model.Channel), args.Get(1).(int64), args.Error(2)
}

func (m *mockChannelRepository) Create(ctx context.Context, channel *model.Channel) error {
	args := m.Called(ctx, channel)
	return args.Error(0)
}

func (m *mockChannelRepository) Update(ctx context.Context, channel *model.Channel) error {
	args := m.Called(ctx, channel)
	return args.Error(0)
}

func (m *mockChannelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockEpisodeRepository struct {
	mock.Mock
}

func (m *mockEpisodeRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Episode, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Episode), args.Error(1)
}

func (m *mockEpisodeRepository) FindByChannelID(ctx context.Context, channelID uuid.UUID, filter repository.EpisodeFilter) ([]model.Episode, int64, error) {
	args := m.Called(ctx, channelID, filter)
	return args.Get(0).([]model.Episode), args.Get(1).(int64), args.Error(2)
}

func (m *mockEpisodeRepository) Create(ctx context.Context, episode *model.Episode) error {
	args := m.Called(ctx, episode)
	return args.Error(0)
}

func (m *mockEpisodeRepository) Update(ctx context.Context, episode *model.Episode) error {
	args := m.Called(ctx, episode)
	return args.Error(0)
}

func (m *mockEpisodeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockLLMClient struct {
	mock.Mock
}

func (m *mockLLMClient) GenerateScript(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	args := m.Called(ctx, systemPrompt, userPrompt)
	return args.String(0), args.Error(1)
}

// テストケース

func TestBuildUserPrompt(t *testing.T) {
	svc := &scriptService{}

	t.Run("チャンネル設定とキャラクターが正しく含まれる", func(t *testing.T) {
		channel := &model.Channel{
			UserPrompt: "テック系ポッドキャスト",
			Characters: []model.Character{
				{Name: "太郎", Persona: "明るいホスト"},
				{Name: "花子", Persona: "知識豊富なゲスト"},
			},
		}

		result := svc.buildUserPrompt(channel, "AI について語る", 10)

		assert.Contains(t, result, "## チャンネル設定")
		assert.Contains(t, result, "テック系ポッドキャスト")
		assert.Contains(t, result, "## 登場人物")
		assert.Contains(t, result, "太郎: 明るいホスト")
		assert.Contains(t, result, "花子: 知識豊富なゲスト")
		assert.Contains(t, result, "## エピソードの長さ")
		assert.Contains(t, result, "10分")
		assert.Contains(t, result, "## 今回のテーマ")
		assert.Contains(t, result, "AI について語る")
	})

	t.Run("チャンネル設定が空の場合は省略される", func(t *testing.T) {
		channel := &model.Channel{
			UserPrompt: "",
			Characters: []model.Character{
				{Name: "太郎"},
			},
		}

		result := svc.buildUserPrompt(channel, "テスト", 5)

		assert.NotContains(t, result, "## チャンネル設定")
		assert.Contains(t, result, "## 登場人物")
		assert.Contains(t, result, "- 太郎")
	})

	t.Run("キャラクターのペルソナが空の場合は名前のみ", func(t *testing.T) {
		channel := &model.Channel{
			Characters: []model.Character{
				{Name: "太郎", Persona: ""},
				{Name: "花子", Persona: "ゲスト"},
			},
		}

		result := svc.buildUserPrompt(channel, "テスト", 10)

		assert.Contains(t, result, "- 太郎\n")
		assert.Contains(t, result, "- 花子: ゲスト")
	})
}

func TestGenerateScript_Validation(t *testing.T) {
	ctx := context.Background()

	t.Run("無効な userID でエラー", func(t *testing.T) {
		svc := &scriptService{}

		_, err := svc.GenerateScript(ctx, "invalid-uuid", uuid.New().String(), uuid.New().String(), "test", nil)

		assert.Error(t, err)
	})

	t.Run("無効な channelID でエラー", func(t *testing.T) {
		svc := &scriptService{}

		_, err := svc.GenerateScript(ctx, uuid.New().String(), "invalid-uuid", uuid.New().String(), "test", nil)

		assert.Error(t, err)
	})

	t.Run("無効な episodeID でエラー", func(t *testing.T) {
		svc := &scriptService{}

		_, err := svc.GenerateScript(ctx, uuid.New().String(), uuid.New().String(), "invalid-uuid", "test", nil)

		assert.Error(t, err)
	})
}

func TestGenerateScript_ChannelNotFound(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	channelID := uuid.New()
	episodeID := uuid.New()

	mockChannelRepo := new(mockChannelRepository)
	mockChannelRepo.On("FindByID", ctx, channelID).Return(nil, apperror.ErrNotFound.WithMessage("Channel not found"))

	svc := &scriptService{
		channelRepo: mockChannelRepo,
	}

	_, err := svc.GenerateScript(ctx, userID.String(), channelID.String(), episodeID.String(), "test", nil)

	assert.Error(t, err)
	var appErr *apperror.AppError
	assert.True(t, errors.As(err, &appErr))
	mockChannelRepo.AssertExpectations(t)
}

func TestGenerateScript_Forbidden(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	otherUserID := uuid.New()
	channelID := uuid.New()
	episodeID := uuid.New()

	mockChannelRepo := new(mockChannelRepository)
	mockChannelRepo.On("FindByID", ctx, channelID).Return(&model.Channel{
		ID:     channelID,
		UserID: otherUserID, // 別のユーザーのチャンネル
	}, nil)

	svc := &scriptService{
		channelRepo: mockChannelRepo,
	}

	_, err := svc.GenerateScript(ctx, userID.String(), channelID.String(), episodeID.String(), "test", nil)

	assert.Error(t, err)
	var appErr *apperror.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperror.ErrForbidden.Code, appErr.Code)
	mockChannelRepo.AssertExpectations(t)
}

func TestGenerateScript_EpisodeNotFound(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	channelID := uuid.New()
	episodeID := uuid.New()

	mockChannelRepo := new(mockChannelRepository)
	mockChannelRepo.On("FindByID", ctx, channelID).Return(&model.Channel{
		ID:     channelID,
		UserID: userID,
	}, nil)

	mockEpisodeRepo := new(mockEpisodeRepository)
	mockEpisodeRepo.On("FindByID", ctx, episodeID).Return(nil, apperror.ErrNotFound.WithMessage("Episode not found"))

	svc := &scriptService{
		channelRepo: mockChannelRepo,
		episodeRepo: mockEpisodeRepo,
	}

	_, err := svc.GenerateScript(ctx, userID.String(), channelID.String(), episodeID.String(), "test", nil)

	assert.Error(t, err)
	mockChannelRepo.AssertExpectations(t)
	mockEpisodeRepo.AssertExpectations(t)
}

func TestGenerateScript_EpisodeNotInChannel(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	channelID := uuid.New()
	otherChannelID := uuid.New()
	episodeID := uuid.New()

	mockChannelRepo := new(mockChannelRepository)
	mockChannelRepo.On("FindByID", ctx, channelID).Return(&model.Channel{
		ID:     channelID,
		UserID: userID,
	}, nil)

	mockEpisodeRepo := new(mockEpisodeRepository)
	mockEpisodeRepo.On("FindByID", ctx, episodeID).Return(&model.Episode{
		ID:        episodeID,
		ChannelID: otherChannelID, // 別のチャンネルのエピソード
	}, nil)

	svc := &scriptService{
		channelRepo: mockChannelRepo,
		episodeRepo: mockEpisodeRepo,
	}

	_, err := svc.GenerateScript(ctx, userID.String(), channelID.String(), episodeID.String(), "test", nil)

	assert.Error(t, err)
	var appErr *apperror.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperror.ErrNotFound.Code, appErr.Code)
	mockChannelRepo.AssertExpectations(t)
	mockEpisodeRepo.AssertExpectations(t)
}

func TestGenerateScript_LLMError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	channelID := uuid.New()
	episodeID := uuid.New()
	characterID := uuid.New()

	mockChannelRepo := new(mockChannelRepository)
	mockChannelRepo.On("FindByID", ctx, channelID).Return(&model.Channel{
		ID:     channelID,
		UserID: userID,
		Characters: []model.Character{
			{ID: characterID, Name: "太郎"},
		},
	}, nil)

	mockEpisodeRepo := new(mockEpisodeRepository)
	mockEpisodeRepo.On("FindByID", ctx, episodeID).Return(&model.Episode{
		ID:        episodeID,
		ChannelID: channelID,
	}, nil)

	mockLLM := new(mockLLMClient)
	mockLLM.On("GenerateScript", ctx, mock.Anything, mock.Anything).Return("", apperror.ErrGenerationFailed.WithMessage("LLM error"))

	svc := &scriptService{
		channelRepo: mockChannelRepo,
		episodeRepo: mockEpisodeRepo,
		llmClient:   mockLLM,
	}

	_, err := svc.GenerateScript(ctx, userID.String(), channelID.String(), episodeID.String(), "test", nil)

	assert.Error(t, err)
	mockChannelRepo.AssertExpectations(t)
	mockEpisodeRepo.AssertExpectations(t)
	mockLLM.AssertExpectations(t)
}

func TestGenerateScript_ParseError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	channelID := uuid.New()
	episodeID := uuid.New()
	characterID := uuid.New()

	mockChannelRepo := new(mockChannelRepository)
	mockChannelRepo.On("FindByID", ctx, channelID).Return(&model.Channel{
		ID:     channelID,
		UserID: userID,
		Characters: []model.Character{
			{ID: characterID, Name: "太郎"},
		},
	}, nil)

	mockEpisodeRepo := new(mockEpisodeRepository)
	mockEpisodeRepo.On("FindByID", ctx, episodeID).Return(&model.Episode{
		ID:        episodeID,
		ChannelID: channelID,
	}, nil)

	mockLLM := new(mockLLMClient)
	// 不正な形式の出力（話者名が許可リストにない）
	mockLLM.On("GenerateScript", ctx, mock.Anything, mock.Anything).Return("不明な話者: こんにちは", nil)

	svc := &scriptService{
		channelRepo: mockChannelRepo,
		episodeRepo: mockEpisodeRepo,
		llmClient:   mockLLM,
	}

	_, err := svc.GenerateScript(ctx, userID.String(), channelID.String(), episodeID.String(), "test", nil)

	assert.Error(t, err)
	var appErr *apperror.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperror.ErrGenerationFailed.Code, appErr.Code)
	mockChannelRepo.AssertExpectations(t)
	mockEpisodeRepo.AssertExpectations(t)
	mockLLM.AssertExpectations(t)
}

func TestGenerateScript_DurationMinutesDefault(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	channelID := uuid.New()
	episodeID := uuid.New()
	characterID := uuid.New()

	mockChannelRepo := new(mockChannelRepository)
	mockChannelRepo.On("FindByID", ctx, channelID).Return(&model.Channel{
		ID:     channelID,
		UserID: userID,
		Characters: []model.Character{
			{ID: characterID, Name: "太郎"},
		},
	}, nil)

	mockEpisodeRepo := new(mockEpisodeRepository)
	mockEpisodeRepo.On("FindByID", ctx, episodeID).Return(&model.Episode{
		ID:        episodeID,
		ChannelID: channelID,
	}, nil)

	var capturedUserPrompt string
	mockLLM := new(mockLLMClient)
	mockLLM.On("GenerateScript", ctx, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		capturedUserPrompt = args.String(2)
	}).Return("", apperror.ErrGenerationFailed) // LLM エラーで早期終了させてトランザクションをスキップ

	svc := &scriptService{
		channelRepo: mockChannelRepo,
		episodeRepo: mockEpisodeRepo,
		llmClient:   mockLLM,
	}

	// durationMinutes を nil で渡す（デフォルト値 10 が使われるはず）
	_, _ = svc.GenerateScript(ctx, userID.String(), channelID.String(), episodeID.String(), "test", nil)

	// LLM に渡されたプロンプトにデフォルト値の 10 分が含まれていることを確認
	assert.Contains(t, capturedUserPrompt, "10分")
	mockLLM.AssertExpectations(t)
}

func TestGenerateScript_DurationMinutesCustom(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	channelID := uuid.New()
	episodeID := uuid.New()
	characterID := uuid.New()

	mockChannelRepo := new(mockChannelRepository)
	mockChannelRepo.On("FindByID", ctx, channelID).Return(&model.Channel{
		ID:     channelID,
		UserID: userID,
		Characters: []model.Character{
			{ID: characterID, Name: "太郎"},
		},
	}, nil)

	mockEpisodeRepo := new(mockEpisodeRepository)
	mockEpisodeRepo.On("FindByID", ctx, episodeID).Return(&model.Episode{
		ID:        episodeID,
		ChannelID: channelID,
	}, nil)

	var capturedUserPrompt string
	mockLLM := new(mockLLMClient)
	mockLLM.On("GenerateScript", ctx, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		capturedUserPrompt = args.String(2)
	}).Return("", apperror.ErrGenerationFailed) // LLM エラーで早期終了させてトランザクションをスキップ

	svc := &scriptService{
		channelRepo: mockChannelRepo,
		episodeRepo: mockEpisodeRepo,
		llmClient:   mockLLM,
	}

	// durationMinutes を 30 で指定
	duration := 30
	_, _ = svc.GenerateScript(ctx, userID.String(), channelID.String(), episodeID.String(), "test", &duration)

	// LLM に渡されたプロンプトに指定した 30 分が含まれていることを確認
	assert.Contains(t, capturedUserPrompt, "30分")
	mockLLM.AssertExpectations(t)
}
