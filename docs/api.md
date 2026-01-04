# Anycast API 設計

## 概要

- **ベース URL**: `/api/v1`
- **形式**: REST API
- **データ形式**: JSON
- **認証**: Bearer Token（JWT）

---

## API 一覧

| メソッド | パス | 説明 | 実装 |
|----------|------|------|:----:|
| **システム** | - | - | - |
| GET | `/health` | ヘルスチェック | ✅ |
| GET | `/swagger/*` | Swagger UI（開発環境のみ） | ✅ |
| **[Auth（認証）](#auth認証)** | - | - | - |
| POST | `/api/v1/auth/register` | [ユーザー登録](#ユーザー登録) | ✅ |
| POST | `/api/v1/auth/login` | [メール/パスワード認証](#メールパスワード認証) | ✅ |
| POST | `/api/v1/auth/oauth/google` | [Google OAuth 認証](#google-oauth-認証) | ✅ |
| **[Users（ユーザー）](#usersユーザー)** | - | - | - |
| GET | `/api/v1/users/:userId` | [ユーザー取得](#ユーザー取得) | |
| **[Channels](#channels)** | - | - | - |
| GET | `/api/v1/channels` | [チャンネル一覧取得](#チャンネル一覧取得) | |
| GET | `/api/v1/channels/:channelId` | [チャンネル取得](#チャンネル取得) | ✅ |
| POST | `/api/v1/channels` | [チャンネル作成](#チャンネル作成) | ✅ |
| PATCH | `/api/v1/channels/:channelId` | [チャンネル更新](#チャンネル更新) | ✅ |
| DELETE | `/api/v1/channels/:channelId` | [チャンネル削除](#チャンネル削除) | ✅ |
| **[Characters](#characters)** | - | - | - |
| GET | `/api/v1/channels/:channelId/characters` | [キャラクター一覧取得](#キャラクター一覧取得) | |
| POST | `/api/v1/channels/:channelId/characters` | [キャラクター作成](#キャラクター作成) | |
| PATCH | `/api/v1/channels/:channelId/characters/:characterId` | [キャラクター更新](#キャラクター更新) | |
| DELETE | `/api/v1/channels/:channelId/characters/:characterId` | [キャラクター削除](#キャラクター削除) | |
| **[Search（検索）](#search検索)** | - | - | - |
| GET | `/api/v1/search/channels` | [チャンネル検索](#チャンネル検索) | |
| GET | `/api/v1/search/episodes` | [エピソード検索](#エピソード検索) | |
| **[Likes（お気に入り）](#likesお気に入り)** | - | - | - |
| POST | `/api/v1/episodes/:episodeId/like` | [お気に入り登録](#お気に入り登録) | |
| DELETE | `/api/v1/episodes/:episodeId/like` | [お気に入り解除](#お気に入り解除) | |
| GET | `/api/v1/me/likes` | [お気に入りしたエピソード一覧](#お気に入りしたエピソード一覧) | |
| **[Bookmarks（後で見る）](#bookmarks後で見る)** | - | - | - |
| POST | `/api/v1/episodes/:episodeId/bookmark` | [ブックマーク登録](#ブックマーク登録) | |
| DELETE | `/api/v1/episodes/:episodeId/bookmark` | [ブックマーク解除](#ブックマーク解除) | |
| GET | `/api/v1/me/bookmarks` | [ブックマークしたエピソード一覧](#ブックマークしたエピソード一覧) | |
| **[Playback History（再生履歴）](#playback-history再生履歴)** | - | - | - |
| PUT | `/api/v1/episodes/:episodeId/playback` | [再生履歴を更新](#再生履歴を更新) | |
| DELETE | `/api/v1/episodes/:episodeId/playback` | [再生履歴を削除](#再生履歴を削除) | |
| GET | `/api/v1/me/playback-history` | [再生履歴一覧を取得](#再生履歴一覧を取得) | |
| **[Follows（フォロー）](#followsフォロー)** | - | - | - |
| POST | `/api/v1/episodes/:episodeId/follow` | [フォロー登録](#フォロー登録) | |
| DELETE | `/api/v1/episodes/:episodeId/follow` | [フォロー解除](#フォロー解除) | |
| GET | `/api/v1/me/follows` | [フォロー中のエピソード一覧](#フォロー中のエピソード一覧) | |
| **[Me（自分のリソース）](#me自分のリソース)** | - | - | - |
| GET | `/api/v1/me` | [現在のユーザー取得](#現在のユーザー取得) | ✅ |
| PATCH | `/api/v1/me` | [ユーザー情報更新](#ユーザー情報更新) | |
| GET | `/api/v1/me/channels` | [自分のチャンネル一覧](#自分のチャンネル一覧取得) | ✅ |
| GET | `/api/v1/me/channels/:channelId/episodes` | [自分のチャンネルのエピソード一覧](#自分のチャンネルのエピソード一覧取得) | ✅ |
| **[Episodes](#episodes)** | - | - | - |
| GET | `/api/v1/channels/:channelId/episodes` | [エピソード一覧取得](#エピソード一覧取得公開用) | |
| GET | `/api/v1/channels/:channelId/episodes/:episodeId` | [エピソード取得](#エピソード取得) | |
| POST | `/api/v1/channels/:channelId/episodes` | [エピソード作成](#エピソード作成) | ✅ |
| PATCH | `/api/v1/channels/:channelId/episodes/:episodeId` | [エピソード更新](#エピソード更新) | ✅ |
| DELETE | `/api/v1/channels/:channelId/episodes/:episodeId` | [エピソード削除](#エピソード削除) | ✅ |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/publish` | [エピソード公開](#エピソード公開) | ✅ |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/unpublish` | [エピソード非公開](#エピソード非公開) | ✅ |
| **[Script（台本）](#script台本)** | - | - | - |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/script/import` | [台本テキスト取り込み](#台本テキスト取り込み) | |
| GET | `/api/v1/channels/:channelId/episodes/:episodeId/script/export` | [台本テキスト出力](#台本テキスト出力) | |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/script/generate` | [台本を AI で生成](#台本を-ai-で生成) | |
| **[ScriptLines（台本行）](#scriptlines台本行)** | - | - | - |
| GET | `/api/v1/channels/:channelId/episodes/:episodeId/script/lines` | [台本行一覧取得](#台本行一覧取得) | ✅ |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/script/lines` | [行追加](#行追加) | |
| PATCH | `/api/v1/channels/:channelId/episodes/:episodeId/script/lines/:lineId` | [行更新](#行更新) | |
| DELETE | `/api/v1/channels/:channelId/episodes/:episodeId/script/lines/:lineId` | [行削除](#行削除) | |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/script/reorder` | [行並び替え](#行並び替え) | |
| **[Audio（音声生成）](#audio音声生成)** | - | - | - |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/script/lines/:lineId/audio/generate` | [行単位音声生成](#行単位音声生成) | |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/audio/generate` | [エピソード全体音声生成](#エピソード全体音声生成) | |
| **[Audios（音声ファイル）](#audios音声ファイル)** | - | - | - |
| POST | `/api/v1/audios` | [音声アップロード](#音声アップロード) | |
| GET | `/api/v1/audios/:audioId` | [音声取得](#音声取得) | |
| DELETE | `/api/v1/audios/:audioId` | [音声削除](#音声削除) | |
| **[Images（画像ファイル）](#images画像ファイル)** | - | - | - |
| POST | `/api/v1/images` | [画像アップロード](#画像アップロード) | |
| GET | `/api/v1/images/:imageId` | [画像取得](#画像取得) | |
| DELETE | `/api/v1/images/:imageId` | [画像削除](#画像削除) | |
| **[Voices（ボイス）](#voicesボイス)** | - | - | - |
| GET | `/api/v1/voices` | [ボイス一覧取得](#ボイス一覧取得) | ✅ |
| GET | `/api/v1/voices/:voiceId` | [ボイス取得](#ボイス取得) | ✅ |
| **[Categories（カテゴリ）](#categoriesカテゴリ)** | - | - | - |
| GET | `/api/v1/categories` | [カテゴリ一覧取得](#カテゴリ一覧取得) | ✅ |
| **[Sound Effects（効果音）](#sound-effects効果音)** | - | - | - |
| GET | `/api/v1/sound-effects` | [効果音一覧取得](#効果音一覧取得) | |
| GET | `/api/v1/sound-effects/:sfxId` | [効果音取得](#効果音取得) | |

---

## 共通仕様

### レスポンス形式

**成功時:**
```json
{
  "data": { ... }
}
```

**エラー時:**
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "エラーメッセージ"
  }
}
```

### ページネーション

一覧取得 API は以下のクエリパラメータをサポート:

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| limit | int | 20 | 取得件数（最大 100） |
| offset | int | 0 | オフセット |

**レスポンス:**
```json
{
  "data": [ ... ],
  "pagination": {
    "total": 100,
    "limit": 20,
    "offset": 0
  }
}
```

### 権限

| 権限レベル | 説明 |
|------------|------|
| Guest | ログインなしでアクセス可能（現時点では該当なし） |
| Public | ログイン済みユーザーであれば誰でもアクセス可能 |
| Owner | 自身のリソースのみ操作可能 |
| Admin | 運営のみ操作可能 |

**エンドポイント別権限:**

| リソース | 参照 | 作成 | 更新 | 削除 |
|----------|:----:|:----:|:----:|:----:|
| Users | Owner | - | Owner | Owner |
| Channels | Public | Owner | Owner | Owner |
| Characters | Public | Owner | Owner | Owner |
| Episodes | Public | Owner | Owner | Owner |
| Script / ScriptLines | Public | Owner | Owner | Owner |
| Likes | Owner | Owner | - | Owner |
| Bookmarks | Owner | Owner | - | Owner |
| Playback History | Owner | Owner | Owner | Owner |
| Follows | Owner | Owner | - | Owner |
| Audio（生成） | - | Owner | - | - |
| Audios（アップロード） | Owner | Owner | - | Owner |
| Images（アップロード） | Owner | Owner | - | Owner |
| Voices | Public | Admin | Admin | Admin |
| Categories | Public | Admin | Admin | Admin |
| Sound Effects | Public | Admin | Admin | Admin |

### 公開状態によるアクセス制御

チャンネルとエピソードには公開状態（`publishedAt`）があり、参照時のアクセス制御に影響する。

| エンドポイント | オーナー | 他ユーザー |
|---------------|----------|------------|
| `GET /channels` | - | 公開中のみ |
| `GET /channels/:channelId` | 全て | 公開中のみ |
| `GET /channels/:channelId/episodes` | - | 公開中のみ |
| `GET /channels/:channelId/episodes/:episodeId` | 全て | 公開中のみ |
| `GET /search/channels` | - | 公開中のみ |
| `GET /search/episodes` | - | 公開中のみ |
| `GET /me/channels` | 全て | - |
| `GET /me/channels/:channelId/episodes` | 全て | - |

- **公開中**: `publishedAt IS NOT NULL AND publishedAt <= NOW()`
- **非公開（下書き）**: `publishedAt IS NULL`
- **予約公開**: `publishedAt > NOW()`（将来的に対応可能）

---

## Auth（認証）

認証は Auth.js（フロントエンド）で処理し、バックエンドはユーザーの作成・検証とトークン検証を担当。

### ユーザー登録

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

### メール/パスワード認証

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

### Google OAuth 認証

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

### 現在のユーザー取得

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

### ユーザー情報更新

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

## Users（ユーザー）

### ユーザー取得

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

## Channels

### チャンネル一覧取得

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

### チャンネル取得

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
    "scriptPrompt": "明るく楽しい雰囲気で...",
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

> **Note:** `scriptPrompt` はオーナーのみに表示されます。他ユーザーがアクセスした場合は含まれません。

### チャンネル作成

```
POST /channels
```

**リクエスト:**
```json
{
  "name": "チャンネル名",
  "description": "説明",
  "scriptPrompt": "明るく楽しい雰囲気で...",
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

### チャンネル更新

```
PATCH /channels/:channelId
```

**リクエスト:**
```json
{
  "name": "新しいチャンネル名",
  "description": "新しい説明",
  "scriptPrompt": "明るく楽しい雰囲気で...",
  "categoryId": "uuid",
  "artworkImageId": "uuid",
  "publishedAt": "2025-01-01T00:00:00Z"
}
```

- `publishedAt`: 公開日時を設定（`null` で非公開化）

### チャンネル削除

```
DELETE /channels/:channelId
```

---

## Characters

### キャラクター一覧取得

```
GET /channels/:channelId/characters
```

### キャラクター作成

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

### キャラクター更新

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

### キャラクター削除

```
DELETE /channels/:channelId/characters/:characterId
```

**バリデーション:**
- チャンネルのキャラクター数が 1 人の場合は削除不可（最低 1 人必要）

---

## Search（検索）

フリーワード検索用のエンドポイント。公開中のコンテンツのみ検索可能。

### チャンネル検索

```
GET /search/channels
```

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| q | string | **必須** | 検索キーワード（name, description を対象） |
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
  ],
  "pagination": {
    "total": 100,
    "limit": 20,
    "offset": 0
  }
}
```

### エピソード検索

```
GET /search/episodes
```

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| q | string | **必須** | 検索キーワード（title, description を対象） |
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
      "channel": {
        "id": "uuid",
        "name": "チャンネル名"
      },
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

## Likes（お気に入り）

エピソードへのお気に入り機能。

### お気に入り登録

```
POST /episodes/:episodeId/like
```

**レスポンス（201 Created）:**
```json
{
  "data": {
    "id": "uuid",
    "episodeId": "uuid",
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```

**エラー（409 Conflict）:**
```json
{
  "error": {
    "code": "ALREADY_LIKED",
    "message": "既にお気に入り済みです"
  }
}
```

### お気に入り解除

```
DELETE /episodes/:episodeId/like
```

**レスポンス（204 No Content）:**
レスポンスボディなし

**エラー（404 Not Found）:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "お気に入りが見つかりません"
  }
}
```

### お気に入りしたエピソード一覧

```
GET /auth/me/likes
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
      "episode": {
        "id": "uuid",
        "title": "エピソードタイトル",
        "description": "説明",
        "channel": {
          "id": "uuid",
          "name": "チャンネル名",
          "artwork": { "id": "uuid", "url": "..." }
        },
        "publishedAt": "2025-01-01T00:00:00Z"
      },
      "likedAt": "2025-01-01T00:00:00Z"
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

## Bookmarks（後で見る）

エピソードへの「後で見る」機能。

### ブックマーク登録

```
POST /episodes/:episodeId/bookmark
```

**レスポンス（201 Created）:**
```json
{
  "data": {
    "id": "uuid",
    "episodeId": "uuid",
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```

**エラー（409 Conflict）:**
```json
{
  "error": {
    "code": "ALREADY_BOOKMARKED",
    "message": "既にブックマーク済みです"
  }
}
```

### ブックマーク解除

```
DELETE /episodes/:episodeId/bookmark
```

**レスポンス（204 No Content）:**
レスポンスボディなし

**エラー（404 Not Found）:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "ブックマークが見つかりません"
  }
}
```

### ブックマークしたエピソード一覧

```
GET /auth/me/bookmarks
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
      "episode": {
        "id": "uuid",
        "title": "エピソードタイトル",
        "description": "説明",
        "channel": {
          "id": "uuid",
          "name": "チャンネル名",
          "artwork": { "id": "uuid", "url": "..." }
        },
        "publishedAt": "2025-01-01T00:00:00Z"
      },
      "bookmarkedAt": "2025-01-01T00:00:00Z"
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

## Playback History（再生履歴）

エピソードの再生履歴を管理する。

### 再生履歴を更新

```
PUT /episodes/:episodeId/playback
```

再生位置や完了状態を更新する。履歴が存在しない場合は新規作成（Upsert）。

**リクエスト:**
```json
{
  "progressMs": 120000,
  "completed": false
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| progressMs | int | | 再生位置（ミリ秒） |
| completed | bool | | 再生完了フラグ |

**レスポンス（200 OK）:**
```json
{
  "data": {
    "episodeId": "uuid",
    "progressMs": 120000,
    "completed": false,
    "playedAt": "2025-01-01T00:00:00Z"
  }
}
```

### 再生履歴を削除

```
DELETE /episodes/:episodeId/playback
```

**レスポンス（204 No Content）:**
レスポンスボディなし

**エラー（404 Not Found）:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "再生履歴が見つかりません"
  }
}
```

### 再生履歴一覧を取得

```
GET /auth/me/playback-history
```

最近再生した順で一覧を取得。

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| completed | bool | - | 完了状態でフィルタ |
| limit | int | 20 | 取得件数（最大 100） |
| offset | int | 0 | オフセット |

**レスポンス:**
```json
{
  "data": [
    {
      "episode": {
        "id": "uuid",
        "title": "エピソードタイトル",
        "description": "説明",
        "fullAudio": { "id": "uuid", "url": "...", "durationMs": 180000 },
        "channel": {
          "id": "uuid",
          "name": "チャンネル名",
          "artwork": { "id": "uuid", "url": "..." }
        },
        "publishedAt": "2025-01-01T00:00:00Z"
      },
      "progressMs": 120000,
      "completed": false,
      "playedAt": "2025-01-01T00:00:00Z"
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

## Follows（フォロー）

他のユーザーのエピソードをフォローする機能。自分のチャンネルのエピソードはフォロー不可。

### フォロー登録

```
POST /episodes/:episodeId/follow
```

**レスポンス（201 Created）:**
```json
{
  "data": {
    "id": "uuid",
    "episodeId": "uuid",
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```

**エラー（400 Bad Request）:**
```json
{
  "error": {
    "code": "SELF_FOLLOW_NOT_ALLOWED",
    "message": "自分のエピソードはフォローできません"
  }
}
```

**エラー（409 Conflict）:**
```json
{
  "error": {
    "code": "ALREADY_FOLLOWED",
    "message": "既にフォロー済みです"
  }
}
```

### フォロー解除

```
DELETE /episodes/:episodeId/follow
```

**レスポンス（204 No Content）:**
レスポンスボディなし

**エラー（404 Not Found）:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "フォローが見つかりません"
  }
}
```

### フォロー中のエピソード一覧

```
GET /auth/me/follows
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
      "episode": {
        "id": "uuid",
        "title": "エピソードタイトル",
        "description": "説明",
        "channel": {
          "id": "uuid",
          "name": "チャンネル名",
          "artwork": { "id": "uuid", "url": "..." }
        },
        "publishedAt": "2025-01-01T00:00:00Z"
      },
      "followedAt": "2025-01-01T00:00:00Z"
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

## Me（自分のリソース）

### 自分のチャンネル一覧取得

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
      "scriptPrompt": "明るく楽しい雰囲気で...",
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

### 自分のチャンネルのエピソード一覧取得

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
      "scriptPrompt": "今回のテーマについて詳しく解説する",
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

## Episodes

### エピソード一覧取得（公開用）

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

### エピソード取得

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
    "scriptPrompt": "今回のテーマについて詳しく解説する",
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

> **Note:** `scriptPrompt` はオーナーのみに表示されます。他ユーザーがアクセスした場合は含まれません。

### エピソード作成

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

### エピソード更新

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

> **Note:** `scriptPrompt` は台本生成時に自動で保存されます。直接編集する場合は API から設定可能ですが、通常は台本生成 API 経由で更新されます。
>
> **Note:** 公開状態の変更は専用エンドポイント（[エピソード公開](#エピソード公開) / [エピソード非公開](#エピソード非公開)）を使用してください。

### エピソード削除

```
DELETE /channels/:channelId/episodes/:episodeId
```

### エピソード公開

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

### エピソード非公開

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

## Script（台本）

### 台本テキスト取り込み

```
POST /channels/:channelId/episodes/:episodeId/script/import
```

**リクエスト:**
```json
{
  "text": "太郎: こんにちは\n花子: やあ\n__SILENCE__: 800\n__SFX__: chime"
}
```

**レスポンス（成功時）:**
```json
{
  "data": {
    "lines": [
      { "id": "uuid", "lineOrder": 0, "lineType": "speech", ... }
    ]
  }
}
```

**レスポンス（エラー時）:**
```json
{
  "error": {
    "code": "SCRIPT_PARSE_ERROR",
    "message": "台本のパースに失敗しました",
    "details": [
      { "line": 3, "reason": "不明な話者: 三郎" },
      { "line": 5, "reason": "__SFX__ の値が不正です" }
    ]
  }
}
```

### 台本テキスト出力

```
GET /channels/:channelId/episodes/:episodeId/script/export
```

**レスポンス:**
```json
{
  "data": {
    "text": "太郎: こんにちは\n花子: やあ\n__SILENCE__: 800"
  }
}
```

### 台本を AI で生成

```
POST /channels/:channelId/episodes/:episodeId/script/generate
```

**リクエスト:**
```json
{
  "prompt": "今日の天気について楽しく話す"
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| prompt | string | ◯ | テーマやシナリオ。URL が含まれていれば RAG で内容を取得して台本生成に利用 |

> **Note:** `prompt` はエピソードの `scriptPrompt` として自動保存されます。

**レスポンス:**
```json
{
  "data": {
    "lines": [ ... ]
  }
}
```

---

## ScriptLines（台本行）

### 台本行一覧取得

```
GET /channels/:channelId/episodes/:episodeId/script/lines
```

指定したエピソードの台本行一覧を `lineOrder` 順で取得する。

**レスポンス:**
```json
{
  "data": [
    {
      "id": "uuid",
      "lineOrder": 0,
      "lineType": "speech",
      "speaker": { "id": "uuid", "name": "太郎" },
      "text": "こんにちは",
      "emotion": null,
      "audio": { "id": "uuid", "url": "...", "durationMs": 2500 },
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    },
    {
      "id": "uuid",
      "lineOrder": 1,
      "lineType": "silence",
      "durationMs": 800,
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    },
    {
      "id": "uuid",
      "lineOrder": 2,
      "lineType": "sfx",
      "sfx": { "id": "uuid", "name": "chime" },
      "volume": 0.8,
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    }
  ]
}
```

### 行追加

```
POST /channels/:channelId/episodes/:episodeId/script/lines
```

**リクエスト（speech）:**
```json
{
  "lineType": "speech",
  "speakerId": "uuid",
  "text": "こんにちは",
  "emotion": "嬉しい",
  "insertAfter": "uuid"
}
```

- `speakerId`: 同じ Channel に属する Character の ID を指定
- `emotion`: 感情・喋り方の指定（任意）

**リクエスト（silence）:**
```json
{
  "lineType": "silence",
  "durationMs": 800,
  "insertAfter": "uuid"
}
```

**リクエスト（sfx）:**
```json
{
  "lineType": "sfx",
  "sfxId": "uuid",
  "volume": 0.8,
  "insertAfter": "uuid"
}
```

- `insertAfter`: 指定した行の後に挿入。null の場合は先頭に挿入。

### 行更新

```
PATCH /channels/:channelId/episodes/:episodeId/script/lines/:lineId
```

**リクエスト:**
```json
{
  "text": "新しいセリフ",
  "emotion": "笑いながら"
}
```

### 行削除

```
DELETE /channels/:channelId/episodes/:episodeId/script/lines/:lineId
```

### 行並び替え

```
POST /channels/:channelId/episodes/:episodeId/script/reorder
```

**リクエスト:**
```json
{
  "lineIds": ["uuid1", "uuid2", "uuid3"]
}
```

---

## Audio（音声生成）

### 行単位音声生成

```
POST /channels/:channelId/episodes/:episodeId/script/lines/:lineId/audio/generate
```

**レスポンス:**
```json
{
  "data": {
    "audio": {
      "id": "uuid",
      "url": "https://storage.example.com/audio.mp3",
      "durationMs": 2500
    }
  }
}
```

### エピソード全体音声生成

```
POST /channels/:channelId/episodes/:episodeId/audio/generate
```

**レスポンス:**
```json
{
  "data": {
    "audio": {
      "id": "uuid",
      "url": "https://storage.example.com/full-episode.mp3",
      "durationMs": 180000
    }
  }
}
```

---

## Audios（音声ファイル）

### 音声アップロード

```
POST /audios
```

**リクエスト:** `multipart/form-data`

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| file | File | ◯ | アップロードするファイル（mp3 など） |

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "mimeType": "audio/mpeg",
    "url": "https://storage.example.com/file.mp3",
    "filename": "bgm.mp3",
    "fileSize": 1024000,
    "durationMs": 180000
  }
}
```

### 音声取得

```
GET /audios/:audioId
```

### 音声削除

```
DELETE /audios/:audioId
```

---

## Images（画像ファイル）

### 画像アップロード

```
POST /images
```

**リクエスト:** `multipart/form-data`

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| file | File | ◯ | アップロードするファイル（png, jpeg など） |

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "mimeType": "image/png",
    "url": "https://storage.example.com/artwork.png",
    "filename": "artwork.png",
    "fileSize": 512000
  }
}
```

### 画像取得

```
GET /images/:imageId
```

### 画像削除

```
DELETE /images/:imageId
```

---

## Voices（ボイス）

システム管理のマスタデータ。参照のみ可能。

### ボイス一覧取得

```
GET /voices
```

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| provider | string | - | プロバイダでフィルタ（例: google） |
| gender | string | - | 性別でフィルタ（male / female / neutral） |

**レスポンス:**
```json
{
  "data": [
    {
      "id": "uuid",
      "provider": "google",
      "providerVoiceId": "ja-JP-Wavenet-C",
      "name": "ja-JP-Wavenet-C",
      "gender": "male",
      "isActive": true
    }
  ]
}
```

### ボイス取得

```
GET /voices/:voiceId
```

---

## Categories（カテゴリ）

システム管理のマスタデータ。参照のみ可能。

### カテゴリ一覧取得

```
GET /categories
```

**レスポンス:**
```json
{
  "data": [
    {
      "id": "uuid",
      "slug": "technology",
      "name": "テクノロジー",
      "sortOrder": 0,
      "isActive": true
    }
  ]
}
```

---

## Sound Effects（効果音）

システム管理のマスタデータ。参照のみ可能。

### 効果音一覧取得

```
GET /sound-effects
```

**レスポンス:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "chime",
      "description": "チャイム音",
      "audio": { "id": "uuid", "url": "...", "durationMs": 1500 }
    }
  ]
}
```

### 効果音取得

```
GET /sound-effects/:sfxId
```

---

## エラーコード一覧

| コード | HTTP Status | 説明 |
|--------|-------------|------|
| VALIDATION_ERROR | 400 | バリデーションエラー |
| RESERVED_NAME | 400 | 予約語を使用している |
| SCRIPT_PARSE_ERROR | 400 | 台本のパースに失敗 |
| UNAUTHORIZED | 401 | 認証が必要 |
| INVALID_CREDENTIALS | 401 | メールアドレスまたはパスワードが正しくない |
| FORBIDDEN | 403 | アクセス権限がない |
| NOT_FOUND | 404 | リソースが見つからない |
| DUPLICATE_EMAIL | 409 | メールアドレスが既に登録済み |
| DUPLICATE_USERNAME | 409 | ユーザー名が既に使用されている |
| DUPLICATE_NAME | 409 | 名前が重複している |
| ALREADY_LIKED | 409 | 既にお気に入り済み |
| ALREADY_BOOKMARKED | 409 | 既にブックマーク済み |
| ALREADY_FOLLOWED | 409 | 既にフォロー済み |
| SELF_FOLLOW_NOT_ALLOWED | 400 | 自分のエピソードはフォロー不可 |
| SFX_IN_USE | 409 | 効果音が使用中のため削除不可 |
| INTERNAL_ERROR | 500 | サーバー内部エラー |
| GENERATION_FAILED | 500 | 音声/台本の生成に失敗 |
| MEDIA_UPLOAD_FAILED | 500 | メディアアップロードに失敗 |
