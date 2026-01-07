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
        "avatar": {
          "id": "uuid",
          "url": "https://storage.example.com/images/xxx.png?signature=..."
        },
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
  "characters": {
    "connect": [
      { "id": "uuid" }
    ],
    "create": [
      {
        "name": "新しいキャラ",
        "persona": "明るく元気な性格",
        "avatarId": "uuid",
        "voiceId": "uuid"
      }
    ]
  }
}
```

`characters` オブジェクトの各フィールド:
- **connect**: 既存キャラクターを紐づける（自分が所有するキャラクターの ID を指定）
- **create**: 新規キャラクターを作成して紐づける

**バリデーション:**

| フィールド | ルール |
|------------|--------|
| name | 必須、255文字以内 |
| description | 必須、2000文字以内 |
| userPrompt | 必須、2000文字以内 |
| categoryId | 必須、UUID 形式 |
| characters | 必須、connect と create の合計が 1〜2 件 |
| characters.connect[].id | 必須、UUID 形式、自分が所有するキャラクターのみ |
| characters.create[].name | 必須、255文字以内、同一ユーザー内で一意、`__` 始まり禁止 |
| characters.create[].persona | 2000文字以内 |
| characters.create[].avatarId | UUID 形式 |
| characters.create[].voiceId | 必須、UUID 形式、is_active = true のボイスのみ |

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
| name | 必須、255文字以内 |
| description | 必須、2000文字以内 |
| userPrompt | 必須、2000文字以内 |
| categoryId | 必須、UUID 形式 |

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
        "avatar": {
          "id": "uuid",
          "url": "https://storage.example.com/images/xxx.png?signature=..."
        },
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

# チャンネルへのキャラクター紐づけ

チャンネルに登場させるキャラクターを管理する。キャラクター自体の CRUD は [Characters API](./characters.md) を参照。

## チャンネルのキャラクター紐づけ更新

```
PUT /channels/:channelId/characters
```

チャンネルに紐づけるキャラクターを設定する。既存の紐づけは全て置き換えられる。

**リクエスト:**
```json
{
  "characters": {
    "connect": [
      { "id": "uuid" }
    ],
    "create": [
      {
        "name": "新しいキャラ",
        "persona": "明るく元気な性格",
        "avatarId": "uuid",
        "voiceId": "uuid"
      }
    ]
  }
}
```

`characters` オブジェクトの各フィールド:
- **connect**: 既存キャラクターを紐づける（自分が所有するキャラクターの ID を指定）
- **create**: 新規キャラクターを作成して紐づける

**バリデーション:**

| フィールド | ルール |
|------------|--------|
| characters | 必須、connect と create の合計が 1〜2 件 |
| characters.connect[].id | 必須、UUID 形式、自分が所有するキャラクターのみ |
| characters.create[].name | 必須、255文字以内、同一ユーザー内で一意、`__` 始まり禁止 |
| characters.create[].persona | 2000文字以内 |
| characters.create[].avatarId | UUID 形式 |
| characters.create[].voiceId | 必須、UUID 形式、is_active = true のボイスのみ |

**レスポンス（200 OK）:**
```json
{
  "data": {
    "id": "uuid",
    "name": "チャンネル名",
    "characters": [
      {
        "id": "uuid",
        "name": "太郎",
        "persona": "明るい性格",
        "avatar": {
          "id": "uuid",
          "url": "https://storage.example.com/images/xxx.png?signature=..."
        },
        "voice": {
          "id": "uuid",
          "name": "ja-JP-Wavenet-C",
          "provider": "google",
          "gender": "male"
        }
      }
    ]
  }
}
```

**エラー（400 Bad Request）:**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "キャラクターは1〜2人必要です"
  }
}
```

**エラー（404 Not Found）:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "指定されたキャラクターが見つかりません"
  }
}
```
