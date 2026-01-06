package audio

import "bytes"

// MP3 データから再生時間（ミリ秒）を推定する
// ビットレートベースの推定を使用するため、VBR の場合は誤差が生じる可能性がある
func GetMP3DurationMs(data []byte) int {
	// MP3 は通常 128kbps なので、サイズから推定
	// duration (秒) = ファイルサイズ (バイト) * 8 / ビットレート (bps)
	const defaultBitrate = 128000 // 128 kbps

	reader := bytes.NewReader(data)
	fileSize := reader.Size()
	durationSeconds := float64(fileSize*8) / float64(defaultBitrate)

	return int(durationSeconds * 1000)
}
