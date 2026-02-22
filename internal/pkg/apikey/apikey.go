package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	prefix    = "ak_"
	randBytes = 32
	// 表示用プレフィックスの長さ（ak_ + 先頭8文字 + ...）
	displayPrefixHexLen = 8
)

// GenerateResult は API キー生成結果を保持する
type GenerateResult struct {
	PlainText string
	Hash      string
	Prefix    string
}

// Generate は新しい API キーを生成し、平文・ハッシュ・表示用プレフィックスを返す
func Generate() (GenerateResult, error) {
	b := make([]byte, randBytes)
	if _, err := rand.Read(b); err != nil {
		return GenerateResult{}, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	hexStr := hex.EncodeToString(b)
	plainText := prefix + hexStr

	return GenerateResult{
		PlainText: plainText,
		Hash:      HashKey(plainText),
		Prefix:    plainText[:len(prefix)+displayPrefixHexLen] + "...",
	}, nil
}

// HashKey は API キーの平文から SHA-256 ハッシュを計算する
func HashKey(plainText string) string {
	h := sha256.Sum256([]byte(plainText))
	return hex.EncodeToString(h[:])
}
