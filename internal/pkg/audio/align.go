package audio

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode"
)

// WordTimestamp は単語とそのタイムスタンプを保持する（stt パッケージからの循環参照を避けるため再定義）
type WordTimestamp struct {
	Word      string
	StartTime time.Duration
	EndTime   time.Duration
}

// LineBoundary は行の時間境界を保持する
type LineBoundary struct {
	StartTime time.Duration
	EndTime   time.Duration
}

// DP アライメントのスコア定数
const (
	dpMatchScore    = 2
	dpMismatchScore = -1
	dpGapScore      = -1
)

// maxDivergenceRatio は元テキストと STT テキストの文字数比の最大許容乖離。
// DP アライメントはローカルな偏差に強いため、文字カウント方式より緩い閾値を設定する。
const maxDivergenceRatio = 0.5

// AlignTextToTimestamps は既知のテキスト行と STT の単語タイムスタンプを照合し、
// 各行の時間境界を返す
//
// アルゴリズム:
//  1. 元テキストと STT テキストの正規化文字列を構築
//  2. Needleman-Wunsch DP アライメントで文字レベルの最適対応を計算
//  3. アライメント結果から各行の最終対応文字を特定し、対応する STT 単語のタイムスタンプで境界を設定
//
// TTS がテキストをスキップ・追加した場合も DP 上で gap として処理され、
// 前後の対応文字が正しくマッチするため境界がずれない。
func AlignTextToTimestamps(lines []string, words []WordTimestamp) ([]LineBoundary, error) {
	if len(lines) == 0 {
		return nil, fmt.Errorf("テキスト行が空です")
	}
	if len(words) == 0 {
		return nil, fmt.Errorf("単語タイムスタンプが空です")
	}

	// Phase 1: 元テキストの正規化文字列を構築し、行境界位置を記録
	var origRunes []rune
	// lineEndIndices[i] は行 i の末尾の次のインデックス（exclusive）
	lineEndIndices := make([]int, len(lines))
	for i, line := range lines {
		normalized := []rune(normalizeText(line))
		origRunes = append(origRunes, normalized...)
		lineEndIndices[i] = len(origRunes)
	}

	// Phase 2: STT テキストの正規化文字列を構築し、各文字→単語のマッピングを記録
	var sttRunes []rune
	var sttRuneToWordIdx []int
	for wIdx, w := range words {
		normalized := []rune(normalizeText(w.Word))
		for range normalized {
			sttRuneToWordIdx = append(sttRuneToWordIdx, wIdx)
		}
		sttRunes = append(sttRunes, normalized...)
	}

	// 乖離チェック
	if len(origRunes) > 0 {
		ratio := float64(len(sttRunes)) / float64(len(origRunes))
		if math.Abs(ratio-1.0) > maxDivergenceRatio {
			return nil, fmt.Errorf(
				"STT 認識結果とテキストの文字数が大きく乖離しています（元: %d, STT: %d, 比率: %.2f）",
				len(origRunes), len(sttRunes), ratio,
			)
		}
	}

	// 行が1つの場合は全単語をカバー
	if len(lines) == 1 {
		return []LineBoundary{{
			StartTime: words[0].StartTime,
			EndTime:   words[len(words)-1].EndTime,
		}}, nil
	}

	// Phase 3: DP アライメントで文字レベルの最適対応を計算
	mapping := dpAlignment(origRunes, sttRunes)

	// Phase 4: アライメント結果から行境界を特定
	results := make([]LineBoundary, len(lines))
	results[0].StartTime = words[0].StartTime

	for lineIdx := 0; lineIdx < len(lines)-1; lineIdx++ {
		lineStart := 0
		if lineIdx > 0 {
			lineStart = lineEndIndices[lineIdx-1]
		}
		lineEnd := lineEndIndices[lineIdx]
		nextLineEnd := lineEndIndices[lineIdx+1]

		// 現在の行の最後にマッチした STT 位置を探す
		lastSttPos := -1
		for i := lineEnd - 1; i >= lineStart; i-- {
			if mapping[i] >= 0 {
				lastSttPos = mapping[i]
				break
			}
		}

		// 次の行の最初にマッチした STT 位置を探す
		firstSttPos := -1
		for i := lineEnd; i < nextLineEnd; i++ {
			if mapping[i] >= 0 {
				firstSttPos = mapping[i]
				break
			}
		}

		// カットポイントを決定
		switch {
		case lastSttPos >= 0 && firstSttPos >= 0:
			// 通常ケース: 次の行の最初の単語の直前でカット
			// TTS が追加したテキストは現在の行側に含まれる
			firstWordIdx := sttRuneToWordIdx[firstSttPos]
			if firstWordIdx > 0 {
				prevWordIdx := firstWordIdx - 1
				midpoint := words[prevWordIdx].EndTime +
					(words[firstWordIdx].StartTime-words[prevWordIdx].EndTime)/2
				results[lineIdx].EndTime = midpoint
				results[lineIdx+1].StartTime = midpoint
			} else {
				results[lineIdx].EndTime = words[firstWordIdx].StartTime
				results[lineIdx+1].StartTime = words[firstWordIdx].StartTime
			}

		case lastSttPos >= 0:
			// 次の行にマッチがない — 現在の行の末尾で区切り
			cutWordIdx := sttRuneToWordIdx[lastSttPos]
			if cutWordIdx+1 < len(words) {
				midpoint := words[cutWordIdx].EndTime +
					(words[cutWordIdx+1].StartTime-words[cutWordIdx].EndTime)/2
				results[lineIdx].EndTime = midpoint
				results[lineIdx+1].StartTime = midpoint
			} else {
				results[lineIdx].EndTime = words[cutWordIdx].EndTime
				results[lineIdx+1].StartTime = words[cutWordIdx].EndTime
			}

		case firstSttPos >= 0:
			// 現在の行にマッチがない — 次の行の先頭で区切り
			firstWordIdx := sttRuneToWordIdx[firstSttPos]
			results[lineIdx].EndTime = words[firstWordIdx].StartTime
			results[lineIdx+1].StartTime = words[firstWordIdx].StartTime

		default:
			// 両方にマッチがない — 0 duration のセグメント
			results[lineIdx].EndTime = results[lineIdx].StartTime
			results[lineIdx+1].StartTime = results[lineIdx].StartTime
		}
	}

	// 最後の行の EndTime
	results[len(lines)-1].EndTime = words[len(words)-1].EndTime

	return results, nil
}

