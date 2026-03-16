# インフラストラクチャ設計

Anycast Backend のインフラ構成と各サービスの設計。
システム設定（タイムアウト・グレースフルシャットダウン）については [system.md](system.md) を参照。

## 構成図

```
┌──────────────────────────────────────────────────────────────────────────┐
│                           Railway                                        │
│                                                                          │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐               │
│  │   Backend    │    │  PostgreSQL  │    │    Redis     │               │
│  │  (Go / Gin)  │───▶│     16       │    │      7       │               │
│  │              │───▶│              │    │              │               │
│  └──────┬───────┘    └──────────────┘    └──────────────┘               │
│         │                                                                │
└─────────┼────────────────────────────────────────────────────────────────┘
          │
          │ HTTPS
          ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                        Google Cloud Platform                             │
│                                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                  │
│  │  Cloud Tasks │  │     GCS      │  │  Vertex AI   │                  │
│  │ (ジョブキュー) │  │ (メディア保存) │  │              │                  │
│  └──────────────┘  └──────────────┘  │ ・Gemini TTS  │                  │
│                                       │ ・Gemini LLM  │                  │
│  ┌──────────────┐                    │ ・Gemini Image│                  │
│  │ Speech-to-   │                    │ ・Cloud STT   │                  │
│  │ Text v2      │                    └──────────────┘                  │
│  └──────────────┘                                                       │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────┐
│                       外部 SaaS                                          │
│                                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │
│  │   OpenAI     │  │   Claude     │  │  ElevenLabs  │  │   Slack    │ │
│  │ (LLM/画像)   │  │   (LLM)     │  │   (TTS)      │  │ (通知)     │ │
│  └──────────────┘  └──────────────┘  └──────────────┘  └────────────┘ │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## ホスティング: Railway

Backend・PostgreSQL・Redis をすべて Railway 上で運用する。

### Backend

| 項目 | 値 |
|------|------|
| ビルド | Dockerfile（マルチステージビルド、Go 1.24-bookworm） |
| ランタイム | debian:bookworm-slim + FFmpeg |
| ヘルスチェック | `GET /health`（タイムアウト 30 秒） |
| 再起動ポリシー | 失敗時に最大 3 回まで再起動 |
| 起動手順 | golang-migrate で DB マイグレーション → サーバー起動 |

### PostgreSQL

| 項目 | 値 |
|------|------|
| バージョン | 16 |
| 接続 | 環境変数 `DATABASE_URL` |
| ローカル | Docker Compose（ポート 5433） |

### Redis

| 項目 | 値 |
|------|------|
| バージョン | 7 |
| 用途 | キャッシュ（ボイス・チャンネル・カテゴリ・エピソード） |
| 接続 | 環境変数 `REDIS_URL` |
| 未設定時 | no-op クライアント（キャッシュ無効） |
| ローカル | Docker Compose（ポート 6379） |

---

## Google Cloud Platform

### Cloud Tasks（非同期ジョブキュー）

台本生成・音声生成の非同期ジョブをキューイングする。
タスク完了後に Backend のワーカーエンドポイントへ OIDC 認証付きで HTTP コールバックする。

| 項目 | 値 |
|------|------|
| ロケーション | asia-northeast1（デフォルト） |
| キュー名 | audio-generation-queue（デフォルト、台本・音声共通） |
| 認証 | OIDC（Service Account） |
| ワーカー URL | `{GOOGLE_CLOUD_TASKS_WORKER_URL}/audio` または `/script` |
| ローカル代替 | goroutine で直接実行（`GOOGLE_CLOUD_TASKS_WORKER_URL` 未設定時） |

### Google Cloud Storage（メディア保存）

音声ファイル・画像ファイルの永続化ストレージ。

| 項目 | 値 |
|------|------|
| バケット | 環境変数 `GOOGLE_CLOUD_STORAGE_BUCKET_NAME` |
| 音声パス | `audios/{audioID}.mp3` |
| 画像パス | `images/{imageID}{ext}` |
| アクセス | 署名付き URL（V4 スキーム、有効期限 1 時間） |

### Vertex AI

GCP の AI サービスは Vertex AI 経由で利用する。

| サービス | モデル | ロケーション | 用途 |
|----------|--------|-------------|------|
| Gemini TTS | gemini-2.5-pro-tts | global | 音声合成 |
| Gemini LLM | gemini-2.5-flash | asia-northeast1 | 台本生成（Phase 設定で切替） |
| Gemini Image | gemini-2.5-flash-image | us-central1 | 画像生成 |

### Cloud Speech-to-Text v2

マルチスピーカー音声生成時の行分割に使用する。
詳細は [audio-generation-pipeline.md](audio-generation-pipeline.md) を参照。

| 項目 | 値 |
|------|------|
| API | Speech-to-Text v2 |
| モデル | long |
| 言語 | ja-JP |
| 用途 | TTS 音声の単語レベルタイムスタンプ取得 |

---

## 外部 SaaS

### LLM（台本生成）

`llm.Registry` で複数プロバイダを管理し、台本生成の Phase ごとにプロバイダを切り替える。
API キーが設定されたプロバイダのみ起動時に自動登録される。

| プロバイダ | 有効化条件 | モデル |
|-----------|-----------|--------|
| OpenAI | `OPENAI_API_KEY` 設定時 | GPT-5.2 |
| Claude | `CLAUDE_API_KEY` 設定時 | Claude Sonnet 4.6 |
| Gemini | `GOOGLE_CLOUD_PROJECT_ID` 設定時 | Gemini 2.5 Flash |

Phase 別のプロバイダ割り当ては [script-prompt-workflow.md](script-prompt-workflow.md) を参照。

### TTS（音声合成）

`tts.Registry` で複数プロバイダを管理する。
キャラクターの Voice に紐づく Provider から動的にプロバイダを選択する。

| プロバイダ | 有効化条件 | 出力形式 |
|-----------|-----------|----------|
| Gemini TTS | `GOOGLE_CLOUD_PROJECT_ID` 設定時 | PCM 24kHz → MP3 変換 |
| ElevenLabs | `ELEVENLABS_API_KEY` 設定時 | MP3 44.1kHz 128kbps |

### 画像生成

`imagegen.Registry` で複数プロバイダを管理する。
環境変数 `IMAGE_GEN_PROVIDER` でプロバイダを切り替える。

| プロバイダ | 有効化条件 | モデル |
|-----------|-----------|--------|
| Gemini（デフォルト） | `GOOGLE_CLOUD_PROJECT_ID` 設定時 | gemini-2.5-flash-image |
| OpenAI | `OPENAI_API_KEY` 設定時 | gpt-image-1（デフォルト） |

### Slack（通知）

各種イベントを Slack Webhook で通知する。
対応する環境変数が未設定の場合、その通知チャンネルは無効化される。

| 通知 | 環境変数 | トリガー |
|------|---------|---------|
| フィードバック | `SLACK_FEEDBACK_WEBHOOK_URL` | フィードバック送信時 |
| お問い合わせ | `SLACK_CONTACT_WEBHOOK_URL` | お問い合わせ送信時 |
| アラート | `SLACK_ALERT_WEBHOOK_URL` | ジョブ失敗時 |
| 新規登録 | `SLACK_REGISTRATION_WEBHOOK_URL` | ユーザー登録時 |

---

## リアルタイム通信: WebSocket

ジョブ（台本生成・音声生成）の進捗をリアルタイムでクライアントに通知する。

| 項目 | 値 |
|------|------|
| エンドポイント | `GET /ws/jobs?token={jwt}` |
| 認証 | JWT トークン（クエリパラメータ） |
| Write 待機 | 10 秒 |
| Pong 待機 | 60 秒 |
| Ping 間隔 | 54 秒 |
| 最大メッセージサイズ | 512 バイト |
| 送信方式 | ユーザー ID 単位（`SendToUser`） |

---

## マルチプロバイダアーキテクチャ

LLM・TTS・画像生成はすべて **Registry パターン** で複数プロバイダを管理する。

```
Registry
  ├── Provider A（API キー設定時に自動登録）
  ├── Provider B（API キー設定時に自動登録）
  └── Provider C（API キー設定時に自動登録）
