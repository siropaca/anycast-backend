# Channels（チャンネル）

## チャンネル一覧取得

```
GET /channels
```

公開中のチャンネルのみ取得可能。自分のチャンネル（非公開含む）は `GET /me/channels` を使用。

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| categoryId | uuid | - | カテゴリ ID でフィルタ |
| limit | int | 20 | 取得件数（最大 100） |
| offset | int | 0 | オフセット |

**レスポンス:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "チャンネル名",
      "description": "説明",
      "category": { "id": "uuid", "slug": "technology", "name": "テクノロジー" },
      "artwork": { "id": "uuid", "url": "..." },
      "publishedAt": "2025-01-01T00:00:00Z",
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    }
  ]
}
```

---

## チャンネル取得

```
GET /channels/:channelId
```

公開中のチャンネル、または自分のチャンネルのみ取得可能。

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "name": "チャンネル名",
    "description": "説明",
    "userPrompt": "明るく楽しい雰囲気で...",
    "category": { "id": "uuid", "slug": "technology", "name": "テクノロジー" },
    "artwork": { "id": "uuid", "url": "..." },
    "characters": [
      {
        "id": "uuid",
        "name": "太郎",
        "persona": "明るい性格",
        "voice": {
          "id": "uuid",
          "name": "ja-JP-Wavenet-C",
          "provider": "google",
          "gender": "male"
        }
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

## チャンネル作成

```
POST /channels
```

**リクエスト:**
```json
{
  "name": "チャンネル名",
  "description": "説明",
  "userPrompt": "明るく楽しい雰囲気で...",
  "categoryId": "uuid",
  "artworkImageId": "uuid",
  "characters": [
    {
      "name": "太郎",
      "persona": "明るく元気な性格",
      "voiceId": "uuid"
    }
  ]
}
```

**バリデーション:**
| フィールド | ルール |
|------------|--------|
| name | 必須、255文字以内 |
| description | 必須、2000文字以内 |
| userPrompt | 必須、2000文字以内 |
| categoryId | 必須、UUID 形式 |
| characters | 必須、1〜2人 |
| characters[].name | 必須、255文字以内 |
| characters[].persona | 2000文字以内 |
| characters[].voiceId | 必須、UUID 形式 |

---

## チャンネル更新

```
PATCH /channels/:channelId
```

**リクエスト:**
```json
{
  "name": "新しいチャンネル名",
  "description": "新しい説明",
  "userPrompt": "明るく楽しい雰囲気で...",
  "categoryId": "uuid",
  "artworkImageId": "uuid"
}
```

**バリデーション:**
| フィールド | ルール |
|------------|--------|
| name | 255文字以内 |
| description | 2000文字以内 |
| userPrompt | 2000文字以内 |
| categoryId | UUID 形式 |

> **Note:** 公開状態の変更は専用エンドポイント（[チャンネル公開](#チャンネル公開) / [チャンネル非公開](#チャンネル非公開)）を使用してください。

---

## チャンネル削除

```
DELETE /channels/:channelId
```

---

## チャンネル公開

```
POST /channels/:channelId/publish
```

チャンネルを公開状態にする。`publishedAt` を省略すると現在時刻で即時公開、指定すると予約公開になる。

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
    "name": "チャンネル名",
    "description": "説明",
    "publishedAt": "2025-01-01T00:00:00Z",
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

---

## チャンネル非公開

```
POST /channels/:channelId/unpublish
```

チャンネルを非公開（下書き）状態に戻す。

**レスポンス（200 OK）:**
```json
{
  "data": {
    "id": "uuid",
    "name": "チャンネル名",
    "description": "説明",
    "publishedAt": null,
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

---

## 自分のチャンネル一覧取得

```
GET /me/channels
```

自分のチャンネル一覧を取得（非公開含む）。

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
      "name": "チャンネル名",
      "description": "説明",
      "userPrompt": "明るく楽しい雰囲気で...",
      "category": { "id": "uuid", "slug": "technology", "name": "テクノロジー" },
      "artwork": { "id": "uuid", "url": "..." },
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

## 自分のチャンネル取得

```
GET /me/channels/:channelId
```

自分のチャンネルを取得（非公開含む）。編集画面での使用を想定。

**パスパラメータ:**

| パラメータ | 型 | 説明 |
|------------|-----|------|
| channelId | uuid | チャンネル ID |

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "name": "チャンネル名",
    "description": "説明",
    "userPrompt": "明るく楽しい雰囲気で...",
    "category": { "id": "uuid", "slug": "technology", "name": "テクノロジー" },
    "artwork": { "id": "uuid", "url": "..." },
    "characters": [
      {
        "id": "uuid",
        "name": "太郎",
        "persona": "明るい性格",
        "voice": {
          "id": "uuid",
          "name": "ja-JP-Wavenet-C",
          "gender": "male"
        }
      }
    ],
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
    "message": "チャンネルが見つかりません"
  }
}
```

---

# Characters（キャラクター）

## キャラクター一覧取得

```
GET /channels/:channelId/characters
```

---

## キャラクター作成

```
POST /channels/:channelId/characters
```

**リクエスト:**
```json
{
  "name": "太郎",
  "persona": "明るく元気な性格。語尾に「だよね」をつける。",
  "voiceId": "uuid"
}
```

**バリデーション:**
- name: 必須、同一チャンネル内で一意、`__` で始まる名前は禁止
- voiceId: 必須、is_active = true のボイスのみ指定可能
- チャンネルのキャラクター数が 2 人を超える場合はエラー

---

## キャラクター更新

```
PATCH /channels/:channelId/characters/:characterId
```

**リクエスト:**
```json
{
  "name": "新しい名前",
  "persona": "新しいペルソナ",
  "voiceId": "uuid"
}
```

---

## キャラクター削除

```
DELETE /channels/:channelId/characters/:characterId
```

**バリデーション:**
- チャンネルのキャラクター数が 1 人の場合は削除不可（最低 1 人必要）
