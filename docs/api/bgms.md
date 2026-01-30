# BGMs（BGM）

ユーザーが所有する BGM。複数のエピソードで使い回すことができる。

## BGM 一覧取得

```
GET /me/bgms
```

自分の BGM 一覧を取得。

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| include_system | boolean | false | true の場合、システム BGM も含める |
| limit | int | 20 | 取得件数（最大 100） |
| offset | int | 0 | オフセット |

**レスポンス（include_system=false）:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "お気に入り曲",
      "isSystem": false,
      "audio": {
        "id": "uuid",
        "url": "https://storage.example.com/audios/xxx.mp3?signature=...",
        "durationMs": 180000
      },
      "episodes": [
        {
          "id": "uuid",
          "title": "第1回 はじめまして",
          "channel": {
            "id": "uuid",
            "name": "マイチャンネル"
          }
        }
      ],
      "channels": [
        {
          "id": "uuid",
          "name": "マイチャンネル"
        }
      ],
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 5,
    "limit": 20,
    "offset": 0
  }
}
```

**レスポンス（include_system=true）:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "お気に入り曲",
      "isSystem": false,
      "audio": {
        "id": "uuid",
        "url": "https://storage.example.com/audios/xxx.mp3?signature=...",
        "durationMs": 180000
      },
      "episodes": [
        {
          "id": "uuid",
          "title": "第1回 はじめまして",
          "channel": {
            "id": "uuid",
            "name": "マイチャンネル"
          }
        }
      ],
      "channels": [
        {
          "id": "uuid",
          "name": "マイチャンネル"
        }
      ],
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    },
    {
      "id": "uuid",
      "name": "明るいポップ",
      "isSystem": true,
      "audio": {
        "id": "uuid",
        "url": "https://storage.example.com/audios/xxx.mp3?signature=...",
        "durationMs": 180000
      },
      "episodes": [],
      "channels": [],
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 15,
    "limit": 20,
    "offset": 0
  }
}
```

> **Note:** `include_system=true` の場合、ユーザー BGM → システム BGM の順で返却。ユーザー BGM は `created_at` 降順、システム BGM は `sort_order` 順。

---

## BGM 取得

```
GET /me/bgms/:bgmId
```

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "name": "お気に入り曲",
    "isSystem": false,
    "audio": {
      "id": "uuid",
      "url": "https://storage.example.com/audios/xxx.mp3?signature=...",
      "durationMs": 180000
    },
    "episodes": [
      {
        "id": "uuid",
        "title": "第1回 はじめまして",
        "channel": {
          "id": "uuid",
          "name": "マイチャンネル"
        }
      }
    ],
    "channels": [
      {
        "id": "uuid",
        "name": "マイチャンネル"
      }
    ],
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

---

## BGM 作成

```
POST /me/bgms
```

**リクエスト:**
```json
{
  "name": "お気に入り曲",
  "audioId": "uuid"
}
```

**バリデーション:**

| フィールド | ルール |
|------------|--------|
| name | 必須、255 文字以内、同一ユーザー内で一意 |
| audioId | 必須、UUID 形式、存在する音声ファイルのみ指定可能 |

**レスポンス（201 Created）:**
```json
{
  "data": {
    "id": "uuid",
    "name": "お気に入り曲",
    "isSystem": false,
    "audio": {
      "id": "uuid",
      "url": "https://storage.example.com/audios/xxx.mp3?signature=...",
      "durationMs": 180000
    },
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

**エラー（409 Conflict）:**
```json
{
  "error": {
    "code": "DUPLICATE_NAME",
    "message": "同じ名前の BGM が既に存在します"
  }
}
```

---

## BGM 更新

```
PATCH /me/bgms/:bgmId
```

**リクエスト:**
```json
{
  "name": "新しい名前"
}
```

**バリデーション:**

| フィールド | ルール |
|------------|--------|
| name | 255 文字以内、同一ユーザー内で一意 |

> **Note:** audioId は更新不可。音声ファイルを変更したい場合は、新しい BGM を作成する。

**レスポンス（200 OK）:**
```json
{
  "data": {
    "id": "uuid",
    "name": "新しい名前",
    "isSystem": false,
    "audio": {
      "id": "uuid",
      "url": "https://storage.example.com/audios/xxx.mp3?signature=...",
      "durationMs": 180000
    },
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

---

## BGM 削除

```
DELETE /me/bgms/:bgmId
```

**制約:**
- いずれかのエピソードで使用中の BGM は削除不可

**レスポンス（204 No Content）:**
レスポンスボディなし

**エラー（409 Conflict）:**
```json
{
  "error": {
    "code": "BGM_IN_USE",
    "message": "この BGM は使用中のため削除できません"
  }
}
```

---

# System BGMs（システム BGM）

管理者が提供するシステム BGM。全ユーザーが利用可能。

## システム BGM 一覧取得

```
GET /system-bgms
```

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
      "name": "明るいポップ",
      "audio": {
        "id": "uuid",
        "url": "https://storage.example.com/audios/xxx.mp3?signature=...",
        "durationMs": 180000
      },
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 10,
    "limit": 20,
    "offset": 0
  }
}
```

---

## システム BGM 取得

```
GET /system-bgms/:bgmId
```

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "name": "明るいポップ",
    "audio": {
      "id": "uuid",
      "url": "https://storage.example.com/audios/xxx.mp3?signature=...",
      "durationMs": 180000
    },
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

---

## システム BGM 作成

```
POST /system-bgms
```

**権限:** Admin

**リクエスト:**
```json
{
  "name": "明るいポップ",
  "audioId": "uuid",
  "sortOrder": 0
}
```

**バリデーション:**

| フィールド | ルール |
|------------|--------|
| name | 必須、255 文字以内、一意 |
| audioId | 必須、UUID 形式、存在する音声ファイルのみ指定可能 |
| sortOrder | 0 以上の整数 |

**レスポンス（201 Created）:**
```json
{
  "data": {
    "id": "uuid",
    "name": "明るいポップ",
    "audio": {
      "id": "uuid",
      "url": "https://storage.example.com/audios/xxx.mp3?signature=...",
      "durationMs": 180000
    },
    "sortOrder": 0,
    "isActive": true,
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

**エラー（409 Conflict）:**
```json
{
  "error": {
    "code": "DUPLICATE_NAME",
    "message": "同じ名前のシステム BGM が既に存在します"
  }
}
```

---

## システム BGM 更新

```
PATCH /system-bgms/:bgmId
```

**権限:** Admin

**リクエスト:**
```json
{
  "name": "新しい名前",
  "sortOrder": 1,
  "isActive": false
}
```

**バリデーション:**

| フィールド | ルール |
|------------|--------|
| name | 255 文字以内、一意 |
| sortOrder | 0 以上の整数 |
| isActive | boolean |

> **Note:** audioId は更新不可。音声ファイルを変更したい場合は、新しいシステム BGM を作成する。

**レスポンス（200 OK）:**
```json
{
  "data": {
    "id": "uuid",
    "name": "新しい名前",
    "audio": {
      "id": "uuid",
      "url": "https://storage.example.com/audios/xxx.mp3?signature=...",
      "durationMs": 180000
    },
    "sortOrder": 1,
    "isActive": false,
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

---

## システム BGM 削除

```
DELETE /system-bgms/:bgmId
```

**権限:** Admin

**制約:**
- いずれかのエピソードで使用中のシステム BGM は削除不可

**レスポンス（204 No Content）:**
レスポンスボディなし

**エラー（409 Conflict）:**
```json
{
  "error": {
    "code": "BGM_IN_USE",
    "message": "この BGM は使用中のため削除できません"
  }
}
```
