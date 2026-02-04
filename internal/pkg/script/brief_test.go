package script

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectTalkMode(t *testing.T) {
	t.Run("0人は monologue", func(t *testing.T) {
		assert.Equal(t, TalkModeMonologue, DetectTalkMode(0))
	})

	t.Run("1人は monologue", func(t *testing.T) {
		assert.Equal(t, TalkModeMonologue, DetectTalkMode(1))
	})

	t.Run("2人は dialogue", func(t *testing.T) {
		assert.Equal(t, TalkModeDialogue, DetectTalkMode(2))
	})

	t.Run("3人は dialogue", func(t *testing.T) {
		assert.Equal(t, TalkModeDialogue, DetectTalkMode(3))
	})
}

func TestNormalizeBrief(t *testing.T) {
	t.Run("2人のキャラクターで dialogue になる", func(t *testing.T) {
		input := BriefInput{
			EpisodeTitle:       "テストエピソード",
			EpisodeDescription: "テスト説明",
			EpisodeGoal:        "テスト目標",
			DurationMinutes:    10,
			ChannelName:        "テストチャンネル",
			ChannelDescription: "チャンネル説明",
			ChannelCategory:    "テクノロジー",
			ChannelStyleGuide:  "カジュアルに",
			Characters: []BriefInputCharacter{
				{Name: "太郎", Gender: "male", Persona: "明るい性格"},
				{Name: "花子", Gender: "female", Persona: "知的な性格"},
			},
			MasterGuide: "ユーザー指示",
			Theme:       "AIの未来",
			WithEmotion: true,
		}

		brief := NormalizeBrief(input)

		assert.Equal(t, "テストエピソード", brief.Episode.Title)
		assert.Equal(t, "テスト説明", brief.Episode.Description)
		assert.Equal(t, "テスト目標", brief.Episode.Goal)
		assert.Equal(t, 10, brief.Episode.DurationMinutes)
		assert.Equal(t, "テストチャンネル", brief.Channel.Name)
		assert.Equal(t, "チャンネル説明", brief.Channel.Description)
		assert.Equal(t, "テクノロジー", brief.Channel.Category)
		assert.Equal(t, "カジュアルに", brief.Channel.StyleGuide)
		assert.Len(t, brief.Characters, 2)
		assert.Equal(t, "太郎", brief.Characters[0].Name)
		assert.Equal(t, "male", brief.Characters[0].Gender)
		assert.Equal(t, "明るい性格", brief.Characters[0].Persona)
		assert.Equal(t, "花子", brief.Characters[1].Name)
		assert.Equal(t, "ユーザー指示", brief.MasterGuide)
		assert.Equal(t, "AIの未来", brief.Theme)
		assert.Equal(t, TalkModeDialogue, brief.Constraints.TalkMode)
		assert.True(t, brief.Constraints.WithEmotion)
		assert.True(t, brief.Constraints.TTSOptimized)
		assert.Empty(t, brief.Constraints.Avoid)
	})

	t.Run("1人のキャラクターで monologue になる", func(t *testing.T) {
		input := BriefInput{
			EpisodeTitle:    "ソロエピソード",
			DurationMinutes: 5,
			ChannelName:     "ソロチャンネル",
			ChannelCategory: "教育",
			Characters: []BriefInputCharacter{
				{Name: "太郎", Gender: "male"},
			},
			Theme:       "プログラミング入門",
			WithEmotion: false,
		}

		brief := NormalizeBrief(input)

		assert.Equal(t, TalkModeMonologue, brief.Constraints.TalkMode)
		assert.Len(t, brief.Characters, 1)
		assert.False(t, brief.Constraints.WithEmotion)
	})

	t.Run("空の省略可能フィールド", func(t *testing.T) {
		input := BriefInput{
			EpisodeTitle:    "最小構成",
			DurationMinutes: 10,
			ChannelName:     "テスト",
			ChannelCategory: "その他",
			Characters: []BriefInputCharacter{
				{Name: "太郎", Gender: "male"},
			},
			Theme: "テスト",
		}

		brief := NormalizeBrief(input)

		assert.Empty(t, brief.Episode.Description)
		assert.Empty(t, brief.Episode.Goal)
		assert.Empty(t, brief.Channel.Description)
		assert.Empty(t, brief.Channel.StyleGuide)
		assert.Empty(t, brief.MasterGuide)
	})
}

func TestBrief_ToJSON(t *testing.T) {
	t.Run("正常に JSON 変換できる", func(t *testing.T) {
		brief := Brief{
			Episode: BriefEpisode{
				Title:           "テスト",
				DurationMinutes: 10,
			},
			Channel: BriefChannel{
				Name:     "テスト",
				Category: "テクノロジー",
			},
			Characters: []BriefCharacter{
				{Name: "太郎", Gender: "male"},
			},
			Theme: "テスト",
			Constraints: BriefConstraints{
				TalkMode:     TalkModeDialogue,
				TTSOptimized: true,
				Avoid:        []string{},
			},
		}

		jsonStr, err := brief.ToJSON()
		require.NoError(t, err)

		// JSON として有効か確認
		var parsed map[string]interface{}
		err = json.Unmarshal([]byte(jsonStr), &parsed)
		require.NoError(t, err)
		assert.Contains(t, parsed, "episode")
		assert.Contains(t, parsed, "channel")
		assert.Contains(t, parsed, "characters")
		assert.Contains(t, parsed, "theme")
		assert.Contains(t, parsed, "constraints")
	})
}