```

- API キーが設定されたプロバイダのみ起動時に登録
- 未設定のプロバイダは登録されず、使用不可
- 実行時にプロバイダを動的に選択（Voice の Provider 属性、Phase 設定、環境変数）

---

## ローカル開発時の構成

| コンポーネント | 本番 | ローカル代替 |
|--------------|------|-------------|
| PostgreSQL | Railway | Docker Compose（ポート 5433） |
| Redis | Railway | Docker Compose（ポート 6379）、または未設定で無効化 |
| Cloud Tasks | GCP | goroutine 直接実行 |
| GCS | GCP | GCP 接続（ローカル代替なし） |
| Slack | Webhook | 未設定で無効化 |
| ホットリロード | - | Air（`.air.toml`） |

---

## 関連ドキュメント

| ドキュメント | 説明 |
|-------------|------|
| [system.md](system.md) | タイムアウト・グレースフルシャットダウン・外部サービス詳細設定 |
| [audio-generation-pipeline.md](audio-generation-pipeline.md) | 音声生成パイプラインの詳細 |
| [script-generate-async-api.md](script-generate-async-api.md) | 台本生成 API の詳細設計 |
| [ADR-018](../adr/018-monolith-with-cloud-tasks.md) | モノリス + Cloud Tasks 構成の技術選定 |
| [ADR-020](../adr/020-graceful-shutdown.md) | グレースフルシャットダウンの実装 |
| [ADR-021](../adr/021-cache-redis.md) | Redis キャッシュ基盤の導入 |
