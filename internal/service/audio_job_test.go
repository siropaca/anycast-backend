package service

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/infrastructure/tts"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// AudioJobRepository のモック
type mockAudioJobRepository struct {
	mock.Mock
}

func (m *mockAudioJobRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.AudioJob, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AudioJob), args.Error(1)
}

func (m *mockAudioJobRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filter repository.AudioJobFilter) ([]model.AudioJob, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.AudioJob), args.Error(1)
}

func (m *mockAudioJobRepository) FindByEpisodeID(ctx context.Context, episodeID uuid.UUID) ([]model.AudioJob, error) {
	args := m.Called(ctx, episodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.AudioJob), args.Error(1)
}

func (m *mockAudioJobRepository) FindPendingByEpisodeID(ctx context.Context, episodeID uuid.UUID) (*model.AudioJob, error) {
	args := m.Called(ctx, episodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AudioJob), args.Error(1)
}

func (m *mockAudioJobRepository) Create(ctx context.Context, job *model.AudioJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *mockAudioJobRepository) Update(ctx context.Context, job *model.AudioJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *mockAudioJobRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestAudioJobService_GetJob(t *testing.T) {
	userID := uuid.New()
	jobID := uuid.New()
	episodeID := uuid.New()
	channelID := uuid.New()

	t.Run("ジョブを取得できる", func(t *testing.T) {
		mockRepo := new(mockAudioJobRepository)
		job := &model.AudioJob{
			ID:        jobID,
			EpisodeID: episodeID,
			UserID:    userID,
			Status:    model.AudioJobStatusPending,
			Progress:  0,
			Episode: model.Episode{
				ID:    episodeID,
				Title: "Test Episode",
				Channel: model.Channel{
					ID:   channelID,
					Name: "Test Channel",
				},
			},
		}
		mockRepo.On("FindByID", mock.Anything, jobID).Return(job, nil)

		svc := &audioJobService{audioJobRepo: mockRepo}
		result, err := svc.GetJob(context.Background(), userID.String(), jobID.String())

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, jobID, result.ID)
		assert.Equal(t, "pending", result.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("他のユーザーのジョブは取得できない", func(t *testing.T) {
		mockRepo := new(mockAudioJobRepository)
		otherUserID := uuid.New()
		job := &model.AudioJob{
			ID:        jobID,
			EpisodeID: episodeID,
			UserID:    otherUserID, // 別のユーザー
			Status:    model.AudioJobStatusPending,
		}
		mockRepo.On("FindByID", mock.Anything, jobID).Return(job, nil)

		svc := &audioJobService{audioJobRepo: mockRepo}
		result, err := svc.GetJob(context.Background(), userID.String(), jobID.String())

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "アクセス権限")
		mockRepo.AssertExpectations(t)
	})

	t.Run("存在しないジョブはエラーを返す", func(t *testing.T) {
		mockRepo := new(mockAudioJobRepository)
		mockRepo.On("FindByID", mock.Anything, jobID).Return(nil, apperror.ErrNotFound.WithMessage("Job not found"))

		svc := &audioJobService{audioJobRepo: mockRepo}
		result, err := svc.GetJob(context.Background(), userID.String(), jobID.String())

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("無効な jobID はエラーを返す", func(t *testing.T) {
		mockRepo := new(mockAudioJobRepository)

		svc := &audioJobService{audioJobRepo: mockRepo}
		result, err := svc.GetJob(context.Background(), userID.String(), "invalid-uuid")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("無効な userID はエラーを返す", func(t *testing.T) {
		mockRepo := new(mockAudioJobRepository)

		svc := &audioJobService{audioJobRepo: mockRepo}
		result, err := svc.GetJob(context.Background(), "invalid-uuid", jobID.String())

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestAudioJobService_ListMyJobs(t *testing.T) {
	userID := uuid.New()
	jobID1 := uuid.New()
	jobID2 := uuid.New()
	episodeID := uuid.New()
	channelID := uuid.New()

	t.Run("ジョブ一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockAudioJobRepository)
		jobs := []model.AudioJob{
			{
				ID:        jobID1,
				EpisodeID: episodeID,
				UserID:    userID,
				Status:    model.AudioJobStatusCompleted,
				Progress:  100,
				Episode: model.Episode{
					ID:    episodeID,
					Title: "Episode 1",
					Channel: model.Channel{
						ID:   channelID,
						Name: "Channel",
					},
				},
			},
			{
				ID:        jobID2,
				EpisodeID: episodeID,
				UserID:    userID,
				Status:    model.AudioJobStatusPending,
				Progress:  0,
				Episode: model.Episode{
					ID:    episodeID,
					Title: "Episode 2",
					Channel: model.Channel{
						ID:   channelID,
						Name: "Channel",
					},
				},
			},
		}
		filter := repository.AudioJobFilter{}
		mockRepo.On("FindByUserID", mock.Anything, userID, filter).Return(jobs, nil)

		svc := &audioJobService{audioJobRepo: mockRepo}
		result, err := svc.ListMyJobs(context.Background(), userID.String(), filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ステータスでフィルタできる", func(t *testing.T) {
		mockRepo := new(mockAudioJobRepository)
		status := model.AudioJobStatusPending
		filter := repository.AudioJobFilter{Status: &status}
		jobs := []model.AudioJob{
			{
				ID:        jobID1,
				EpisodeID: episodeID,
				UserID:    userID,
				Status:    model.AudioJobStatusPending,
				Progress:  0,
				Episode: model.Episode{
					ID:    episodeID,
					Title: "Episode 1",
					Channel: model.Channel{
						ID:   channelID,
						Name: "Channel",
					},
				},
			},
		}
		mockRepo.On("FindByUserID", mock.Anything, userID, filter).Return(jobs, nil)

		svc := &audioJobService{audioJobRepo: mockRepo}
		result, err := svc.ListMyJobs(context.Background(), userID.String(), filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 1)
		assert.Equal(t, "pending", result.Data[0].Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("空のジョブ一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockAudioJobRepository)
		filter := repository.AudioJobFilter{}
		mockRepo.On("FindByUserID", mock.Anything, userID, filter).Return([]model.AudioJob{}, nil)

		svc := &audioJobService{audioJobRepo: mockRepo}
		result, err := svc.ListMyJobs(context.Background(), userID.String(), filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Data)
		mockRepo.AssertExpectations(t)
	})

	t.Run("無効な userID はエラーを返す", func(t *testing.T) {
		mockRepo := new(mockAudioJobRepository)
		filter := repository.AudioJobFilter{}

		svc := &audioJobService{audioJobRepo: mockRepo}
		result, err := svc.ListMyJobs(context.Background(), "invalid-uuid", filter)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("リポジトリがエラーを返すとエラーを返す", func(t *testing.T) {
		mockRepo := new(mockAudioJobRepository)
		filter := repository.AudioJobFilter{}
		mockRepo.On("FindByUserID", mock.Anything, userID, filter).Return(nil, apperror.ErrInternal.WithMessage("Database error"))

		svc := &audioJobService{audioJobRepo: mockRepo}
		result, err := svc.ListMyJobs(context.Background(), userID.String(), filter)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestAudioJobStatus(t *testing.T) {
	t.Run("AudioJobStatus 定数が正しい", func(t *testing.T) {
		assert.Equal(t, model.AudioJobStatus("pending"), model.AudioJobStatusPending)
		assert.Equal(t, model.AudioJobStatus("processing"), model.AudioJobStatusProcessing)
		assert.Equal(t, model.AudioJobStatus("completed"), model.AudioJobStatusCompleted)
		assert.Equal(t, model.AudioJobStatus("failed"), model.AudioJobStatusFailed)
	})
}

func TestSplitTurnsIntoChunks(t *testing.T) {
	t.Run("空のターンの場合は nil を返す", func(t *testing.T) {
		chunks := splitTurnsIntoChunks(nil, 100)
		assert.Nil(t, chunks)

		chunks = splitTurnsIntoChunks([]tts.SpeakerTurn{}, 100)
		assert.Nil(t, chunks)
	})

	t.Run("制限内のターンは 1 チャンクになる", func(t *testing.T) {
		turns := []tts.SpeakerTurn{
			{Speaker: "speaker1", Text: "Hello"},
			{Speaker: "speaker2", Text: "World"},
		}
		chunks := splitTurnsIntoChunks(turns, 100)

		assert.Len(t, chunks, 1)
		assert.Len(t, chunks[0], 2)
	})

	t.Run("制限を超えるターンは複数チャンクに分割される", func(t *testing.T) {
		turns := []tts.SpeakerTurn{
			{Speaker: "speaker1", Text: "Hello World"},       // 11 + 15 = 26 bytes
			{Speaker: "speaker2", Text: "How are you"},       // 11 + 15 = 26 bytes
			{Speaker: "speaker1", Text: "I am fine"},         // 9 + 15 = 24 bytes
			{Speaker: "speaker2", Text: "Good to hear that"}, // 17 + 15 = 32 bytes
		}
		// 制限を 55 バイトに設定（ポーズ分 15 バイトを含む）
		// チャンク1: 26 + 26 = 52 bytes (< 55)
		// チャンク2: 24 bytes (52 + 24 = 76 > 55 なので新チャンク)
		// チャンク3: 32 bytes (24 + 32 = 56 > 55 なので新チャンク)
		chunks := splitTurnsIntoChunks(turns, 55)

		assert.Len(t, chunks, 3)
		assert.Len(t, chunks[0], 2)
		assert.Len(t, chunks[1], 1)
		assert.Len(t, chunks[2], 1)
	})

	t.Run("emotion を含むターンのバイト数が正しく計算される", func(t *testing.T) {
		emotion := "happy"
		turns := []tts.SpeakerTurn{
			{Speaker: "speaker1", Text: "Hello", Emotion: &emotion}, // 5 + 5 + 3 + 15 = 28 bytes
			{Speaker: "speaker2", Text: "World"},                    // 5 + 15 = 20 bytes
		}
		// 制限を 30 バイトに設定
		chunks := splitTurnsIntoChunks(turns, 30)

		assert.Len(t, chunks, 2) // emotion 付きのため分割される (28 + 20 = 48 > 30)
	})

	t.Run("日本語テキストが正しく分割される", func(t *testing.T) {
		// 日本語は UTF-8 で 1 文字 3 バイト、ポーズ分 15 バイトを加算
		turns := []tts.SpeakerTurn{
			{Speaker: "speaker1", Text: "こんにちは"},  // 15 + 15 = 30 bytes
			{Speaker: "speaker2", Text: "お元気ですか"}, // 18 + 15 = 33 bytes
			{Speaker: "speaker1", Text: "元気です"},   // 12 + 15 = 27 bytes
		}
		// 制限を 65 バイトに設定
		chunks := splitTurnsIntoChunks(turns, 65)

		assert.Len(t, chunks, 2)
		assert.Len(t, chunks[0], 2) // 30 + 33 = 63 bytes (< 65)
		assert.Len(t, chunks[1], 1) // 27 bytes (63 + 27 = 90 > 65 なので新チャンク)
	})

	t.Run("大量のターンが正しく分割される", func(t *testing.T) {
		// 100 ターン、各 50 バイト（+ ポーズ 15 バイト = 65 バイト）
		turns := make([]tts.SpeakerTurn, 100)
		for i := range turns {
			turns[i] = tts.SpeakerTurn{
				Speaker: "speaker1",
				Text:    strings.Repeat("a", 50),
			}
		}
		// 制限を 500 バイトに設定 → 約 7 ターンずつ (65 * 7 = 455)
		chunks := splitTurnsIntoChunks(turns, 500)

		// 各チャンクが制限を超えていないことを確認（ポーズ分を含む）
		for _, chunk := range chunks {
			totalBytes := 0
			for _, turn := range chunk {
				totalBytes += len(turn.Text) + turnPauseOverheadBytes
			}
			assert.LessOrEqual(t, totalBytes, 500)
		}
		// 全ターンが含まれていることを確認
		totalTurns := 0
		for _, chunk := range chunks {
			totalTurns += len(chunk)
		}
		assert.Equal(t, 100, totalTurns)
	})
}
