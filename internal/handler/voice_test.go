package handler

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/siropaca/anycast-backend/internal/model"
)

func TestToVoiceResponse(t *testing.T) {
	t.Run("Voice モデルを VoiceResponse に変換できる", func(t *testing.T) {
		id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		voice := &model.Voice{
			ID:              id,
			Provider:        "google",
			ProviderVoiceID: "en-US-Neural2-A",
			Name:            "American English Female",
			Gender:          "female",
			IsActive:        true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		resp := toVoiceResponse(voice)

		assert.Equal(t, id, resp.ID)
		assert.Equal(t, "google", resp.Provider)
		assert.Equal(t, "en-US-Neural2-A", resp.ProviderVoiceID)
		assert.Equal(t, "American English Female", resp.Name)
		assert.Equal(t, "female", resp.Gender)
		assert.True(t, resp.IsActive)
	})

	t.Run("IsActive が false の場合も正しく変換される", func(t *testing.T) {
		voice := &model.Voice{
			ID:       uuid.New(),
			IsActive: false,
		}

		resp := toVoiceResponse(voice)

		assert.False(t, resp.IsActive)
	})
}

func TestToVoiceResponses(t *testing.T) {
	t.Run("空のスライスを変換すると空のスライスを返す", func(t *testing.T) {
		voices := []model.Voice{}

		resp := toVoiceResponses(voices)

		assert.Empty(t, resp)
	})

	t.Run("複数の Voice を変換できる", func(t *testing.T) {
		id1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		id2 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
		voices := []model.Voice{
			{
				ID:              id1,
				Provider:        "google",
				ProviderVoiceID: "voice-1",
				Name:            "Voice 1",
				Gender:          "male",
				IsActive:        true,
			},
			{
				ID:              id2,
				Provider:        "amazon",
				ProviderVoiceID: "voice-2",
				Name:            "Voice 2",
				Gender:          "female",
				IsActive:        false,
			},
		}

		resp := toVoiceResponses(voices)

		assert.Len(t, resp, 2)
		assert.Equal(t, id1, resp[0].ID)
		assert.Equal(t, "google", resp[0].Provider)
		assert.Equal(t, id2, resp[1].ID)
		assert.Equal(t, "amazon", resp[1].Provider)
	})

	t.Run("変換結果の長さが入力と一致する", func(t *testing.T) {
		voices := make([]model.Voice, 5)
		for i := range voices {
			voices[i] = model.Voice{ID: uuid.New()}
		}

		resp := toVoiceResponses(voices)

		assert.Len(t, resp, len(voices))
	})
}
