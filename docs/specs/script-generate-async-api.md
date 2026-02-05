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
| status | string | フィルタ: pending, processing, canceling, completed, failed, canceled |

### ジョブキャンセル

```
POST /script-jobs/{jobId}/cancel
```

**認証**: 必須

**説明**: 台本生成ジョブをキャンセルする。

- `pending` 状態のジョブは即座に `canceled` に遷移
- `processing` 状態のジョブは `canceling` に遷移し、次のチェックポイントで中断

**レスポンス**: `200 OK`

```json
{
  "success": true
}
```

**エラー**:

| コード | 説明 |
|-------|------|
| 400 | キャンセル不可（既にキャンセル中/済み、完了済み、失敗済み） |
| 403 | ジョブへのアクセス権限なし |
| 404 | ジョブが存在しない |

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

// キャンセル中通知
{
  "type": "script_canceling",
  "payload": {
    "jobId": "..."
  }
}

// キャンセル完了通知
{
  "type": "script_canceled",
  "payload": {
    "jobId": "..."
  }
}

// ヘルスチェック応答
{"type": "pong"}
```

## ジョブステータス

```
pending ────▶ processing ───▶ completed
    │              │
    │              └──────────▶ failed
    │              │
    ▼              ▼
canceled      canceling ───▶ canceled
                       ───▶ failed
```

| ステータス | 説明 |
|-----------|------|
| pending | ジョブ作成済み、処理待ち |
| processing | 台本生成処理中 |
| canceling | キャンセル要求を受け付け、処理中断中 |
| completed | 処理完了 |
| failed | 処理失敗 |
| canceled | キャンセル完了 |

## 処理フロー

### 進捗の目安

多段階ワークフローで処理する。詳細は [台本生成プロンプトワークフロー仕様](./script-prompt-workflow.md) を参照。

| 進捗 | Phase | 処理内容 |
|------|-------|---------|
| 0% | - | ジョブ作成 |
| 5% | - | チャンネル・エピソード・ユーザー情報の読み込み |
| 10% | Phase 1 | ブリーフ正規化 |
| 15% | Phase 2 | 素材+アウトライン生成（LLM 1回目）開始 |
| 35% | Phase 2 | 素材+アウトライン生成 完了 |
| 40% | Phase 3 | 台本ドラフト生成（LLM 2回目）開始 |
| 75% | Phase 3 | 台本ドラフト生成 完了 |
| 80% | Phase 4 | QA 定量チェック |
| 85% | Phase 4 | パッチ修正（LLM 3回目、条件付き） |
| 90% | - | 台本パース・DB 保存 |
| 95% | - | 完了処理 |
| 100% | - | 完了 |

### 処理詳細

1. **データ読み込み**: チャンネル、エピソード、ユーザー情報を取得
2. **ブリーフ正規化** (Phase 1): User/Channel/Episode/Character 情報を構造化スロットに正規化
3. **素材+アウトライン生成** (Phase 2): LLM で具体例・落とし穴・疑問を生成し、3ブロック構成のアウトラインを設計
4. **台本ドラフト生成** (Phase 3): アウトラインと素材を元に台本を生成
5. **QA 検証+パッチ修正** (Phase 4): コードで定量チェック → 不合格時のみ LLM で局所修正
6. **データ保存**: 既存の台本行を削除し、新しい台本行をバッチ作成

### システムプロンプト

`talk_mode`（dialogue/monologue）と感情タグの有無で Phase 別のプロンプトを使い分ける。
詳細は [台本生成プロンプトワークフロー仕様](./script-prompt-workflow.md) を参照。

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
- 1 つのセリフは 20〜80 文字程度（許容範囲: 10〜120 文字）
- セリフの末尾に句点（。）は付けない
- セリフ中に句点を含めない（1行に1文）
- 自然なフィラー（「えーと」「まあ」等）を適度に含める
- 設定箇所: `internal/service/script_prompts.go`

## 外部サービス

### LLM API

マルチプロバイダ対応（OpenAI / Claude / Gemini）。`llm.Registry` で複数プロバイダのクライアントを管理し、Phase ごとに使用するプロバイダを切り替え可能。

- API キーが設定されたプロバイダが起動時に自動登録される
- Phase ごとの設定（プロバイダ・Temperature）は `internal/service/script_prompts.go` の `PhaseConfig` で定義
- 起動時に Phase 設定で必要なプロバイダが未登録の場合はエラーで起動失敗する

Phase 別設定:
| Phase | Provider | Temperature | 理由 |
|-------|----------|-------------|------|
| Phase 2 | OpenAI | 0.9 | 創造的な素材生成 |
| Phase 3 | OpenAI | 0.7 | 内容の忠実性と自然さのバランス |
| Phase 4 | OpenAI | 0.5 | 局所修正のため低め |

- プロバイダ設定箇所: `internal/service/script_prompts.go`（`PhaseConfig`）
- クライアント実装: `internal/infrastructure/llm/`

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
CREATE TYPE script_job_status AS ENUM ('pending', 'processing', 'canceling', 'completed', 'failed', 'canceled');

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
| `internal/service/script_job.go` | 多段階ワークフロー実行ロジック |
| `internal/service/script_prompts.go` | Phase 2/3/4 のシステムプロンプト定義 |
| `internal/repository/script_job.go` | データベースアクセス |
| `internal/model/script_job.go` | データモデル |
| `internal/infrastructure/llm/client.go` | LLM クライアントインターフェース |
| `internal/infrastructure/llm/registry.go` | LLM プロバイダ Registry |
| `internal/infrastructure/cloudtasks/client.go` | Cloud Tasks クライアント |
| `internal/infrastructure/websocket/hub.go` | WebSocket ハブ |
| `internal/pkg/script/parser.go` | 台本テキストパーサー |
| `internal/pkg/script/brief.go` | ブリーフ正規化（Phase 1） |
| `internal/pkg/script/grounding.go` | Phase 2 出力構造体とパーサー |
| `internal/pkg/script/validator.go` | QA 定量チェック（Phase 4） |
| `internal/pkg/script/json_extractor.go` | LLM 出力からの JSON 抽出 |
