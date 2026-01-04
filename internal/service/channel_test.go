package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/siropaca/anycast-backend/internal/model"
)

func TestToChannelResponse(t *testing.T) {
	now := time.Now()
	categoryID := uuid.New()
	channelID := uuid.New()
	userID := uuid.New()
	artworkID := uuid.New()
	voiceID := uuid.New()
	characterID := uuid.New()

	baseChannel := &model.Channel{
		ID:          channelID,
		UserID:      userID,
		Name:        "Test Channel",
		Description: "Test Description",
		UserPrompt:  "Test User Prompt",
		CategoryID:  categoryID,
		PublishedAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
		Category: model.Category{
			ID:   categoryID,
			Slug: "technology",
			Name: "テクノロジー",
		},
		Characters: []model.Character{
			{
				ID:      characterID,
				Name:    "太郎",
				Persona: "明るい性格",
				VoiceID: voiceID,
				Voice: model.Voice{
					ID:     voiceID,
					Name:   "ja-JP-Wavenet-C",
					Gender: model.GenderMale,
				},
			},
		},
	}

	t.Run("isOwner が true の場合、userPrompt が含まれる", func(t *testing.T) {
		resp := toChannelResponse(baseChannel, true)

		assert.Equal(t, channelID, resp.ID)
		assert.Equal(t, "Test Channel", resp.Name)
		assert.Equal(t, "Test Description", resp.Description)
		assert.Equal(t, "Test User Prompt", resp.UserPrompt)
		assert.Equal(t, categoryID, resp.Category.ID)
		assert.Equal(t, "technology", resp.Category.Slug)
		assert.Equal(t, "テクノロジー", resp.Category.Name)
		assert.NotNil(t, resp.PublishedAt)
		assert.Len(t, resp.Characters, 1)
		assert.Equal(t, characterID, resp.Characters[0].ID)
		assert.Equal(t, "太郎", resp.Characters[0].Name)
		assert.Equal(t, "明るい性格", resp.Characters[0].Persona)
		assert.Equal(t, voiceID, resp.Characters[0].Voice.ID)
		assert.Equal(t, "ja-JP-Wavenet-C", resp.Characters[0].Voice.Name)
		assert.Equal(t, "male", resp.Characters[0].Voice.Gender)
	})

	t.Run("isOwner が false の場合、userPrompt が空文字になる", func(t *testing.T) {
		resp := toChannelResponse(baseChannel, false)

		assert.Equal(t, channelID, resp.ID)
		assert.Equal(t, "Test Channel", resp.Name)
		assert.Equal(t, "Test Description", resp.Description)
		assert.Equal(t, "", resp.UserPrompt)
	})

	t.Run("Artwork が nil の場合、レスポンスの Artwork も nil", func(t *testing.T) {
		channel := *baseChannel
		channel.Artwork = nil

		resp := toChannelResponse(&channel, true)

		assert.Nil(t, resp.Artwork)
	})

	t.Run("Artwork がある場合、正しく変換される", func(t *testing.T) {
		channel := *baseChannel
		channel.ArtworkID = &artworkID
		channel.Artwork = &model.Image{
			ID:  artworkID,
			URL: "https://example.com/artwork.png",
		}

		resp := toChannelResponse(&channel, true)

		assert.NotNil(t, resp.Artwork)
		assert.Equal(t, artworkID, resp.Artwork.ID)
		assert.Equal(t, "https://example.com/artwork.png", resp.Artwork.URL)
	})

	t.Run("PublishedAt が nil の場合、レスポンスの PublishedAt も nil", func(t *testing.T) {
		channel := *baseChannel
		channel.PublishedAt = nil

		resp := toChannelResponse(&channel, true)

		assert.Nil(t, resp.PublishedAt)
	})
}

func TestToChannelResponses(t *testing.T) {
	now := time.Now()
	categoryID := uuid.New()

	channels := []model.Channel{
		{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			Name:        "Channel 1",
			Description: "Description 1",
			UserPrompt:  "Prompt 1",
			CategoryID:  categoryID,
			CreatedAt:   now,
			UpdatedAt:   now,
			Category: model.Category{
				ID:   categoryID,
				Slug: "tech",
				Name: "Tech",
			},
		},
		{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			Name:        "Channel 2",
			Description: "Description 2",
			UserPrompt:  "Prompt 2",
			CategoryID:  categoryID,
			CreatedAt:   now,
			UpdatedAt:   now,
			Category: model.Category{
				ID:   categoryID,
				Slug: "tech",
				Name: "Tech",
			},
		},
	}

	t.Run("複数チャンネルを正しく変換する", func(t *testing.T) {
		result := toChannelResponses(channels)

		assert.Len(t, result, 2)
		assert.Equal(t, "Channel 1", result[0].Name)
		assert.Equal(t, "Channel 2", result[1].Name)
	})

	t.Run("オーナーとして扱われるため userPrompt が含まれる", func(t *testing.T) {
		result := toChannelResponses(channels)

		assert.Equal(t, "Prompt 1", result[0].UserPrompt)
		assert.Equal(t, "Prompt 2", result[1].UserPrompt)
	})

	t.Run("空のスライスの場合、空のスライスを返す", func(t *testing.T) {
		result := toChannelResponses([]model.Channel{})

		assert.Len(t, result, 0)
		assert.NotNil(t, result)
	})
}
