# Episodes（エピソード）

## エピソード一覧取得（公開用）

```
GET /channels/:channelId/episodes
```

公開中のエピソードのみ取得可能。自分のチャンネルのエピソード（非公開含む）は `GET /me/channels/:channelId/episodes` を使用。

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| limit | int | 20 | 取得件数（最大 100） |
| offset | int | 0 | オフセット |

**レスポンス:**
```json
{
  "data": [
    {
      "id": "uuid",
      "title": "エピソードタイトル",
      "description": "エピソードの説明",
      "fullAudio": { "id": "uuid", "url": "...", "durationMs": 180000 },
      "publishedAt": "2025-01-01T00:00:00Z",
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 100,
    "limit": 20,
    "offset": 0
  }
}
```

---

## エピソード取得

```
GET /channels/:channelId/episodes/:episodeId
```

公開中のエピソード、または自分のエピソードのみ取得可能。

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "title": "エピソードタイトル",
    "description": "エピソードの説明",
    "userPrompt": "今回のテーマについて詳しく解説する",
    "bgm": { "id": "uuid", "url": "..." },
    "fullAudio": { "id": "uuid", "url": "..." },
    "script": [
      {
        "id": "uuid",
        "lineOrder": 0,
        "lineType": "speech",
        "speaker": { "id": "uuid", "name": "太郎" },
        "text": "こんにちは",
        "emotion": null,
        "audio": { "id": "uuid", "url": "..." }
      },
      {
        "id": "uuid",
        "lineOrder": 1,
        "lineType": "silence",
        "durationMs": 800
      },
      {
        "id": "uuid",
        "lineOrder": 2,
        "lineType": "sfx",
        "sfx": { "id": "uuid", "name": "chime" },
        "volume": 0.8
      }
    ],
    "publishedAt": "2025-01-01T00:00:00Z",
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

> **Note:** `userPrompt` はオーナーのみに表示されます。他ユーザーがアクセスした場合は含まれません。

---

## エピソード作成

```
POST /channels/:channelId/episodes
```

**リクエスト:**
```json
{
  "title": "エピソードタイトル",
  "description": "エピソードの説明",
  "artworkImageId": "uuid",
  "bgmAudioId": "uuid"
}
```

**バリデーション:**
| フィールド | ルール |
|------------|--------|
| title | 必須、255文字以内 |
| description | 2000文字以内 |

---

## エピソード更新

```
PATCH /channels/:channelId/episodes/:episodeId
```

**リクエスト:**
```json
{
  "title": "新しいタイトル",
  "description": "新しい説明",
  "artworkImageId": "uuid",
  "bgmAudioId": "uuid"
}
```

**バリデーション:**
| フィールド | ルール |
|------------|--------|
| title | 255文字以内 |
| description | 2000文字以内 |
| userPrompt | 2000文字以内 |

> **Note:** `userPrompt` は台本生成時に自動で保存されます。直接編集する場合は API から設定可能ですが、通常は台本生成 API 経由で更新されます。
>
> **Note:** 公開状態の変更は専用エンドポイント（[エピソード公開](#エピソード公開) / [エピソード非公開](#エピソード非公開)）を使用してください。

---

## エピソード削除

```
DELETE /channels/:channelId/episodes/:episodeId
```

---

## エピソード公開

```
POST /channels/:channelId/episodes/:episodeId/publish
```

エピソードを公開状態にする。`publishedAt` を省略すると現在時刻で即時公開、指定すると予約公開になる。

**リクエスト:**
```json
{
  "publishedAt": "2025-01-01T00:00:00Z"
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| publishedAt | string | | 公開日時（RFC3339 形式）。省略時は現在時刻で即時公開 |

**レスポンス（200 OK）:**
```json
{
  "data": {
    "id": "uuid",
    "title": "エピソードタイトル",
    "description": "エピソードの説明",
    "publishedAt": "2025-01-01T00:00:00Z",
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

---

## エピソード非公開

```
POST /channels/:channelId/episodes/:episodeId/unpublish
```

エピソードを非公開（下書き）状態に戻す。

**レスポンス（200 OK）:**
```json
{
  "data": {
    "id": "uuid",
    "title": "エピソードタイトル",
    "description": "エピソードの説明",
    "publishedAt": null,
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

---

## 自分のチャンネルのエピソード一覧取得

```
GET /me/channels/:channelId/episodes
```

自分のチャンネルに紐付くエピソード一覧を取得（非公開含む）。編集画面での使用を想定。

**パスパラメータ:**

| パラメータ | 型 | 説明 |
|------------|-----|------|
| channelId | uuid | チャンネル ID |

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| status | string | - | 公開状態でフィルタ: `published` / `draft` |
| limit | int | 20 | 取得件数（最大 100） |
| offset | int | 0 | オフセット |

**レスポンス:**
```json
{
  "data": [
    {
      "id": "uuid",
      "title": "エピソードタイトル",
      "description": "エピソードの説明",
      "userPrompt": "今回のテーマについて詳しく解説する",
      "fullAudio": { "id": "uuid", "url": "...", "durationMs": 180000 },
      "publishedAt": "2025-01-01T00:00:00Z",
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 100,
    "limit": 20,
    "offset": 0
  }
}
```

**エラー（403 Forbidden）:**
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "このチャンネルへのアクセス権限がありません"
  }
}
```

**エラー（404 Not Found）:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "チャンネルが見つかりません"
  }
}
```

---

## 自分のチャンネルのエピソード取得

```
GET /me/channels/:channelId/episodes/:episodeId
```

自分のチャンネルに紐付くエピソードを取得（非公開含む）。編集画面での使用を想定。

**パスパラメータ:**

| パラメータ | 型 | 説明 |
|------------|-----|------|
| channelId | uuid | チャンネル ID |
| episodeId | uuid | エピソード ID |

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "title": "エピソードタイトル",
    "description": "エピソードの説明",
    "userPrompt": "今回のテーマについて詳しく解説する",
    "artwork": { "id": "uuid", "url": "..." },
    "fullAudio": { "id": "uuid", "url": "...", "durationMs": 180000 },
    "publishedAt": "2025-01-01T00:00:00Z",
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

**エラー（403 Forbidden）:**
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "このチャンネルへのアクセス権限がありません"
  }
}
```

**エラー（404 Not Found）:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "エピソードが見つかりません"
  }
}
```
