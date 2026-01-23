package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
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
