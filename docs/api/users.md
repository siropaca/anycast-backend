# Users（ユーザー）

## ユーザー取得

```
GET /users/:username
```

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "username": "user_name",
    "displayName": "ユーザー名",
    "bio": "自己紹介文",
    "avatar": { "id": "uuid", "url": "..." },
    "headerImage": { "id": "uuid", "url": "..." },
    "channels": [
      {
        "id": "uuid",
        "name": "チャンネル名",
        "description": "説明",
        "category": { "id": "uuid", "slug": "technology", "name": "テクノロジー" },
        "artwork": { "id": "uuid", "url": "..." },
        "episodeCount": 12,
        "publishedAt": "2025-01-01T00:00:00Z",
        "createdAt": "2025-01-01T00:00:00Z",
        "updatedAt": "2025-01-01T00:00:00Z"
      }
    ],
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```

> **Note:**
> - 他ユーザーの情報は公開プロフィールのみ（email は非公開）
> - `channels` には公開中のチャンネルのみ含まれます
> - `episodeCount` は公開済みエピソードの件数です

---

## 現在のユーザー取得

```
GET /me
```

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "user_name",
    "displayName": "ユーザー名",
    "bio": "自己紹介文",
    "avatar": { "id": "uuid", "url": "..." },
    "headerImage": { "id": "uuid", "url": "..." },
    "userPrompt": "台本生成の基本方針...",
    "hasPassword": true,
    "oauthProviders": ["google"],
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```

---

## ユーザー情報更新

```
PATCH /me
```

**リクエスト:**
```json
{
  "username": "new_username",
  "displayName": "新しい名前",
  "bio": "自己紹介文",
  "avatarImageId": "uuid",
  "headerImageId": "uuid"
}
```

---

## ユーザープロンプト更新

```
PATCH /me/prompt
```

ユーザーの台本生成用プロンプト（基本方針）を更新する。このプロンプトはユーザーのすべてのチャンネル・エピソードの台本生成に適用される。

**リクエスト:**
```json
{
  "userPrompt": "台本生成の基本方針..."
}
```

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "user_name",
    "displayName": "ユーザー名",
    "bio": "自己紹介文",
    "avatar": { "id": "uuid", "url": "..." },
    "headerImage": { "id": "uuid", "url": "..." },
    "userPrompt": "台本生成の基本方針...",
    "hasPassword": true,
    "oauthProviders": ["google"],
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```

### プロンプトの適用順序

台本生成時、プロンプトは以下の順序で結合（追記）される：

1. **User.userPrompt** - ユーザーの基本方針
2. **Channel.userPrompt** - チャンネル固有の方針
