# 音声生成 API (非同期)

このドキュメントでは、エピソードの台本から音声ファイルを非同期で生成する API の仕様を記載する。

## 概要

音声生成は長時間（数十秒〜数分）かかる処理のため、非同期ジョブとして実行する。  
クライアントはジョブを作成後、ポーリングまたは WebSocket でジョブの進捗・完了を監視できる。

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
      │             │  PostgreSQL │     │  Google Cloud    │
      │             │             │     │  TTS / Storage   │
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

エピソードの音声生成ジョブを作成する。

```
POST /channels/{channelId}/episodes/{episodeId}/audio/generate-async
```

**認証**: 必須

**リクエストボディ**:

| フィールド | 型 | 必須 | 説明 |
|-----------|------|------|------|
| voiceStyle | string | - | 音声のスタイル指示（最大 500 文字） |
| bgmVolumeDb | number | - | BGM 音量（-60 〜 0 dB、デフォルト: -15） |
| fadeOutMs | number | - | フェードアウト時間（0 〜 30000 ms、デフォルト: 3000） |
| paddingStartMs | number | - | 音声開始前の余白（0 〜 10000 ms、デフォルト: 500） |
| paddingEndMs | number | - | 音声終了後の余白（0 〜 10000 ms、デフォルト: 1000） |

**レスポンス**: `202 Accepted`

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "episodeId": "660e8400-e29b-41d4-a716-446655440001",
  "status": "pending",
  "progress": 0,
  "voiceStyle": "",
  "bgmVolumeDb": -15,
  "fadeOutMs": 3000,
  "paddingStartMs": 500,
  "paddingEndMs": 1000,
  "episode": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "title": "エピソードタイトル"
  },
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

**エラー**:

| コード | 説明 |
|-------|------|
| 400 | バリデーションエラー（台本なし、既に処理中のジョブあり等） |
| 403 | チャンネルへのアクセス権限なし |
| 404 | エピソードが存在しない |

### ジョブ詳細取得

```
GET /audio-jobs/{jobId}
```

**認証**: 必須

**レスポンス**: `200 OK`

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "episodeId": "660e8400-e29b-41d4-a716-446655440001",
  "status": "completed",
  "progress": 100,
  "voiceStyle": "",
  "bgmVolumeDb": -15,
  "fadeOutMs": 3000,
  "paddingStartMs": 500,
  "paddingEndMs": 1000,
  "episode": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "title": "エピソードタイトル"
  },
  "resultAudio": {
    "id": "770e8400-e29b-41d4-a716-446655440002",
    "url": "https://storage.googleapis.com/...",
    "durationMs": 120000
  },
  "startedAt": "2024-01-01T00:00:01Z",
  "completedAt": "2024-01-01T00:00:30Z",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:30Z"
}
```

### ユーザーのジョブ一覧取得

```
GET /me/audio-jobs
```

**認証**: 必須

**クエリパラメータ**:

| パラメータ | 型 | 説明 |
|-----------|------|------|
| status | string | フィルタ: pending, processing, completed, failed |

### 内部ワーカーエンドポイント

Cloud Tasks から呼び出される。

```
POST /internal/worker/audio
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
台本生成ジョブと共通のエンドポイントを使用する。

```
GET /ws/jobs?token={jwt}
```

### クライアント → サーバー

```json
// ジョブの購読開始
{"type": "subscribe", "jobId": "..."}

// ジョブの購読解除
{"type": "unsubscribe", "jobId": "..."}

// ヘルスチェック
{"type": "ping"}
```

### サーバー → クライアント

