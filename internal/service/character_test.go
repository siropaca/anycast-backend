package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/model"
)

func TestToCharacterResponses(t *testing.T) {
	now := time.Now()
	voiceID1 := uuid.New()
	voiceID2 := uuid.New()
	charID1 := uuid.New()
	charID2 := uuid.New()
	userID := uuid.New()

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
		},
	}

	t.Run("複数キャラクターを正しく変換する", func(t *testing.T) {
		svc := &characterService{}
		result := svc.toCharacterResponses(characters)

		assert.Len(t, result, 2)
		assert.Equal(t, charID1, result[0].ID)
		assert.Equal(t, "太郎", result[0].Name)
		assert.Equal(t, "明るい性格", result[0].Persona)
		assert.Equal(t, voiceID1, result[0].Voice.ID)
		assert.Equal(t, "ja-JP-Wavenet-C", result[0].Voice.Name)
		assert.Equal(t, "google", result[0].Voice.Provider)
		assert.Equal(t, "male", result[0].Voice.Gender)
		assert.Equal(t, now, result[0].CreatedAt)
		assert.Equal(t, now, result[0].UpdatedAt)

		assert.Equal(t, charID2, result[1].ID)
		assert.Equal(t, "花子", result[1].Name)
		assert.Equal(t, "落ち着いた性格", result[1].Persona)
		assert.Equal(t, voiceID2, result[1].Voice.ID)
		assert.Equal(t, "ja-JP-Wavenet-B", result[1].Voice.Name)
		assert.Equal(t, "google", result[1].Voice.Provider)
		assert.Equal(t, "female", result[1].Voice.Gender)
	})

	t.Run("空のスライスの場合、空のスライスを返す", func(t *testing.T) {
		svc := &characterService{}
		result := svc.toCharacterResponses([]model.Character{})

		assert.Len(t, result, 0)
		assert.NotNil(t, result)
		assert.Equal(t, []response.CharacterResponse{}, result)
	})
}
