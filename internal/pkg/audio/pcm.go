package audio

import "encoding/binary"

// EncodeWAV は s16le PCM データに WAV ヘッダーを付与して WAV 形式に変換する
func EncodeWAV(pcmData []byte, sampleRate, channels, bytesPerSample int) []byte {
	dataSize := len(pcmData)
	headerSize := 44
	buf := make([]byte, headerSize+dataSize)

	// RIFF header
	copy(buf[0:4], "RIFF")
	binary.LittleEndian.PutUint32(buf[4:8], uint32(headerSize-8+dataSize))
	copy(buf[8:12], "WAVE")

	// fmt chunk
	copy(buf[12:16], "fmt ")
	binary.LittleEndian.PutUint32(buf[16:20], 16) // chunk size
	binary.LittleEndian.PutUint16(buf[20:22], 1)  // PCM format
	binary.LittleEndian.PutUint16(buf[22:24], uint16(channels))
	binary.LittleEndian.PutUint32(buf[24:28], uint32(sampleRate))
	binary.LittleEndian.PutUint32(buf[28:32], uint32(sampleRate*channels*bytesPerSample)) // byte rate
	binary.LittleEndian.PutUint16(buf[32:34], uint16(channels*bytesPerSample))            // block align
	binary.LittleEndian.PutUint16(buf[34:36], uint16(bytesPerSample*8))                   // bits per sample

	// data chunk
	copy(buf[36:40], "data")
	binary.LittleEndian.PutUint32(buf[40:44], uint32(dataSize))
	copy(buf[headerSize:], pcmData)

	return buf
}

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
