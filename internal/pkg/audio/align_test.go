package audio

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeText(t *testing.T) {
	t.Run("句読点と空白を除去する", func(t *testing.T) {
		result := normalizeText("こんにちは、世界！ 今日は良い天気ですね。")

		assert.Equal(t, "こんにちは世界今日は良い天気ですね", result)
	})

	t.Run("ブラケット記号を除去する", func(t *testing.T) {
		result := normalizeText("[laughing] こんにちは")

		assert.Equal(t, "laughingこんにちは", result)
	})

	t.Run("空文字列の場合は空を返す", func(t *testing.T) {
		result := normalizeText("")

		assert.Equal(t, "", result)
	})
}

func TestAlignTextToTimestamps(t *testing.T) {
	t.Run("2行のテキストを正しくアライメントする", func(t *testing.T) {
		lines := []string{
			"こんにちは世界",
			"今日は天気が良い",
		}
		words := []WordTimestamp{
			{Word: "こんにちは", StartTime: 0, EndTime: 500 * time.Millisecond},
			{Word: "世界", StartTime: 500 * time.Millisecond, EndTime: 1000 * time.Millisecond},
			{Word: "今日", StartTime: 1200 * time.Millisecond, EndTime: 1500 * time.Millisecond},
			{Word: "は", StartTime: 1500 * time.Millisecond, EndTime: 1600 * time.Millisecond},
			{Word: "天気", StartTime: 1600 * time.Millisecond, EndTime: 1900 * time.Millisecond},
			{Word: "が", StartTime: 1900 * time.Millisecond, EndTime: 2000 * time.Millisecond},
			{Word: "良い", StartTime: 2000 * time.Millisecond, EndTime: 2500 * time.Millisecond},
		}

		boundaries, err := AlignTextToTimestamps(lines, words)

		require.NoError(t, err)
		require.Len(t, boundaries, 2)
		// 行1: "こんにちは世界" → "世界" と "今日" の間のギャップ(200ms)の中間点で分割
		assert.Equal(t, time.Duration(0), boundaries[0].StartTime)
		assert.Equal(t, 1100*time.Millisecond, boundaries[0].EndTime)
		// 行2: "今日は天気が良い" → 中間点から開始
		assert.Equal(t, 1100*time.Millisecond, boundaries[1].StartTime)
		assert.Equal(t, 2500*time.Millisecond, boundaries[1].EndTime)
	})

	t.Run("句読点を含むテキストを正しくアライメントする", func(t *testing.T) {
		lines := []string{
			"こんにちは、世界。",
			"今日は天気が良い！",
		}
		words := []WordTimestamp{
			{Word: "こんにちは", StartTime: 0, EndTime: 500 * time.Millisecond},
			{Word: "世界", StartTime: 600 * time.Millisecond, EndTime: 1000 * time.Millisecond},
			{Word: "今日は", StartTime: 1200 * time.Millisecond, EndTime: 1500 * time.Millisecond},
			{Word: "天気が", StartTime: 1500 * time.Millisecond, EndTime: 1800 * time.Millisecond},
			{Word: "良い", StartTime: 1800 * time.Millisecond, EndTime: 2200 * time.Millisecond},
		}

		boundaries, err := AlignTextToTimestamps(lines, words)

		require.NoError(t, err)
		require.Len(t, boundaries, 2)
		// "世界" と "今日は" の間のギャップ(200ms)の中間点で分割
		assert.Equal(t, 1100*time.Millisecond, boundaries[0].EndTime)
		assert.Equal(t, 1100*time.Millisecond, boundaries[1].StartTime)
		assert.Equal(t, 2200*time.Millisecond, boundaries[1].EndTime)
	})

	t.Run("1行の場合は全単語をカバーする", func(t *testing.T) {
		lines := []string{"こんにちは世界"}
		words := []WordTimestamp{
			{Word: "こんにちは", StartTime: 100 * time.Millisecond, EndTime: 500 * time.Millisecond},
			{Word: "世界", StartTime: 500 * time.Millisecond, EndTime: 1000 * time.Millisecond},
		}

		boundaries, err := AlignTextToTimestamps(lines, words)

		require.NoError(t, err)
		require.Len(t, boundaries, 1)
		assert.Equal(t, 100*time.Millisecond, boundaries[0].StartTime)
		assert.Equal(t, 1000*time.Millisecond, boundaries[0].EndTime)
	})

	t.Run("空の行の場合はエラーを返す", func(t *testing.T) {
		_, err := AlignTextToTimestamps([]string{}, []WordTimestamp{{Word: "test"}})

		assert.Error(t, err)
	})

	t.Run("空の単語の場合はエラーを返す", func(t *testing.T) {
		_, err := AlignTextToTimestamps([]string{"test"}, []WordTimestamp{})

		assert.Error(t, err)
	})

	t.Run("STT の文字数が元テキストより多い場合に DP アライメントで正しく分割する", func(t *testing.T) {
		lines := []string{
			"あいう",
			"えおか",
		}
		words := []WordTimestamp{
			{Word: "あいうう", StartTime: 0, EndTime: 1 * time.Second},
			{Word: "えおか", StartTime: 1200 * time.Millisecond, EndTime: 2 * time.Second},
		}

		boundaries, err := AlignTextToTimestamps(lines, words)

		require.NoError(t, err)
		require.Len(t, boundaries, 2)
		// "あいうう" と "えおか" の間のギャップ(200ms)の中間点で分割
		assert.Equal(t, 1100*time.Millisecond, boundaries[0].EndTime)
		assert.Equal(t, 1100*time.Millisecond, boundaries[1].StartTime)
		assert.Equal(t, 2*time.Second, boundaries[1].EndTime)
	})

	t.Run("STT と元テキストの文字数が大きく乖離している場合はエラーを返す", func(t *testing.T) {
		lines := []string{"こんにちは"}
		words := []WordTimestamp{
			{Word: "こんにちは世界今日は天気が良いですね", StartTime: 0, EndTime: 1 * time.Second},
		}

		_, err := AlignTextToTimestamps(lines, words)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "乖離")
	})

	t.Run("STT のトークン分割が異なっても DP アライメントで正しく分割する", func(t *testing.T) {
		// STT が「きました」を「きまし」「た」に分割するケース。
		// DP アライメントにより「た」が行1に正しく対応し、
		// 「た」と「酒蔵」の間の中間点でカットする。
		lines := []string{
			"わかってきました",
			"酒蔵って行ったことない",
		}
		words := []WordTimestamp{
			{Word: "わかって", StartTime: 0, EndTime: 400 * time.Millisecond},
			{Word: "きまし", StartTime: 400 * time.Millisecond, EndTime: 700 * time.Millisecond},
			{Word: "た", StartTime: 700 * time.Millisecond, EndTime: 800 * time.Millisecond},
			{Word: "酒蔵", StartTime: 1100 * time.Millisecond, EndTime: 1400 * time.Millisecond},
			{Word: "って", StartTime: 1400 * time.Millisecond, EndTime: 1550 * time.Millisecond},
			{Word: "行った", StartTime: 1550 * time.Millisecond, EndTime: 1800 * time.Millisecond},
			{Word: "こと", StartTime: 1800 * time.Millisecond, EndTime: 2000 * time.Millisecond},
			{Word: "ない", StartTime: 2100 * time.Millisecond, EndTime: 2400 * time.Millisecond},
		}

		boundaries, err := AlignTextToTimestamps(lines, words)

		require.NoError(t, err)
		require.Len(t, boundaries, 2)
		// 「た」(EndTime=800ms) と「酒蔵」(StartTime=1100ms) の間のギャップ(300ms)で分割
		// 中間点: 800 + (1100-800)/2 = 950ms
		assert.Equal(t, time.Duration(0), boundaries[0].StartTime)
		assert.Equal(t, 950*time.Millisecond, boundaries[0].EndTime)
		assert.Equal(t, 950*time.Millisecond, boundaries[1].StartTime)
		assert.Equal(t, 2400*time.Millisecond, boundaries[1].EndTime)
	})

	t.Run("STT の文字脱落があっても DP アライメントで正しく分割する", func(t *testing.T) {
		// STT が助詞「の」を脱落させて文字数が1少ない。
		// DP アライメントにより「の」は gap として処理され、
		// 前後の文字が正しく対応するため境界がずれない。
		lines := []string{
			"日本酒の美味しさがわかってきました",
			"酒蔵って行ったことはない",
		}
		words := []WordTimestamp{
			// STT: 「の」が脱落して「日本酒美味しさ」と認識
			{Word: "日本酒", StartTime: 0, EndTime: 300 * time.Millisecond},
			{Word: "美味しさ", StartTime: 320 * time.Millisecond, EndTime: 600 * time.Millisecond},
			{Word: "が", StartTime: 600 * time.Millisecond, EndTime: 680 * time.Millisecond},
			{Word: "わかって", StartTime: 700 * time.Millisecond, EndTime: 1000 * time.Millisecond},
			{Word: "きました", StartTime: 1000 * time.Millisecond, EndTime: 1300 * time.Millisecond},
			// ↓ 文間ポーズ 300ms
			{Word: "酒蔵", StartTime: 1600 * time.Millisecond, EndTime: 1800 * time.Millisecond},
			{Word: "って", StartTime: 1800 * time.Millisecond, EndTime: 1950 * time.Millisecond},
			{Word: "行った", StartTime: 1950 * time.Millisecond, EndTime: 2200 * time.Millisecond},
			{Word: "こと", StartTime: 2200 * time.Millisecond, EndTime: 2400 * time.Millisecond},
			{Word: "は", StartTime: 2400 * time.Millisecond, EndTime: 2500 * time.Millisecond},
			{Word: "ない", StartTime: 2500 * time.Millisecond, EndTime: 2700 * time.Millisecond},
		}

		boundaries, err := AlignTextToTimestamps(lines, words)

		require.NoError(t, err)
		require.Len(t, boundaries, 2)
		// 「きました」(EndTime=1300ms) と「酒蔵」(StartTime=1600ms) の間で分割
		assert.Equal(t, time.Duration(0), boundaries[0].StartTime)
		assert.Equal(t, 1450*time.Millisecond, boundaries[0].EndTime)
		assert.Equal(t, 1450*time.Millisecond, boundaries[1].StartTime)
		assert.Equal(t, 2700*time.Millisecond, boundaries[1].EndTime)
	})

	t.Run("STT が2行目を短縮認識しても DP アライメントで正しく分割する", func(t *testing.T) {
		// STT が2行目を短縮認識（「にいった」→「いた」）。
		// DP アライメントにより「た」が行1に正しく対応し、
		// 「た」と「さけぐら」の間の中間点でカットする。
		lines := []string{
			"わかってきました", // 8文字
			"さけぐらにいった", // 8文字
		}
		words := []WordTimestamp{
			{Word: "わかって", StartTime: 0, EndTime: 400 * time.Millisecond},
			{Word: "きまし", StartTime: 400 * time.Millisecond, EndTime: 700 * time.Millisecond},
			{Word: "た", StartTime: 700 * time.Millisecond, EndTime: 800 * time.Millisecond},
			{Word: "さけぐら", StartTime: 1100 * time.Millisecond, EndTime: 1400 * time.Millisecond},
			{Word: "いた", StartTime: 1400 * time.Millisecond, EndTime: 1700 * time.Millisecond},
		}
		// STT 総文字数: 4+3+1+4+2 = 14, 元: 16, scale = 0.875
		// scaledBoundary[0] = round(8 * 0.875) = 7
		// 中間点: 800 + (1100-800)/2 = 950ms

		boundaries, err := AlignTextToTimestamps(lines, words)

		require.NoError(t, err)
		require.Len(t, boundaries, 2)
		assert.Equal(t, time.Duration(0), boundaries[0].StartTime)
		assert.Equal(t, 950*time.Millisecond, boundaries[0].EndTime)
		assert.Equal(t, 950*time.Millisecond, boundaries[1].StartTime)
		assert.Equal(t, 1700*time.Millisecond, boundaries[1].EndTime)
	})

	t.Run("3行のテキストを正しくアライメントする", func(t *testing.T) {
		lines := []string{
			"こんにちはミホです",
			"最近日本酒がすきです",
			"おすすめを教えて",
		}
		words := []WordTimestamp{
			{Word: "こんにちは", StartTime: 0, EndTime: 500 * time.Millisecond},
			{Word: "ミホ", StartTime: 500 * time.Millisecond, EndTime: 700 * time.Millisecond},
			{Word: "です", StartTime: 700 * time.Millisecond, EndTime: 900 * time.Millisecond},
			// ↓ 文間ポーズ 300ms
			{Word: "最近", StartTime: 1200 * time.Millisecond, EndTime: 1400 * time.Millisecond},
			{Word: "日本酒", StartTime: 1400 * time.Millisecond, EndTime: 1700 * time.Millisecond},
			{Word: "が", StartTime: 1700 * time.Millisecond, EndTime: 1800 * time.Millisecond},
			{Word: "すき", StartTime: 1800 * time.Millisecond, EndTime: 2000 * time.Millisecond},
			{Word: "です", StartTime: 2000 * time.Millisecond, EndTime: 2200 * time.Millisecond},
			// ↓ 文間ポーズ 300ms
			{Word: "おすすめ", StartTime: 2500 * time.Millisecond, EndTime: 2800 * time.Millisecond},
			{Word: "を", StartTime: 2800 * time.Millisecond, EndTime: 2900 * time.Millisecond},
			{Word: "教えて", StartTime: 2900 * time.Millisecond, EndTime: 3200 * time.Millisecond},
		}

		boundaries, err := AlignTextToTimestamps(lines, words)

		require.NoError(t, err)
		require.Len(t, boundaries, 3)
		// 行1-2 境界: 「です」(900ms) と「最近」(1200ms) の間 → 中間点 1050ms
		assert.Equal(t, time.Duration(0), boundaries[0].StartTime)
		assert.Equal(t, 1050*time.Millisecond, boundaries[0].EndTime)
		// 行2-3 境界: 「です」(2200ms) と「おすすめ」(2500ms) の間 → 中間点 2350ms
		assert.Equal(t, 1050*time.Millisecond, boundaries[1].StartTime)
		assert.Equal(t, 2350*time.Millisecond, boundaries[1].EndTime)
		assert.Equal(t, 2350*time.Millisecond, boundaries[2].StartTime)
		assert.Equal(t, 3200*time.Millisecond, boundaries[2].EndTime)
	})
	t.Run("TTS が行の冒頭フレーズをスキップした場合に DP アライメントで正しく分割する", func(t *testing.T) {
		// TTS が行1の「よろしくお願いします」をスキップして「実は私...」から読み始めるケース。
		// 文字カウント方式では全後続境界がシフトするが、DP アライメントではスキップが
		// gap として局所的に処理されるため、行2以降の境界がずれない。
		lines := []string{
			"よろしくお願いします。実は私、最近プログラミングを始めたいなって思ってるんですけど。",
			"えっ、ユウキさんでも悩んだんですか。",
			"なるほど。例えばWebサイトを作りたい場合はどうなりますか。",
		}
		words := []WordTimestamp{
			// 行1: 「よろしくお願いします」がスキップされて「実は私」から始まる
			{Word: "実は", StartTime: 0, EndTime: 300 * time.Millisecond},
			{Word: "私", StartTime: 300 * time.Millisecond, EndTime: 500 * time.Millisecond},
			{Word: "最近", StartTime: 500 * time.Millisecond, EndTime: 700 * time.Millisecond},
			{Word: "プログラミング", StartTime: 700 * time.Millisecond, EndTime: 1200 * time.Millisecond},
			{Word: "を", StartTime: 1200 * time.Millisecond, EndTime: 1300 * time.Millisecond},
			{Word: "始めたいなって", StartTime: 1300 * time.Millisecond, EndTime: 1700 * time.Millisecond},
			{Word: "思ってるんですけど", StartTime: 1700 * time.Millisecond, EndTime: 2200 * time.Millisecond},
			// ↓ 文間ポーズ 300ms
			{Word: "えっ", StartTime: 2500 * time.Millisecond, EndTime: 2700 * time.Millisecond},
			{Word: "ユウキ", StartTime: 2700 * time.Millisecond, EndTime: 2900 * time.Millisecond},
			{Word: "さんでも", StartTime: 2900 * time.Millisecond, EndTime: 3200 * time.Millisecond},
			{Word: "悩んだんですか", StartTime: 3200 * time.Millisecond, EndTime: 3700 * time.Millisecond},
			// ↓ 文間ポーズ 300ms
			{Word: "なるほど", StartTime: 4000 * time.Millisecond, EndTime: 4300 * time.Millisecond},
			{Word: "例えば", StartTime: 4300 * time.Millisecond, EndTime: 4500 * time.Millisecond},
			{Word: "Web", StartTime: 4500 * time.Millisecond, EndTime: 4600 * time.Millisecond},
			{Word: "サイト", StartTime: 4600 * time.Millisecond, EndTime: 4800 * time.Millisecond},
			{Word: "を", StartTime: 4800 * time.Millisecond, EndTime: 4900 * time.Millisecond},
			{Word: "作りたい", StartTime: 4900 * time.Millisecond, EndTime: 5200 * time.Millisecond},
			{Word: "場合は", StartTime: 5200 * time.Millisecond, EndTime: 5400 * time.Millisecond},
			{Word: "どうなりますか", StartTime: 5400 * time.Millisecond, EndTime: 5800 * time.Millisecond},
		}

		boundaries, err := AlignTextToTimestamps(lines, words)

		require.NoError(t, err)
		require.Len(t, boundaries, 3)

		// 行1-2 境界: 「思ってるんですけど」(EndTime=2200ms) と「えっ」(StartTime=2500ms) の中間点
		assert.Equal(t, time.Duration(0), boundaries[0].StartTime)
		assert.Equal(t, 2350*time.Millisecond, boundaries[0].EndTime)
		// 行2-3 境界: 「悩んだんですか」(EndTime=3700ms) と「なるほど」(StartTime=4000ms) の中間点
		assert.Equal(t, 2350*time.Millisecond, boundaries[1].StartTime)
		assert.Equal(t, 3850*time.Millisecond, boundaries[1].EndTime)
		assert.Equal(t, 3850*time.Millisecond, boundaries[2].StartTime)
		assert.Equal(t, 5800*time.Millisecond, boundaries[2].EndTime)
	})

	t.Run("TTS がテキストを追加した場合に DP アライメントで正しく分割する", func(t *testing.T) {
		// TTS が行1と行2の間に「そうですね」を追加したケース。
		// DP アライメントにより追加テキストは gap として処理され、
		// 元の行境界が正しく保たれる。
		lines := []string{
			"こんにちは世界",
			"今日は天気が良い",
		}
		words := []WordTimestamp{
			{Word: "こんにちは", StartTime: 0, EndTime: 500 * time.Millisecond},
			{Word: "世界", StartTime: 500 * time.Millisecond, EndTime: 1000 * time.Millisecond},
			// TTS が追加した「そうですね」
			{Word: "そうですね", StartTime: 1000 * time.Millisecond, EndTime: 1400 * time.Millisecond},
			// ↓ 文間ポーズ
			{Word: "今日", StartTime: 1600 * time.Millisecond, EndTime: 1800 * time.Millisecond},
			{Word: "は", StartTime: 1800 * time.Millisecond, EndTime: 1900 * time.Millisecond},
			{Word: "天気", StartTime: 1900 * time.Millisecond, EndTime: 2100 * time.Millisecond},
			{Word: "が", StartTime: 2100 * time.Millisecond, EndTime: 2200 * time.Millisecond},
			{Word: "良い", StartTime: 2200 * time.Millisecond, EndTime: 2500 * time.Millisecond},
		}

		boundaries, err := AlignTextToTimestamps(lines, words)

		require.NoError(t, err)
		require.Len(t, boundaries, 2)
		// 行1-2 境界: 「世界」の EndTime(1000ms) と「そうですね」の StartTime(1000ms) は同じ。
		// 追加テキストは行1側に含まれ、「そうですね」(EndTime=1400ms) と「今日」(StartTime=1600ms) の中間点でカット。
		assert.Equal(t, time.Duration(0), boundaries[0].StartTime)
		assert.Equal(t, 1500*time.Millisecond, boundaries[0].EndTime)
		assert.Equal(t, 1500*time.Millisecond, boundaries[1].StartTime)
		assert.Equal(t, 2500*time.Millisecond, boundaries[1].EndTime)
	})

	t.Run("TTS が行全体をスキップした場合にスキップ行が無音セグメントとなる", func(t *testing.T) {
		// TTS が行2「さようなら」を完全にスキップしたケース。
		// 行2のセグメントには word0 と word1 の間の無音区間が含まれる。
		// 行1 末尾は word0-word1 間の中間点（650ms）、行3 先頭は word1 開始（800ms）。
		lines := []string{
			"こんにちは",
			"さようなら",
			"またね",
		}
		words := []WordTimestamp{
			{Word: "こんにちは", StartTime: 0, EndTime: 500 * time.Millisecond},
			// ↓ 行2「さようなら」がスキップ
			{Word: "またね", StartTime: 800 * time.Millisecond, EndTime: 1200 * time.Millisecond},
		}

		boundaries, err := AlignTextToTimestamps(lines, words)

		require.NoError(t, err)
		require.Len(t, boundaries, 3)
		// 行1: こんにちは — word0 と word1 の中間点でカット
		assert.Equal(t, time.Duration(0), boundaries[0].StartTime)
		assert.Equal(t, 650*time.Millisecond, boundaries[0].EndTime)
		// 行2: スキップされた行は word0-word1 間の無音区間を含む
		assert.Equal(t, 650*time.Millisecond, boundaries[1].StartTime)
		assert.Equal(t, 800*time.Millisecond, boundaries[1].EndTime)
		// 行3: またね
		assert.Equal(t, 800*time.Millisecond, boundaries[2].StartTime)
		assert.Equal(t, 1200*time.Millisecond, boundaries[2].EndTime)
	})
}

