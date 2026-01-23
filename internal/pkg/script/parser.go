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
