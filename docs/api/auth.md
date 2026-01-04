# Auth（認証）

認証は Auth.js（フロントエンド）で処理し、バックエンドはユーザーの作成・検証とトークン検証を担当。

## ユーザー登録

```
POST /auth/register
```

**リクエスト:**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "displayName": "ユーザー名"
}
```

**バリデーション:**
| フィールド | ルール |
|------------|--------|
| email | 必須、有効なメールアドレス形式 |
| password | 必須、8〜100文字 |
| displayName | 必須、20文字以内 |

> **Note:** `username` は `displayName` から自動生成されます。スペースはアンダースコアに変換され、重複時はランダムな番号がサフィックスとして付与されます。

**レスポンス（201 Created）:**
```json
{
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "user_name",
    "displayName": "ユーザー名",
    "avatarUrl": null
  }
}
```

**エラー（409 Conflict）:**
```json
{
  "error": {
    "code": "DUPLICATE_EMAIL",
    "message": "このメールアドレスは既に使用されています"
  }
}
```

---

## メール/パスワード認証

```
POST /auth/login
```

**リクエスト:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**レスポンス（200 OK）:**
```json
{
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "user_name",
    "displayName": "ユーザー名",
    "avatarUrl": "https://..."
  }
}
```

**エラー（401 Unauthorized）:**
```json
{
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "メールアドレスまたはパスワードが正しくありません"
  }
}
```

---

## Google OAuth 認証

```
POST /auth/oauth/google
```

ユーザーが存在しない場合は新規作成、存在する場合はトークン情報を更新。

**リクエスト:**
```json
{
  "providerUserId": "google-provider-id",
  "email": "user@gmail.com",
  "displayName": "ユーザー名",
  "accessToken": "...",
  "refreshToken": "...",
  "expiresAt": 1234567890
}
```

> **Note:** `username` は `displayName` から自動生成されます（新規ユーザー作成時のみ）。

**レスポンス（200 OK / 201 Created）:**
```json
{
  "data": {
    "id": "uuid",
    "email": "user@gmail.com",
    "username": "user_name",
    "displayName": "ユーザー名",
    "avatarUrl": "https://..."
  }
}
```
