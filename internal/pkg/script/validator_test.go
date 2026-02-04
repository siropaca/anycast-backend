package script

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate_LineLengths(t *testing.T) {
	t.Run("正常な長さのセリフは合格", func(t *testing.T) {
		lines := []ParsedLine{
			{SpeakerName: "太郎", Text: "これは正常な長さのセリフですね"},
			{SpeakerName: "花子", Text: "はい、そうですね。十分な長さがありますね"},
		}
		config := ValidatorConfig{TalkMode: TalkModeDialogue, DurationMinutes: 1}
		result := Validate(lines, config)

		hasLineLengthIssue := false
		for _, issue := range result.Issues {
			if issue.Check == "line_length" {
				hasLineLengthIssue = true
			}
		}
		assert.False(t, hasLineLengthIssue)
	})

	t.Run("短すぎるセリフは不合格", func(t *testing.T) {
		lines := []ParsedLine{
			{SpeakerName: "太郎", Text: "短い"},
		}
		config := ValidatorConfig{TalkMode: TalkModeMonologue, DurationMinutes: 1}
		result := Validate(lines, config)

		hasLineLengthIssue := false
		for _, issue := range result.Issues {
			if issue.Check == "line_length" {
				hasLineLengthIssue = true
			}
		}
		assert.True(t, hasLineLengthIssue)
	})

	t.Run("長すぎるセリフは不合格", func(t *testing.T) {
		longText := strings.Repeat("あ", 121)
		lines := []ParsedLine{
			{SpeakerName: "太郎", Text: longText},
		}
		config := ValidatorConfig{TalkMode: TalkModeMonologue, DurationMinutes: 1}
		result := Validate(lines, config)

		hasLineLengthIssue := false
		for _, issue := range result.Issues {
			if issue.Check == "line_length" {
				hasLineLengthIssue = true
			}
		}
		assert.True(t, hasLineLengthIssue)
	})
}

func TestValidate_TrailingPeriod(t *testing.T) {
	t.Run("句点なしは合格", func(t *testing.T) {
		lines := []ParsedLine{
			{SpeakerName: "太郎", Text: "これは句点なしのセリフです"},
		}
		config := ValidatorConfig{TalkMode: TalkModeMonologue, DurationMinutes: 1}
		result := Validate(lines, config)

		hasTrailingPeriod := false
		for _, issue := range result.Issues {
			if issue.Check == "trailing_period" {
				hasTrailingPeriod = true
			}
		}
		assert.False(t, hasTrailingPeriod)
	})

	t.Run("句点ありは不合格", func(t *testing.T) {
		lines := []ParsedLine{
			{SpeakerName: "太郎", Text: "これは句点ありのセリフです。"},
		}
		config := ValidatorConfig{TalkMode: TalkModeMonologue, DurationMinutes: 1}
		result := Validate(lines, config)

		hasTrailingPeriod := false
		for _, issue := range result.Issues {
			if issue.Check == "trailing_period" {
				hasTrailingPeriod = true
			}
		}
		assert.True(t, hasTrailingPeriod)
	})
}

func TestValidate_PeriodInText(t *testing.T) {
	t.Run("セリフ中に句点なしは合格", func(t *testing.T) {
		lines := []ParsedLine{
			{SpeakerName: "太郎", Text: "これはセリフ中に句点がないテストです"},
		}
		config := ValidatorConfig{TalkMode: TalkModeMonologue, DurationMinutes: 1}
		result := Validate(lines, config)

		hasPeriodInText := false
		for _, issue := range result.Issues {
			if issue.Check == "period_in_text" {
				hasPeriodInText = true
			}
		}
		assert.False(t, hasPeriodInText)
	})

	t.Run("セリフ中に句点ありは不合格", func(t *testing.T) {
		lines := []ParsedLine{
			{SpeakerName: "太郎", Text: "これは前半です。これは後半です"},
		}
		config := ValidatorConfig{TalkMode: TalkModeMonologue, DurationMinutes: 1}
		result := Validate(lines, config)

		hasPeriodInText := false
		for _, issue := range result.Issues {
			if issue.Check == "period_in_text" {
				hasPeriodInText = true
			}
		}
		assert.True(t, hasPeriodInText)
	})
}

func TestValidate_MinimumLines(t *testing.T) {
	t.Run("十分な行数は合格", func(t *testing.T) {
		lines := make([]ParsedLine, 40)
		for i := range lines {
			lines[i] = ParsedLine{SpeakerName: "太郎", Text: "これはテスト用のセリフですよ"}
		}
		config := ValidatorConfig{TalkMode: TalkModeMonologue, DurationMinutes: 10}
		result := Validate(lines, config)

		hasMinLines := false
		for _, issue := range result.Issues {
			if issue.Check == "minimum_lines" {
				hasMinLines = true
			}
		}
		assert.False(t, hasMinLines)
	})

	t.Run("行数不足は不合格", func(t *testing.T) {
		lines := []ParsedLine{
			{SpeakerName: "太郎", Text: "これはテスト用のセリフです"},
		}
		config := ValidatorConfig{TalkMode: TalkModeMonologue, DurationMinutes: 10}
		result := Validate(lines, config)

		hasMinLines := false
		for _, issue := range result.Issues {
			if issue.Check == "minimum_lines" {
				hasMinLines = true
			}
		}
		assert.True(t, hasMinLines)
	})
}

