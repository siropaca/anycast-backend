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
| Gemini ロケーション | 環境変数 `GEMINI_LLM_LOCATION`（デフォルト: asia-northeast1） | Gemini LLM のリージョン |
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

### Google Cloud Text-to-Speech（Gemini TTS）

Gemini 2.5 Pro TTS（Vertex AI バックエンド）を使用した音声合成。マルチスピーカー合成に対応し、32k token の長い台本をサポート。

| 設定 | 値 | 説明 |
|------|------|------|
| モデル | gemini-2.5-pro-tts | Gemini TTS モデル |
| 言語コード | ja-JP | 日本語固定 |
| ロケーション | 環境変数 `GOOGLE_CLOUD_TTS_LOCATION`（デフォルト: global） | Vertex AI エンドポイントのリージョン |

- 設定箇所: `internal/infrastructure/tts/`

### AI 画像生成（マルチプロバイダ）

チャンネル・エピソードのアートワークをテキストプロンプトから生成する。環境変数 `IMAGE_GEN_PROVIDER` でプロバイダを切り替え可能。

| 設定 | 値 | 説明 |
|------|------|------|
| プロバイダ | 環境変数 `IMAGE_GEN_PROVIDER`（デフォルト: gemini） | 画像生成プロバイダ（`gemini` / `openai`） |

#### Gemini（デフォルト）

Gemini 2.5 Flash Image（Vertex AI バックエンド）を使用。

| 設定 | 値 | 説明 |
|------|------|------|
| モデル | gemini-2.5-flash-image | Gemini 画像生成モデル |
| ロケーション | 環境変数 `GEMINI_IMAGE_GEN_LOCATION`（デフォルト: us-central1） | Vertex AI エンドポイントのリージョン |
| 出力形式 | ResponseModalities: TEXT, IMAGE | テキストと画像の両方を生成可能 |

#### OpenAI

OpenAI の画像生成 API を使用。

| 設定 | 値 | 説明 |
|------|------|------|
| モデル | 環境変数 `OPENAI_IMAGE_GEN_MODEL`（デフォルト: gpt-image-1） | OpenAI 画像生成モデル |

- 設定箇所: `internal/infrastructure/imagegen/`

### Google Cloud Storage（GCS）

メディアファイル（音声・画像）の保存先。署名付き URL による安全なアクセスを提供。

| 設定 | 値 | 説明 |
|------|------|------|
| 署名付き URL 有効期限 | 1 時間 | V4 署名スキームを使用 |
| 音声パス形式 | `audios/{audioID}.mp3` | MP3 固定 |
| 画像パス形式 | `images/{imageID}{ext}` | 拡張子は元ファイルに準拠 |

- 設定箇所: `internal/infrastructure/storage/`

### Slack 通知

各種イベントを Slack に Webhook で通知する。対応する環境変数が空の場合、その通知は無効化される。

| 設定 | 環境変数 | 説明 |
|------|------|------|
| フィードバック通知 | `SLACK_FEEDBACK_WEBHOOK_URL` | ユーザーからのフィードバック送信時に通知 |
| お問い合わせ通知 | `SLACK_CONTACT_WEBHOOK_URL` | お問い合わせ送信時に通知 |
| アラート通知 | `SLACK_ALERT_WEBHOOK_URL` | ジョブ失敗時にアラート通知 |

- 設定箇所: `internal/infrastructure/slack/`

### トレーサー

台本生成の各 Phase のデータ（プロンプト、レスポンス、中間成果物）を出力するユーティリティ。環境変数 `TRACE_MODE` で動作を制御する。

| モード | 説明 |
|--------|------|
| `none`（デフォルト） | 何も出力しない |
| `log` | slog の Debug レベルでトレース出力 |
| `file` | `tmp/traces/{エピソードタイトル}/` 配下に Phase ごとの Markdown ファイルを出力 |

- 設定箇所: `internal/pkg/tracer/`、`internal/service/script_job.go`

## タイムアウトの関係性

```
クライアント ─── HTTP サーバー ─── LLM API（OpenAI / Claude / Gemini）
                  │                    │
            WriteTimeout: 180s    Timeout: 120s
```

- HTTP サーバーの `WriteTimeout`（180秒）> LLM API のタイムアウト（120秒）
- これにより、LLM 生成中にサーバー側でタイムアウトすることを防いでいる