// dpAlignment は Needleman-Wunsch アルゴリズムで2つの文字列の最適アライメントを計算し、
// 元テキストの各文字が STT テキストのどの位置に対応するかを返す。
// 対応がない文字（gap）は -1 になる。
func dpAlignment(orig, stt []rune) []int {
	n := len(orig)
	m := len(stt)

	// DP テーブル: dp[i][j] = orig[:i] と stt[:j] の最適アライメントスコア
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}

	// 初期化: 先頭からの gap
	for i := 1; i <= n; i++ {
		dp[i][0] = dp[i-1][0] + dpGapScore
	}
	for j := 1; j <= m; j++ {
		dp[0][j] = dp[0][j-1] + dpGapScore
	}

	// DP テーブルを埋める
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			score := dpMismatchScore
			if orig[i-1] == stt[j-1] {
				score = dpMatchScore
			}
			dp[i][j] = max(
				dp[i-1][j-1]+score,    // match/mismatch（対角）
				dp[i-1][j]+dpGapScore, // gap in STT（元テキストの文字をスキップ）
				dp[i][j-1]+dpGapScore, // gap in orig（STT の文字をスキップ）
			)
		}
	}

	// トレースバック: 最適パスを辿り、元テキスト → STT の対応マッピングを構築
	mapping := make([]int, n)
	for i := range mapping {
		mapping[i] = -1
	}

	i, j := n, m
	for i > 0 && j > 0 {
		score := dpMismatchScore
		if orig[i-1] == stt[j-1] {
			score = dpMatchScore
		}

		switch dp[i][j] {
		case dp[i-1][j-1] + score:
			// match/mismatch: 両方の文字を対応させる
			mapping[i-1] = j - 1
			i--
			j--
		case dp[i-1][j] + dpGapScore:
			// gap in STT: 元テキストの文字に対応なし（TTS がスキップ）
			i--
		default:
			// gap in orig: STT の文字に対応なし（TTS が追加）
			j--
		}
	}

	return mapping
}

