# Characters（キャラクター）

ユーザーが所有するキャラクター。複数のチャンネルで使い回すことができる。

## キャラクター一覧取得

```
GET /me/characters
```

自分のキャラクター一覧を取得。

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
      "name": "太郎",
      "persona": "明るく元気な性格",
      "avatar": {
        "id": "uuid",
        "url": "https://storage.example.com/images/xxx.png?signature=..."
      },
      "voice": {
        "id": "uuid",
        "name": "ja-JP-Wavenet-C",
        "provider": "google",
        "gender": "male"
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

## キャラクター取得

```
GET /me/characters/:characterId
```

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "name": "太郎",
    "persona": "明るく元気な性格",
    "avatar": {
      "id": "uuid",
      "url": "https://storage.example.com/images/xxx.png?signature=..."
    },
    "voice": {
      "id": "uuid",
      "name": "ja-JP-Wavenet-C",
      "provider": "google",
      "gender": "male"
    },
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

---

## キャラクター作成

```
POST /me/characters
```

**リクエスト:**
```json
{
  "name": "太郎",
  "persona": "明るく元気な性格。語尾に「だよね」をつける。",
  "avatarId": "uuid",
  "voiceId": "uuid"
}
```

**バリデーション:**

| フィールド | ルール |
|------------|--------|
| name | 必須、255文字以内、同一ユーザー内で一意、`__` で始まる名前は禁止 |
| persona | 2000文字以内 |
| avatarId | UUID 形式、存在する画像のみ指定可能 |
| voiceId | 必須、UUID 形式、is_active = true のボイスのみ指定可能 |

**レスポンス（201 Created）:**
```json
{
  "data": {
    "id": "uuid",
    "name": "太郎",
    "persona": "明るく元気な性格。語尾に「だよね」をつける。",
    "avatar": {
      "id": "uuid",
      "url": "https://storage.example.com/images/xxx.png?signature=..."
    },
    "voice": {
      "id": "uuid",
      "name": "ja-JP-Wavenet-C",
      "provider": "google",
      "gender": "male"
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
    "message": "同じ名前のキャラクターが既に存在します"
  }
}
```

---

## キャラクター更新

```
PATCH /me/characters/:characterId
```

**リクエスト:**
```json
{
  "name": "新しい名前",
  "persona": "新しいペルソナ",
  "avatarId": "uuid",
  "voiceId": "uuid"
}
```

**バリデーション:**

| フィールド | ルール |
|------------|--------|
| name | 255文字以内、同一ユーザー内で一意、`__` で始まる名前は禁止 |
| persona | 2000文字以内 |
| avatarId | UUID 形式、存在する画像のみ指定可能 |
| voiceId | UUID 形式、is_active = true のボイスのみ指定可能 |

**レスポンス（200 OK）:**
```json
{
  "data": {
    "id": "uuid",
    "name": "新しい名前",
    "persona": "新しいペルソナ",
    "avatar": {
      "id": "uuid",
      "url": "https://storage.example.com/images/xxx.png?signature=..."
    },
    "voice": {
      "id": "uuid",
      "name": "ja-JP-Wavenet-C",
      "provider": "google",
      "gender": "male"
    },
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

---

## キャラクター削除

```
DELETE /me/characters/:characterId
```

**制約:**
- いずれかのチャンネルで使用中のキャラクターは削除不可

**レスポンス（204 No Content）:**
レスポンスボディなし

**エラー（409 Conflict）:**
```json
{
  "error": {
    "code": "CHARACTER_IN_USE",
    "message": "このキャラクターは使用中のため削除できません"
  }
}
```
