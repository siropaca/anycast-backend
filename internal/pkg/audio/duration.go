package audio

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// GetDurationMs は音声データの再生時間（ミリ秒）を取得する
// ffprobe を使用して正確な再生時間を取得する
// エラー時は 0 を返す（エラーハンドリングが必要な場合は GetDurationMsE を使用）
func GetDurationMs(data []byte) int {
	durationMs, _ := GetDurationMsE(data)
	return durationMs
}

// GetDurationMsE は音声データの再生時間（ミリ秒）を取得する（エラー付き）
// ffprobe を使用して正確な再生時間を取得する
func GetDurationMsE(data []byte) (int, error) {
	// 一時ファイルを作成
	tmpFile, err := os.CreateTemp("", "audio-duration-*")
	if err != nil {
		return 0, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		return 0, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// ffprobe で再生時間を取得
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		tmpFile.Name(),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w (stderr: %s)", err, stderr.String())
	}

	// 秒数をパースしてミリ秒に変換
	durationStr := bytes.TrimSpace(stdout.Bytes())
	durationSec, err := strconv.ParseFloat(string(durationStr), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return int(durationSec * 1000), nil
}
