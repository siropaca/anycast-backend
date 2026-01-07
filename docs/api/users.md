# Users（ユーザー）

## ユーザー取得

```
GET /users/:userId
```

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "username": "user_name",
    "displayName": "ユーザー名",
    "avatar": { "id": "uuid", "url": "..." },
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```

※ 他ユーザーの情報は公開プロフィールのみ（email は非公開）

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
    "avatar": { "id": "uuid", "url": "..." },
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
  "avatarImageId": "uuid"
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
    "avatar": { "id": "uuid", "url": "..." },
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
3. **Episode.userPrompt** - エピソード固有の設定
