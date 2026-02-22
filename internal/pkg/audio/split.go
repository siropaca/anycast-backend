package audio

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
)

// SilenceInterval は無音区間を表す
type SilenceInterval struct {
	StartSec float64
	EndSec   float64
}

// PCMSplitConfig は PCM 無音分割のパラメータ
type PCMSplitConfig struct {
	SampleRate       int     // サンプルレート（例: 24000）
	Channels         int     // チャンネル数（例: 1）
	BytesPerSample   int     // 1サンプルあたりのバイト数（例: 2 = s16le）
	NoiseDB          float64 // 無音とみなすノイズレベル（例: -30.0）
	MinSilenceSec    float64 // 無音の最小継続時間（例: 0.3）
	ExpectedSegments int     // 期待するセグメント数（0 の場合は全無音区間で分割）
}

// DetectSilenceIntervals は PCM データから無音区間を検出して返す
//
// ffmpeg の silencedetect フィルターを使用して無音区間を検出する。
// config の SampleRate, Channels, BytesPerSample, NoiseDB, MinSilenceSec を使用する。
func DetectSilenceIntervals(pcmData []byte, config PCMSplitConfig) ([]SilenceInterval, error) {
	if len(pcmData) == 0 {
		return nil, fmt.Errorf("PCM データが空です")
	}

	// PCM を一時ファイルに書き出し
	tmpFile, err := os.CreateTemp("", "pcm-silence-*.raw")
	if err != nil {
		return nil, fmt.Errorf("一時ファイルの作成に失敗しました: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(pcmData); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("一時ファイルへの書き込みに失敗しました: %w", err)
	}
	tmpFile.Close()

	// ffmpeg silencedetect を実行
	args := []string{
		"-f", "s16le",
		"-ar", strconv.Itoa(config.SampleRate),
		"-ac", strconv.Itoa(config.Channels),
		"-i", tmpFile.Name(),
		"-af", fmt.Sprintf("silencedetect=noise=%gdB:d=%g", config.NoiseDB, config.MinSilenceSec),
		"-f", "null",
		"-",
	}

	cmd := exec.Command("ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg silencedetect に失敗しました: %w (stderr: %s)", err, stderr.String())
	}

	return parseSilenceDetectOutput(stderr.String()), nil
}

// SplitPCMBySilence は PCM データを無音区間で分割する
//
// ffmpeg の silencedetect フィルターを使用して無音区間を検出し、
// 無音区間の中間点でカットして複数のセグメントに分割する。
func SplitPCMBySilence(pcmData []byte, config PCMSplitConfig) ([][]byte, error) {
	intervals, err := DetectSilenceIntervals(pcmData, config)
	if err != nil {
		return nil, err
	}

	if len(intervals) == 0 {
		// 無音区間がない場合は全体を1セグメントとして返す
		return [][]byte{pcmData}, nil
	}

	// ExpectedSegments が指定されている場合、必要なカットポイント数に絞る
	// 無音区間の長い順（＝文と文の間のポーズ）を優先して選択する
	if config.ExpectedSegments > 1 && len(intervals) > config.ExpectedSegments-1 {
		intervals = selectTopSilenceIntervals(intervals, config.ExpectedSegments-1)
	}

	// 無音区間の中間点をカットポイントとする
	bytesPerSec := config.SampleRate * config.Channels * config.BytesPerSample
	blockAlign := config.Channels * config.BytesPerSample

	var segments [][]byte
	prevOffset := 0

	for _, interval := range intervals {
		midSec := (interval.StartSec + interval.EndSec) / 2.0
		byteOffset := int(math.Round(midSec * float64(bytesPerSec)))

		// ブロックアライメントに合わせる
		byteOffset = (byteOffset / blockAlign) * blockAlign

		// 範囲チェック
		if byteOffset <= prevOffset || byteOffset >= len(pcmData) {
			continue
		}

		segments = append(segments, pcmData[prevOffset:byteOffset])
		prevOffset = byteOffset
	}

	// 最後のセグメント
	if prevOffset < len(pcmData) {
		segments = append(segments, pcmData[prevOffset:])
	}

	return segments, nil
}

// silencedetect 出力のパース用正規表現
var (
	silenceStartRe = regexp.MustCompile(`silence_start:\s*([\d.]+)`)
	silenceEndRe   = regexp.MustCompile(`silence_end:\s*([\d.]+)`)
)

// selectTopSilenceIntervals は無音区間を長さ順でソートし、上位 n 個を時系列順で返す
func selectTopSilenceIntervals(intervals []SilenceInterval, n int) []SilenceInterval {
	if n >= len(intervals) {
		return intervals
	}

	// コピーして長さ降順でソート
	ranked := make([]SilenceInterval, len(intervals))
	copy(ranked, intervals)
	sort.Slice(ranked, func(i, j int) bool {
		durI := ranked[i].EndSec - ranked[i].StartSec
		durJ := ranked[j].EndSec - ranked[j].StartSec
		return durI > durJ
	})

	// 上位 n 個を取得
	selected := ranked[:n]

	// 時系列順に戻す
	sort.Slice(selected, func(i, j int) bool {
		return selected[i].StartSec < selected[j].StartSec
	})

	return selected
}

// parseSilenceDetectOutput は ffmpeg silencedetect の stderr 出力をパースする
func parseSilenceDetectOutput(stderr string) []SilenceInterval {
	starts := silenceStartRe.FindAllStringSubmatch(stderr, -1)
	ends := silenceEndRe.FindAllStringSubmatch(stderr, -1)

	// start と end をペアにする
	count := len(starts)
	if len(ends) < count {
		count = len(ends)
	}

	intervals := make([]SilenceInterval, 0, count)
	for i := 0; i < count; i++ {
		startSec, err1 := strconv.ParseFloat(starts[i][1], 64)
		endSec, err2 := strconv.ParseFloat(ends[i][1], 64)
		if err1 != nil || err2 != nil {
			continue
		}
		intervals = append(intervals, SilenceInterval{
			StartSec: startSec,
			EndSec:   endSec,
		})
	}

	return intervals
}