// SnapBoundariesToSilence は STT で得た行境界を最寄りの無音区間中間点にスナップする
//
// 各内部カットポイント（boundaries[i].EndTime / boundaries[i+1].StartTime）について、
// maxSnapDistance 以内で最も長い無音区間の中間点にスナップする。
// 文間ポーズは文中の句読点ポーズより長いため、最長の無音を選ぶことで
// 誤スナップを防ぐ。
// 先頭境界の StartTime と末尾境界の EndTime は変更しない。
func SnapBoundariesToSilence(boundaries []LineBoundary, silences []SilenceInterval, maxSnapDistance time.Duration) []LineBoundary {
	if len(boundaries) <= 1 || len(silences) == 0 {
		return boundaries
	}

	result := make([]LineBoundary, len(boundaries))
	copy(result, boundaries)

	// 使用済みの無音区間インデックスを追跡（同じ無音に複数カットがスナップしないようにする）
	usedSilences := make(map[int]bool)

	// 内部カットポイントを処理（最初の StartTime と最後の EndTime は変更しない）
	for i := 0; i < len(result)-1; i++ {
		cutTime := result[i].EndTime
		cutSec := cutTime.Seconds()

		bestIdx := -1
		bestDuration := 0.0

		for j, s := range silences {
			if usedSilences[j] {
				continue
			}
			// 距離は無音区間の最寄りの端で計算する（区間内なら距離0）
			var distSec float64
			if cutSec < s.StartSec {
				distSec = s.StartSec - cutSec
			} else if cutSec > s.EndSec {
				distSec = cutSec - s.EndSec
			}
			dist := time.Duration(distSec * float64(time.Second))
			duration := s.EndSec - s.StartSec
			// 範囲内で最も長い無音区間を選択する
			if dist <= maxSnapDistance && duration > bestDuration {
				bestIdx = j
				bestDuration = duration
			}
		}

		if bestIdx >= 0 {
			midSec := (silences[bestIdx].StartSec + silences[bestIdx].EndSec) / 2.0
			snappedTime := time.Duration(midSec * float64(time.Second))
			result[i].EndTime = snappedTime
			result[i+1].StartTime = snappedTime
			usedSilences[bestIdx] = true
		}
	}

	// スナップ後に時系列順が崩れた場合は元の境界にフォールバック
	// 0-duration セグメント（StartTime == EndTime）はスキップ行で合法的に発生するため、
	// 厳密な逆転（negative duration）のみをチェックする
	for i := 0; i < len(result)-1; i++ {
		if result[i].EndTime < result[i].StartTime || result[i+1].StartTime > result[i+1].EndTime {
			return boundaries
		}
	}

	return result
}

// SplitPCMByTimestamps は PCM データをタイムスタンプ境界で分割する
func SplitPCMByTimestamps(pcmData []byte, boundaries []LineBoundary, sampleRate, channels, bytesPerSample int) [][]byte {
	bytesPerSec := sampleRate * channels * bytesPerSample
	blockAlign := channels * bytesPerSample

	segments := make([][]byte, len(boundaries))
	for i, b := range boundaries {
		startByte := int(b.StartTime.Seconds() * float64(bytesPerSec))
		endByte := int(b.EndTime.Seconds() * float64(bytesPerSec))

		// ブロックアライメントに合わせる
		startByte = (startByte / blockAlign) * blockAlign
		endByte = (endByte / blockAlign) * blockAlign

		// 範囲チェック
		if startByte < 0 {
			startByte = 0
		}
		if endByte > len(pcmData) {
			endByte = len(pcmData)
		}
		if startByte >= endByte {
			segments[i] = []byte{}
			continue
		}

		segments[i] = pcmData[startByte:endByte]
	}

	return segments
}

// normalizeText はテキストから句読点・空白・記号を除去して正規化する
func normalizeText(text string) string {
	var b strings.Builder
	for _, r := range text {
		if shouldKeepRune(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// shouldKeepRune は正規化時に保持すべき文字かどうかを判定する
func shouldKeepRune(r rune) bool {
	// 空白は除去
	if unicode.IsSpace(r) {
		return false
	}
	// 句読点・記号は除去
	if unicode.IsPunct(r) {
		return false
	}
	// 日本語の句読点（Unicode カテゴリ外のもの）
	switch r {
	case '、', '。', '「', '」', '『', '』', '（', '）', '・', '…', '〜', '！', '？',
		'，', '．', '[', ']', ':', '：':
		return false
	}
	return true
}
