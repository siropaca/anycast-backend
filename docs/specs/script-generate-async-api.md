# 台本生成 API (非同期)

このドキュメントでは、エピソードの台本を AI で非同期生成する API の仕様を記載する。

## 概要

台本生成は LLM（OpenAI GPT）を使用した処理のため、数秒〜数十秒かかる。
非同期ジョブとして実行し、クライアントはポーリングまたは WebSocket でジョブの進捗・完了を監視できる。

## システム構成

```
┌─────────────┐     ┌─────────────┐     ┌──────────────────┐
│   Client    │────▶│  API Server │────▶│  Google Cloud    │
│             │     │             │     │  Tasks           │
└─────────────┘     └─────────────┘     └──────────────────┘
      │                   │                      │
      │                   │                      ▼
      │                   │             ┌──────────────────┐
      │                   │◀────────────│  Worker          │
      │                   │             │  Endpoint        │
      │                   │             └──────────────────┘
      │                   │                      │
      │                   ▼                      ▼
      │             ┌─────────────┐     ┌──────────────────┐
      │             │  PostgreSQL │     │  OpenAI API      │
      │             │             │     │  (GPT)           │
      │             └─────────────┘     └──────────────────┘
      │                   │
      ▼                   │
┌─────────────┐           │
│  WebSocket  │◀──────────┘
│  (進捗通知)  │
└─────────────┘
```

## API エンドポイント

### ジョブ作成

エピソードの台本生成ジョブを作成する。

```
POST /channels/{channelId}/episodes/{episodeId}/script/generate-async
```

**認証**: 必須

**リクエストボディ**:

| フィールド | 型 | 必須 | 説明 |
|-----------|------|------|------|
| prompt | string | ○ | 台本のテーマ・内容の指示（最大 2000 文字） |
| durationMinutes | number | - | エピソードの長さ（3 〜 30 分、デフォルト: 10） |
| withEmotion | boolean | - | 感情タグを付与するか（デフォルト: false） |

**レスポンス**: `202 Accepted`

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "episodeId": "660e8400-e29b-41d4-a716-446655440001",
    "status": "pending",
    "progress": 0,
    "prompt": "今日は AI の未来について語り合おう",
    "durationMinutes": 10,
    "withEmotion": false,
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
}
```

**エラー**:

| コード | 説明 |
|-------|------|
| 400 | バリデーションエラー（prompt なし、キャラクター未設定、既に処理中のジョブあり等） |
| 403 | チャンネルへのアクセス権限なし |
| 404 | エピソードが存在しない |

### ジョブ詳細取得

```
GET /script-jobs/{jobId}
```

**認証**: 必須

**レスポンス**: `200 OK`

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "episodeId": "660e8400-e29b-41d4-a716-446655440001",
    "status": "completed",
    "progress": 100,
    "prompt": "今日は AI の未来について語り合おう",
    "durationMinutes": 10,
    "withEmotion": false,
    "episode": {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "title": "エピソードタイトル",
      "channel": {
        "id": "770e8400-e29b-41d4-a716-446655440002",
        "name": "チャンネル名"
      }
    },
    "scriptLinesCount": 42,
    "startedAt": "2024-01-01T00:00:01Z",
    "completedAt": "2024-01-01T00:00:15Z",
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:15Z"
  }
}
```

### ユーザーのジョブ一覧取得

```
GET /me/script-jobs
```

**認証**: 必須

**クエリパラメータ**:

| パラメータ | 型 | 説明 |
|-----------|------|------|
| status | string | フィルタ: pending, processing, completed, failed |

### 内部ワーカーエンドポイント

Cloud Tasks から呼び出される。

```
POST /internal/worker/script
```

**認証**: Cloud Tasks Service Account (OIDC)

**リクエストボディ**:

```json
{
  "jobId": "550e8400-e29b-41d4-a716-446655440000"
}
```

## WebSocket

リアルタイムで進捗を受け取るための WebSocket エンドポイント。
音声生成ジョブと共通のエンドポイントを使用する。

```
GET /ws/jobs?token={jwt}
```

### クライアント → サーバー

```json
// ジョブの購読開始
{"type": "subscribe", "payload": {"jobId": "..."}}

// ジョブの購読解除
{"type": "unsubscribe", "payload": {"jobId": "..."}}

// ヘルスチェック
{"type": "ping"}
```

### サーバー → クライアント

```json
// 進捗更新
{
  "type": "script_progress",
  "payload": {
    "jobId": "...",
    "progress": 50,
    "message": "AI で台本を生成中..."
  }
}

// 完了通知
{
  "type": "script_completed",
  "payload": {
    "jobId": "...",
    "scriptLinesCount": 42
  }
}

// 失敗通知
{
  "type": "script_failed",
  "payload": {
    "jobId": "...",
    "errorCode": "GENERATION_FAILED",
    "errorMessage": "生成された台本のパースに失敗しました"
  }
}

// ヘルスチェック応答
{"type": "pong"}
```

## ジョブステータス

```
pending ───▶ processing ───▶ completed
                  │
                  └──────────▶ failed
```

| ステータス | 説明 |
|-----------|------|
| pending | ジョブ作成済み、処理待ち |
| processing | 台本生成処理中 |
| completed | 処理完了 |
| failed | 処理失敗 |

