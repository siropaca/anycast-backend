package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
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
	sfxID := uuid.New()
	text := "テストテキスト"
	emotion := "happy"
	durationMs := 3000
	volume := decimal.NewFromFloat(0.75)

	baseScriptLine := &model.ScriptLine{
		ID:         lineID,
		EpisodeID:  episodeID,
		LineOrder:  1,
		LineType:   model.LineTypeSpeech,
		SpeakerID:  &speakerID,
		Text:       &text,
		Emotion:    &emotion,
		DurationMs: &durationMs,
		SfxID:      &sfxID,
		Volume:     &volume,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	t.Run("基本的な変換が正しく行われる", func(t *testing.T) {
		svc := &scriptLineService{}

		sl := *baseScriptLine
		sl.Speaker = nil
		sl.Sfx = nil

		resp := svc.toScriptLineResponse(&sl)

		assert.Equal(t, lineID, resp.ID)
		assert.Equal(t, 1, resp.LineOrder)
		assert.Equal(t, "speech", resp.LineType)
		assert.Equal(t, &text, resp.Text)
		assert.Equal(t, &emotion, resp.Emotion)
		assert.Equal(t, &durationMs, resp.DurationMs)
		assert.Equal(t, now, resp.CreatedAt)
		assert.Equal(t, now, resp.UpdatedAt)
	})

	t.Run("Volume が正しく float64 に変換される", func(t *testing.T) {
		svc := &scriptLineService{}

		sl := *baseScriptLine
		sl.Speaker = nil
		sl.Sfx = nil

		resp := svc.toScriptLineResponse(&sl)

		assert.NotNil(t, resp.Volume)
		assert.Equal(t, 0.75, *resp.Volume)
	})

	t.Run("Volume が nil の場合、レスポンスの Volume も nil", func(t *testing.T) {
		svc := &scriptLineService{}

		sl := *baseScriptLine
		sl.Volume = nil
		sl.Speaker = nil
		sl.Sfx = nil

		resp := svc.toScriptLineResponse(&sl)

		assert.Nil(t, resp.Volume)
	})

	t.Run("Speaker がある場合、正しく変換される", func(t *testing.T) {
		svc := &scriptLineService{}

		sl := *baseScriptLine
		sl.Speaker = &model.Character{
			ID:   speakerID,
			Name: "テストスピーカー",
		}
		sl.Sfx = nil

		resp := svc.toScriptLineResponse(&sl)

		assert.NotNil(t, resp.Speaker)
		assert.Equal(t, speakerID, resp.Speaker.ID)
		assert.Equal(t, "テストスピーカー", resp.Speaker.Name)
	})

	t.Run("Speaker が nil の場合、レスポンスの Speaker も nil", func(t *testing.T) {
		svc := &scriptLineService{}

		sl := *baseScriptLine
		sl.Speaker = nil
		sl.Sfx = nil

		resp := svc.toScriptLineResponse(&sl)

		assert.Nil(t, resp.Speaker)
	})

	t.Run("Sfx がある場合、正しく変換される", func(t *testing.T) {
		svc := &scriptLineService{}

		sl := *baseScriptLine
		sl.Speaker = nil
		sl.Sfx = &model.SoundEffect{
			ID:   sfxID,
			Name: "テスト効果音",
		}

		resp := svc.toScriptLineResponse(&sl)

		assert.NotNil(t, resp.Sfx)
		assert.Equal(t, sfxID, resp.Sfx.ID)
		assert.Equal(t, "テスト効果音", resp.Sfx.Name)
	})

	t.Run("Sfx が nil の場合、レスポンスの Sfx も nil", func(t *testing.T) {
		svc := &scriptLineService{}

		sl := *baseScriptLine
		sl.Speaker = nil
		sl.Sfx = nil

		resp := svc.toScriptLineResponse(&sl)

		assert.Nil(t, resp.Sfx)
	})

	t.Run("Text が nil の場合、レスポンスの Text も nil", func(t *testing.T) {
		svc := &scriptLineService{}

		sl := *baseScriptLine
		sl.Text = nil
		sl.Speaker = nil
		sl.Sfx = nil

		resp := svc.toScriptLineResponse(&sl)

		assert.Nil(t, resp.Text)
	})

	t.Run("Emotion が nil の場合、レスポンスの Emotion も nil", func(t *testing.T) {
		svc := &scriptLineService{}

		sl := *baseScriptLine
		sl.Emotion = nil
		sl.Speaker = nil
		sl.Sfx = nil

		resp := svc.toScriptLineResponse(&sl)

		assert.Nil(t, resp.Emotion)
	})

	t.Run("DurationMs が nil の場合、レスポンスの DurationMs も nil", func(t *testing.T) {
		svc := &scriptLineService{}

		sl := *baseScriptLine
		sl.DurationMs = nil
		sl.Speaker = nil
		sl.Sfx = nil

		resp := svc.toScriptLineResponse(&sl)

		assert.Nil(t, resp.DurationMs)
	})
}

func TestToScriptLineResponses(t *testing.T) {
	now := time.Now()
	episodeID := uuid.New()
	text1 := "テキスト1"
	text2 := "テキスト2"

	scriptLines := []model.ScriptLine{
		{
			ID:        uuid.New(),
			EpisodeID: episodeID,
			LineOrder: 1,
			LineType:  model.LineTypeSpeech,
			Text:      &text1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        uuid.New(),
			EpisodeID: episodeID,
			LineOrder: 2,
			LineType:  model.LineTypeSfx,
			Text:      &text2,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	t.Run("複数の台本行を正しく変換する", func(t *testing.T) {
		svc := &scriptLineService{}

		result := svc.toScriptLineResponses(scriptLines)

		assert.Len(t, result, 2)
		assert.Equal(t, 1, result[0].LineOrder)
		assert.Equal(t, 2, result[1].LineOrder)
		assert.Equal(t, "speech", result[0].LineType)
		assert.Equal(t, "sfx", result[1].LineType)
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
