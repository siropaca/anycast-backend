package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// mockStorageClient は channel_test.go で定義済み

// mockScriptLineRepository は ScriptLineRepository のモック
type mockScriptLineRepository struct {
	mock.Mock
}

func (m *mockScriptLineRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ScriptLine, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ScriptLine), args.Error(1)
}

func (m *mockScriptLineRepository) FindByEpisodeID(ctx context.Context, episodeID uuid.UUID) ([]model.ScriptLine, error) {
	args := m.Called(ctx, episodeID)
	return args.Get(0).([]model.ScriptLine), args.Error(1)
}

func (m *mockScriptLineRepository) FindByEpisodeIDWithVoice(ctx context.Context, episodeID uuid.UUID) ([]model.ScriptLine, error) {
	args := m.Called(ctx, episodeID)
	return args.Get(0).([]model.ScriptLine), args.Error(1)
}

func (m *mockScriptLineRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockScriptLineRepository) DeleteByEpisodeID(ctx context.Context, episodeID uuid.UUID) error {
	args := m.Called(ctx, episodeID)
	return args.Error(0)
}

func (m *mockScriptLineRepository) CreateBatch(ctx context.Context, scriptLines []model.ScriptLine) ([]model.ScriptLine, error) {
	args := m.Called(ctx, scriptLines)
	return args.Get(0).([]model.ScriptLine), args.Error(1)
}

func (m *mockScriptLineRepository) Update(ctx context.Context, scriptLine *model.ScriptLine) error {
	args := m.Called(ctx, scriptLine)
	return args.Error(0)
}

func TestToScriptLineResponse(t *testing.T) {
	now := time.Now()
	lineID := uuid.New()
	episodeID := uuid.New()
	speakerID := uuid.New()
	voiceID := uuid.New()
	emotion := "happy"

	baseScriptLine := &model.ScriptLine{
		ID:        lineID,
		EpisodeID: episodeID,
		LineOrder: 1,
		SpeakerID: speakerID,
		Text:      "テストテキスト",
		Emotion:   &emotion,
		CreatedAt: now,
		UpdatedAt: now,
		Speaker: model.Character{
			ID:      speakerID,
			Name:    "テストスピーカー",
			Persona: "テスト用のペルソナ",
			Voice: model.Voice{
				ID:       voiceID,
				Name:     "テストボイス",
				Provider: "google",
				Gender:   model.GenderFemale,
			},
		},
	}

	t.Run("基本的な変換が正しく行われる", func(t *testing.T) {
		svc := &scriptLineService{}

		sl := *baseScriptLine

		resp := svc.toScriptLineResponse(&sl)

		assert.Equal(t, lineID, resp.ID)
		assert.Equal(t, 1, resp.LineOrder)
		assert.Equal(t, "テストテキスト", resp.Text)
		assert.Equal(t, &emotion, resp.Emotion)
		assert.Equal(t, now, resp.CreatedAt)
		assert.Equal(t, now, resp.UpdatedAt)
	})

	t.Run("Speaker が正しく変換される", func(t *testing.T) {
		svc := &scriptLineService{}

		sl := *baseScriptLine

		resp := svc.toScriptLineResponse(&sl)

		assert.Equal(t, speakerID, resp.Speaker.ID)
		assert.Equal(t, "テストスピーカー", resp.Speaker.Name)
		assert.Equal(t, "テスト用のペルソナ", resp.Speaker.Persona)
		assert.Equal(t, voiceID, resp.Speaker.Voice.ID)
		assert.Equal(t, "テストボイス", resp.Speaker.Voice.Name)
		assert.Equal(t, "google", resp.Speaker.Voice.Provider)
		assert.Equal(t, "female", resp.Speaker.Voice.Gender)
	})

	t.Run("Emotion が nil の場合、レスポンスの Emotion も nil", func(t *testing.T) {
		svc := &scriptLineService{}

		sl := *baseScriptLine
		sl.Emotion = nil

		resp := svc.toScriptLineResponse(&sl)

		assert.Nil(t, resp.Emotion)
	})
}

