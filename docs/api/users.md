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
