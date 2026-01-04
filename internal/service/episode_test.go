package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/siropaca/anycast-backend/internal/model"
)

func TestToEpisodeResponse(t *testing.T) {
	now := time.Now()
	episodeID := uuid.New()
	channelID := uuid.New()
	audioID := uuid.New()
	description := "Test Description"
	scriptPrompt := "Test Script Prompt"

	baseEpisode := &model.Episode{
		ID:           episodeID,
		ChannelID:    channelID,
		Title:        "Test Episode",
		Description:  &description,
		ScriptPrompt: &scriptPrompt,
		PublishedAt:  &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	t.Run("基本的な変換が正しく行われる", func(t *testing.T) {
		resp := toEpisodeResponse(baseEpisode)

		assert.Equal(t, episodeID, resp.ID)
		assert.Equal(t, "Test Episode", resp.Title)
		assert.Equal(t, &description, resp.Description)
		assert.Equal(t, &scriptPrompt, resp.ScriptPrompt)
		assert.NotNil(t, resp.PublishedAt)
		assert.Equal(t, now, resp.CreatedAt)
		assert.Equal(t, now, resp.UpdatedAt)
	})

	t.Run("FullAudio が nil の場合、レスポンスの FullAudio も nil", func(t *testing.T) {
		episode := *baseEpisode
		episode.FullAudio = nil

		resp := toEpisodeResponse(&episode)

		assert.Nil(t, resp.FullAudio)
	})

	t.Run("FullAudio がある場合、正しく変換される", func(t *testing.T) {
		episode := *baseEpisode
		episode.FullAudioID = &audioID
		episode.FullAudio = &model.Audio{
			ID:         audioID,
			URL:        "https://example.com/audio.mp3",
			DurationMs: 180000,
		}

		resp := toEpisodeResponse(&episode)

		assert.NotNil(t, resp.FullAudio)
		assert.Equal(t, audioID, resp.FullAudio.ID)
		assert.Equal(t, "https://example.com/audio.mp3", resp.FullAudio.URL)
		assert.Equal(t, 180000, resp.FullAudio.DurationMs)
	})

	t.Run("Description が nil の場合、レスポンスの Description も nil", func(t *testing.T) {
		episode := *baseEpisode
		episode.Description = nil

		resp := toEpisodeResponse(&episode)

		assert.Nil(t, resp.Description)
	})

	t.Run("PublishedAt が nil の場合、レスポンスの PublishedAt も nil", func(t *testing.T) {
		episode := *baseEpisode
		episode.PublishedAt = nil

		resp := toEpisodeResponse(&episode)

		assert.Nil(t, resp.PublishedAt)
	})
}

func TestToEpisodeResponses(t *testing.T) {
	now := time.Now()
	channelID := uuid.New()
	desc1 := "Description 1"
	desc2 := "Description 2"
	prompt1 := "Prompt 1"
	prompt2 := "Prompt 2"

	episodes := []model.Episode{
		{
			ID:           uuid.New(),
			ChannelID:    channelID,
			Title:        "Episode 1",
			Description:  &desc1,
			ScriptPrompt: &prompt1,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           uuid.New(),
			ChannelID:    channelID,
			Title:        "Episode 2",
			Description:  &desc2,
			ScriptPrompt: &prompt2,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	t.Run("複数エピソードを正しく変換する", func(t *testing.T) {
		result := toEpisodeResponses(episodes)

		assert.Len(t, result, 2)
		assert.Equal(t, "Episode 1", result[0].Title)
		assert.Equal(t, "Episode 2", result[1].Title)
	})

	t.Run("scriptPrompt が含まれる", func(t *testing.T) {
		result := toEpisodeResponses(episodes)

		assert.Equal(t, &prompt1, result[0].ScriptPrompt)
		assert.Equal(t, &prompt2, result[1].ScriptPrompt)
	})

	t.Run("空のスライスの場合、空のスライスを返す", func(t *testing.T) {
		result := toEpisodeResponses([]model.Episode{})

		assert.Len(t, result, 0)
		assert.NotNil(t, result)
	})
}
