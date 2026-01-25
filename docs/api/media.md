# Audio（音声生成）

## 非同期音声生成

エピソードの全台本行から音声を非同期で生成します。  
Gemini TTS の multi-speaker 機能を使用し、各キャラクターの声で 1 つの音声ファイルを生成します。  
BGM が設定されている場合は FFmpeg でミキシングを行います。

```
POST /channels/:channelId/episodes/:episodeId/audio/generate-async
```

**リクエスト:**
```json
{
  "voiceStyle": "Read aloud in a warm, welcoming tone",
  "bgmVolumeDb": -15,
  "fadeOutMs": 3000,
  "paddingStartMs": 500,
  "paddingEndMs": 1000
}
```

| フィールド | 型 | 必須 | デフォルト | 説明 |
|------------|-----|:----:|------------|------|
| voiceStyle | string | | - | 音声生成のスタイル指示（500文字以内） |
| bgmVolumeDb | number | | -15 | BGM 音量（dB）。0 で原音量、負の値で音量を下げる |
| fadeOutMs | int | | 3000 | BGM のフェードアウト時間（ms） |
| paddingStartMs | int | | 500 | 音声開始前の BGM のみの余白時間（ms） |
| paddingEndMs | int | | 1000 | 音声終了後の BGM のみの余白時間（ms） |

**処理フロー:**

1. ジョブレコードを作成（status: `pending`）
2. Cloud Tasks にジョブをキューイング（ローカル開発時は goroutine で直接実行）
3. クライアントに即座にジョブ情報を返却（202 Accepted）
4. ワーカーが非同期で以下を実行:
   - エピソードの全台本行を取得
   - 各キャラクターの Voice 設定を収集
   - Gemini TTS multi-speaker API で音声を生成
   - BGM が設定されている場合、FFmpeg でミキシング
   - エピソードの `fullAudio` として保存
   - WebSocket で完了通知

**レスポンス（202 Accepted）:**
```json
{
  "data": {
    "jobId": "uuid",
    "status": "pending",
    "progress": 0
  }
}
```

**エラー:**

| コード | 説明 |
|--------|------|
| VALIDATION_ERROR | 台本に speech 行が存在しない |
| JOB_ENQUEUE_FAILED | ジョブのキューイングに失敗 |

---

## 音声生成ジョブ取得

```
GET /audio-jobs/:jobId
```

指定したジョブの状態を取得します。

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "episodeId": "uuid",
    "status": "completed",
    "progress": 100,
    "voiceStyle": "Read aloud in a warm, welcoming tone",
    "audio": {
      "id": "uuid",
      "url": "https://storage.example.com/full-episode.mp3",
      "mimeType": "audio/mpeg",
      "fileSize": 1024000,
      "durationMs": 180000
    },
    "errorCode": null,
    "errorMessage": null,
    "startedAt": "2025-01-01T00:00:00Z",
    "completedAt": "2025-01-01T00:00:10Z",
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:10Z"
  }
}
```

**ステータス:**

| ステータス | 説明 |
|------------|------|
| pending | キュー待ち |
| processing | 処理中 |
| canceling | キャンセル中 |
| completed | 完了 |
| failed | 失敗 |
| canceled | キャンセル完了 |

---

## 音声生成ジョブキャンセル

```
POST /audio-jobs/:jobId/cancel
```

音声生成ジョブをキャンセルします。

- `pending` 状態のジョブは即座に `canceled` に遷移
- `processing` 状態のジョブは `canceling` に遷移し、次のチェックポイントで中断

**レスポンス（200 OK）:**
```json
{
  "success": true
}
```

**エラー:**

| コード | 説明 |
|--------|------|
| VALIDATION_ERROR | キャンセル不可（既にキャンセル中/済み、完了済み、失敗済み） |
| FORBIDDEN | ジョブへのアクセス権限なし |
| NOT_FOUND | ジョブが存在しない |

---

## 自分の音声生成ジョブ一覧

```
GET /me/audio-jobs
```

自分が作成した音声生成ジョブの一覧を取得します。

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| status | string | - | ステータスでフィルタ: `pending` / `processing` / `canceling` / `completed` / `failed` / `canceled` |

**レスポンス:**
```json
{
  "data": [
    {
      "id": "uuid",
      "episodeId": "uuid",
      "status": "processing",
      "progress": 45,
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:05Z"
    }
  ]
}
```

---

## 音声アップロード

```
POST /audios
```

**リクエスト:** `multipart/form-data`

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| file | File | ◯ | アップロードする音声ファイル（mp3, wav, ogg, aac, m4a） |

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "mimeType": "audio/mpeg",
    "url": "https://storage.example.com/audio.mp3",
    "filename": "bgm.mp3",
    "fileSize": 1024000,
    "durationMs": 180000
  }
}
```

> **Note:** `durationMs` は MP3 形式の場合のみビットレートベースで推定されます。その他の形式では 0 が返されます。

---

# WebSocket

## WebSocket 接続

音声生成・台本生成ジョブの進捗・完了をリアルタイムで受信できます。

```
WS /ws/jobs?token=<JWT>
```

**クエリパラメータ:**

| パラメータ | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| token | string | ◯ | JWT トークン |

### クライアント → サーバー

**ジョブの購読開始:**
```json
{
  "type": "subscribe",
  "payload": {
    "jobId": "uuid"
  }
}
```

**ジョブの購読解除:**
```json
{
  "type": "unsubscribe",
  "payload": {
    "jobId": "uuid"
  }
}
```

**Ping（接続確認）:**
```json
{
  "type": "ping"
}
```

### サーバー → クライアント

**進捗通知:**
```json
{
  "type": "progress",
  "payload": {
    "jobId": "uuid",
    "progress": 45,
    "message": "音声を生成中..."
  }
}
```

**完了通知:**
```json
{
  "type": "completed",
  "payload": {
    "jobId": "uuid",
    "audio": {
      "id": "uuid",
      "url": "https://storage.example.com/full-episode.mp3",
      "durationMs": 180000
    }
  }
}
```

**エラー通知:**
```json
{
  "type": "failed",
  "payload": {
    "jobId": "uuid",
    "errorCode": "GENERATION_FAILED",
    "errorMessage": "音声生成に失敗しました"
  }
}
```

**キャンセル中通知:**
```json
{
  "type": "audio_canceling",
  "payload": {
    "jobId": "uuid"
  }
}
```

**キャンセル完了通知:**
```json
{
  "type": "audio_canceled",
  "payload": {
    "jobId": "uuid"
  }
}
```

**Pong（Ping への応答）:**
```json
{
  "type": "pong"
}
```

---

# Images（画像ファイル）

## 画像アップロード

```
POST /images
```

**リクエスト:** `multipart/form-data`

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| file | File | ◯ | アップロードするファイル（png, jpeg など） |

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "mimeType": "image/png",
    "url": "https://storage.example.com/artwork.png",
    "filename": "artwork.png",
    "fileSize": 512000
  }
}
```