func TestDpAlignment(t *testing.T) {
	t.Run("完全一致の文字列を正しくマッピングする", func(t *testing.T) {
		orig := []rune("あいうえお")
		stt := []rune("あいうえお")

		mapping := dpAlignment(orig, stt)

		require.Len(t, mapping, 5)
		for i := 0; i < 5; i++ {
			assert.Equal(t, i, mapping[i])
		}
	})

	t.Run("STT に gap がある場合に元テキストの対応文字が -1 になる", func(t *testing.T) {
		orig := []rune("あいうえお")
		stt := []rune("あいえお") // 「う」がスキップ

		mapping := dpAlignment(orig, stt)

		require.Len(t, mapping, 5)
		assert.Equal(t, 0, mapping[0])  // あ → 0
		assert.Equal(t, 1, mapping[1])  // い → 1
		assert.Equal(t, -1, mapping[2]) // う → gap
		assert.Equal(t, 2, mapping[3])  // え → 2
		assert.Equal(t, 3, mapping[4])  // お → 3
	})

	t.Run("STT に追加文字がある場合もマッピングが正しい", func(t *testing.T) {
		orig := []rune("あいうえお")
		stt := []rune("あいXうえお") // 「X」が追加

		mapping := dpAlignment(orig, stt)

		require.Len(t, mapping, 5)
		assert.Equal(t, 0, mapping[0]) // あ → 0
		assert.Equal(t, 1, mapping[1]) // い → 1
		assert.Equal(t, 3, mapping[2]) // う → 3（X をスキップ）
		assert.Equal(t, 4, mapping[3]) // え → 4
		assert.Equal(t, 5, mapping[4]) // お → 5
	})

	t.Run("先頭のテキストがスキップされた場合も正しくマッピングする", func(t *testing.T) {
		orig := []rune("よろしくこんにちは")
		stt := []rune("こんにちは") // 「よろしく」がスキップ

		mapping := dpAlignment(orig, stt)

		require.Len(t, mapping, 9)
		// 「よろしく」は gap
		for i := 0; i < 4; i++ {
			assert.Equal(t, -1, mapping[i])
		}
		// 「こんにちは」は正しくマッピング
		for i := 4; i < 9; i++ {
			assert.Equal(t, i-4, mapping[i])
		}
	})
}

