# Script（台本）

## 台本を AI で生成（非同期）

台本生成は LLM（OpenAI GPT）を使用した処理のため、非同期ジョブとして実行されます。
詳細は [台本生成 API（非同期）仕様書](../specs/script-generate-async-api.md) を参照してください。

```
POST /channels/:channelId/episodes/:episodeId/script/generate-async
```

**リクエスト:**
```json
{
  "prompt": "今日の天気について楽しく話す",
  "durationMinutes": 10,
  "withEmotion": true
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| prompt | string | ◯ | テーマやシナリオ（2000文字以内）。URL が含まれていれば RAG で内容を取得して台本生成に利用 |
| durationMinutes | int | | エピソードの長さ（分）。3〜30の範囲で指定。デフォルト: 10 |
| withEmotion | bool | | 感情を付与するかどうか。デフォルト: false |

**レスポンス（202 Accepted）:**
```json
{
  "data": {
    "id": "uuid",
    "episodeId": "uuid",
    "status": "pending",
    "progress": 0,
    "prompt": "今日の天気について楽しく話す",
    "durationMinutes": 10,
    "withEmotion": true,
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

---

## 最新完了済み台本生成ジョブ取得

```
GET /channels/:channelId/episodes/:episodeId/script-jobs/latest
```

エピソードの最新の完了済み台本生成ジョブを取得します。
完了済みジョブが存在しない場合は `data: null` を返します（404 ではありません）。

**レスポンス（200 OK）:**

完了済みジョブがある場合:
```json
{
  "data": {
    "id": "uuid",
    "episodeId": "uuid",
    "status": "completed",
    "progress": 100,
    "prompt": "今日の天気について楽しく話す",
    "durationMinutes": 10,
    "withEmotion": true,
    "episode": {
      "id": "uuid",
      "title": "エピソードタイトル",
      "channel": {
        "id": "uuid",
        "name": "チャンネル名"
      }
    },
    "scriptLinesCount": 42,
    "startedAt": "2025-01-01T00:00:00Z",
    "completedAt": "2025-01-01T00:00:15Z",
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:15Z"
  }
}
```

完了済みジョブがない場合:
```json
{
  "data": null
}
```

**エラー:**

| コード | 説明 |
|--------|------|
| FORBIDDEN | チャンネルへのアクセス権限なし |
| NOT_FOUND | チャンネル・エピソードが存在しない |

---

## 台本生成ジョブ取得

```
GET /script-jobs/:jobId
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
    "prompt": "今日の天気について楽しく話す",
    "durationMinutes": 10,
    "withEmotion": true,
    "episode": {
      "id": "uuid",
      "title": "エピソードタイトル",
      "channel": {
        "id": "uuid",
        "name": "チャンネル名"
      }
    },
    "scriptLinesCount": 42,
    "startedAt": "2025-01-01T00:00:00Z",
    "completedAt": "2025-01-01T00:00:15Z",
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:15Z"
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

## 台本生成ジョブキャンセル

```
POST /script-jobs/:jobId/cancel
```

台本生成ジョブをキャンセルします。

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

## 自分の台本生成ジョブ一覧

```
GET /me/script-jobs
```

自分が作成した台本生成ジョブの一覧を取得します。

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
      "prompt": "今日の天気について楽しく話す",
      "durationMinutes": 10,
      "withEmotion": true,
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:05Z"
    }
  ]
}
```

---

## 台本テキスト取り込み

```
POST /channels/:channelId/episodes/:episodeId/script/import
```

テキスト形式の台本をインポートする。既存の台本がある場合は全て削除される。

**リクエスト:**
```json
{
  "text": "太郎: こんにちは\n花子: [excited] やあ"
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| text | string | ◯ | 台本テキスト |

**テキストフォーマット:**

```
話者名: [感情] セリフ
```

- `[感情]` は省略可能
- 話者名はチャンネルに登録されているキャラクター名のみ使用可能

**例:**
```
太郎: こんにちは
花子: [excited] やあ
太郎: 今日はいい天気だね
```

**レスポンス（成功時）:**
```json
{
  "data": [
    { "id": "uuid", "lineOrder": 0, "speaker": { ... }, "text": "...", ... }
  ]
}
```

**レスポンス（エラー時）:**
```json
{
  "error": {
    "code": "SCRIPT_PARSE_ERROR",
    "message": "台本のパースに失敗しました",
    "details": [
      { "line": 3, "reason": "不明な話者: 三郎" }
    ]
  }
}
```

---

## 台本テキスト出力

```
GET /channels/:channelId/episodes/:episodeId/script/export
```

台本をテキストファイルとしてダウンロードする。出力されたテキストはそのままインポート可能。

**レスポンス:**
- Content-Type: `text/plain; charset=utf-8`
- Content-Disposition: `attachment; filename="エピソード名.txt"; filename*=UTF-8''...`

```
太郎: こんにちは
花子: [excited] やあ
```

---

# ScriptLines（台本行）

## 台本行一覧取得

```
GET /channels/:channelId/episodes/:episodeId/script/lines
```

指定したエピソードの台本行一覧を `lineOrder` 順で取得する。

**レスポンス:**
```json
{
  "data": [
    {
      "id": "uuid",
      "lineOrder": 0,
      "speaker": {
        "id": "uuid",
        "name": "太郎",
        "persona": "明るく元気な性格",
        "voice": { "id": "uuid", "name": "Voice1", "provider": "google", "gender": "male" }
      },
      "text": "こんにちは",
      "emotion": null,
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    },
    {
      "id": "uuid",
      "lineOrder": 1,
      "speaker": {
        "id": "uuid",
        "name": "花子",
        "persona": "落ち着いた知的な性格",
        "voice": { "id": "uuid", "name": "Voice2", "provider": "google", "gender": "female" }
      },
      "text": "やあ",
      "emotion": "excited",
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    }
  ]
}
```

---

## 行追加

```
POST /channels/:channelId/episodes/:episodeId/script/lines
```

指定した位置に新しい台本行を追加する。

**リクエスト:**
```json
{
  "speakerId": "uuid",
  "text": "セリフのテキスト",
  "emotion": "excited",
  "afterLineId": "uuid"
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| speakerId | string | ◯ | 話者（キャラクター）の ID |
| text | string | ◯ | セリフのテキスト |
| emotion | string | | 感情表現（省略可） |
| afterLineId | string | | この行の後に挿入する。`null` または省略時は先頭に挿入 |

**レスポンス:**
```json
{
  "id": "uuid",
  "lineOrder": 1,
  "speaker": {
    "id": "uuid",
    "name": "太郎",
    "persona": "明るく元気な性格",
    "voice": { "id": "uuid", "name": "Voice1", "provider": "google", "gender": "male" }
  },
  "text": "セリフのテキスト",
  "emotion": "excited",
  "createdAt": "2025-01-01T00:00:00Z",
  "updatedAt": "2025-01-01T00:00:00Z"
}
```

**エラー:**
- `400 Bad Request`: バリデーションエラー
- `403 Forbidden`: チャンネルのオーナーでない場合
- `404 Not Found`: `afterLineId` で指定した行が存在しない場合

---

## 行更新

```
PATCH /channels/:channelId/episodes/:episodeId/script/lines/:lineId
```

**リクエスト:**
```json
{
  "speakerId": "uuid",
  "text": "新しいセリフ",
  "emotion": "laughing"
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| speakerId | string | | 話者（キャラクター）の ID。チャンネルに紐づいたキャラクターのみ指定可能 |
| text | string | | セリフのテキスト |
| emotion | string | | 感情表現。空文字を指定すると削除 |

**レスポンス:**
```json
{
  "id": "uuid",
  "lineOrder": 1,
  "speaker": {
    "id": "uuid",
    "name": "太郎",
    "persona": "明るく元気な性格",
    "voice": { "id": "uuid", "name": "Voice1", "provider": "google", "gender": "male" }
  },
  "text": "新しいセリフ",
  "emotion": "laughing",
  "createdAt": "2025-01-01T00:00:00Z",
  "updatedAt": "2025-01-01T00:00:00Z"
}
```

**エラー:**
- `400 Bad Request`: バリデーションエラー（指定されたspeakerIdがチャンネルに紐づいていない等）
- `403 Forbidden`: チャンネルのオーナーでない場合
- `404 Not Found`: 台本行が存在しない場合

---

## 行削除

```
DELETE /channels/:channelId/episodes/:episodeId/script/lines/:lineId
```

指定した台本行を削除する。

**レスポンス:**
- `204 No Content`: 削除成功
- `403 Forbidden`: チャンネルのオーナーでない場合
- `404 Not Found`: 台本行が存在しない場合

---

## 全行削除

```
DELETE /channels/:channelId/episodes/:episodeId/script/lines
```

指定したエピソードの台本行をすべて削除する。

**レスポンス:**
- `204 No Content`: 削除成功（台本行が 0 件の場合も 204）
- `403 Forbidden`: チャンネルのオーナーでない場合
- `404 Not Found`: チャンネルまたはエピソードが存在しない場合

---

## 行並び替え

```
POST /channels/:channelId/episodes/:episodeId/script/reorder
```

台本行の順序を変更する。指定された順序で `lineOrder` を再割り当てする。

**リクエスト:**
```json
{
  "lineIds": ["uuid-1", "uuid-2", "uuid-3"]
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| lineIds | string[] | ◯ | 並び替え後の順序で台本行 ID を指定 |

**処理内容:**
1. `lineIds` の配列順に `lineOrder` を 0, 1, 2, ... と再割り当て
2. 指定された全ての行が対象エピソードに属していることを検証

**レスポンス:**
```json
{
  "data": [
    {
      "id": "uuid-1",
      "lineOrder": 0,
      "speaker": { ... },
      "text": "...",
      "emotion": null,
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    },
    {
      "id": "uuid-2",
      "lineOrder": 1,
      "speaker": { ... },
      "text": "...",
      "emotion": "excited",
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    }
  ]
}
```

**エラー:**
- `400 Bad Request`: バリデーションエラー（空配列、重複 ID など）
- `403 Forbidden`: チャンネルのオーナーでない場合
- `404 Not Found`: 指定した行が存在しない、または対象エピソードに属していない場合