func TestToScriptLineResponses(t *testing.T) {
	now := time.Now()
	episodeID := uuid.New()
	speakerID := uuid.New()
	voiceID := uuid.New()

	scriptLines := []model.ScriptLine{
		{
			ID:        uuid.New(),
			EpisodeID: episodeID,
			LineOrder: 1,
			SpeakerID: speakerID,
			Text:      "テキスト1",
			CreatedAt: now,
			UpdatedAt: now,
			Speaker: model.Character{
				ID:   speakerID,
				Name: "テストスピーカー",
				Voice: model.Voice{
					ID:       voiceID,
					Provider: "google",
					Gender:   model.GenderMale,
				},
			},
		},
		{
			ID:        uuid.New(),
			EpisodeID: episodeID,
			LineOrder: 2,
			SpeakerID: speakerID,
			Text:      "テキスト2",
			CreatedAt: now,
			UpdatedAt: now,
			Speaker: model.Character{
				ID:   speakerID,
				Name: "テストスピーカー",
				Voice: model.Voice{
					ID:       voiceID,
					Provider: "google",
					Gender:   model.GenderMale,
				},
			},
		},
	}

	t.Run("複数の台本行を正しく変換する", func(t *testing.T) {
		svc := &scriptLineService{}

		result := svc.toScriptLineResponses(scriptLines)

		assert.Len(t, result, 2)
		assert.Equal(t, 1, result[0].LineOrder)
		assert.Equal(t, 2, result[1].LineOrder)
		assert.Equal(t, "テキスト1", result[0].Text)
		assert.Equal(t, "テキスト2", result[1].Text)
	})

	t.Run("空のスライスの場合、空のスライスを返す", func(t *testing.T) {
		svc := &scriptLineService{}

		result := svc.toScriptLineResponses([]model.ScriptLine{})

		assert.Len(t, result, 0)
		assert.NotNil(t, result)
	})
}