func TestSnapBoundariesToSilence(t *testing.T) {
	t.Run("カットポイントを無音区間中間点にスナップする", func(t *testing.T) {
		boundaries := []LineBoundary{
			{StartTime: 0, EndTime: 1000 * time.Millisecond},
			{StartTime: 1000 * time.Millisecond, EndTime: 2500 * time.Millisecond},
		}
		silences := []SilenceInterval{
			{StartSec: 0.9, EndSec: 1.1}, // 中間点: 1.0s → STT境界と一致
		}

		result := SnapBoundariesToSilence(boundaries, silences, 500*time.Millisecond)

		require.Len(t, result, 2)
		// 無音区間 0.9-1.1 の中間点 1.0s にスナップ
		assert.Equal(t, 1000*time.Millisecond, result[0].EndTime)
		assert.Equal(t, 1000*time.Millisecond, result[1].StartTime)
		// 先頭と末尾は変更されない
		assert.Equal(t, time.Duration(0), result[0].StartTime)
		assert.Equal(t, 2500*time.Millisecond, result[1].EndTime)
	})

	t.Run("STT 境界から離れた無音区間にスナップする", func(t *testing.T) {
		boundaries := []LineBoundary{
			{StartTime: 0, EndTime: 1000 * time.Millisecond},
			{StartTime: 1000 * time.Millisecond, EndTime: 2500 * time.Millisecond},
		}
		silences := []SilenceInterval{
			{StartSec: 1.1, EndSec: 1.3}, // 中間点: 1.2s（STT境界から200ms離れている）
		}

		result := SnapBoundariesToSilence(boundaries, silences, 500*time.Millisecond)

		require.Len(t, result, 2)
		// 無音区間 1.1-1.3 の中間点 1.2s にスナップ
		assert.Equal(t, 1200*time.Millisecond, result[0].EndTime)
		assert.Equal(t, 1200*time.Millisecond, result[1].StartTime)
	})

	t.Run("maxSnapDistance を超える無音区間はスナップしない", func(t *testing.T) {
		boundaries := []LineBoundary{
			{StartTime: 0, EndTime: 1000 * time.Millisecond},
			{StartTime: 1000 * time.Millisecond, EndTime: 3000 * time.Millisecond},
		}
		silences := []SilenceInterval{
			{StartSec: 2.0, EndSec: 2.2}, // 中間点: 2.1s（STT境界から1100ms離れている）
		}

		result := SnapBoundariesToSilence(boundaries, silences, 500*time.Millisecond)

		require.Len(t, result, 2)
		// スナップされず元の境界のまま
		assert.Equal(t, 1000*time.Millisecond, result[0].EndTime)
		assert.Equal(t, 1000*time.Millisecond, result[1].StartTime)
	})

	t.Run("silences が空の場合は元の境界を返す", func(t *testing.T) {
		boundaries := []LineBoundary{
			{StartTime: 0, EndTime: 1000 * time.Millisecond},
			{StartTime: 1000 * time.Millisecond, EndTime: 2000 * time.Millisecond},
		}

		result := SnapBoundariesToSilence(boundaries, nil, 500*time.Millisecond)

		assert.Equal(t, boundaries, result)
	})

	t.Run("1行の場合は元の境界を返す", func(t *testing.T) {
		boundaries := []LineBoundary{
			{StartTime: 0, EndTime: 2000 * time.Millisecond},
		}
		silences := []SilenceInterval{
			{StartSec: 1.0, EndSec: 1.2},
		}

		result := SnapBoundariesToSilence(boundaries, silences, 500*time.Millisecond)

		assert.Equal(t, boundaries, result)
	})

	t.Run("同じ無音区間に複数のカットがスナップしない", func(t *testing.T) {
		boundaries := []LineBoundary{
			{StartTime: 0, EndTime: 900 * time.Millisecond},
			{StartTime: 900 * time.Millisecond, EndTime: 1100 * time.Millisecond},
			{StartTime: 1100 * time.Millisecond, EndTime: 2000 * time.Millisecond},
		}
		silences := []SilenceInterval{
			{StartSec: 0.95, EndSec: 1.05}, // 中間点: 1.0s — 両方のカットに近い
		}

		result := SnapBoundariesToSilence(boundaries, silences, 500*time.Millisecond)

		require.Len(t, result, 3)
		// 最初のカット(900ms)が無音中間点(1000ms)にスナップ
		assert.Equal(t, 1000*time.Millisecond, result[0].EndTime)
		assert.Equal(t, 1000*time.Millisecond, result[1].StartTime)
		// 2つ目のカット(1100ms)は同じ無音が使用済みのためスナップされない
		assert.Equal(t, 1100*time.Millisecond, result[1].EndTime)
		assert.Equal(t, 1100*time.Millisecond, result[2].StartTime)
	})

	t.Run("3行のテキストで2つのカットポイントをそれぞれスナップする", func(t *testing.T) {
		boundaries := []LineBoundary{
			{StartTime: 0, EndTime: 1000 * time.Millisecond},
			{StartTime: 1000 * time.Millisecond, EndTime: 2200 * time.Millisecond},
			{StartTime: 2200 * time.Millisecond, EndTime: 3500 * time.Millisecond},
		}
		silences := []SilenceInterval{
			{StartSec: 0.9, EndSec: 1.1},  // 中間点: 1.0s
			{StartSec: 2.1, EndSec: 2.35}, // 中間点: 2.225s
		}

		result := SnapBoundariesToSilence(boundaries, silences, 500*time.Millisecond)

		require.Len(t, result, 3)
		assert.Equal(t, 1000*time.Millisecond, result[0].EndTime)
		assert.Equal(t, 1000*time.Millisecond, result[1].StartTime)
		assert.Equal(t, 2225*time.Millisecond, result[1].EndTime)
		assert.Equal(t, 2225*time.Millisecond, result[2].StartTime)
		assert.Equal(t, 3500*time.Millisecond, result[2].EndTime)
	})

	t.Run("範囲内に複数の無音がある場合は最長の無音にスナップする", func(t *testing.T) {
		// STT の文字カウントドリフトで境界が次の行に入り込んだケース:
		// STT 境界が行8の途中にあり、近くに短い句読点ポーズ（0.15s）と
		// 少し遠いが長い文間ポーズ（0.6s）がある
		boundaries := []LineBoundary{
			{StartTime: 0, EndTime: 5200 * time.Millisecond},
			{StartTime: 5200 * time.Millisecond, EndTime: 8000 * time.Millisecond},
		}
		silences := []SilenceInterval{
			{StartSec: 4.8, EndSec: 5.4},   // 文間ポーズ（0.6s）— 端は 200ms 手前
			{StartSec: 5.35, EndSec: 5.50}, // 句読点ポーズ（0.15s）— 端は 150ms 先
		}

		result := SnapBoundariesToSilence(boundaries, silences, 500*time.Millisecond)

		require.Len(t, result, 2)
		// 最長の無音（4.8-5.4, 中間点 5.1s）にスナップされるべき
		assert.Equal(t, 5100*time.Millisecond, result[0].EndTime)
		assert.Equal(t, 5100*time.Millisecond, result[1].StartTime)
	})

	t.Run("長い無音区間でも端が近ければスナップする", func(t *testing.T) {
		// speaker2 の実例: STT境界 3900ms、無音 4217-5131ms（中間点 4674ms）
		// 中間点までは 774ms だが、無音開始点までは 317ms なのでスナップすべき
		boundaries := []LineBoundary{
			{StartTime: 300 * time.Millisecond, EndTime: 3900 * time.Millisecond},
			{StartTime: 3900 * time.Millisecond, EndTime: 8300 * time.Millisecond},
		}
		silences := []SilenceInterval{
			{StartSec: 0.0, EndSec: 0.297},
			{StartSec: 1.706, EndSec: 2.132},
			{StartSec: 4.217, EndSec: 5.131}, // 文境界の無音
			{StartSec: 5.863, EndSec: 6.437},
		}

		result := SnapBoundariesToSilence(boundaries, silences, 500*time.Millisecond)

		require.Len(t, result, 2)
		// 無音 4.217-5.131 の中間点 4.674s にスナップされるべき
		assert.InDelta(t, 4674, result[0].EndTime.Milliseconds(), 1)
		assert.InDelta(t, 4674, result[1].StartTime.Milliseconds(), 1)
	})
}

