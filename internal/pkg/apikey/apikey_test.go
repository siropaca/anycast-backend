package apikey

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	result, err := Generate()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	// 平文が ak_ プレフィックスで始まること
	if !strings.HasPrefix(result.PlainText, "ak_") {
		t.Errorf("PlainText should start with 'ak_', got %q", result.PlainText)
	}

	// 平文の長さ: "ak_" (3) + 64 hex chars (32 bytes) = 67
	if len(result.PlainText) != 67 {
		t.Errorf("PlainText length should be 67, got %d", len(result.PlainText))
	}

	// ハッシュが 64 文字（SHA-256 hex）であること
	if len(result.Hash) != 64 {
		t.Errorf("Hash length should be 64, got %d", len(result.Hash))
	}

	// プレフィックスが "ak_" + 8文字 + "..." であること
	if !strings.HasPrefix(result.Prefix, "ak_") {
		t.Errorf("Prefix should start with 'ak_', got %q", result.Prefix)
	}
	if !strings.HasSuffix(result.Prefix, "...") {
		t.Errorf("Prefix should end with '...', got %q", result.Prefix)
	}

	// HashKey が同じ平文から同じハッシュを返すこと
	hash := HashKey(result.PlainText)
	if hash != result.Hash {
		t.Errorf("HashKey() returned different hash: got %q, want %q", hash, result.Hash)
	}
}

func TestGenerate_Uniqueness(t *testing.T) {
	r1, err := Generate()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	r2, err := Generate()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	if r1.PlainText == r2.PlainText {
		t.Error("two generated keys should be different")
	}

	if r1.Hash == r2.Hash {
		t.Error("two generated hashes should be different")
	}
}

func TestHashKey(t *testing.T) {
	// 同じ入力に対して同じハッシュを返すこと
	key := "ak_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	h1 := HashKey(key)
	h2 := HashKey(key)

	if h1 != h2 {
		t.Error("HashKey should return the same hash for the same input")
	}

	// 異なる入力に対して異なるハッシュを返すこと
	h3 := HashKey(key + "x")
	if h1 == h3 {
		t.Error("HashKey should return different hashes for different inputs")
	}
}