func TestScriptLineService_Delete(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	channelID := uuid.New()
	episodeID := uuid.New()
	lineID := uuid.New()

	t.Run("正常に削除できる", func(t *testing.T) {
		mockChannelRepo := new(mockChannelRepository)
		mockEpisodeRepo := new(mockEpisodeRepository)
		mockScriptLineRepo := new(mockScriptLineRepository)

		mockChannelRepo.On("FindByID", ctx, channelID).Return(&model.Channel{
			ID:     channelID,
			UserID: userID,
		}, nil)

		mockEpisodeRepo.On("FindByID", ctx, episodeID).Return(&model.Episode{
			ID:        episodeID,
			ChannelID: channelID,
		}, nil)

		mockScriptLineRepo.On("FindByID", ctx, lineID).Return(&model.ScriptLine{
			ID:        lineID,
			EpisodeID: episodeID,
		}, nil)

		mockScriptLineRepo.On("Delete", ctx, lineID).Return(nil)

		svc := &scriptLineService{
			channelRepo:    mockChannelRepo,
			episodeRepo:    mockEpisodeRepo,
			scriptLineRepo: mockScriptLineRepo,
		}

		err := svc.Delete(ctx, userID.String(), channelID.String(), episodeID.String(), lineID.String())

		assert.NoError(t, err)
		mockChannelRepo.AssertExpectations(t)
		mockEpisodeRepo.AssertExpectations(t)
		mockScriptLineRepo.AssertExpectations(t)
	})

	t.Run("無効な userID でエラー", func(t *testing.T) {
		svc := &scriptLineService{}

		err := svc.Delete(ctx, "invalid-uuid", channelID.String(), episodeID.String(), lineID.String())

		assert.Error(t, err)
	})

	t.Run("無効な channelID でエラー", func(t *testing.T) {
		svc := &scriptLineService{}

		err := svc.Delete(ctx, userID.String(), "invalid-uuid", episodeID.String(), lineID.String())

		assert.Error(t, err)
	})

	t.Run("無効な episodeID でエラー", func(t *testing.T) {
		svc := &scriptLineService{}

		err := svc.Delete(ctx, userID.String(), channelID.String(), "invalid-uuid", lineID.String())

		assert.Error(t, err)
	})

	t.Run("無効な lineID でエラー", func(t *testing.T) {
		svc := &scriptLineService{}

		err := svc.Delete(ctx, userID.String(), channelID.String(), episodeID.String(), "invalid-uuid")

		assert.Error(t, err)
	})

	t.Run("チャンネルが見つからない場合エラー", func(t *testing.T) {
		mockChannelRepo := new(mockChannelRepository)
		mockChannelRepo.On("FindByID", ctx, channelID).Return(nil, apperror.ErrNotFound.WithMessage("Channel not found"))

		svc := &scriptLineService{
			channelRepo: mockChannelRepo,
		}

		err := svc.Delete(ctx, userID.String(), channelID.String(), episodeID.String(), lineID.String())

		assert.Error(t, err)
		var appErr *apperror.AppError
		assert.True(t, errors.As(err, &appErr))
		mockChannelRepo.AssertExpectations(t)
	})

	t.Run("チャンネルのオーナーでない場合エラー", func(t *testing.T) {
		otherUserID := uuid.New()
		mockChannelRepo := new(mockChannelRepository)
		mockChannelRepo.On("FindByID", ctx, channelID).Return(&model.Channel{
			ID:     channelID,
			UserID: otherUserID,
		}, nil)

		svc := &scriptLineService{
			channelRepo: mockChannelRepo,
		}

		err := svc.Delete(ctx, userID.String(), channelID.String(), episodeID.String(), lineID.String())

		assert.Error(t, err)
		var appErr *apperror.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, apperror.ErrForbidden.Code, appErr.Code)
		mockChannelRepo.AssertExpectations(t)
	})

	t.Run("エピソードが見つからない場合エラー", func(t *testing.T) {
		mockChannelRepo := new(mockChannelRepository)
		mockEpisodeRepo := new(mockEpisodeRepository)

		mockChannelRepo.On("FindByID", ctx, channelID).Return(&model.Channel{
			ID:     channelID,
			UserID: userID,
		}, nil)

		mockEpisodeRepo.On("FindByID", ctx, episodeID).Return(nil, apperror.ErrNotFound.WithMessage("Episode not found"))

		svc := &scriptLineService{
			channelRepo: mockChannelRepo,
			episodeRepo: mockEpisodeRepo,
		}

		err := svc.Delete(ctx, userID.String(), channelID.String(), episodeID.String(), lineID.String())

		assert.Error(t, err)
		mockChannelRepo.AssertExpectations(t)
		mockEpisodeRepo.AssertExpectations(t)
	})

	t.Run("エピソードが別のチャンネルに属している場合エラー", func(t *testing.T) {
		otherChannelID := uuid.New()
		mockChannelRepo := new(mockChannelRepository)
		mockEpisodeRepo := new(mockEpisodeRepository)

		mockChannelRepo.On("FindByID", ctx, channelID).Return(&model.Channel{
			ID:     channelID,
			UserID: userID,
		}, nil)

		mockEpisodeRepo.On("FindByID", ctx, episodeID).Return(&model.Episode{
			ID:        episodeID,
			ChannelID: otherChannelID,
		}, nil)

		svc := &scriptLineService{
			channelRepo: mockChannelRepo,
			episodeRepo: mockEpisodeRepo,
		}

		err := svc.Delete(ctx, userID.String(), channelID.String(), episodeID.String(), lineID.String())

		assert.Error(t, err)
		var appErr *apperror.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, apperror.ErrNotFound.Code, appErr.Code)
		mockChannelRepo.AssertExpectations(t)
		mockEpisodeRepo.AssertExpectations(t)
	})

	t.Run("台本行が見つからない場合エラー", func(t *testing.T) {
		mockChannelRepo := new(mockChannelRepository)
		mockEpisodeRepo := new(mockEpisodeRepository)
		mockScriptLineRepo := new(mockScriptLineRepository)

		mockChannelRepo.On("FindByID", ctx, channelID).Return(&model.Channel{
			ID:     channelID,
			UserID: userID,
		}, nil)

		mockEpisodeRepo.On("FindByID", ctx, episodeID).Return(&model.Episode{
			ID:        episodeID,
			ChannelID: channelID,
		}, nil)

		mockScriptLineRepo.On("FindByID", ctx, lineID).Return(nil, apperror.ErrNotFound.WithMessage("Script line not found"))

		svc := &scriptLineService{
			channelRepo:    mockChannelRepo,
			episodeRepo:    mockEpisodeRepo,
			scriptLineRepo: mockScriptLineRepo,
		}

		err := svc.Delete(ctx, userID.String(), channelID.String(), episodeID.String(), lineID.String())

		assert.Error(t, err)
		mockChannelRepo.AssertExpectations(t)
		mockEpisodeRepo.AssertExpectations(t)
		mockScriptLineRepo.AssertExpectations(t)
	})

	t.Run("台本行が別のエピソードに属している場合エラー", func(t *testing.T) {
		otherEpisodeID := uuid.New()
		mockChannelRepo := new(mockChannelRepository)
		mockEpisodeRepo := new(mockEpisodeRepository)
		mockScriptLineRepo := new(mockScriptLineRepository)

		mockChannelRepo.On("FindByID", ctx, channelID).Return(&model.Channel{
			ID:     channelID,
			UserID: userID,
		}, nil)

		mockEpisodeRepo.On("FindByID", ctx, episodeID).Return(&model.Episode{
			ID:        episodeID,
			ChannelID: channelID,
		}, nil)

		mockScriptLineRepo.On("FindByID", ctx, lineID).Return(&model.ScriptLine{
			ID:        lineID,
			EpisodeID: otherEpisodeID,
		}, nil)

		svc := &scriptLineService{
			channelRepo:    mockChannelRepo,
			episodeRepo:    mockEpisodeRepo,
			scriptLineRepo: mockScriptLineRepo,
		}

		err := svc.Delete(ctx, userID.String(), channelID.String(), episodeID.String(), lineID.String())

		assert.Error(t, err)
		var appErr *apperror.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, apperror.ErrNotFound.Code, appErr.Code)
		mockChannelRepo.AssertExpectations(t)
		mockEpisodeRepo.AssertExpectations(t)
		mockScriptLineRepo.AssertExpectations(t)
	})
}
