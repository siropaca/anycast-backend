package script

import (
	"fmt"
	"regexp"
	"strings"
)

// JSON コードブロックを抽出する正規表現
var (
	jsonCodeBlockRegex = regexp.MustCompile("(?s)```json\\s*\\n(.+?)\\n\\s*```")
	codeBlockRegex     = regexp.MustCompile("(?s)```\\s*\\n(.+?)\\n\\s*```")
)

// ExtractJSON は LLM の出力テキストから JSON 部分を抽出する
//
// 抽出順序:
//  1. ```json ... ``` コードブロック
//  2. ``` ... ``` コードブロック
//  3. 最初の { から最後の } まで
//  4. いずれもなければエラー
func ExtractJSON(text string) (string, error) {
	// 1. ```json ... ``` コードブロック
	if matches := jsonCodeBlockRegex.FindStringSubmatch(text); len(matches) > 1 {
		return strings.TrimSpace(matches[1]), nil
	}

	// 2. ``` ... ``` コードブロック
	if matches := codeBlockRegex.FindStringSubmatch(text); len(matches) > 1 {
		return strings.TrimSpace(matches[1]), nil
	}

	// 3. 最初の { から最後の } まで
	firstBrace := strings.Index(text, "{")
	lastBrace := strings.LastIndex(text, "}")
	if firstBrace != -1 && lastBrace != -1 && lastBrace > firstBrace {
		return strings.TrimSpace(text[firstBrace : lastBrace+1]), nil
	}

	return "", fmt.Errorf("JSON が見つかりません")
}
