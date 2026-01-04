package prompt

import (
	"regexp"
	"strings"
)

var (
	// 連続する改行（2つ以上）を1つに
	multipleNewlines = regexp.MustCompile(`\n{2,}`)
	// 行頭のスペース・タブを除去
	leadingSpaces = regexp.MustCompile(`(?m)^[ \t]+`)
)

// プロンプトの不要な空白を除去してトークン数を削減する
// 可読性を維持するため、過度な圧縮は行わない
func Compress(s string) string {
	s = strings.TrimSpace(s)
	s = leadingSpaces.ReplaceAllString(s, "")
	s = multipleNewlines.ReplaceAllString(s, "\n")
	return s
}
