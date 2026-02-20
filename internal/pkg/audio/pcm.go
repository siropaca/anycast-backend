package audio

// GenerateSilencePCM は指定ミリ秒の無音 PCM データ (s16le) を生成する
func GenerateSilencePCM(durationMs, sampleRate, channels, bytesPerSample int) []byte {
	totalSamples := sampleRate * durationMs / 1000
	totalBytes := totalSamples * channels * bytesPerSample
	return make([]byte, totalBytes)
}

// ConcatPCM は複数の PCM データを結合する
// PCM はヘッダーなしの生データなので単純に連結できる
func ConcatPCM(segments [][]byte) []byte {
	totalLen := 0
	for _, seg := range segments {
		totalLen += len(seg)
	}
	result := make([]byte, 0, totalLen)
	for _, seg := range segments {
		result = append(result, seg...)
	}
	return result
}