```json
// 進捗更新
{
  "type": "audio_progress",
  "payload": {
    "jobId": "...",
    "progress": 50,
    "message": "BGM をミキシング中..."
  }
}

// 完了通知
{
  "type": "audio_completed",
  "payload": {
    "jobId": "...",
    "audio": {
      "id": "...",
      "durationMs": 120000
    }
  }
}

// 失敗通知
{
  "type": "audio_failed",
  "payload": {
    "jobId": "...",
    "errorCode": "GENERATION_FAILED",
    "errorMessage": "音声の生成に失敗しました"
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
| processing | 音声生成処理中 |
| completed | 処理完了 |
| failed | 処理失敗 |

## 処理フロー

### 進捗の目安

| 進捗 | 処理内容 |
|------|---------|
| 0% | ジョブ作成 |
| 10% | エピソード・台本の読み込み |
| 20% | TTS 音声生成開始 |
| 25-45% | 複数チャンクの TTS 処理（チャンク数に応じて） |
| 45% | 音声チャンクの結合 |
| 50% | TTS 処理完了 |
| 55-70% | BGM ミキシング |
| 85% | 音声ファイルのアップロード |
| 95% | エピソード情報の更新 |
| 100% | 完了 |

### 処理詳細

1. **台本読み込み**: エピソードに紐づく台本を取得
2. **話者マッピング**: キャラクターを TTS の話者 ID にマッピング
3. **テキスト分割**: TTS の入力制限（3500 バイト）に収まるようチャンク分割
4. **TTS 合成**: Google Cloud TTS (Gemini TTS) で音声合成
5. **チャンク結合**: 複数チャンクの場合、FFmpeg で結合
6. **BGM ミキシング**: FFmpeg で BGM と音声をミックス
7. **アップロード**: 生成した音声ファイルを GCS にアップロード
8. **エピソード更新**: `fullAudioId` と `audioOutdated` を更新

## 外部サービス

### Google Cloud Text-to-Speech

| 設定 | 値 | 説明 |
|------|------|------|
| モデル | gemini-2.5-pro-tts | Gemini TTS モデル |
| 言語 | ja-JP | 日本語 |
| 出力形式 | MP3 | libmp3lame エンコーダー |
| テキスト上限 | 3500 バイト | Google TTS の 4000 バイト制限に対する安全マージン |
| リトライ回数 | 3 回 | エラー時の最大リトライ |
| リトライ間隔 | 1秒, 2秒, 3秒 | 指数バックオフ |

- 設定箇所: `internal/infrastructure/tts/google_tts_client.go`

### Google Cloud Tasks

| 設定 | 値 | 説明 |
|------|------|------|
| ロケーション | asia-northeast1 | デフォルト |
| キュー名 | async-jobs | 非同期ジョブ用キュー（台本・音声生成共通） |
| 認証 | OIDC | Service Account による認証 |
| ワーカー URL | {baseURL}/audio | ベース URL + `/audio` |

- 設定箇所: `internal/infrastructure/cloudtasks/client.go`
- ベース URL は環境変数 `GOOGLE_CLOUD_TASKS_WORKER_URL` で設定
- Cloud Tasks が未設定の場合（ローカル開発）は goroutine で直接実行

### Google Cloud Storage

| 設定 | 値 | 説明 |
|------|------|------|
| パス | audios/{audioId}.mp3 | 音声ファイルの保存パス |
| 署名付き URL 有効期限 | 1 時間 | V4 スキーム使用 |

- 設定箇所: `internal/infrastructure/storage/gcs_client.go`

## エラーコード

| コード | HTTP | 説明 |
|-------|------|------|
| VALIDATION_ERROR | 400 | 台本なし、重複ジョブ等 |
| UNAUTHORIZED | 401 | 認証エラー |
| FORBIDDEN | 403 | アクセス権限なし |
| NOT_FOUND | 404 | リソースが存在しない |
| GENERATION_FAILED | 500 | TTS 生成失敗 |
| MEDIA_UPLOAD_FAILED | 500 | ファイルアップロード失敗 |
| INTERNAL_ERROR | 500 | その他の内部エラー |

## データベーススキーマ

### audio_jobs テーブル

```sql
CREATE TABLE audio_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    episode_id UUID NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status audio_job_status NOT NULL DEFAULT 'pending',
    progress INTEGER NOT NULL DEFAULT 0,
    voice_style TEXT NOT NULL DEFAULT '',

    -- BGM ミキシング設定
    bgm_volume_db DECIMAL(5, 2) NOT NULL DEFAULT -15.0,
    fade_out_ms INTEGER NOT NULL DEFAULT 3000,
    padding_start_ms INTEGER NOT NULL DEFAULT 500,
    padding_end_ms INTEGER NOT NULL DEFAULT 1000,

    -- 結果
    result_audio_id UUID REFERENCES audios (id) ON DELETE SET NULL,
    error_message TEXT,
    error_code VARCHAR(50),

    -- タイムスタンプ
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TYPE audio_job_status AS ENUM ('pending', 'processing', 'completed', 'failed');
```

### audios テーブル

```sql
CREATE TABLE audios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mime_type VARCHAR(100) NOT NULL,
    path VARCHAR(1024) NOT NULL,
    filename VARCHAR(255) NOT NULL,
    file_size INTEGER NOT NULL,
    duration_ms INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## 関連ファイル

| ファイル | 説明 |
|---------|------|
| `internal/handler/audio_job.go` | REST API ハンドラー |
| `internal/handler/worker.go` | ワーカーエンドポイント |
| `internal/handler/websocket.go` | WebSocket ハンドラー |
| `internal/service/audio_job.go` | ビジネスロジック |
| `internal/service/ffmpeg.go` | FFmpeg 処理 |
| `internal/repository/audio_job.go` | データベースアクセス |
| `internal/model/audio_job.go` | データモデル |
| `internal/infrastructure/tts/google_tts_client.go` | TTS クライアント |
| `internal/infrastructure/cloudtasks/client.go` | Cloud Tasks クライアント |
| `internal/infrastructure/storage/gcs_client.go` | GCS クライアント |
| `internal/infrastructure/websocket/hub.go` | WebSocket ハブ |
