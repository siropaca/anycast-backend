# Anycast API 設計

## 概要

- **ベース URL**: `/api/v1`
- **形式**: REST API
- **データ形式**: JSON
- **認証**: Bearer Token（JWT）

---

## API 一覧

| メソッド | パス | 説明 | 実装 | 詳細 |
|----------|------|------|:----:|------|
| **システム** | - | - | - | - |
| GET | `/health` | ヘルスチェック | ✅ | - |
| GET | `/swagger/*` | Swagger UI（開発環境のみ） | ✅ | - |
| **Auth（認証）** | - | - | - | [auth.md](./auth.md) |
| POST | `/api/v1/auth/register` | ユーザー登録 | ✅ | [詳細](./auth.md#ユーザー登録) |
| POST | `/api/v1/auth/login` | メール/パスワード認証 | ✅ | [詳細](./auth.md#メールパスワード認証) |
| POST | `/api/v1/auth/oauth/google` | Google OAuth 認証 | ✅ | [詳細](./auth.md#google-oauth-認証) |
| **Users（ユーザー）** | - | - | - | [users.md](./users.md) |
| GET | `/api/v1/users/:userId` | ユーザー取得 | | [詳細](./users.md#ユーザー取得) |
| GET | `/api/v1/me` | 現在のユーザー取得 | ✅ | [詳細](./users.md#現在のユーザー取得) |
| PATCH | `/api/v1/me` | ユーザー情報更新 | | [詳細](./users.md#ユーザー情報更新) |
| **Channels** | - | - | - | [channels.md](./channels.md) |
| GET | `/api/v1/channels` | チャンネル一覧取得 | | [詳細](./channels.md#チャンネル一覧取得) |
| GET | `/api/v1/channels/:channelId` | チャンネル取得 | ✅ | [詳細](./channels.md#チャンネル取得) |
| POST | `/api/v1/channels` | チャンネル作成 | ✅ | [詳細](./channels.md#チャンネル作成) |
| PATCH | `/api/v1/channels/:channelId` | チャンネル更新 | ✅ | [詳細](./channels.md#チャンネル更新) |
| DELETE | `/api/v1/channels/:channelId` | チャンネル削除 | ✅ | [詳細](./channels.md#チャンネル削除) |
| POST | `/api/v1/channels/:channelId/publish` | チャンネル公開 | ✅ | [詳細](./channels.md#チャンネル公開) |
| POST | `/api/v1/channels/:channelId/unpublish` | チャンネル非公開 | ✅ | [詳細](./channels.md#チャンネル非公開) |
| GET | `/api/v1/me/channels` | 自分のチャンネル一覧 | ✅ | [詳細](./channels.md#自分のチャンネル一覧取得) |
| GET | `/api/v1/me/channels/:channelId` | 自分のチャンネル取得 | ✅ | [詳細](./channels.md#自分のチャンネル取得) |
| **Characters** | - | - | - | [characters.md](./characters.md) |
| GET | `/api/v1/me/characters` | キャラクター一覧取得 | ✅ | [詳細](./characters.md#キャラクター一覧取得) |
| GET | `/api/v1/me/characters/:characterId` | キャラクター取得 | | [詳細](./characters.md#キャラクター取得) |
| POST | `/api/v1/me/characters` | キャラクター作成 | | [詳細](./characters.md#キャラクター作成) |
| PATCH | `/api/v1/me/characters/:characterId` | キャラクター更新 | | [詳細](./characters.md#キャラクター更新) |
| DELETE | `/api/v1/me/characters/:characterId` | キャラクター削除 | | [詳細](./characters.md#キャラクター削除) |
| PUT | `/api/v1/channels/:channelId/characters` | チャンネルのキャラクター紐づけ更新 | | [詳細](./channels.md#チャンネルのキャラクター紐づけ更新) |
| **Episodes** | - | - | - | [episodes.md](./episodes.md) |
| GET | `/api/v1/channels/:channelId/episodes` | エピソード一覧取得 | | [詳細](./episodes.md#エピソード一覧取得公開用) |
| GET | `/api/v1/channels/:channelId/episodes/:episodeId` | エピソード取得 | | [詳細](./episodes.md#エピソード取得) |
| POST | `/api/v1/channels/:channelId/episodes` | エピソード作成 | ✅ | [詳細](./episodes.md#エピソード作成) |
| PATCH | `/api/v1/channels/:channelId/episodes/:episodeId` | エピソード更新 | ✅ | [詳細](./episodes.md#エピソード更新) |
| DELETE | `/api/v1/channels/:channelId/episodes/:episodeId` | エピソード削除 | ✅ | [詳細](./episodes.md#エピソード削除) |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/publish` | エピソード公開 | ✅ | [詳細](./episodes.md#エピソード公開) |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/unpublish` | エピソード非公開 | ✅ | [詳細](./episodes.md#エピソード非公開) |
| GET | `/api/v1/me/channels/:channelId/episodes` | 自分のチャンネルのエピソード一覧 | ✅ | [詳細](./episodes.md#自分のチャンネルのエピソード一覧取得) |
| GET | `/api/v1/me/channels/:channelId/episodes/:episodeId` | 自分のチャンネルのエピソード取得 | ✅ | [詳細](./episodes.md#自分のチャンネルのエピソード取得) |
| **Script（台本）** | - | - | - | [script.md](./script.md) |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/script/generate` | 台本を AI で生成 | ✅ | [詳細](./script.md#台本を-ai-で生成) |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/script/import` | 台本テキスト取り込み | ✅ | [詳細](./script.md#台本テキスト取り込み) |
| GET | `/api/v1/channels/:channelId/episodes/:episodeId/script/export` | 台本テキスト出力 | ✅ | [詳細](./script.md#台本テキスト出力) |
| GET | `/api/v1/channels/:channelId/episodes/:episodeId/script/lines` | 台本行一覧取得 | ✅ | [詳細](./script.md#台本行一覧取得) |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/script/lines` | 行追加 | | [詳細](./script.md#行追加) |
| PATCH | `/api/v1/channels/:channelId/episodes/:episodeId/script/lines/:lineId` | 行更新 | | [詳細](./script.md#行更新) |
| DELETE | `/api/v1/channels/:channelId/episodes/:episodeId/script/lines/:lineId` | 行削除 | ✅ | [詳細](./script.md#行削除) |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/script/reorder` | 行並び替え | | [詳細](./script.md#行並び替え) |
| **Audio（音声生成）** | - | - | - | [media.md](./media.md) |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/script/lines/:lineId/audio/generate` | 行単位音声生成 | ✅ | [詳細](./media.md#行単位音声生成) |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/audio/generate` | エピソード全体音声生成 | | [詳細](./media.md#エピソード全体音声生成) |
| **Images（画像ファイル）** | - | - | - | [media.md](./media.md#images画像ファイル) |
| POST | `/api/v1/images` | 画像アップロード | | [詳細](./media.md#画像アップロード) |
| **Search（検索）** | - | - | - | [engagement.md](./engagement.md) |
| GET | `/api/v1/search/channels` | チャンネル検索 | | [詳細](./engagement.md#チャンネル検索) |
| GET | `/api/v1/search/episodes` | エピソード検索 | | [詳細](./engagement.md#エピソード検索) |
| GET | `/api/v1/search/users` | ユーザー検索 | | [詳細](./engagement.md#ユーザー検索) |
| **Likes（お気に入り）** | - | - | - | [engagement.md](./engagement.md#likesお気に入り) |
| POST | `/api/v1/episodes/:episodeId/like` | お気に入り登録 | | [詳細](./engagement.md#お気に入り登録) |
| DELETE | `/api/v1/episodes/:episodeId/like` | お気に入り解除 | | [詳細](./engagement.md#お気に入り解除) |
| GET | `/api/v1/me/likes` | お気に入りしたエピソード一覧 | | [詳細](./engagement.md#お気に入りしたエピソード一覧) |
| **Bookmarks（後で見る）** | - | - | - | [engagement.md](./engagement.md#bookmarks後で見る) |
| POST | `/api/v1/episodes/:episodeId/bookmark` | ブックマーク登録 | | [詳細](./engagement.md#ブックマーク登録) |
| DELETE | `/api/v1/episodes/:episodeId/bookmark` | ブックマーク解除 | | [詳細](./engagement.md#ブックマーク解除) |
| GET | `/api/v1/me/bookmarks` | ブックマークしたエピソード一覧 | | [詳細](./engagement.md#ブックマークしたエピソード一覧) |
| **Playback History（再生履歴）** | - | - | - | [engagement.md](./engagement.md#playback-history再生履歴) |
| PUT | `/api/v1/episodes/:episodeId/playback` | 再生履歴を更新 | | [詳細](./engagement.md#再生履歴を更新) |
| DELETE | `/api/v1/episodes/:episodeId/playback` | 再生履歴を削除 | | [詳細](./engagement.md#再生履歴を削除) |
| GET | `/api/v1/me/playback-history` | 再生履歴一覧を取得 | | [詳細](./engagement.md#再生履歴一覧を取得) |
| **Follows（フォロー）** | - | - | - | [engagement.md](./engagement.md#followsフォロー) |
| POST | `/api/v1/episodes/:episodeId/follow` | フォロー登録 | | [詳細](./engagement.md#フォロー登録) |
| DELETE | `/api/v1/episodes/:episodeId/follow` | フォロー解除 | | [詳細](./engagement.md#フォロー解除) |
| GET | `/api/v1/me/follows` | フォロー中のエピソード一覧 | | [詳細](./engagement.md#フォロー中のエピソード一覧) |
| **Voices（ボイス）** | - | - | - | [master.md](./master.md) |
| GET | `/api/v1/voices` | ボイス一覧取得 | ✅ | [詳細](./master.md#ボイス一覧取得) |
| GET | `/api/v1/voices/:voiceId` | ボイス取得 | ✅ | [詳細](./master.md#ボイス取得) |
| **Categories（カテゴリ）** | - | - | - | [master.md](./master.md#categoriesカテゴリ) |
| GET | `/api/v1/categories` | カテゴリ一覧取得 | ✅ | [詳細](./master.md#カテゴリ一覧取得) |
| **Sound Effects（効果音）** | - | - | - | [master.md](./master.md#sound-effects効果音) |
| GET | `/api/v1/sound-effects` | 効果音一覧取得 | | [詳細](./master.md#効果音一覧取得) |
| GET | `/api/v1/sound-effects/:sfxId` | 効果音取得 | | [詳細](./master.md#効果音取得) |
| **Admin（管理者）** | - | - | - | [admin.md](./admin.md) |
| POST | `/admin/cleanup/orphaned-media` | 孤児メディアファイル削除 | ✅ | [詳細](./admin.md#孤児メディアファイル削除) |

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
| Characters | Owner | Owner | Owner | Owner |
| Episodes | Public | Owner | Owner | Owner |
| Script / ScriptLines | Owner | Owner | Owner | Owner |
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
| CHARACTER_IN_USE | 409 | キャラクターが使用中のため削除不可 |
| INTERNAL_ERROR | 500 | サーバー内部エラー |
| GENERATION_FAILED | 500 | 音声/台本の生成に失敗 |
| MEDIA_UPLOAD_FAILED | 500 | メディアアップロードに失敗 |
