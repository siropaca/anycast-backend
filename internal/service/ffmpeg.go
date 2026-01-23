package service

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

// FFmpegService は FFmpeg を使用した音声処理サービスのインターフェース
type FFmpegService interface {
	MixAudioWithBGM(ctx context.Context, params MixParams) ([]byte, error)
	ConcatAudio(ctx context.Context, audioChunks [][]byte) ([]byte, error)
}

// MixParams は音声ミキシングのパラメータ
type MixParams struct {
	VoiceData       []byte  // ナレーション音声データ
	BGMData         []byte  // BGM 音声データ
	VoiceDurationMs int     // ナレーション長 (ms)
	BGMVolumeDB     float64 // BGM 音量 (dB)
	FadeOutMs       int     // フェードアウト時間 (ms)
	PaddingStartMs  int     // 音声開始前の余白 (ms)
	PaddingEndMs    int     // 音声終了後の余白 (ms)
}

type ffmpegService struct{}

// NewFFmpegService は FFmpegService の実装を返す
func NewFFmpegService() FFmpegService {
	return &ffmpegService{}
}

// MixAudioWithBGM はナレーションと BGM をミキシングする
//
// フィルタグラフ:
// [BGM] → aloop(無限ループ) → volume(-15dB) → afade(フェードアウト) → atrim(カット)
//
//	↓
//
// [Voice] → adelay(前余白) ──────────────────────────────────────────→ amix → [Output]
func (s *ffmpegService) MixAudioWithBGM(ctx context.Context, params MixParams) ([]byte, error) {
	log := logger.FromContext(ctx)

	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "ffmpeg-mix-*")
	if err != nil {
		log.Error("failed to create temp dir", "error", err)
		return nil, apperror.ErrInternal.WithMessage("一時ディレクトリの作成に失敗しました").WithError(err)
	}
	defer os.RemoveAll(tmpDir)

	// 一時ファイルに書き込み
	voicePath := filepath.Join(tmpDir, "voice.mp3")
	bgmPath := filepath.Join(tmpDir, "bgm.mp3")
	outputPath := filepath.Join(tmpDir, "output.mp3")

	if err := os.WriteFile(voicePath, params.VoiceData, 0o644); err != nil {
		log.Error("failed to write voice file", "error", err)
		return nil, apperror.ErrInternal.WithMessage("音声ファイルの書き込みに失敗しました").WithError(err)
	}

	if err := os.WriteFile(bgmPath, params.BGMData, 0o644); err != nil {
		log.Error("failed to write bgm file", "error", err)
		return nil, apperror.ErrInternal.WithMessage("BGM ファイルの書き込みに失敗しました").WithError(err)
	}

	// 総出力時間を計算（単位: 秒）
	totalDurationMs := params.PaddingStartMs + params.VoiceDurationMs + params.PaddingEndMs
	totalDurationSec := float64(totalDurationMs) / 1000.0

	// フェードアウト開始時間を計算（単位: 秒）
	fadeStartSec := totalDurationSec - float64(params.FadeOutMs)/1000.0
	if fadeStartSec < 0 {
		fadeStartSec = 0
	}

	// フィルタグラフを構築
	// [0:a] = voice, [1:a] = bgm
	filterComplex := fmt.Sprintf(
		// BGM: ループ → 音量調整 → フェードアウト → 長さカット
		"[1:a]aloop=loop=-1:size=2e+09,volume=%sdB,afade=t=out:st=%s:d=%s,atrim=0:%s[bgm];"+
			// Voice: 前余白を追加
			"[0:a]adelay=%d|%d[voice];"+
			// ミックス (BGM は voice に合わせて調整)
			"[bgm][voice]amix=inputs=2:duration=first:dropout_transition=0[out]",
		formatFloat(params.BGMVolumeDB),
		formatFloat(fadeStartSec),
		formatFloat(float64(params.FadeOutMs)/1000.0),
		formatFloat(totalDurationSec),
		params.PaddingStartMs,
		params.PaddingStartMs,
	)

	// FFmpeg コマンドを実行
	args := []string{
		"-i", voicePath,
		"-i", bgmPath,
		"-filter_complex", filterComplex,
		"-map", "[out]",
		"-c:a", "libmp3lame",
		"-b:a", "192k",
		"-y",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	log.Info("executing ffmpeg", "args", args)

	if err := cmd.Run(); err != nil {
		log.Error("ffmpeg failed", "error", err, "stderr", stderr.String())
		return nil, apperror.ErrInternal.WithMessage("音声ミキシングに失敗しました").WithError(err)
	}

	// 出力ファイルを読み込み
	outputData, err := os.ReadFile(outputPath)
	if err != nil {
		log.Error("failed to read output file", "error", err)
		return nil, apperror.ErrInternal.WithMessage("出力ファイルの読み込みに失敗しました").WithError(err)
	}

	return outputData, nil
}

// ConcatAudio は複数の音声データを連結する
//
// FFmpeg の concat demuxer を使用して、複数の MP3 ファイルをシームレスに結合する。
//
// @param audioChunks - 連結する音声データの配列
// @returns 連結された音声データ
func (s *ffmpegService) ConcatAudio(ctx context.Context, audioChunks [][]byte) ([]byte, error) {
	log := logger.FromContext(ctx)

	if len(audioChunks) == 0 {
		return nil, apperror.ErrValidation.WithMessage("結合する音声データがありません")
	}

	// 1つだけの場合はそのまま返す
	if len(audioChunks) == 1 {
		return audioChunks[0], nil
	}

	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "ffmpeg-concat-*")
	if err != nil {
		log.Error("failed to create temp dir", "error", err)
		return nil, apperror.ErrInternal.WithMessage("一時ディレクトリの作成に失敗しました").WithError(err)
	}
	defer os.RemoveAll(tmpDir)

	// 各チャンクを一時ファイルに書き込み、concat リストを作成
	var concatList bytes.Buffer
	for i, chunk := range audioChunks {
		chunkPath := filepath.Join(tmpDir, fmt.Sprintf("chunk_%03d.mp3", i))
		if err := os.WriteFile(chunkPath, chunk, 0o644); err != nil {
			log.Error("failed to write chunk file", "error", err, "index", i)
			return nil, apperror.ErrInternal.WithMessage("音声チャンクの書き込みに失敗しました").WithError(err)
		}
		// concat demuxer 用のリスト形式
		concatList.WriteString(fmt.Sprintf("file '%s'\n", chunkPath))
	}

	// concat リストファイルを作成
	listPath := filepath.Join(tmpDir, "concat_list.txt")
	if err := os.WriteFile(listPath, concatList.Bytes(), 0o644); err != nil {
		log.Error("failed to write concat list", "error", err)
		return nil, apperror.ErrInternal.WithMessage("結合リストの書き込みに失敗しました").WithError(err)
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

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	log.Info("executing ffmpeg concat", "chunks", len(audioChunks))

	if err := cmd.Run(); err != nil {
		log.Error("ffmpeg concat failed", "error", err, "stderr", stderr.String())
		return nil, apperror.ErrInternal.WithMessage("音声の結合に失敗しました").WithError(err)
	}

	// 出力ファイルを読み込み
	outputData, err := os.ReadFile(outputPath)
	if err != nil {
		log.Error("failed to read output file", "error", err)
		return nil, apperror.ErrInternal.WithMessage("出力ファイルの読み込みに失敗しました").WithError(err)
	}

	log.Info("audio concat completed", "chunks", len(audioChunks), "output_size", len(outputData))

	return outputData, nil
}

// formatFloat は float64 を文字列に変換する（小数点以下3桁）
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 3, 64)
}
