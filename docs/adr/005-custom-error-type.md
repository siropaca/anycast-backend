# ADR-005: カスタムエラー型

## ステータス

Accepted

## コンテキスト

REST API において、一貫したエラーレスポンスを返す必要があった。

以下の要件を満たす必要があった:
- 統一されたエラーレスポンス形式
- エラーコードによる分類
- HTTP ステータスコードとの対応
- 元のエラー情報の保持（デバッグ用）
- 詳細情報の付与（バリデーションエラーなど）

## 決定

**カスタムエラー型 `AppError`** を定義し、アプリケーション全体で使用する。

```go
type AppError struct {
    Code       string      `json:"code"`
    Message    string      `json:"message"`
    HTTPStatus int         `json:"-"`
    Details    interface{} `json:"details,omitempty"`
    Err        error       `json:"-"`
}
```

定義済みエラーを用意し、必要に応じてカスタマイズする:

```go
var ErrNotFound = &AppError{
    Code:       "NOT_FOUND",
    Message:    "Resource not found",
    HTTPStatus: http.StatusNotFound,
}

// 使用例
return apperror.ErrNotFound.WithMessage("Voice not found")
```

## 選択肢

### 選択肢 1: カスタムエラー型

- 一貫したエラーハンドリング
- 型安全
- 拡張性が高い
- 定型コードが必要

### 選択肢 2: 標準エラー + エラーラッピング

- シンプル
- Go の慣習に沿う
- HTTP ステータスとの対応が困難
- エラーレスポンスの構造化が手動

### 選択肢 3: エラーハンドリングライブラリ（pkg/errors など）

- スタックトレース付与
- 広く使われている
- HTTP API 向けの機能が不足
- 追加依存

### 選択肢 4: gRPC ステータスコードの流用

- 豊富なエラーコード
- 詳細情報のサポート
- REST API には過剰
- gRPC 依存

## 理由

1. **API 設計との整合性**: `docs/api.md` で定義したエラーコードを実装に反映
2. **型安全性**: コンパイル時にエラー型を検証可能
3. **柔軟性**: `WithMessage`, `WithDetails`, `WithError` でカスタマイズ可能
4. **デバッグ容易性**: 元のエラーを保持しつつ、クライアントには安全なメッセージを返す

## 結果

- `internal/apperror/error.go` に `AppError` 型を定義
- `internal/apperror/codes.go` に定義済みエラーを列挙
- Handler では `handler.Error(c, err)` で統一的にエラーレスポンスを返す
- 新しいエラーコードが必要な場合は `codes.go` に追加
