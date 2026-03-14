# Anycast Backend

AIポッドキャスト作成・配信プラットフォーム「Anycast」のバックエンド API サーバーです。

## 関連リポジトリ

- [anycast-frontend](https://github.com/siropaca/anycast-frontend) - フロントエンド
- [anycast-mcp-server](https://github.com/siropaca/anycast-mcp-server) - MCP Server

## 技術スタック

- **言語**: Go 1.24
- **フレームワーク**: Gin
- **ORM**: GORM
- **マイグレーション**: golang-migrate
- **API**: REST API
- **API ドキュメント**: Swagger (swaggo/swag)
- **DB**: PostgreSQL
- **ストレージ**: GCS（Google Cloud Storage）
- **TTS**: Gemini TTS / ElevenLabs（プロバイダ切替可能）
- **画像生成**: Gemini / OpenAI（プロバイダ切替可能）
- **LLM**: OpenAI / Claude / Gemini（Phase ごとにプロバイダ切替可能）
- **キャッシュ**: Redis (go-redis/v9)
- **ジョブキュー**: Google Cloud Tasks
- **WebSocket**: gorilla/websocket
- **音声処理**: FFmpeg
- **バージョン管理**: mise
- **静的解析**: golangci-lint
- **ローカル環境**: Docker Compose
- **ホットリロード**: Air
- **ホスティング**: Railway

## アーキテクチャ

レイヤードアーキテクチャ + 軽量 DDD を採用しています。

```
Client (Frontend / MCP Server)
  ↓ HTTP / WebSocket
Router / Middleware
  ↓
Handler
  ↓
Service
  ├→ Repository → PostgreSQL (Railway)
  └→ Infrastructure → External Services
```

| レイヤー | 責務 |
|----------|------|
| Handler | HTTP リクエスト/レスポンス処理 |
| Service | ビジネスロジック |
| Repository | データアクセス（GORM） |
| Infrastructure | 外部サービス連携（LLM / TTS / STT / 画像生成 / ストレージ / ジョブキュー / 通知） |
| Model | ドメインモデル |
| DTO | リクエスト/レスポンス構造体 |

### インフラストラクチャ

| カテゴリ | サービス | 用途 |
|----------|----------|------|
| ホスティング | Railway | Backend / PostgreSQL / Redis |
| LLM | Vertex AI (Gemini) / OpenAI / Claude | 台本生成（Phase ごとにプロバイダ切替可能） |
| TTS | Vertex AI (Gemini TTS) / ElevenLabs | 音声合成（プロバイダ切替可能） |
| STT | Google Cloud Speech API | 音声認識 |
| 画像生成 | Vertex AI (Gemini) / OpenAI | サムネイル等の画像生成（プロバイダ切替可能） |
| ストレージ | Google Cloud Storage (GCS) | 音声・画像ファイルの保存 |
| ジョブキュー | Google Cloud Tasks | 台本生成・音声生成の非同期処理 |
| リアルタイム通信 | WebSocket | ジョブ進捗の通知 |
| 通知 | Slack Webhooks | フィードバック・アラート・お問い合わせ通知 |

> **Note:** Google Cloud の AI 関連サービス（LLM / TTS / STT / 画像生成）は Vertex AI 経由で利用しています。Cloud Tasks はジョブ完了後に Backend の Worker エンドポイントへコールバックし、結果を処理します。ローカル開発時は Cloud Tasks の代わりに goroutine で直接実行されます。

### 設計アプローチ

本プロジェクトでは **ドメインモデル駆動** で設計を行います。

```
ドメインモデル設計 → API 設計 → DB 設計
```

新しい機能を追加する際は、まずドメインモデルを設計し、それを永続化・公開するための手段として API と DB を設計します。  
詳細は [docs/domain-model/INDEX.md](docs/domain-model/INDEX.md) を参照。

アーキテクチャ決定の詳細は [docs/adr/](docs/adr/) を参照。

## セットアップ

### 前提条件

- [mise](https://mise.jdx.dev/) がインストールされていること
- [Docker](https://www.docker.com/) および Docker Compose がインストールされていること
- [FFmpeg](https://ffmpeg.org/) がインストールされていること（音声ミキシング機能で使用）
  - macOS: `brew install ffmpeg`
  - Ubuntu/Debian: `sudo apt install ffmpeg`

### インストール

```bash
# 自動セットアップ（推奨）
make bootstrap  # または make bs

# または手動でセットアップ
mise trust && mise install  # Go とツールのインストール
go mod download             # 依存関係のダウンロード
cp .env.example .env        # 環境変数ファイルの作成
```

> **Note**: シェルで `mise activate` を実行するか、`.bashrc` / `.zshrc` に設定を追加してください。

### 環境変数

| 変数 | 説明 | デフォルト |
|------|------|------------|
| `PORT` | サーバーのポート番号 | 8081 |
| `DATABASE_URL` | PostgreSQL 接続 URL | - |
| `DB_LOG_LEVEL` | DB クエリのログレベル（silent / error / warn / info） | silent |
| `REDIS_URL` | Redis 接続 URL（未設定時はキャッシュ無効） | - |
| `APP_ENV` | 環境（development / production） | development |
| `AUTH_SECRET` | JWT 検証用シークレットキー（フロントエンドの AUTH_SECRET と同じ値） | - |
| `CORS_ALLOWED_ORIGINS` | CORS 許可オリジン（カンマ区切りで複数指定可能） | http://localhost:3210 |
| `OPENAI_API_KEY` | OpenAI API キー（設定すると OpenAI プロバイダが有効化される） | - |
| `CLAUDE_API_KEY` | Claude API キー（設定すると Claude プロバイダが有効化される） | - |
| `IMAGE_GEN_PROVIDER` | 画像生成プロバイダ（`gemini` / `openai`） | gemini |
| `OPENAI_IMAGE_GEN_MODEL` | OpenAI 画像生成モデル | gpt-image-1 |
| `GEMINI_IMAGE_GEN_LOCATION` | Gemini 画像生成のロケーション | us-central1 |
| `GEMINI_LLM_LOCATION` | Gemini LLM のロケーション | asia-northeast1 |
| `GOOGLE_CLOUD_PROJECT_ID` | GCP プロジェクト ID | - |
| `GOOGLE_CLOUD_CREDENTIALS_JSON` | サービスアカウントの JSON キー | - |
| `GOOGLE_CLOUD_STORAGE_BUCKET_NAME` | GCS バケット名 | - |
| `GOOGLE_CLOUD_TASKS_LOCATION` | Cloud Tasks ロケーション | asia-northeast1 |
| `GOOGLE_CLOUD_TASKS_QUEUE_NAME` | Cloud Tasks キュー名 | audio-generation-queue |
| `GOOGLE_CLOUD_TASKS_SERVICE_ACCOUNT_EMAIL` | Cloud Tasks サービスアカウントメール | - |
| `GOOGLE_CLOUD_TASKS_WORKER_URL` | ワーカーエンドポイントのベース URL（末尾に `/audio` や `/script` が付与される） | - |
| `GOOGLE_CLOUD_TTS_LOCATION` | Gemini TTS のロケーション | global |
| `ELEVENLABS_API_KEY` | ElevenLabs API キー（設定すると ElevenLabs プロバイダが有効化される） | - |
| `TRACE_MODE` | トレースモード（none / log / file） | none |
| `SLACK_FEEDBACK_WEBHOOK_URL` | Slack Webhook URL（フィードバック通知用、空の場合は通知無効） | - |
| `SLACK_CONTACT_WEBHOOK_URL` | Slack Webhook URL（お問い合わせ通知用、空の場合は通知無効） | - |
| `SLACK_ALERT_WEBHOOK_URL` | Slack Webhook URL（ジョブ失敗アラート通知用、空の場合はアラート無効） | - |
| `SLACK_REGISTRATION_WEBHOOK_URL` | Slack Webhook URL（新規登録通知用、空の場合は通知無効） | - |

> **Note:** `GOOGLE_CLOUD_PROJECT_ID` と `GOOGLE_CLOUD_TASKS_WORKER_URL` が未設定の場合、Cloud Tasks を使わずに goroutine で直接ジョブを実行します（ローカル開発用）。

### DB の起動

```bash
docker compose up -d
```

### マイグレーション

```bash
make migrate-up
```

### 開発サーバーの起動

```bash
make dev
```

サーバーは http://localhost:8081 で起動します。

### ローカル開発時のデバッグ出力

ローカル開発時、以下のディレクトリにデバッグ用ファイルが出力されます（`tmp/` は `.gitignore` に含まれています）。

| ディレクトリ | 内容 | 条件 |
|-------------|------|------|
| `tmp/audio-debug/{jobID}/` | TTS 完了直後のスピーカー別音源（WAV） | `APP_ENV` が `production` 以外 |
| `tmp/traces/{エピソードタイトル}/` | 台本生成の Phase ごとのトレース（Markdown） | `TRACE_MODE=file` |

## API ドキュメント

### Swagger UI

開発サーバー起動後、以下の URL で Swagger UI にアクセスできます。

```
http://localhost:8081/swagger/index.html
```

API の仕様確認やインタラクティブなテストが可能です。

### ドキュメントの更新

Handler に Swagger アノテーションを追加した後、以下のコマンドでドキュメントを再生成します。

```bash
make swagger
```

### API エンドポイント

詳細は [docs/api/](docs/api/) を参照。

## コマンド一覧

| コマンド | 説明 |
|----------|------|
| `make bootstrap` | 開発環境をセットアップ |
| `make dev` | 開発サーバーを起動（ホットリロード） |
| `make run` | サーバーを起動 |
| `make build` | バイナリをビルド |
| `make test` | テストを実行 |
| `make fmt` | コードをフォーマット |
| `make lint` | 静的解析を実行 |
| `make lint-fix` | 静的解析を実行（自動修正あり） |
| `make tidy` | 依存関係を整理 |
| `make clean` | ビルド成果物を削除 |
| `make swagger` | Swagger ドキュメント生成 |
| `make migrate-up` | マイグレーション実行 |
| `make migrate-down` | マイグレーションロールバック |
| `make migrate-reset` | マイグレーションリセット（テーブル全削除 → 再マイグレーション） |
| `make migrate-reset-seed` | マイグレーションリセット + シード投入 |
| `make seed` | シードデータを投入（開発環境用） |
| `make token` | 開発用 JWT トークンを生成 |
| `make cleanup` | 孤児メディアファイルをクリーンアップ（dry-run） |
| `make cleanup-run` | 孤児メディアファイルをクリーンアップ（実行） |

### mise タスク

一部のコマンドは `mise run` でも実行できます。

| コマンド | 説明 |
|----------|------|
| `mise run bootstrap` (`mise run bs`) | 開発環境をセットアップ |
| `mise run clean` | ビルド成果物を削除 |

## ディレクトリ構成

```
.
├── main.go              # エントリーポイント
├── go.mod
├── go.sum
├── Makefile             # コマンド定義
├── .env.example         # 環境変数のサンプル
├── .air.toml            # Air 設定
├── .mise.toml           # mise 設定
├── docker-compose.yml   # ローカル DB
├── railway.toml         # Railway 設定
├── Dockerfile           # Docker ビルド設定
├── scripts/             # セットアップスクリプト
├── migrations/          # マイグレーションファイル
├── seeds/               # シードデータ（開発環境用）
├── docs/                # ドキュメント
│   ├── adr/             # Architecture Decision Records
│   ├── api/             # API 設計
│   ├── domain-model/    # ドメインモデル（集約・エンティティ・値オブジェクト）
│   └── specs/           # 仕様ドキュメント（DB 設計、非同期 API 詳細設計など）
├── swagger/             # Swagger ドキュメント（自動生成）
├── http/                # HTTP リクエストファイル
├── internal/            # 内部パッケージ
│   ├── apperror/        # カスタムエラー型
│   ├── config/          # 設定管理
│   ├── di/              # DI コンテナ
│   ├── dto/             # Data Transfer Objects
│   ├── handler/         # ハンドラー
│   ├── infrastructure/  # 外部サービス連携
│   │   ├── cloudtasks/  # Cloud Tasks クライアント
│   │   ├── imagegen/    # 画像生成クライアント（Gemini / OpenAI）
│   │   ├── llm/         # LLM クライアント（OpenAI / Claude / Gemini）
│   │   ├── slack/       # Slack 通知クライアント
│   │   ├── storage/     # GCS クライアント
│   │   ├── stt/         # STT クライアント（Google Cloud Speech）
│   │   ├── tts/         # TTS クライアント（Gemini / ElevenLabs）
│   │   └── websocket/   # WebSocket Hub
│   ├── middleware/      # ミドルウェア
│   ├── model/           # ドメインモデル
│   ├── pkg/             # 共通ユーティリティ
│   │   ├── apikey/      # API キー生成・検証
│   │   ├── audio/       # 音声ファイル処理
│   │   ├── cache/       # Redis キャッシュ
│   │   ├── crypto/      # パスワードハッシュ
│   │   ├── db/          # DB 接続
│   │   ├── jwt/         # JWT トークン管理
│   │   ├── logger/      # 構造化ログ
│   │   ├── prompt/      # プロンプト圧縮
│   │   ├── script/      # 台本パーサー
│   │   ├── token/       # トークン生成
│   │   ├── optional/    # nullable フィールド用ユーティリティ
│   │   ├── tracer/      # 台本生成トレーサー
│   │   └── uuid/        # UUID パース
│   ├── repository/      # データアクセス層
│   ├── router/          # ルーティング
│   └── service/         # ビジネスロジック層
├── README.md
├── CLAUDE.md
└── AGENTS.md
```
