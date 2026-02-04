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

### LLM API（マルチプロバイダ）

`llm.Registry` で複数プロバイダ（OpenAI / Claude / Gemini）のクライアントを管理する。API キーが設定されたプロバイダが起動時に自動登録される。Phase ごとに使用するプロバイダを `internal/service/script_prompts.go` の `PhaseConfig` で切り替え可能。

| 設定 | 値 | 説明 |
|------|------|------|
| OpenAI モデル | GPT-5.2 | OpenAI のデフォルトモデル |
| Claude モデル | Claude Sonnet 4 | Claude のデフォルトモデル |
| Gemini モデル | Gemini 2.5 Flash | Gemini のデフォルトモデル |
| タイムアウト | 120秒 | API リクエストのタイムアウト |
| リトライ回数 | 3回 | エラー時の最大リトライ回数 |
| リトライ間隔 | 1秒, 2秒, 3秒 | 指数バックオフ（attempt × 1秒） |

Phase 別設定（`internal/service/script_prompts.go`）:

| Phase | Provider | Temperature | 用途 |
|-------|----------|-------------|------|
| Phase 2 | OpenAI | 0.9 | 素材+アウトライン生成 |
| Phase 3 | OpenAI | 0.7 | 台本ドラフト生成 |
| Phase 4 | OpenAI | 0.5 | QA パッチ修正 |

- 設定箇所: `internal/infrastructure/llm/`、`internal/service/script_prompts.go`

### Google Cloud Text-to-Speech

（未実装）

### Google Cloud Storage

（未実装）

## タイムアウトの関係性

```
クライアント ─── HTTP サーバー ─── LLM API（OpenAI / Claude / Gemini）
                  │                    │
            WriteTimeout: 180s    Timeout: 120s
```

- HTTP サーバーの `WriteTimeout`（180秒）> LLM API のタイムアウト（120秒）
- これにより、LLM 生成中にサーバー側でタイムアウトすることを防いでいる
