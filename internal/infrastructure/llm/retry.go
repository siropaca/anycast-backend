package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

const (
	// リトライ回数
	maxRetries = 3
)

// retryWithBackoff は LLM API 呼び出しをリトライ付きで実行する
//
// 空文字列レスポンスもリトライ対象。最大 maxRetries 回リトライし、
// 線形バックオフ（attempt 秒）で待機する。
func retryWithBackoff(ctx context.Context, providerName string, fn func() (string, error)) (string, error) {
	log := logger.FromContext(ctx)

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Debug("generating with LLM", "provider", providerName, "attempt", attempt)

		result, err := fn()
		if err != nil {
			lastErr = err
			log.Warn("LLM API error", "provider", providerName, "attempt", attempt, "error", err)

			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}

			log.Error("LLM API failed after retries", "provider", providerName, "error", err)
			return "", apperror.ErrGenerationFailed.WithMessage("台本の生成に失敗しました").WithError(err)
		}

		if result == "" {
			lastErr = fmt.Errorf("empty response from %s", providerName)
			log.Warn("LLM response is empty", "provider", providerName, "attempt", attempt)

			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}

			log.Error("LLM returned empty response after retries", "provider", providerName)
			return "", apperror.ErrGenerationFailed.WithMessage("台本の生成に失敗しました: レスポンスがありません")
		}

		log.Debug("LLM generation successful", "provider", providerName, "content_length", len(result))
		return result, nil
	}

	return "", apperror.ErrGenerationFailed.WithMessage("台本の生成に失敗しました").WithError(lastErr)
}
