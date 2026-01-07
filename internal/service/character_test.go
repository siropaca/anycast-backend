package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/model"
)

func TestToCharacterWithChannelsResponses(t *testing.T) {
	now := time.Now()
	voiceID1 := uuid.New()
	voiceID2 := uuid.New()
	charID1 := uuid.New()
	charID2 := uuid.New()
	userID := uuid.New()
	channelID1 := uuid.New()
	channelID2 := uuid.New()

	characters := []model.Character{
		{
			ID:        charID1,
			UserID:    userID,
			Name:      "太郎",
			Persona:   "明るい性格",
			VoiceID:   voiceID1,
			CreatedAt: now,
			UpdatedAt: now,
			Voice: model.Voice{
				ID:       voiceID1,
				Name:     "ja-JP-Wavenet-C",
				Provider: "google",
				Gender:   model.GenderMale,
			},
			ChannelCharacters: []model.ChannelCharacter{
				{
					ID:          uuid.New(),
					ChannelID:   channelID1,
					CharacterID: charID1,
					Channel: model.Channel{
						ID:   channelID1,
						Name: "テストチャンネル1",
					},
				},
				{
					ID:          uuid.New(),
					ChannelID:   channelID2,
					CharacterID: charID1,
					Channel: model.Channel{
						ID:   channelID2,
						Name: "テストチャンネル2",
					},
				},
			},
		},
		{
			ID:        charID2,
			UserID:    userID,
			Name:      "花子",
			Persona:   "落ち着いた性格",
			VoiceID:   voiceID2,
			CreatedAt: now,
			UpdatedAt: now,
			Voice: model.Voice{
				ID:       voiceID2,
				Name:     "ja-JP-Wavenet-B",
				Provider: "google",
				Gender:   model.GenderFemale,
			},
			ChannelCharacters: []model.ChannelCharacter{},
		},
	}

	t.Run("複数キャラクターを正しく変換する", func(t *testing.T) {
		svc := &characterService{}
		ctx := context.Background()
		result, err := svc.toCharacterWithChannelsResponses(ctx, characters)

		assert.NoError(t, err)
		assert.Len(t, result, 2)

		// 1人目のキャラクター
		assert.Equal(t, charID1, result[0].ID)
		assert.Equal(t, "太郎", result[0].Name)
		assert.Equal(t, "明るい性格", result[0].Persona)
		assert.Equal(t, voiceID1, result[0].Voice.ID)
		assert.Equal(t, "ja-JP-Wavenet-C", result[0].Voice.Name)
		assert.Equal(t, "google", result[0].Voice.Provider)
		assert.Equal(t, "male", result[0].Voice.Gender)
		assert.Equal(t, now, result[0].CreatedAt)
		assert.Equal(t, now, result[0].UpdatedAt)

		// チャンネル情報
		assert.Len(t, result[0].Channels, 2)
		assert.Equal(t, channelID1, result[0].Channels[0].ID)
		assert.Equal(t, "テストチャンネル1", result[0].Channels[0].Name)
		assert.Equal(t, channelID2, result[0].Channels[1].ID)
		assert.Equal(t, "テストチャンネル2", result[0].Channels[1].Name)

		// 2人目のキャラクター
		assert.Equal(t, charID2, result[1].ID)
		assert.Equal(t, "花子", result[1].Name)
		assert.Equal(t, "落ち着いた性格", result[1].Persona)
		assert.Equal(t, voiceID2, result[1].Voice.ID)
		assert.Equal(t, "ja-JP-Wavenet-B", result[1].Voice.Name)
		assert.Equal(t, "google", result[1].Voice.Provider)
		assert.Equal(t, "female", result[1].Voice.Gender)

		// チャンネルなし
		assert.Len(t, result[1].Channels, 0)
	})

	t.Run("空のスライスの場合、空のスライスを返す", func(t *testing.T) {
		svc := &characterService{}
		ctx := context.Background()
		result, err := svc.toCharacterWithChannelsResponses(ctx, []model.Character{})

		assert.NoError(t, err)
		assert.Len(t, result, 0)
		assert.NotNil(t, result)
		assert.Equal(t, []response.CharacterWithChannelsResponse{}, result)
	})
}
