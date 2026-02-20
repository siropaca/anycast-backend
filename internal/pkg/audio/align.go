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

// tightGapThreshold は単語間が密接に接続されていると見なす最大ギャップ。
// これより小さいギャップは STT による単語分割の結果と判断し、先読み吸収の対象とする。
const tightGapThreshold = 80 * time.Millisecond

// AlignTextToTimestamps は既知のテキスト行と STT の単語タイムスタンプを照合し、
// 各行の時間境界を返す
//
// アルゴリズム:
//  1. 全行のテキストを正規化して累積文字数で行境界を計算
//  2. STT 単語を累積的に消費し、文字数が行境界に到達した単語をカットポイントとする
//  3. カットポイント直後のギャップが極小の場合、次の有意なギャップまで先読み吸収する
//  4. カットポイントでは隣接単語間の中間点を使用し、音声の欠落を防ぐ
func AlignTextToTimestamps(lines []string, words []WordTimestamp) ([]LineBoundary, error) {
	if len(lines) == 0 {
		return nil, fmt.Errorf("テキスト行が空です")
	}
	if len(words) == 0 {
		return nil, fmt.Errorf("単語タイムスタンプが空です")
	}

	// 各行の正規化文字数を計算
	lineLengths := make([]int, len(lines))
	for i, line := range lines {
		lineLengths[i] = len([]rune(normalizeText(line)))
	}

	// 累積文字数で行境界を計算（行 i の終端 = sum(lineLengths[0..i])）
	boundaries := make([]int, len(lines))
	cumulative := 0
	for i, length := range lineLengths {
		cumulative += length
		boundaries[i] = cumulative
	}

	totalOriginalChars := cumulative

	// STT の総文字数を計算
	sttTotalChars := 0
	for _, w := range words {
		sttTotalChars += len([]rune(normalizeText(w.Word)))
	}

	// 乖離チェック
	if totalOriginalChars > 0 {
		ratio := float64(sttTotalChars) / float64(totalOriginalChars)
		if math.Abs(ratio-1.0) > 0.3 {
			return nil, fmt.Errorf(
				"STT 認識結果とテキストの文字数が大きく乖離しています（元: %d, STT: %d, 比率: %.2f）",
				totalOriginalChars, sttTotalChars, ratio,
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

	// STT と元テキストの文字数差を補正するため、境界を STT 文字空間にスケーリング
	scaledBoundaries := make([]int, len(boundaries))
	if totalOriginalChars > 0 && sttTotalChars > 0 {
		scale := float64(sttTotalChars) / float64(totalOriginalChars)
		for i, b := range boundaries {
			scaledBoundaries[i] = int(math.Round(float64(b) * scale))
		}
	} else {
		copy(scaledBoundaries, boundaries)
	}

	// STT 単語の正規化文字を累積的に消費して行境界を特定
	results := make([]LineBoundary, len(lines))
	charCount := 0
	lineIdx := 0
	wordIdx := 0

	results[0].StartTime = words[0].StartTime

	for wordIdx < len(words) && lineIdx < len(lines)-1 {
		wordChars := len([]rune(normalizeText(words[wordIdx].Word)))
		charCount += wordChars
		wordIdx++

		// 累積文字数がスケーリング済み行境界に到達したらカットポイントを確定
		if charCount >= scaledBoundaries[lineIdx] {
			cutIdx := wordIdx - 1

			// 先読み吸収: カット直後のギャップが極小（< 80ms）の場合、
			// STT が単語を分割した可能性がある（例: 「きまし」+「た」）。
			// 次の有意なギャップまで単語を吸収して正しい文境界まで進める。
			for cutIdx+1 < len(words) {
				gapAfterCut := words[cutIdx+1].StartTime - words[cutIdx].EndTime
				if gapAfterCut >= tightGapThreshold {
					break // 有意なギャップに到達、ここでカット
				}
				// ギャップが小さい — この先に有意なギャップがあるか確認
				foundBoundary := false
				for j := cutIdx + 2; j < len(words) && j <= cutIdx+5; j++ {
					if words[j].StartTime-words[j-1].EndTime >= tightGapThreshold {
						foundBoundary = true
						break
					}
				}
				if !foundBoundary {
					break // 近くに有意なギャップがないので吸収しない
				}
				cutIdx++
				charCount += len([]rune(normalizeText(words[cutIdx].Word)))
			}
			wordIdx = cutIdx + 1

			// カットポイント: 隣接単語間の中間点を使用して音声の欠落を防ぐ
			if cutIdx+1 < len(words) {
				midpoint := words[cutIdx].EndTime +
					(words[cutIdx+1].StartTime-words[cutIdx].EndTime)/2
				results[lineIdx].EndTime = midpoint
				results[lineIdx+1].StartTime = midpoint
			} else {
				results[lineIdx].EndTime = words[cutIdx].EndTime
				if lineIdx+1 < len(lines) {
					results[lineIdx+1].StartTime = words[cutIdx].EndTime
				}
			}

			lineIdx++
		}
	}

	// 残りの行が未割り当ての場合（STT 単語が足りない）
	if lineIdx < len(lines) {
		lastEnd := words[len(words)-1].EndTime
		for i := lineIdx; i < len(lines); i++ {
			if results[i].StartTime == 0 && i > 0 {
				results[i].StartTime = lastEnd
			}
			results[i].EndTime = lastEnd
		}
	}

	return results, nil
}

// SnapBoundariesToSilence は STT で得た行境界を最寄りの無音区間中間点にスナップする
//
// 各内部カットポイント（boundaries[i].EndTime / boundaries[i+1].StartTime）について、
// maxSnapDistance 以内の最寄り無音区間の中間点にスナップする。
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
		bestDist := maxSnapDistance + 1 // 初期値は最大距離を超える値

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
			if dist <= maxSnapDistance && dist < bestDist {
				bestIdx = j
				bestDist = dist
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
	for i := 0; i < len(result)-1; i++ {
		if result[i].EndTime <= result[i].StartTime || result[i+1].StartTime >= result[i+1].EndTime {
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