func TestSplitPCMByTimestamps(t *testing.T) {
	t.Run("タイムスタンプ境界で PCM を分割する", func(t *testing.T) {
		// 2秒分の PCM データ（24kHz, mono, s16le = 96000 bytes）
		pcmData := make([]byte, 96000)
		for i := range pcmData {
			pcmData[i] = byte(i % 256)
		}

		boundaries := []LineBoundary{
			{StartTime: 0, EndTime: 1 * time.Second},
			{StartTime: 1 * time.Second, EndTime: 2 * time.Second},
		}

		segments := SplitPCMByTimestamps(pcmData, boundaries, 24000, 1, 2)

		require.Len(t, segments, 2)
		assert.Equal(t, 48000, len(segments[0])) // 1秒 = 48000 bytes
		assert.Equal(t, 48000, len(segments[1]))
	})

	t.Run("範囲外のタイムスタンプはクリップされる", func(t *testing.T) {
		pcmData := make([]byte, 48000) // 1秒分

		boundaries := []LineBoundary{
			{StartTime: 0, EndTime: 2 * time.Second}, // 実際のデータより長い
		}

		segments := SplitPCMByTimestamps(pcmData, boundaries, 24000, 1, 2)

		require.Len(t, segments, 1)
		assert.Equal(t, 48000, len(segments[0])) // データ末尾でクリップ
	})
}
