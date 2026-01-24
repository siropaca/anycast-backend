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
	"github.com/siropaca/anycast-backend/internal/pkg/audio"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

// FFmpegService は FFmpeg を使用した音声処理サービスのインターフェースを表す
type FFmpegService interface {
	MixAudioWithBGM(ctx context.Context, params MixParams) ([]byte, error)
	ConcatAudio(ctx context.Context, audioChunks [][]byte) ([]byte, error)
	// ConvertToMP3 は音声データを MP3 に変換する
	// format: 入力形式（"pcm" または "ogg"）
	ConvertToMP3(ctx context.Context, audioData []byte, format string, sampleRateHz int) ([]byte, error)
}

// MixParams は音声ミキシングのパラメータを表す
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

// NewFFmpegService は ffmpegService を生成して FFmpegService として返す
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
		log.Error("一時ディレクトリの作成に失敗しました", "error", err)
		return nil, apperror.ErrInternal.WithMessage("一時ディレクトリの作成に失敗しました").WithError(err)
	}
	defer os.RemoveAll(tmpDir)

	// 一時ファイルに書き込み
	voicePath := filepath.Join(tmpDir, "voice.mp3")
	bgmPath := filepath.Join(tmpDir, "bgm.mp3")
	outputPath := filepath.Join(tmpDir, "output.mp3")

	if err := os.WriteFile(voicePath, params.VoiceData, 0o644); err != nil {
		log.Error("音声ファイルの書き込みに失敗しました", "error", err)
		return nil, apperror.ErrInternal.WithMessage("音声ファイルの書き込みに失敗しました").WithError(err)
	}

	if err := os.WriteFile(bgmPath, params.BGMData, 0o644); err != nil {
		log.Error("BGM ファイルの書き込みに失敗しました", "error", err)
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

	log.Info("FFmpeg を実行中", "args", args)

	if err := cmd.Run(); err != nil {
		log.Error("FFmpeg が失敗しました", "error", err, "stderr", stderr.String())
		return nil, apperror.ErrInternal.WithMessage("音声ミキシングに失敗しました").WithError(err)
	}

	// 出力ファイルを読み込み
	outputData, err := os.ReadFile(outputPath)
	if err != nil {
		log.Error("出力ファイルの読み込みに失敗しました", "error", err)
		return nil, apperror.ErrInternal.WithMessage("出力ファイルの読み込みに失敗しました").WithError(err)
	}

	return outputData, nil
}

// ConcatAudio は複数の音声データを連結する
// audioChunks に連結する音声データの配列を渡す。
func (s *ffmpegService) ConcatAudio(ctx context.Context, audioChunks [][]byte) ([]byte, error) {
	log := logger.FromContext(ctx)

	log.Info("FFmpeg 結合を実行中", "chunks", len(audioChunks))

	outputData, err := audio.Concat(audioChunks)
	if err != nil {
		log.Error("FFmpeg 結合が失敗しました", "error", err)
		return nil, apperror.ErrInternal.WithMessage("音声の結合に失敗しました").WithError(err)
	}

	log.Info("音声結合が完了しました", "chunks", len(audioChunks), "output_size", len(outputData))

	return outputData, nil
}

// formatFloat は float64 を文字列に変換する（小数点以下 3 桁）
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 3, 64)
}

// ConvertToMP3 は音声データを MP3 に変換する
// format: 入力形式（"pcm" または "ogg"）
func (s *ffmpegService) ConvertToMP3(ctx context.Context, audioData []byte, format string, sampleRateHz int) ([]byte, error) {
	log := logger.FromContext(ctx)

	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "ffmpeg-convert-*")
	if err != nil {
		log.Error("一時ディレクトリの作成に失敗しました", "error", err)
		return nil, apperror.ErrInternal.WithMessage("一時ディレクトリの作成に失敗しました").WithError(err)
	}
	defer os.RemoveAll(tmpDir)

	// 入力ファイル名と FFmpeg 引数を format に応じて決定
	var inputPath string
	var inputArgs []string

	switch format {
	case "pcm":
		inputPath = filepath.Join(tmpDir, "input.pcm")
		// LINEAR16 PCM は signed 16-bit little-endian、mono
		inputArgs = []string{
			"-f", "s16le",
			"-ar", strconv.Itoa(sampleRateHz),
			"-ac", "1",
			"-i", inputPath,
		}
	case "ogg":
		inputPath = filepath.Join(tmpDir, "input.ogg")
		inputArgs = []string{"-i", inputPath}
	default:
		return nil, apperror.ErrValidation.WithMessage("サポートされていない音声フォーマットです: " + format)
	}

	outputPath := filepath.Join(tmpDir, "output.mp3")

	if err := os.WriteFile(inputPath, audioData, 0o644); err != nil {
		log.Error("入力ファイルの書き込みに失敗しました", "error", err)
		return nil, apperror.ErrInternal.WithMessage("入力ファイルの書き込みに失敗しました").WithError(err)
	}

	// FFmpeg コマンドを実行
	args := append(inputArgs, "-c:a", "libmp3lame", "-b:a", "192k", "-y", outputPath)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	log.Info("FFmpeg 変換を実行中", "format", format, "input_size", len(audioData))

	if err := cmd.Run(); err != nil {
		log.Error("FFmpeg 変換が失敗しました", "error", err, "stderr", stderr.String())
		return nil, apperror.ErrInternal.WithMessage(format + " から MP3 への変換に失敗しました").WithError(err)
	}

	// 出力ファイルを読み込み
	outputData, err := os.ReadFile(outputPath)
	if err != nil {
		log.Error("出力ファイルの読み込みに失敗しました", "error", err)
		return nil, apperror.ErrInternal.WithMessage("出力ファイルの読み込みに失敗しました").WithError(err)
	}

	log.Info("音声変換が完了しました", "format", format, "output_size", len(outputData))

	return outputData, nil
}
