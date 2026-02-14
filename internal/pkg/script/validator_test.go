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

func TestValidate_TotalCharacterCount(t *testing.T) {
	// ヘルパー: 指定文字数のセリフを生成する（20〜80文字の範囲で分散）
	buildLines := func(totalChars int) []ParsedLine {
		var lines []ParsedLine
		remaining := totalChars
		speakers := []string{"太郎", "花子"}
		i := 0
		for remaining > 0 {
			// 40文字ずつ追加（最後は残り全部）
			charCount := 40
			if remaining < charCount {
				charCount = remaining
			}
			if charCount < 10 {
				// 最低10文字を保証するため、足りない場合はスキップ
				break
			}
			text := strings.Repeat("あ", charCount)
			lines = append(lines, ParsedLine{
				SpeakerName: speakers[i%2],
				Text:        text,
			})
			remaining -= charCount
			i++
		}
		return lines
	}

	t.Run("目標範囲内の文字数は合格", func(t *testing.T) {
		// 5分 × 300文字 = 1500文字が目標
		lines := buildLines(1500)
		result := checkTotalCharacterCount(lines, 5)
		assert.Empty(t, result)
	})

	t.Run("目標の80%ちょうどは合格", func(t *testing.T) {
		// 5分 × 300 × 0.8 = 1200文字
		lines := buildLines(1200)
		result := checkTotalCharacterCount(lines, 5)
		assert.Empty(t, result)
	})

	t.Run("目標の120%ちょうどは合格", func(t *testing.T) {
		// 5分 × 300 × 1.2 = 1800文字
		lines := buildLines(1800)
		result := checkTotalCharacterCount(lines, 5)
		assert.Empty(t, result)
	})

	t.Run("文字数不足は不合格", func(t *testing.T) {
		// 5分 × 300 × 0.8 = 1200 → 1100は不足
		lines := buildLines(1100)
		result := checkTotalCharacterCount(lines, 5)
		assert.Len(t, result, 1)
		assert.Equal(t, "total_character_count", result[0].Check)
		assert.Contains(t, result[0].Message, "不足")
	})

	t.Run("文字数過多は不合格", func(t *testing.T) {
		// 5分 × 300 × 1.2 = 1800 → 1900は過多
		lines := buildLines(1900)
		result := checkTotalCharacterCount(lines, 5)
		assert.Len(t, result, 1)
		assert.Equal(t, "total_character_count", result[0].Check)
		assert.Contains(t, result[0].Message, "多すぎ")
	})

	t.Run("durationMinutes が 0 の場合はスキップ", func(t *testing.T) {
		lines := buildLines(100)
		result := checkTotalCharacterCount(lines, 0)
		assert.Empty(t, result)
	})
}

func TestValidate_AllPass(t *testing.T) {
	t.Run("全チェック合格で Passed が true", func(t *testing.T) {
		// 5分 × 4行 = 20行以上必要
		// 5分 × 300文字 = 1500文字、±20% = 1200〜1800文字
		// 文長にゆらぎを持たせる（標準偏差5文字以上）
		lines := make([]ParsedLine, 0, 30)
		texts := []string{
			"こんにちは、今日もポッドキャストを始めていきましょう、よろしくお願いします",
			"はい、よろしくお願いします、今日のテーマはすごく興味深いので楽しみにしていました",
			"今日は人工知能について話していこうと思います、最近本当に進歩がすごいですよね",
			"えーと、具体的にはどういうところが進歩しているんでしょうか、もう少し詳しく教えてもらえるとありがたいです",
			"まあ簡単に言うと、画像認識や自然言語処理の精度が格段に上がってきているんですよ",
			"なるほどね、それは実際に使ってみると本当に実感できますよね、私も驚きました",
			"そうなんです、意外と身近なところで活用されているんですよ",
			"ちょっと驚きましたね、スマートフォンのカメラ機能にも使われているとは思いませんでした",
			"具体的な例を挙げると、医療分野では画像診断の精度が人間の医師と同等レベルに達しています",
			"もちろん課題もたくさんあって、倫理的な問題や雇用への影響なども考えなければいけません",
			"確かにそれは大事なポイントですよね、技術の進歩だけではなく社会的な影響も見ないといけない",
			"ありがとうございます、とても分かりやすい説明でした、リスナーの皆さんにも伝わったと思います",
		}

		for i := 0; i < 30; i++ {
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
