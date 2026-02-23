package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// episodeService 用のモック（channel_test.go と同じパッケージなので重複定義はできない）
// channel_test.go で定義した mockStorageClient を再利用

func TestToEpisodeResponse(t *testing.T) {
	now := time.Now()
	episodeID := uuid.New()
	channelID := uuid.New()
	ownerID := uuid.New()
	audioID := uuid.New()
	artworkID := uuid.New()
	description := "Test Description"

	owner := &model.User{
		ID:          ownerID,
		Username:    "testowner",
		DisplayName: "Test Owner",
	}

	baseEpisode := &model.Episode{
		ID:          episodeID,
		ChannelID:   channelID,
		Title:       "Test Episode",
		Description: description,
		PublishedAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	t.Run("基本的な変換が正しく行われる", func(t *testing.T) {
		mockStorage := new(mockStorageClient)
		svc := &episodeService{storageClient: mockStorage}
		ctx := context.Background()

		resp, err := svc.toEpisodeResponse(ctx, baseEpisode, owner, nil)

		assert.NoError(t, err)
		assert.Equal(t, episodeID, resp.ID)
		assert.Equal(t, ownerID, resp.Owner.ID)
		assert.Equal(t, "testowner", resp.Owner.Username)
		assert.Equal(t, "Test Owner", resp.Owner.DisplayName)
		assert.Nil(t, resp.Owner.Avatar)
		assert.Equal(t, "Test Episode", resp.Title)
		assert.Equal(t, description, resp.Description)
		assert.NotNil(t, resp.PublishedAt)
		assert.Equal(t, now, resp.CreatedAt)
		assert.Equal(t, now, resp.UpdatedAt)
	})

	t.Run("FullAudio が nil の場合、レスポンスの FullAudio も nil", func(t *testing.T) {
		mockStorage := new(mockStorageClient)
		svc := &episodeService{storageClient: mockStorage}
		ctx := context.Background()

		episode := *baseEpisode
		episode.FullAudio = nil

		resp, err := svc.toEpisodeResponse(ctx, &episode, owner, nil)

		assert.NoError(t, err)
		assert.Nil(t, resp.FullAudio)
	})

	t.Run("FullAudio がある場合、署名 URL が生成される", func(t *testing.T) {
		mockStorage := new(mockStorageClient)
		mockStorage.On("GenerateSignedURL", mock.Anything, "audios/full-audio.mp3", storage.SignedURLExpirationAudio).Return("https://signed-url.example.com/full-audio.mp3", nil)
		svc := &episodeService{storageClient: mockStorage}
		ctx := context.Background()

		episode := *baseEpisode
		episode.FullAudioID = &audioID
		episode.FullAudio = &model.Audio{
			ID:         audioID,
			Path:       "audios/full-audio.mp3",
			MimeType:   "audio/mpeg",
			FileSize:   1024000,
			DurationMs: 180000,
		}

		resp, err := svc.toEpisodeResponse(ctx, &episode, owner, nil)

		assert.NoError(t, err)
		assert.NotNil(t, resp.FullAudio)
		assert.Equal(t, audioID, resp.FullAudio.ID)
		assert.Equal(t, "https://signed-url.example.com/full-audio.mp3", resp.FullAudio.URL)
		assert.Equal(t, 180000, resp.FullAudio.DurationMs)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Artwork がある場合、署名 URL が生成される", func(t *testing.T) {
		mockStorage := new(mockStorageClient)
		mockStorage.On("GenerateSignedURL", mock.Anything, "images/artwork.png", storage.SignedURLExpirationImage).Return("https://signed-url.example.com/artwork.png", nil)
		svc := &episodeService{storageClient: mockStorage}
		ctx := context.Background()

		episode := *baseEpisode
		episode.ArtworkID = &artworkID
		episode.Artwork = &model.Image{
			ID:   artworkID,
			Path: "images/artwork.png",
		}

		resp, err := svc.toEpisodeResponse(ctx, &episode, owner, nil)

		assert.NoError(t, err)
		assert.NotNil(t, resp.Artwork)
		assert.Equal(t, artworkID, resp.Artwork.ID)
		assert.Equal(t, "https://signed-url.example.com/artwork.png", resp.Artwork.URL)
		mockStorage.AssertExpectations(t)
	})

	t.Run("PublishedAt が nil の場合、レスポンスの PublishedAt も nil", func(t *testing.T) {
		mockStorage := new(mockStorageClient)
		svc := &episodeService{storageClient: mockStorage}
		ctx := context.Background()

		episode := *baseEpisode
		episode.PublishedAt = nil

		resp, err := svc.toEpisodeResponse(ctx, &episode, owner, nil)

		assert.NoError(t, err)
		assert.Nil(t, resp.PublishedAt)
	})
}

func TestToEpisodeResponses(t *testing.T) {
	now := time.Now()
	channelID := uuid.New()
	ownerID := uuid.New()

	owner := &model.User{
		ID:          ownerID,
		Username:    "testowner",
		DisplayName: "Test Owner",
	}

	episodes := []model.Episode{
		{
			ID:          uuid.New(),
			ChannelID:   channelID,
			Title:       "Episode 1",
			Description: "Description 1",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			ChannelID:   channelID,
			Title:       "Episode 2",
			Description: "Description 2",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	t.Run("複数エピソードを正しく変換する", func(t *testing.T) {
		mockStorage := new(mockStorageClient)
		mockScriptLineRepo := new(mockScriptLineRepository)
		mockScriptLineRepo.On("CountByEpisodeIDs", mock.Anything, mock.Anything).Return(map[uuid.UUID]int{
			episodes[0].ID: 5,
			episodes[1].ID: 3,
		}, nil)
		svc := &episodeService{storageClient: mockStorage, scriptLineRepo: mockScriptLineRepo}
		ctx := context.Background()

		result, err := svc.toEpisodeResponses(ctx, episodes, owner)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "Episode 1", result[0].Title)
		assert.Equal(t, "Episode 2", result[1].Title)
		assert.Equal(t, 5, result[0].ScriptLineCount)
		assert.Equal(t, 3, result[1].ScriptLineCount)
	})

	t.Run("空のスライスの場合、空のスライスを返す", func(t *testing.T) {
		mockStorage := new(mockStorageClient)
		mockScriptLineRepo := new(mockScriptLineRepository)
		mockScriptLineRepo.On("CountByEpisodeIDs", mock.Anything, mock.Anything).Return(map[uuid.UUID]int{}, nil)
		svc := &episodeService{storageClient: mockStorage, scriptLineRepo: mockScriptLineRepo}
		ctx := context.Background()

		result, err := svc.toEpisodeResponses(ctx, []model.Episode{}, owner)

		assert.NoError(t, err)
		assert.Len(t, result, 0)
		assert.NotNil(t, result)
	})
}
