package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// ScriptJobRepository のモック
type mockScriptJobRepository struct {
	mock.Mock
}

func (m *mockScriptJobRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ScriptJob, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ScriptJob), args.Error(1)
}

func (m *mockScriptJobRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filter repository.ScriptJobFilter) ([]model.ScriptJob, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.ScriptJob), args.Error(1)
}

func (m *mockScriptJobRepository) FindByEpisodeID(ctx context.Context, episodeID uuid.UUID) ([]model.ScriptJob, error) {
	args := m.Called(ctx, episodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.ScriptJob), args.Error(1)
}

func (m *mockScriptJobRepository) FindPendingByEpisodeID(ctx context.Context, episodeID uuid.UUID) (*model.ScriptJob, error) {
	args := m.Called(ctx, episodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ScriptJob), args.Error(1)
}

func (m *mockScriptJobRepository) Create(ctx context.Context, job *model.ScriptJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *mockScriptJobRepository) Update(ctx context.Context, job *model.ScriptJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *mockScriptJobRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockScriptJobRepository) UpdateProgress(ctx context.Context, id uuid.UUID, progress int) error {
	args := m.Called(ctx, id, progress)
	return args.Error(0)
}

func TestScriptJobStatus(t *testing.T) {
	t.Run("ScriptJobStatus 定数が正しい", func(t *testing.T) {
		assert.Equal(t, model.ScriptJobStatus("pending"), model.ScriptJobStatusPending)
		assert.Equal(t, model.ScriptJobStatus("processing"), model.ScriptJobStatusProcessing)
		assert.Equal(t, model.ScriptJobStatus("canceling"), model.ScriptJobStatusCanceling)
		assert.Equal(t, model.ScriptJobStatus("completed"), model.ScriptJobStatusCompleted)
		assert.Equal(t, model.ScriptJobStatus("failed"), model.ScriptJobStatusFailed)
		assert.Equal(t, model.ScriptJobStatus("canceled"), model.ScriptJobStatusCanceled)
	})
}

func TestScriptJobService_CancelJob(t *testing.T) {
	userID := uuid.New()
	jobID := uuid.New()
	episodeID := uuid.New()

	t.Run("pending ジョブをキャンセルすると canceled になる", func(t *testing.T) {
		mockRepo := new(mockScriptJobRepository)
		job := &model.ScriptJob{
			ID:        jobID,
			EpisodeID: episodeID,
			UserID:    userID,
			Status:    model.ScriptJobStatusPending,
		}
		mockRepo.On("FindByID", mock.Anything, jobID).Return(job, nil)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(j *model.ScriptJob) bool {
			return j.Status == model.ScriptJobStatusCanceled
		})).Return(nil)

		svc := &scriptJobService{scriptJobRepo: mockRepo}
		err := svc.CancelJob(context.Background(), userID.String(), jobID.String())

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("processing ジョブをキャンセルすると canceling になる", func(t *testing.T) {
		mockRepo := new(mockScriptJobRepository)
		job := &model.ScriptJob{
			ID:        jobID,
			EpisodeID: episodeID,
			UserID:    userID,
			Status:    model.ScriptJobStatusProcessing,
		}
		mockRepo.On("FindByID", mock.Anything, jobID).Return(job, nil)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(j *model.ScriptJob) bool {
			return j.Status == model.ScriptJobStatusCanceling
		})).Return(nil)

		svc := &scriptJobService{scriptJobRepo: mockRepo}
		err := svc.CancelJob(context.Background(), userID.String(), jobID.String())

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("completed ジョブはキャンセルできない", func(t *testing.T) {
		mockRepo := new(mockScriptJobRepository)
		job := &model.ScriptJob{
			ID:        jobID,
			EpisodeID: episodeID,
			UserID:    userID,
			Status:    model.ScriptJobStatusCompleted,
		}
		mockRepo.On("FindByID", mock.Anything, jobID).Return(job, nil)

		svc := &scriptJobService{scriptJobRepo: mockRepo}
		err := svc.CancelJob(context.Background(), userID.String(), jobID.String())

		assert.Error(t, err)
		var appErr *apperror.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, apperror.CodeValidation, appErr.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("他のユーザーのジョブはキャンセルできない", func(t *testing.T) {
		mockRepo := new(mockScriptJobRepository)
		otherUserID := uuid.New()
		job := &model.ScriptJob{
			ID:        jobID,
			EpisodeID: episodeID,
			UserID:    otherUserID,
			Status:    model.ScriptJobStatusPending,
		}
		mockRepo.On("FindByID", mock.Anything, jobID).Return(job, nil)

		svc := &scriptJobService{scriptJobRepo: mockRepo}
		err := svc.CancelJob(context.Background(), userID.String(), jobID.String())

		assert.Error(t, err)
		var appErr *apperror.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, apperror.CodeForbidden, appErr.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestScriptJobService_updateProgress(t *testing.T) {
	jobID := uuid.New()
	userID := uuid.New()
	episodeID := uuid.New()

	t.Run("UpdateProgress は進捗のみを更新しステータスを上書きしない", func(t *testing.T) {
		mockRepo := new(mockScriptJobRepository)
		job := &model.ScriptJob{
			ID:        jobID,
			EpisodeID: episodeID,
			UserID:    userID,
			Status:    model.ScriptJobStatusProcessing,
			Progress:  30,
		}

		// UpdateProgress が呼ばれることを確認（Update ではなく）
		mockRepo.On("UpdateProgress", mock.Anything, jobID, 50).Return(nil)

		svc := &scriptJobService{scriptJobRepo: mockRepo}
		svc.updateProgress(context.Background(), job, 50, "処理中...")

		// job.Progress がメモリ上で更新されていることを確認
		assert.Equal(t, 50, job.Progress)
		// Status は変更されていないことを確認
		assert.Equal(t, model.ScriptJobStatusProcessing, job.Status)
		mockRepo.AssertExpectations(t)
	})
}

func TestScriptJobService_checkCanceled(t *testing.T) {
	jobID := uuid.New()
	userID := uuid.New()
	episodeID := uuid.New()

	t.Run("canceling 状態のジョブは ErrCanceled を返す", func(t *testing.T) {
		mockRepo := new(mockScriptJobRepository)
		job := &model.ScriptJob{
			ID:        jobID,
			EpisodeID: episodeID,
			UserID:    userID,
			Status:    model.ScriptJobStatusCanceling,
		}
		mockRepo.On("FindByID", mock.Anything, jobID).Return(job, nil)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(j *model.ScriptJob) bool {
			return j.Status == model.ScriptJobStatusCanceled
		})).Return(nil)

		svc := &scriptJobService{scriptJobRepo: mockRepo}
		err := svc.checkCanceled(context.Background(), job)

		assert.Error(t, err)
		var appErr *apperror.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, apperror.CodeCanceled, appErr.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("processing 状態のジョブはエラーを返さない", func(t *testing.T) {
		mockRepo := new(mockScriptJobRepository)
		job := &model.ScriptJob{
			ID:        jobID,
			EpisodeID: episodeID,
			UserID:    userID,
			Status:    model.ScriptJobStatusProcessing,
		}
		mockRepo.On("FindByID", mock.Anything, jobID).Return(job, nil)

		svc := &scriptJobService{scriptJobRepo: mockRepo}
		err := svc.checkCanceled(context.Background(), job)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
