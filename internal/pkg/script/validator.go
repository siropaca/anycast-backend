package script

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

// ValidationIssue はバリデーションの問題点
type ValidationIssue struct {
	Check   string `json:"check"`
	Line    int    `json:"line"`
	Message string `json:"message"`
}

// ValidationResult はバリデーション結果
type ValidationResult struct {
	Issues []ValidationIssue
	Passed bool
}

// ValidatorConfig はバリデーターの設定
type ValidatorConfig struct {
	TalkMode        TalkMode
	DurationMinutes int
}

// Validate は台本の品質を定量チェックする
//
// チェック項目:
//   - 共通: セリフ長、句点なし終了、セリフ中句点なし、最低行数、文長のゆらぎ
//   - dialogue: 同一話者連続、話者バランス
//   - monologue: 話者一貫性
func Validate(lines []ParsedLine, config ValidatorConfig) ValidationResult {
	var issues []ValidationIssue

	// 共通チェック
	issues = append(issues, checkLineLengths(lines)...)
	issues = append(issues, checkNoTrailingPeriod(lines)...)
	issues = append(issues, checkNoPeriodInText(lines)...)
	issues = append(issues, checkMinimumLines(lines, config.DurationMinutes)...)
	issues = append(issues, checkLengthVariance(lines)...)

	// talk_mode 別チェック
	switch config.TalkMode {
	case TalkModeDialogue:
		issues = append(issues, checkConsecutiveSpeaker(lines)...)
		issues = append(issues, checkSpeakerBalance(lines)...)
	case TalkModeMonologue:
		issues = append(issues, checkSpeakerConsistency(lines)...)
	}

	return ValidationResult{
		Issues: issues,
		Passed: len(issues) == 0,
	}
}

// checkLineLengths は全セリフが10〜120文字以内かチェックする
func checkLineLengths(lines []ParsedLine) []ValidationIssue {
	var issues []ValidationIssue
	for i, line := range lines {
		length := utf8.RuneCountInString(line.Text)
		if length < 10 {
			issues = append(issues, ValidationIssue{
				Check:   "line_length",
				Line:    i + 1,
				Message: fmt.Sprintf("セリフが短すぎます（%d文字、最低10文字）", length),
			})
		}
		if length > 120 {
			issues = append(issues, ValidationIssue{
				Check:   "line_length",
				Line:    i + 1,
				Message: fmt.Sprintf("セリフが長すぎます（%d文字、最大120文字）", length),
			})
		}
	}
	return issues
}

// checkNoTrailingPeriod はセリフの末尾が句点で終わっていないかチェックする
func checkNoTrailingPeriod(lines []ParsedLine) []ValidationIssue {
	var issues []ValidationIssue
	for i, line := range lines {
		if strings.HasSuffix(line.Text, "。") {
			issues = append(issues, ValidationIssue{
				Check:   "trailing_period",
				Line:    i + 1,
				Message: "セリフの末尾に句点（。）があります",
			})
		}
	}
	return issues
}

// checkNoPeriodInText はセリフ中に句点が含まれていないかチェックする
func checkNoPeriodInText(lines []ParsedLine) []ValidationIssue {
	var issues []ValidationIssue
	for i, line := range lines {
		// 末尾の句点は trailing_period で別途チェックするため、末尾以外を検査
		text := line.Text
		if strings.HasSuffix(text, "。") {
			text = text[:len(text)-len("。")]
		}
		if strings.Contains(text, "。") {
			issues = append(issues, ValidationIssue{
				Check:   "period_in_text",
				Line:    i + 1,
				Message: "セリフ中に句点（。）が含まれています（1行に1文にしてください）",
			})
		}
	}
	return issues
}

// checkMinimumLines は台本が最低行数以上あるかチェックする
//
// duration × 4行/分 以上
func checkMinimumLines(lines []ParsedLine, durationMinutes int) []ValidationIssue {
	minLines := durationMinutes * 4
	if len(lines) < minLines {
		return []ValidationIssue{{
			Check:   "minimum_lines",
			Line:    0,
			Message: fmt.Sprintf("台本の行数が不足しています（%d行、最低%d行）", len(lines), minLines),
		}}
	}
	return nil
}

// checkLengthVariance はセリフ長の標準偏差が5文字以上かチェックする
func checkLengthVariance(lines []ParsedLine) []ValidationIssue {
	if len(lines) < 2 {
		return nil
	}

	lengths := make([]float64, len(lines))
	var sum float64
	for i, line := range lines {
		l := float64(utf8.RuneCountInString(line.Text))
		lengths[i] = l
		sum += l
	}

	mean := sum / float64(len(lengths))
	var varianceSum float64
	for _, l := range lengths {
		diff := l - mean
		varianceSum += diff * diff
	}
	stddev := math.Sqrt(varianceSum / float64(len(lengths)))

	if stddev < 5.0 {
		return []ValidationIssue{{
			Check:   "length_variance",
			Line:    0,
			Message: fmt.Sprintf("セリフ長のゆらぎが不足しています（標準偏差 %.1f文字、最低5文字）", stddev),
		}}
	}
	return nil
}

// checkConsecutiveSpeaker は同一話者が4行以上連続していないかチェックする（dialogue 専用）
func checkConsecutiveSpeaker(lines []ParsedLine) []ValidationIssue {
	var issues []ValidationIssue
	if len(lines) == 0 {
		return nil
	}

	currentSpeaker := lines[0].SpeakerName
	consecutiveCount := 1

	for i := 1; i < len(lines); i++ {
		if lines[i].SpeakerName == currentSpeaker {
			consecutiveCount++
			if consecutiveCount == 4 {
				issues = append(issues, ValidationIssue{
					Check:   "consecutive_speaker",
					Line:    i + 1,
					Message: fmt.Sprintf("同一話者（%s）が4行以上連続しています", currentSpeaker),
				})
			}
		} else {
			currentSpeaker = lines[i].SpeakerName
			consecutiveCount = 1
		}
	}
	return issues
}

// checkSpeakerBalance は各話者のセリフ数が全体の20%以上かチェックする（dialogue 専用）
func checkSpeakerBalance(lines []ParsedLine) []ValidationIssue {
	if len(lines) == 0 {
		return nil
	}

	speakerCounts := make(map[string]int)
	for _, line := range lines {
		speakerCounts[line.SpeakerName]++
	}

	total := len(lines)
	threshold := float64(total) * 0.2

	var issues []ValidationIssue
	for speaker, count := range speakerCounts {
		if float64(count) < threshold {
			issues = append(issues, ValidationIssue{
				Check:   "speaker_balance",
				Line:    0,
				Message: fmt.Sprintf("話者（%s）のセリフ数が全体の20%%未満です（%d/%d = %.0f%%）", speaker, count, total, float64(count)/float64(total)*100),
			})
		}
	}
	return issues
}

// checkSpeakerConsistency は全セリフが同一話者であるかチェックする（monologue 専用）
func checkSpeakerConsistency(lines []ParsedLine) []ValidationIssue {
	if len(lines) == 0 {
		return nil
	}

	firstSpeaker := lines[0].SpeakerName
	for i, line := range lines {
		if line.SpeakerName != firstSpeaker {
			return []ValidationIssue{{
				Check:   "speaker_consistency",
				Line:    i + 1,
				Message: fmt.Sprintf("monologue モードで複数の話者が検出されました（%s と %s）", firstSpeaker, line.SpeakerName),
			}}
		}
	}
	return nil
}
