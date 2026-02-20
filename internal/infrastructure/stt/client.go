package stt

import (
	"context"
	"fmt"
	"time"

	speech "cloud.google.com/go/speech/apiv2"
	speechpb "cloud.google.com/go/speech/apiv2/speechpb"
	"google.golang.org/api/option"
)

// WordTimestamp は単語とそのタイムスタンプを保持する
type WordTimestamp struct {
	Word      string
	StartTime time.Duration
	EndTime   time.Duration
}

// Client は Speech-to-Text のインターフェース
type Client interface {
	// RecognizeWithTimestamps は PCM 音声を文字起こしし、単語レベルのタイムスタンプを返す
	RecognizeWithTimestamps(ctx context.Context, pcmData []byte, sampleRate int) ([]WordTimestamp, error)
}

// googleSTTClient は Google Cloud Speech-to-Text v2 クライアント
type googleSTTClient struct {
	client    *speech.Client
	projectID string
	location  string
}

// NewGoogleSTTClient は Google Cloud Speech-to-Text v2 クライアントを生成する
func NewGoogleSTTClient(ctx context.Context, projectID, location, credentialsJSON string) (Client, error) {
	var opts []option.ClientOption
	if credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(credentialsJSON))) //nolint:staticcheck // TODO: migrate to newer auth method
	}

	c, err := speech.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("Speech-to-Text クライアントの作成に失敗しました: %w", err)
	}

	return &googleSTTClient{
		client:    c,
		projectID: projectID,
		location:  location,
	}, nil
}

// chunkDurationSec はチャンク分割時の1チャンクあたりの秒数
// 同期 Recognize API は最大 60 秒なので余裕を持って 55 秒に設定
const chunkDurationSec = 55

// RecognizeWithTimestamps は PCM 音声（s16le, mono）を文字起こしし、単語レベルのタイムスタンプを返す。
// 音声が 55 秒を超える場合は自動的にチャンク分割して処理する。
func (c *googleSTTClient) RecognizeWithTimestamps(ctx context.Context, pcmData []byte, sampleRate int) ([]WordTimestamp, error) {
	bytesPerSec := sampleRate * 1 * 2 // mono, s16le
	maxChunkBytes := chunkDurationSec * bytesPerSec
	if len(pcmData) <= maxChunkBytes {
		return c.recognizeChunk(ctx, pcmData, sampleRate, 0)
	}

	// チャンク分割して順次処理し、タイムスタンプをオフセット調整してマージ
	chunkBytes := chunkDurationSec * bytesPerSec
	// 2バイトアライメントに合わせる
	chunkBytes = (chunkBytes / 2) * 2

	var allWords []WordTimestamp
	for offset := 0; offset < len(pcmData); offset += chunkBytes {
		end := offset + chunkBytes
		if end > len(pcmData) {
			end = len(pcmData)
		}

		chunkOffset := time.Duration(offset) * time.Second / time.Duration(bytesPerSec)
		words, err := c.recognizeChunk(ctx, pcmData[offset:end], sampleRate, chunkOffset)
		if err != nil {
			return nil, err
		}
		allWords = append(allWords, words...)
	}

	if len(allWords) == 0 {
		return nil, fmt.Errorf("音声認識結果が空です")
	}

	return allWords, nil
}

// recognizeChunk は単一チャンクの PCM データを STT で認識し、タイムスタンプにオフセットを加算して返す
func (c *googleSTTClient) recognizeChunk(ctx context.Context, pcmData []byte, sampleRate int, timeOffset time.Duration) ([]WordTimestamp, error) {
	recognizer := fmt.Sprintf(
		"projects/%s/locations/%s/recognizers/_",
		c.projectID, c.location,
	)

	req := &speechpb.RecognizeRequest{
		Recognizer: recognizer,
		Config: &speechpb.RecognitionConfig{
			DecodingConfig: &speechpb.RecognitionConfig_ExplicitDecodingConfig{
				ExplicitDecodingConfig: &speechpb.ExplicitDecodingConfig{
					Encoding:          speechpb.ExplicitDecodingConfig_LINEAR16,
					SampleRateHertz:   int32(sampleRate),
					AudioChannelCount: 1,
				},
			},
			Model:         "long",
			LanguageCodes: []string{"ja-JP"},
			Features: &speechpb.RecognitionFeatures{
				EnableWordTimeOffsets: true,
			},
		},
		AudioSource: &speechpb.RecognizeRequest_Content{
			Content: pcmData,
		},
	}

	resp, err := c.client.Recognize(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("音声認識に失敗しました: %w", err)
	}

	var words []WordTimestamp
	for _, result := range resp.GetResults() {
		if len(result.GetAlternatives()) == 0 {
			continue
		}
		alt := result.GetAlternatives()[0]
		for _, w := range alt.GetWords() {
			words = append(words, WordTimestamp{
				Word:      w.GetWord(),
				StartTime: w.GetStartOffset().AsDuration() + timeOffset,
				EndTime:   w.GetEndOffset().AsDuration() + timeOffset,
			})
		}
	}

	return words, nil
}
