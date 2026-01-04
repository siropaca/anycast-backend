# システム設定

このドキュメントでは、Anycast Backend のシステムレベルの設定について記載する。

## HTTP サーバー

### タイムアウト設定

| 設定 | 値 | 説明 |
|------|------|------|
| ReadTimeout | 10秒 | リクエストボディの読み取りタイムアウト |
| WriteTimeout | 180秒 | レスポンスの書き込みタイムアウト |
| IdleTimeout | 60秒 | Keep-Alive 接続のアイドルタイムアウト |

- `WriteTimeout` は LLM による台本生成（最大 120秒）を考慮して長めに設定している
- 設定箇所: `main.go`

## 外部サービス

### OpenAI API

| 設定 | 値 | 説明 |
|------|------|------|
| モデル | GPT-4o | 台本生成に使用する LLM モデル |
| タイムアウト | 120秒 | API リクエストのタイムアウト |
| リトライ回数 | 3回 | エラー時の最大リトライ回数 |
| リトライ間隔 | 1秒, 2秒, 3秒 | 指数バックオフ（attempt × 1秒） |

- 設定箇所: `internal/infrastructure/llm/openai_client.go`

### Google Cloud Text-to-Speech

（未実装）

### Google Cloud Storage

（未実装）

## タイムアウトの関係性

```
クライアント ─── HTTP サーバー ─── OpenAI API
                  │                    │
            WriteTimeout: 180s    Timeout: 120s
```

- HTTP サーバーの `WriteTimeout`（180秒）> OpenAI API のタイムアウト（120秒）
- これにより、LLM 生成中にサーバー側でタイムアウトすることを防いでいる