func TestValidate_LengthVariance(t *testing.T) {
	t.Run("ゆらぎがある場合は合格", func(t *testing.T) {
		lines := []ParsedLine{
			{SpeakerName: "太郎", Text: "短いセリフだよね"},
			{SpeakerName: "花子", Text: "これはもう少し長めのセリフですね、いい感じですよ"},
			{SpeakerName: "太郎", Text: "うん"},
			{SpeakerName: "花子", Text: "かなり長めのセリフを入れて標準偏差を上げるためにたくさん書きます"},
		}
		config := ValidatorConfig{TalkMode: TalkModeDialogue, DurationMinutes: 1}
		result := Validate(lines, config)

		hasVariance := false
		for _, issue := range result.Issues {
			if issue.Check == "length_variance" {
				hasVariance = true
			}
		}
		assert.False(t, hasVariance)
	})

	t.Run("同じ長さのセリフばかりは不合格", func(t *testing.T) {
		lines := []ParsedLine{
			{SpeakerName: "太郎", Text: "これはテスト用のセリフです"},
			{SpeakerName: "花子", Text: "これはテスト用のセリフです"},
			{SpeakerName: "太郎", Text: "これはテスト用のセリフです"},
			{SpeakerName: "花子", Text: "これはテスト用のセリフです"},
		}
		config := ValidatorConfig{TalkMode: TalkModeDialogue, DurationMinutes: 1}
		result := Validate(lines, config)

		hasVariance := false
		for _, issue := range result.Issues {
			if issue.Check == "length_variance" {
				hasVariance = true
			}
		}
		assert.True(t, hasVariance)
	})
}

func TestValidate_ConsecutiveSpeaker(t *testing.T) {
	t.Run("3行連続は合格", func(t *testing.T) {
		lines := []ParsedLine{
			{SpeakerName: "太郎", Text: "これはテスト用のセリフです"},
			{SpeakerName: "太郎", Text: "まだ太郎が話していますよ"},
			{SpeakerName: "太郎", Text: "三行目ですけど大丈夫ですね"},
			{SpeakerName: "花子", Text: "ここで花子に交代しますよ"},
		}
		config := ValidatorConfig{TalkMode: TalkModeDialogue, DurationMinutes: 1}
		result := Validate(lines, config)

		hasConsecutive := false
		for _, issue := range result.Issues {
			if issue.Check == "consecutive_speaker" {
				hasConsecutive = true
			}
		}
		assert.False(t, hasConsecutive)
	})

	t.Run("4行連続は不合格", func(t *testing.T) {
		lines := []ParsedLine{
			{SpeakerName: "太郎", Text: "これはテスト用のセリフです"},
			{SpeakerName: "太郎", Text: "まだ太郎が話していますよ"},
			{SpeakerName: "太郎", Text: "三行目ですけどまだ続けます"},
			{SpeakerName: "太郎", Text: "四行目、これはアウトですね"},
		}
		config := ValidatorConfig{TalkMode: TalkModeDialogue, DurationMinutes: 1}
		result := Validate(lines, config)

		hasConsecutive := false
		for _, issue := range result.Issues {
			if issue.Check == "consecutive_speaker" {
				hasConsecutive = true
			}
		}
		assert.True(t, hasConsecutive)
	})

	t.Run("monologue では連続チェックしない", func(t *testing.T) {
		lines := []ParsedLine{
			{SpeakerName: "太郎", Text: "これはテスト用のセリフです"},
			{SpeakerName: "太郎", Text: "まだ太郎が話していますよ"},
			{SpeakerName: "太郎", Text: "三行目ですけどまだ続けます"},
			{SpeakerName: "太郎", Text: "四行目、monologue だから大丈夫"},
		}
		config := ValidatorConfig{TalkMode: TalkModeMonologue, DurationMinutes: 1}
		result := Validate(lines, config)

		hasConsecutive := false
		for _, issue := range result.Issues {
			if issue.Check == "consecutive_speaker" {
				hasConsecutive = true
			}
		}
		assert.False(t, hasConsecutive)
	})
}

