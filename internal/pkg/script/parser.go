package script

import (
	"regexp"
	"strconv"
	"strings"
)

// 行の種別
type LineType string

const (
	LineTypeSpeech  LineType = "speech"
	LineTypeSilence LineType = "silence"
	LineTypeSfx     LineType = "sfx"
)

// 予約語（話者名として使用不可）
const (
	SilenceKeyword = "__SILENCE__"
	SfxKeyword     = "__SFX__"
)

// パース済みの台本行
type ParsedLine struct {
	LineType    LineType // 行の種別
	SpeakerName string   // 話者名（speech 時のみ）
	Text        string   // セリフ（speech 時のみ）
	Emotion     *string  // 感情・喋り方（speech 時のみ、オプション）
	DurationMs  int      // 無音の長さ（silence 時のみ）
	SfxName     string   // 効果音名（sfx 時のみ）
}

// パースエラー
type ParseError struct {
	Line   int    // 行番号（1始まり）
	Reason string // エラー理由
}

// パース結果
type ParseResult struct {
	Lines  []ParsedLine // パース成功した行
	Errors []ParseError // パースエラー
}

// 感情を抽出する正規表現: [感情] パターン
var emotionRegex = regexp.MustCompile(`^\[([^\]]+)\]\s*`)

// 台本テキストをパースして ParsedLine のスライスに変換する
//
// フォーマット:
//
//	話者名: [感情] セリフ
//	__SILENCE__: ミリ秒
//	__SFX__: 効果音名
//
// - 感情は省略可能
// - allowedSpeakers に含まれない話者名はエラー
// - allowedSfx が指定されている場合、含まれない効果音名はエラー（nil の場合はチェックしない）
func Parse(text string, allowedSpeakers, allowedSfx []string) ParseResult {
	result := ParseResult{
		Lines:  []ParsedLine{},
		Errors: []ParseError{},
	}

	// 許可された話者名をマップに変換（高速検索用）
	speakerMap := make(map[string]bool, len(allowedSpeakers))
	for _, s := range allowedSpeakers {
		speakerMap[s] = true
	}

	// 許可された効果音名をマップに変換
	var sfxMap map[string]bool
	if allowedSfx != nil {
		sfxMap = make(map[string]bool, len(allowedSfx))
		for _, s := range allowedSfx {
			sfxMap[s] = true
		}
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

		keyword := strings.TrimSpace(trimmed[:colonIdx])
		content := strings.TrimSpace(trimmed[colonIdx+1:])

		// キーワードが空
		if keyword == "" {
			result.Errors = append(result.Errors, ParseError{
				Line:   lineNum,
				Reason: "話者名が空です",
			})
			continue
		}

		// __SILENCE__ の処理
		if keyword == SilenceKeyword {
			if content == "" {
				result.Errors = append(result.Errors, ParseError{
					Line:   lineNum,
					Reason: "__SILENCE__ の値が空です",
				})
				continue
			}
			durationMs, err := strconv.Atoi(content)
			if err != nil || durationMs <= 0 {
				result.Errors = append(result.Errors, ParseError{
					Line:   lineNum,
					Reason: "__SILENCE__ の値は正の整数である必要があります",
				})
				continue
			}
			result.Lines = append(result.Lines, ParsedLine{
				LineType:   LineTypeSilence,
				DurationMs: durationMs,
			})
			continue
		}

		// __SFX__ の処理
		if keyword == SfxKeyword {
			if content == "" {
				result.Errors = append(result.Errors, ParseError{
					Line:   lineNum,
					Reason: "__SFX__ の値が空です",
				})
				continue
			}
			// 許可リストがある場合はチェック
			if sfxMap != nil && !sfxMap[content] {
				result.Errors = append(result.Errors, ParseError{
					Line:   lineNum,
					Reason: "不明な効果音: " + content,
				})
				continue
			}
			result.Lines = append(result.Lines, ParsedLine{
				LineType: LineTypeSfx,
				SfxName:  content,
			})
			continue
		}

		// speech の処理
		speakerName := keyword

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
			LineType:    LineTypeSpeech,
			SpeakerName: speakerName,
			Text:        text,
			Emotion:     emotion,
		})
	}

	return result
}

// パース結果にエラーがあるかどうかを返す
func (r *ParseResult) HasErrors() bool {
	return len(r.Errors) > 0
}
