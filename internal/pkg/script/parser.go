package script

import (
	"regexp"
	"strings"
)

// ParsedLine はパース済みの台本行
type ParsedLine struct {
	SpeakerName string  // 話者名
	Text        string  // セリフ
	Emotion     *string // 感情・喋り方（オプション）
}

// ParseError はパースエラー
type ParseError struct {
	Line   int    // 行番号（1始まり）
	Reason string // エラー理由
}

// ParseResult はパース結果
type ParseResult struct {
	Lines  []ParsedLine // パース成功した行
	Errors []ParseError // パースエラー
}

// MaxImportLines はインポート時に受け付ける最大行数（空行を除く）
const MaxImportLines = 200

// 感情を抽出する正規表現: [感情] パターン
var emotionRegex = regexp.MustCompile(`^\[([^\]]+)\]\s*`)

// Parse は台本テキストをパースして ParsedLine のスライスに変換する
//
// フォーマット:
//
//	話者名: [感情] セリフ
//
// - 感情は省略可能
// - allowedSpeakers に含まれない話者名はエラー
func Parse(text string, allowedSpeakers []string) ParseResult {
	result := ParseResult{
		Lines:  []ParsedLine{},
		Errors: []ParseError{},
	}

	// 許可された話者名をマップに変換（高速検索用）
	speakerMap := make(map[string]bool, len(allowedSpeakers))
	for _, s := range allowedSpeakers {
		speakerMap[s] = true
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// 空行はスキップ
		if trimmed == "" {
			continue
		}

		// コロンで分割
		colonIdx := strings.Index(trimmed, ":")
		if colonIdx == -1 {
			result.Errors = append(result.Errors, ParseError{
				Line:   lineNum,
				Reason: "コロン(:)が見つかりません",
			})
			continue
		}

		speakerName := strings.TrimSpace(trimmed[:colonIdx])
		content := strings.TrimSpace(trimmed[colonIdx+1:])

		// 話者名が空
		if speakerName == "" {
			result.Errors = append(result.Errors, ParseError{
				Line:   lineNum,
				Reason: "話者名が空です",
			})
			continue
		}

		// 話者名が許可リストにない
		if !speakerMap[speakerName] {
			result.Errors = append(result.Errors, ParseError{
				Line:   lineNum,
				Reason: "不明な話者: " + speakerName,
			})
			continue
		}

		// セリフが空
		if content == "" {
			result.Errors = append(result.Errors, ParseError{
				Line:   lineNum,
				Reason: "セリフが空です",
			})
			continue
		}

		// 感情を抽出
		var emotion *string
		text := content
		if matches := emotionRegex.FindStringSubmatch(content); len(matches) > 1 {
			e := matches[1]
			emotion = &e
			text = strings.TrimSpace(emotionRegex.ReplaceAllString(content, ""))
		}

		// 感情を抽出した後、セリフが空になった場合
		if text == "" {
			result.Errors = append(result.Errors, ParseError{
				Line:   lineNum,
				Reason: "セリフが空です",
			})
			continue
		}

		result.Lines = append(result.Lines, ParsedLine{
			SpeakerName: speakerName,
			Text:        text,
			Emotion:     emotion,
		})
	}

	return result
}

// HasErrors はパース結果にエラーがあるかどうかを返す
func (r *ParseResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// StripEmotionTags は台本テキストから全ての感情タグを除去する
//
// Phase 3 → Phase 4 のハンドオフ時に使用し、Phase 4 が新たに感情タグを追加できるようにする。
//
// @param text - 台本テキスト
// @returns 感情タグを除去した台本テキスト
func StripEmotionTags(text string) string {
	lines := strings.Split(text, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		colonIdx := strings.Index(trimmed, ":")
		if colonIdx == -1 {
			result = append(result, trimmed)
			continue
		}
		speaker := strings.TrimSpace(trimmed[:colonIdx])
		content := strings.TrimSpace(trimmed[colonIdx+1:])
		// 感情タグを除去
		content = emotionRegex.ReplaceAllString(content, "")
		content = strings.TrimSpace(content)
		if content == "" {
			continue
		}
		result = append(result, speaker+": "+content)
	}
	return strings.Join(result, "\n")
}

// CapEmotionTags は台本テキスト内の感情タグ数を上限に収める
//
// 上限を超える場合、後方の感情タグから順に除去する。
//
// @param text - 台本テキスト
// @param maxTags - 感情タグの上限数
// @returns 感情タグ数を上限に収めた台本テキスト
func CapEmotionTags(text string, maxTags int) string {
	lines := strings.Split(text, "\n")

	// 感情タグ付きの行インデックスを収集
	taggedIndices := []int{}
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		colonIdx := strings.Index(trimmed, ":")
		if colonIdx == -1 {
			continue
		}
		content := strings.TrimSpace(trimmed[colonIdx+1:])
		if emotionRegex.MatchString(content) {
			taggedIndices = append(taggedIndices, i)
		}
	}

	if len(taggedIndices) <= maxTags {
		return text
	}

	// 後方から削除（前方の感情タグを優先的に残す）
	toRemove := len(taggedIndices) - maxTags
	removeSet := make(map[int]bool, toRemove)
	for i := len(taggedIndices) - 1; i >= 0 && len(removeSet) < toRemove; i-- {
		removeSet[taggedIndices[i]] = true
	}

	result := make([]string, len(lines))
	for i, line := range lines {
		if removeSet[i] {
			trimmed := strings.TrimSpace(line)
			colonIdx := strings.Index(trimmed, ":")
			speaker := strings.TrimSpace(trimmed[:colonIdx])
			content := strings.TrimSpace(trimmed[colonIdx+1:])
			content = emotionRegex.ReplaceAllString(content, "")
			content = strings.TrimSpace(content)
			result[i] = speaker + ": " + content
		} else {
			result[i] = line
		}
	}
	return strings.Join(result, "\n")
}

// CountNonEmptyLines はテキスト中の空行を除いた行数を返す
func CountNonEmptyLines(text string) int {
	count := 0
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}
