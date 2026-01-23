package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

func TestCleanupOrphanedMedia(t *testing.T) {
	ctx := context.Background()

	t.Run("正常系: dry-run で孤児メディアを検出する", func(t *testing.T) {
		mockAudioRepo := new(mockAudioRepository)
		mockImageRepo := new(mockImageRepository)
		mockStorage := new(mockStorageClient)

		svc := NewCleanupService(mockAudioRepo, mockImageRepo, mockStorage)

		orphanedAudios := []model.Audio{
			{ID: uuid.New(), Path: "audios/orphan1.mp3"},
			{ID: uuid.New(), Path: "audios/orphan2.mp3"},
		}
		orphanedImages := []model.Image{
			{ID: uuid.New(), Path: "images/orphan1.png"},
		}

		mockAudioRepo.On("FindOrphaned", mock.Anything).Return(orphanedAudios, nil)
		mockImageRepo.On("FindOrphaned", mock.Anything).Return(orphanedImages, nil)

		result, err := svc.CleanupOrphanedMedia(ctx, true)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.OrphanedAudios, 2)
		assert.Len(t, result.OrphanedImages, 1)
		assert.Equal(t, 0, result.DeletedAudioCount)
		assert.Equal(t, 0, result.DeletedImageCount)
		mockAudioRepo.AssertExpectations(t)
		mockImageRepo.AssertExpectations(t)
		mockStorage.AssertNotCalled(t, "Delete")
	})

	t.Run("正常系: 孤児メディアを削除する", func(t *testing.T) {
		mockAudioRepo := new(mockAudioRepository)
		mockImageRepo := new(mockImageRepository)
		mockStorage := new(mockStorageClient)

		svc := NewCleanupService(mockAudioRepo, mockImageRepo, mockStorage)

		audioID1 := uuid.New()
		audioID2 := uuid.New()
		imageID1 := uuid.New()

		orphanedAudios := []model.Audio{
			{ID: audioID1, Path: "audios/orphan1.mp3"},
			{ID: audioID2, Path: "audios/orphan2.mp3"},
		}
		orphanedImages := []model.Image{
			{ID: imageID1, Path: "images/orphan1.png"},
		}

		mockAudioRepo.On("FindOrphaned", mock.Anything).Return(orphanedAudios, nil)
		mockImageRepo.On("FindOrphaned", mock.Anything).Return(orphanedImages, nil)
		mockStorage.On("Delete", mock.Anything, "audios/orphan1.mp3").Return(nil)
		mockStorage.On("Delete", mock.Anything, "audios/orphan2.mp3").Return(nil)
		mockStorage.On("Delete", mock.Anything, "images/orphan1.png").Return(nil)
		mockAudioRepo.On("Delete", mock.Anything, audioID1).Return(nil)
		mockAudioRepo.On("Delete", mock.Anything, audioID2).Return(nil)
		mockImageRepo.On("Delete", mock.Anything, imageID1).Return(nil)

		result, err := svc.CleanupOrphanedMedia(ctx, false)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.DeletedAudioCount)
		assert.Equal(t, 1, result.DeletedImageCount)
		assert.Equal(t, 0, result.FailedAudioCount)
		assert.Equal(t, 0, result.FailedImageCount)
		mockAudioRepo.AssertExpectations(t)
		mockImageRepo.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
	})

	t.Run("正常系: 孤児メディアがない場合", func(t *testing.T) {
		mockAudioRepo := new(mockAudioRepository)
		mockImageRepo := new(mockImageRepository)
		mockStorage := new(mockStorageClient)

		svc := NewCleanupService(mockAudioRepo, mockImageRepo, mockStorage)

		mockAudioRepo.On("FindOrphaned", mock.Anything).Return([]model.Audio{}, nil)
		mockImageRepo.On("FindOrphaned", mock.Anything).Return([]model.Image{}, nil)

		result, err := svc.CleanupOrphanedMedia(ctx, false)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.OrphanedAudios, 0)
		assert.Len(t, result.OrphanedImages, 0)
		assert.Equal(t, 0, result.DeletedAudioCount)
		assert.Equal(t, 0, result.DeletedImageCount)
	})

	t.Run("異常系: 孤児 Audio の取得に失敗する", func(t *testing.T) {
		mockAudioRepo := new(mockAudioRepository)
		mockImageRepo := new(mockImageRepository)
		mockStorage := new(mockStorageClient)

		svc := NewCleanupService(mockAudioRepo, mockImageRepo, mockStorage)

		mockAudioRepo.On("FindOrphaned", mock.Anything).Return(nil, errors.New("db error"))

		result, err := svc.CleanupOrphanedMedia(ctx, false)

		assert.Nil(t, result)
		assert.Error(t, err)
	})

	t.Run("異常系: 孤児 Image の取得に失敗する", func(t *testing.T) {
		mockAudioRepo := new(mockAudioRepository)
		mockImageRepo := new(mockImageRepository)
		mockStorage := new(mockStorageClient)

		svc := NewCleanupService(mockAudioRepo, mockImageRepo, mockStorage)

		mockAudioRepo.On("FindOrphaned", mock.Anything).Return([]model.Audio{}, nil)
		mockImageRepo.On("FindOrphaned", mock.Anything).Return(nil, errors.New("db error"))

		result, err := svc.CleanupOrphanedMedia(ctx, false)

		assert.Nil(t, result)
		assert.Error(t, err)
	})

	t.Run("異常系: GCS からの削除に失敗した場合はスキップして続行する", func(t *testing.T) {
		mockAudioRepo := new(mockAudioRepository)
		mockImageRepo := new(mockImageRepository)
		mockStorage := new(mockStorageClient)

		svc := NewCleanupService(mockAudioRepo, mockImageRepo, mockStorage)

		audioID1 := uuid.New()
		audioID2 := uuid.New()

		orphanedAudios := []model.Audio{
			{ID: audioID1, Path: "audios/orphan1.mp3"},
			{ID: audioID2, Path: "audios/orphan2.mp3"},
		}

		mockAudioRepo.On("FindOrphaned", mock.Anything).Return(orphanedAudios, nil)
		mockImageRepo.On("FindOrphaned", mock.Anything).Return([]model.Image{}, nil)
		mockStorage.On("Delete", mock.Anything, "audios/orphan1.mp3").Return(errors.New("gcs error"))
		mockStorage.On("Delete", mock.Anything, "audios/orphan2.mp3").Return(nil)
		mockAudioRepo.On("Delete", mock.Anything, audioID2).Return(nil)

		result, err := svc.CleanupOrphanedMedia(ctx, false)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.DeletedAudioCount)
		assert.Equal(t, 1, result.FailedAudioCount)
	})

	t.Run("異常系: DB からの削除に失敗した場合はスキップして続行する", func(t *testing.T) {
		mockAudioRepo := new(mockAudioRepository)
		mockImageRepo := new(mockImageRepository)
		mockStorage := new(mockStorageClient)

		svc := NewCleanupService(mockAudioRepo, mockImageRepo, mockStorage)

		audioID1 := uuid.New()
		audioID2 := uuid.New()

		orphanedAudios := []model.Audio{
			{ID: audioID1, Path: "audios/orphan1.mp3"},
			{ID: audioID2, Path: "audios/orphan2.mp3"},
		}

		mockAudioRepo.On("FindOrphaned", mock.Anything).Return(orphanedAudios, nil)
		mockImageRepo.On("FindOrphaned", mock.Anything).Return([]model.Image{}, nil)
		mockStorage.On("Delete", mock.Anything, "audios/orphan1.mp3").Return(nil)
		mockStorage.On("Delete", mock.Anything, "audios/orphan2.mp3").Return(nil)
		mockAudioRepo.On("Delete", mock.Anything, audioID1).Return(errors.New("db error"))
		mockAudioRepo.On("Delete", mock.Anything, audioID2).Return(nil)

		result, err := svc.CleanupOrphanedMedia(ctx, false)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.DeletedAudioCount)
		assert.Equal(t, 1, result.FailedAudioCount)
	})

	t.Run("正常系: storageClient が nil の場合は GCS 削除をスキップする", func(t *testing.T) {
		mockAudioRepo := new(mockAudioRepository)
		mockImageRepo := new(mockImageRepository)

		svc := NewCleanupService(mockAudioRepo, mockImageRepo, nil)

		audioID1 := uuid.New()

		orphanedAudios := []model.Audio{
			{ID: audioID1, Path: "audios/orphan1.mp3"},
		}

		mockAudioRepo.On("FindOrphaned", mock.Anything).Return(orphanedAudios, nil)
		mockImageRepo.On("FindOrphaned", mock.Anything).Return([]model.Image{}, nil)
		mockAudioRepo.On("Delete", mock.Anything, audioID1).Return(nil)

		result, err := svc.CleanupOrphanedMedia(ctx, false)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.DeletedAudioCount)
	})
}
