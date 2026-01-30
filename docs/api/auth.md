# Auth（認証）

認証は Auth.js（フロントエンド）で処理し、バックエンドはユーザーの作成・検証とトークン検証を担当。

## トークン仕様

| トークン | 形式 | 有効期限 | 保存場所 |
|----------|------|----------|----------|
| アクセストークン | JWT（HS256） | 1時間 | クライアント側（メモリ） |
| リフレッシュトークン | ランダム文字列 | 30日 | DB + クライアント側 |

- アクセストークンの有効期限が切れた場合、リフレッシュトークンを使用して新しいアクセストークンを取得する
- リフレッシュ時はトークンローテーションを行い、新しいリフレッシュトークンも発行する（旧トークンは無効化）

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
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "username": "user_name",
      "displayName": "ユーザー名",
      "avatarUrl": null
    },
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "refreshToken": "random-token-string"
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
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "username": "user_name",
      "displayName": "ユーザー名",
      "avatarUrl": "https://..."
    },
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "refreshToken": "random-token-string"
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
    "user": {
      "id": "uuid",
      "email": "user@gmail.com",
      "username": "user_name",
      "displayName": "ユーザー名",
      "avatarUrl": "https://..."
    },
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "refreshToken": "random-token-string"
  }
}
```

---

## トークンリフレッシュ

```
POST /auth/refresh
```

リフレッシュトークンを使用して、新しいアクセストークンとリフレッシュトークンを取得する。トークンローテーションにより、使用したリフレッシュトークンは無効化される。

**リクエスト:**
```json
{
  "refreshToken": "current-refresh-token"
}
```

**レスポンス（200 OK）:**
```json
{
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "refreshToken": "new-refresh-token"
  }
}
```

**エラー（401 Unauthorized）:**
```json
{
  "error": {
    "code": "INVALID_REFRESH_TOKEN",
    "message": "リフレッシュトークンが無効です"
  }
}
```

---

## ログアウト

```
POST /auth/logout
```

リフレッシュトークンを無効化する。

**認証:** 必須（Bearer Token）

**リクエスト:**
```json
{
  "refreshToken": "current-refresh-token"
}
```

**レスポンス（204 No Content）:**

レスポンスボディなし

---

## パスワード更新

```
PUT /auth/password
```

認証済みユーザーが自分のパスワードを更新する。

**認証:** 必須（Bearer Token）

**リクエスト:**
```json
{
  "currentPassword": "current_password123",
  "newPassword": "new_password456"
}
```

**バリデーション:**
| フィールド | ルール |
|------------|--------|
| currentPassword | 必須 |
| newPassword | 必須、8〜100文字 |

**レスポンス（204 No Content）:**

レスポンスボディなし

**エラー（401 Unauthorized）:**
```json
{
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "現在のパスワードが正しくありません"
  }
}
```

**エラー（400 Bad Request）:**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "新しいパスワードは8文字以上100文字以下で入力してください"
  }
}
```