func TestValidate_SpeakerBalance(t *testing.T) {
	t.Run("バランスの取れた話者は合格", func(t *testing.T) {
		lines := []ParsedLine{
			{SpeakerName: "太郎", Text: "これはテスト用のセリフです"},
			{SpeakerName: "花子", Text: "はい、花子のセリフですよ"},
			{SpeakerName: "太郎", Text: "再び太郎のセリフですよ"},
			{SpeakerName: "花子", Text: "また花子が話しています"},
			{SpeakerName: "太郎", Text: "太郎が最後に話しますね"},
		}
		config := ValidatorConfig{TalkMode: TalkModeDialogue, DurationMinutes: 1}
		result := Validate(lines, config)

		hasBalance := false
		for _, issue := range result.Issues {
			if issue.Check == "speaker_balance" {
				hasBalance = true
			}
		}
		assert.False(t, hasBalance)
	})

	t.Run("偏った話者は不合格", func(t *testing.T) {
		lines := []ParsedLine{
			{SpeakerName: "太郎", Text: "これはテスト用のセリフです"},
			{SpeakerName: "太郎", Text: "まだ太郎が話していますよ"},
			{SpeakerName: "太郎", Text: "三行目も太郎のセリフです"},
			{SpeakerName: "太郎", Text: "四行目も太郎が話してます"},
			{SpeakerName: "太郎", Text: "五行目もまだ太郎ですよ"},
			{SpeakerName: "花子", Text: "やっと花子の番がきましたね"},
		}
		config := ValidatorConfig{TalkMode: TalkModeDialogue, DurationMinutes: 1}
		result := Validate(lines, config)

		hasBalance := false
		for _, issue := range result.Issues {
			if issue.Check == "speaker_balance" {
				hasBalance = true
			}
		}
		assert.True(t, hasBalance)
	})
}

func TestValidate_SpeakerConsistency(t *testing.T) {
	t.Run("monologue で同一話者は合格", func(t *testing.T) {
		lines := []ParsedLine{
			{SpeakerName: "太郎", Text: "これはテスト用のセリフです"},
			{SpeakerName: "太郎", Text: "まだ太郎が話していますよ"},
			{SpeakerName: "太郎", Text: "ずっと太郎のセリフですね"},
		}
		config := ValidatorConfig{TalkMode: TalkModeMonologue, DurationMinutes: 1}
		result := Validate(lines, config)

		hasConsistency := false
		for _, issue := range result.Issues {
			if issue.Check == "speaker_consistency" {
				hasConsistency = true
			}
		}
		assert.False(t, hasConsistency)
	})

	t.Run("monologue で複数話者は不合格", func(t *testing.T) {
		lines := []ParsedLine{
			{SpeakerName: "太郎", Text: "これはテスト用のセリフです"},
			{SpeakerName: "花子", Text: "花子が混ざってしまいました"},
		}
		config := ValidatorConfig{TalkMode: TalkModeMonologue, DurationMinutes: 1}
		result := Validate(lines, config)

		hasConsistency := false
		for _, issue := range result.Issues {
			if issue.Check == "speaker_consistency" {
				hasConsistency = true
			}
		}
		assert.True(t, hasConsistency)
	})

	t.Run("dialogue では一貫性チェックしない", func(t *testing.T) {
		lines := []ParsedLine{
			{SpeakerName: "太郎", Text: "これはテスト用のセリフです"},
			{SpeakerName: "花子", Text: "dialogue なら問題ありません"},
		}
		config := ValidatorConfig{TalkMode: TalkModeDialogue, DurationMinutes: 1}
		result := Validate(lines, config)

		hasConsistency := false
		for _, issue := range result.Issues {
			if issue.Check == "speaker_consistency" {
				hasConsistency = true
			}
		}
		assert.False(t, hasConsistency)
	})
}

func TestValidate_AllPass(t *testing.T) {
	t.Run("全チェック合格で Passed が true", func(t *testing.T) {
		// 5分 × 4行 = 20行以上必要
		// 文長にゆらぎを持たせる（標準偏差5文字以上）
		lines := make([]ParsedLine, 0, 24)
		texts := []string{
			"これは最初のセリフですね",
			"はい、そうですよ、今日もよろしくお願いしますね、頑張っていきましょう",
			"今日のテーマについて話しましょうか",
			"えーと、具体的にはどういうことなんでしょう、もう少し詳しく教えてもらえるとありがたいです",
			"まあ、簡単に言うとこういうことなんです",
			"なるほどね、それはすごく面白い話ですね、初めて知りましたよ",
			"そうなんです、意外でしょう",
			"ちょっと驚きましたね、それは初めて聞きましたし本当にびっくりしています",
			"もう少し教えてくれますか",
			"もちろん、具体例を出すとこんな感じですね、分かりやすいでしょう",
			"なんか、すごいですね",
			"ありがとう、頑張って説明してみましたけどどうでしたか、伝わりましたかね",
		}

		for i := 0; i < 24; i++ {
			speaker := "太郎"
			if i%2 == 1 {
				speaker = "花子"
			}
			lines = append(lines, ParsedLine{
				SpeakerName: speaker,
				Text:        texts[i%len(texts)],
			})
		}

		config := ValidatorConfig{TalkMode: TalkModeDialogue, DurationMinutes: 5}
		result := Validate(lines, config)
		assert.True(t, result.Passed)
		assert.Empty(t, result.Issues)
	})
}