## 処理フロー

### 進捗の目安

| 進捗 | 処理内容 |
|------|---------|
| 0% | ジョブ作成 |
| 10% | チャンネル・エピソード・ユーザー情報の読み込み |
| 20% | LLM 用プロンプトの構築 |
| 30% | AI による台本生成開始 |
| 70% | 生成テキストのパース |
| 85% | 台本行のデータベース保存 |
| 95% | 完了処理 |
| 100% | 完了 |

### 処理詳細

1. **データ読み込み**: チャンネル、エピソード、ユーザー情報を取得
2. **プロンプト構築**: ユーザー設定、チャンネル設定、キャラクター情報、エピソード情報を組み合わせてプロンプトを生成
3. **LLM 生成**: OpenAI GPT で台本テキストを生成
4. **テキストパース**: 生成されたテキストを `話者名: セリフ` 形式でパース
5. **データ保存**: 既存の台本行を削除し、新しい台本行をバッチ作成
6. **エピソード更新**: `userPrompt` を更新

### システムプロンプト

感情タグの有無で 2 種類のシステムプロンプトを使い分ける。

**出力形式（感情なし）**:
```
話者名: セリフ
```

**出力形式（感情あり）**:
```
話者名: [感情] セリフ
```

### 台本生成ガイドライン

- 1 分あたり約 300 文字程度のセリフ量
- 1 つのセリフは 20〜80 文字程度
- セリフは必ず句点（。）で終わる
- 自然なフィラー（「えーと」「まあ」等）を適度に含める
- 設定箇所: `internal/service/script_job.go`

## 外部サービス

### OpenAI API

| 設定 | 値 | 説明 |
|------|------|------|
| モデル | gpt-4o | GPT-4o モデル |
| Temperature | 0.8 | 創造性パラメータ |
| Max Tokens | 16384 | 最大出力トークン数 |

- 設定箇所: `internal/infrastructure/llm/openai_client.go`

### Google Cloud Tasks

| 設定 | 値 | 説明 |
|------|------|------|
| ロケーション | asia-northeast1 | デフォルト |
| キュー名 | async-jobs | 非同期ジョブ用キュー（台本・音声生成共通） |
| 認証 | OIDC | Service Account による認証 |
| ワーカー URL | {baseURL}/script | ベース URL + `/script` |

- 設定箇所: `internal/infrastructure/cloudtasks/client.go`
- ベース URL は環境変数 `GOOGLE_CLOUD_TASKS_WORKER_URL` で設定
- Cloud Tasks が未設定の場合（ローカル開発）は goroutine で直接実行

## エラーコード

| コード | HTTP | 説明 |
|-------|------|------|
| VALIDATION_ERROR | 400 | prompt なし、キャラクター未設定、重複ジョブ等 |
| UNAUTHORIZED | 401 | 認証エラー |
| FORBIDDEN | 403 | アクセス権限なし |
| NOT_FOUND | 404 | リソースが存在しない |
| GENERATION_FAILED | 500 | LLM 生成失敗、パース失敗 |
| INTERNAL_ERROR | 500 | その他の内部エラー |

## データベーススキーマ

### script_jobs テーブル

```sql
CREATE TYPE script_job_status AS ENUM ('pending', 'processing', 'completed', 'failed');

CREATE TABLE script_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    episode_id UUID NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status script_job_status NOT NULL DEFAULT 'pending',
    progress INTEGER NOT NULL DEFAULT 0,

    -- 生成パラメータ
    prompt TEXT NOT NULL,
    duration_minutes INTEGER NOT NULL DEFAULT 10,
    with_emotion BOOLEAN NOT NULL DEFAULT false,

    -- 結果
    error_message TEXT,
    error_code VARCHAR(50),

    -- タイムスタンプ
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_script_jobs_episode_id ON script_jobs (episode_id);
CREATE INDEX idx_script_jobs_user_id ON script_jobs (user_id);
CREATE INDEX idx_script_jobs_status ON script_jobs (status);
CREATE INDEX idx_script_jobs_created_at ON script_jobs (created_at DESC);
```

### script_lines テーブル

```sql
CREATE TABLE script_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    episode_id UUID NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
    line_order INTEGER NOT NULL,
    speaker_id UUID NOT NULL REFERENCES characters (id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    emotion TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (episode_id, line_order) DEFERRABLE INITIALLY DEFERRED
);
```

## 関連ファイル

| ファイル | 説明 |
|---------|------|
| `internal/handler/script_job.go` | REST API ハンドラー |
| `internal/handler/worker.go` | ワーカーエンドポイント |
| `internal/handler/websocket.go` | WebSocket ハンドラー |
| `internal/service/script_job.go` | ビジネスロジック |
| `internal/repository/script_job.go` | データベースアクセス |
| `internal/model/script_job.go` | データモデル |
| `internal/infrastructure/llm/openai_client.go` | OpenAI クライアント |
| `internal/infrastructure/cloudtasks/client.go` | Cloud Tasks クライアント |
| `internal/infrastructure/websocket/hub.go` | WebSocket ハブ |
| `internal/pkg/script/parser.go` | 台本テキストパーサー |
