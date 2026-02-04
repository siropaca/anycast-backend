package script

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePhase2Output(t *testing.T) {
	validJSON := `{
		"grounding": {
			"definitions": [
				{"term": "AI", "definition": "人工知能"}
			],
			"examples": [
				{"id": "ex1", "situation": "テスト状況", "detail": "詳細"}
			],
			"pitfalls": [
				{"id": "pit1", "misconception": "誤解", "reality": "実際"}
			],
			"questions": [
				{"id": "q1", "question": "質問？"}
			],
			"action_steps": [
				{"id": "act1", "step": "やること"}
			]
		},
		"outline": {
			"opening": {"hook": "掴み"},
			"blocks": [
				{
					"block_number": 1,
					"topic": "トピック1",
					"example_ids": ["ex1"],
					"pitfall_ids": ["pit1"],
					"action_step_ids": ["act1"],
					"question_ids": ["q1"]
				},
				{
					"block_number": 2,
					"topic": "トピック2",
					"example_ids": ["ex1"],
					"pitfall_ids": ["pit1"],
					"action_step_ids": ["act1"],
					"question_ids": ["q1"]
				},
				{
					"block_number": 3,
					"topic": "トピック3",
					"example_ids": ["ex1"],
					"pitfall_ids": ["pit1"],
					"action_step_ids": ["act1"],
					"question_ids": ["q1"]
				}
			],
			"closing": {
				"summary": "まとめ",
				"takeaway": "持ち帰り"
			}
		}
	}`

	t.Run("正常な JSON をパースできる", func(t *testing.T) {
		output, err := ParsePhase2Output(validJSON)
		require.NoError(t, err)
		assert.Len(t, output.Outline.Blocks, 3)
		assert.Equal(t, "掴み", output.Outline.Opening.Hook)
		assert.Equal(t, "まとめ", output.Outline.Closing.Summary)
		assert.Len(t, output.Grounding.Examples, 1)
		assert.Len(t, output.Grounding.Pitfalls, 1)
		assert.Len(t, output.Grounding.ActionSteps, 1)
	})

	t.Run("コードブロック付きの出力をパースできる", func(t *testing.T) {
		text := "以下が結果です：\n```json\n" + validJSON + "\n```\n以上です。"
		output, err := ParsePhase2Output(text)
		require.NoError(t, err)
		assert.Len(t, output.Outline.Blocks, 3)
	})

	t.Run("outline が配列形式でもパースできる", func(t *testing.T) {
		arrayOutlineJSON := `{
			"grounding": {
				"definitions": [{"term": "AI", "definition": "人工知能"}],
				"examples": [{"id": "ex1", "situation": "s", "detail": "d"}],
				"pitfalls": [{"id": "pit1", "misconception": "m", "reality": "r"}],
				"questions": [{"id": "q1", "question": "q"}],
				"action_steps": [{"id": "act1", "step": "s"}]
			},
			"outline": [
				{
					"block_number": 1,
					"topic": "トピック1",
					"example_ids": ["ex1"],
					"pitfall_ids": ["pit1"],
					"action_step_ids": ["act1"],
					"question_ids": ["q1"]
				},
				{
					"block_number": 2,
					"topic": "トピック2",
					"example_ids": ["ex1"],
					"pitfall_ids": ["pit1"],
					"action_step_ids": ["act1"],
					"question_ids": ["q1"]
				},
				{
					"block_number": 3,
					"topic": "トピック3",
					"example_ids": ["ex1"],
					"pitfall_ids": ["pit1"],
					"action_step_ids": ["act1"],
					"question_ids": ["q1"]
				}
			]
		}`
		output, err := ParsePhase2Output(arrayOutlineJSON)
		require.NoError(t, err)
		assert.Len(t, output.Outline.Blocks, 3)
		assert.Equal(t, "トピック1", output.Outline.Blocks[0].Topic)
	})

	t.Run("ブロック数が3でなければエラー", func(t *testing.T) {
		badJSON := `{
			"grounding": {
				"definitions": [],
				"examples": [{"id": "ex1", "situation": "s", "detail": "d"}],
				"pitfalls": [{"id": "pit1", "misconception": "m", "reality": "r"}],
				"questions": [],
				"action_steps": [{"id": "act1", "step": "s"}]
			},
			"outline": {
				"opening": {"hook": "h"},
				"blocks": [
					{"block_number": 1, "topic": "t1", "example_ids": [], "pitfall_ids": [], "action_step_ids": [], "question_ids": []}
				],
				"closing": {"summary": "s", "takeaway": "t"}
			}
		}`
		_, err := ParsePhase2Output(badJSON)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ブロック数が3ではありません")
	})

	t.Run("具体例が空ならエラー", func(t *testing.T) {
		badJSON := `{
			"grounding": {
				"definitions": [],
				"examples": [],
				"pitfalls": [{"id": "pit1", "misconception": "m", "reality": "r"}],
				"questions": [],
				"action_steps": [{"id": "act1", "step": "s"}]
			},
			"outline": {
				"opening": {"hook": "h"},
				"blocks": [
					{"block_number": 1, "topic": "t1", "example_ids": [], "pitfall_ids": [], "action_step_ids": [], "question_ids": []},
					{"block_number": 2, "topic": "t2", "example_ids": [], "pitfall_ids": [], "action_step_ids": [], "question_ids": []},
					{"block_number": 3, "topic": "t3", "example_ids": [], "pitfall_ids": [], "action_step_ids": [], "question_ids": []}
				],
				"closing": {"summary": "s", "takeaway": "t"}
			}
		}`
		_, err := ParsePhase2Output(badJSON)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "具体例が空")
	})

	t.Run("不正な JSON はエラー", func(t *testing.T) {
		_, err := ParsePhase2Output("これは JSON ではありません")
		assert.Error(t, err)
	})

	t.Run("不正な JSON 構造はエラー", func(t *testing.T) {
		_, err := ParsePhase2Output(`{"invalid": true}`)
		assert.Error(t, err)
	})
}
