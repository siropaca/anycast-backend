package audio

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Concat は複数の音声データを連結する
//
// FFmpeg の concat demuxer を使用して、複数の MP3 ファイルをシームレスに結合する。
// audioChunks には連結する音声データの配列を渡す。
func Concat(audioChunks [][]byte) ([]byte, error) {
	if len(audioChunks) == 0 {
		return nil, fmt.Errorf("no audio chunks to concatenate")
	}

	// 1つだけの場合はそのまま返す
	if len(audioChunks) == 1 {
		return audioChunks[0], nil
	}

	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "ffmpeg-concat-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// 各チャンクを一時ファイルに書き込み、concat リストを作成
	var concatList bytes.Buffer
	for i, chunk := range audioChunks {
		chunkPath := filepath.Join(tmpDir, fmt.Sprintf("chunk_%03d.mp3", i))
		if err := os.WriteFile(chunkPath, chunk, 0o644); err != nil {
			return nil, fmt.Errorf("failed to write chunk file %d: %w", i, err)
		}
		// concat demuxer 用のリスト形式
		concatList.WriteString(fmt.Sprintf("file '%s'\n", chunkPath))
	}

	// concat リストファイルを作成
	listPath := filepath.Join(tmpDir, "concat_list.txt")
	if err := os.WriteFile(listPath, concatList.Bytes(), 0o644); err != nil {
		return nil, fmt.Errorf("failed to write concat list: %w", err)
	}

	outputPath := filepath.Join(tmpDir, "output.mp3")

	// FFmpeg コマンドを実行（concat demuxer を使用）
	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", listPath,
		"-c:a", "libmp3lame",
		"-b:a", "192k",
		"-y",
		outputPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg concat failed: %w (stderr: %s)", err, stderr.String())
	}

	// 出力ファイルを読み込み
	outputData, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read output file: %w", err)
	}

	return outputData, nil
}
