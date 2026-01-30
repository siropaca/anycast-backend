# Anycast API 設計

## 概要

- **ベース URL**: `/api/v1`
- **形式**: REST API
- **データ形式**: JSON
- **認証**: Bearer Token（JWT）

---

## API 一覧

| メソッド | パス | 説明 | 権限 | 実装 | 詳細 |
|----------|------|------|:----:|:----:|------|
| **システム** | - | - | - | - | - |
| GET | `/health` | ヘルスチェック | - | ✅ | - |
| GET | `/swagger/*` | Swagger UI（開発環境のみ） | - | ✅ | - |
| **Auth（認証）** | - | - | - | - | [auth.md](./auth.md) |
| POST | `/api/v1/auth/register` | ユーザー登録 | Guest | ✅ | [詳細](./auth.md#ユーザー登録) |
| POST | `/api/v1/auth/login` | メール/パスワード認証 | Guest | ✅ | [詳細](./auth.md#メールパスワード認証) |
| POST | `/api/v1/auth/oauth/google` | Google OAuth 認証 | Guest | ✅ | [詳細](./auth.md#google-oauth-認証) |
| POST | `/api/v1/auth/refresh` | トークンリフレッシュ | Guest | ✅ | [詳細](./auth.md#トークンリフレッシュ) |
| POST | `/api/v1/auth/logout` | ログアウト | Owner | ✅ | [詳細](./auth.md#ログアウト) |
| PUT | `/api/v1/auth/password` | パスワード更新 | Owner | | [詳細](./auth.md#パスワード更新) |
| **Users（ユーザー）** | - | - | - | - | [users.md](./users.md) |
| GET | `/api/v1/users/:userId` | ユーザー取得 | Public | | [詳細](./users.md#ユーザー取得) |
| GET | `/api/v1/me` | 現在のユーザー取得 | Owner | ✅ | [詳細](./users.md#現在のユーザー取得) |
| PATCH | `/api/v1/me` | ユーザー情報更新 | Owner | | [詳細](./users.md#ユーザー情報更新) |
| PATCH | `/api/v1/me/prompt` | ユーザープロンプト更新 | Owner | ✅ | [詳細](./users.md#ユーザープロンプト更新) |
| **Channels** | - | - | - | - | [channels.md](./channels.md) |
| GET | `/api/v1/channels` | チャンネル一覧取得 | Public | | [詳細](./channels.md#チャンネル一覧取得) |
| GET | `/api/v1/channels/:channelId` | チャンネル取得 | Public | ✅ | [詳細](./channels.md#チャンネル取得) |
| POST | `/api/v1/channels` | チャンネル作成 | Owner | ✅ | [詳細](./channels.md#チャンネル作成) |
| PATCH | `/api/v1/channels/:channelId` | チャンネル更新 | Owner | ✅ | [詳細](./channels.md#チャンネル更新) |
| DELETE | `/api/v1/channels/:channelId` | チャンネル削除 | Owner | ✅ | [詳細](./channels.md#チャンネル削除) |
| POST | `/api/v1/channels/:channelId/publish` | チャンネル公開 | Owner | ✅ | [詳細](./channels.md#チャンネル公開) |
| POST | `/api/v1/channels/:channelId/unpublish` | チャンネル非公開 | Owner | ✅ | [詳細](./channels.md#チャンネル非公開) |
| DELETE | `/api/v1/channels/:channelId/default-bgm` | システム BGM 削除 | Owner | ✅ | [詳細](./channels.md#デフォルト-bgm-削除) |
| GET | `/api/v1/me/channels` | 自分のチャンネル一覧 | Owner | ✅ | [詳細](./channels.md#自分のチャンネル一覧取得) |
| GET | `/api/v1/me/channels/:channelId` | 自分のチャンネル取得 | Owner | ✅ | [詳細](./channels.md#自分のチャンネル取得) |
| **Characters** | - | - | - | - | [characters.md](./characters.md) |
| GET | `/api/v1/me/characters` | キャラクター一覧取得 | Owner | ✅ | [詳細](./characters.md#キャラクター一覧取得) |
| GET | `/api/v1/me/characters/:characterId` | キャラクター取得 | Owner | ✅ | [詳細](./characters.md#キャラクター取得) |
| POST | `/api/v1/me/characters` | キャラクター作成 | Owner | ✅ | [詳細](./characters.md#キャラクター作成) |
| PATCH | `/api/v1/me/characters/:characterId` | キャラクター更新 | Owner | ✅ | [詳細](./characters.md#キャラクター更新) |
| DELETE | `/api/v1/me/characters/:characterId` | キャラクター削除 | Owner | ✅ | [詳細](./characters.md#キャラクター削除) |
| PUT | `/api/v1/channels/:channelId/characters` | チャンネルのキャラクター紐づけ更新 | Owner | | [詳細](./channels.md#チャンネルのキャラクター紐づけ更新) |
| **BGMs（BGM）** | - | - | - | - | [bgms.md](./bgms.md) |
| GET | `/api/v1/me/bgms` | BGM 一覧取得 | Owner | ✅ | [詳細](./bgms.md#bgm-一覧取得) |
| GET | `/api/v1/me/bgms/:bgmId` | BGM 取得 | Owner | ✅ | [詳細](./bgms.md#bgm-取得) |
| POST | `/api/v1/me/bgms` | BGM 作成 | Owner | ✅ | [詳細](./bgms.md#bgm-作成) |
| PATCH | `/api/v1/me/bgms/:bgmId` | BGM 更新 | Owner | ✅ | [詳細](./bgms.md#bgm-更新) |
| DELETE | `/api/v1/me/bgms/:bgmId` | BGM 削除 | Owner | ✅ | [詳細](./bgms.md#bgm-削除) |
| **System BGMs（システム BGM）** | - | - | - | - | [bgms.md](./bgms.md#default-bgmsデフォルト-bgm) |
| GET | `/api/v1/system-bgms` | システム BGM 一覧取得 | Public | | [詳細](./bgms.md#デフォルト-bgm-一覧取得) |
| GET | `/api/v1/system-bgms/:bgmId` | システム BGM 取得 | Public | | [詳細](./bgms.md#デフォルト-bgm-取得) |
| POST | `/api/v1/system-bgms` | システム BGM 作成 | Admin | | [詳細](./bgms.md#デフォルト-bgm-作成) |
| PATCH | `/api/v1/system-bgms/:bgmId` | システム BGM 更新 | Admin | | [詳細](./bgms.md#デフォルト-bgm-更新) |
| DELETE | `/api/v1/system-bgms/:bgmId` | システム BGM 削除 | Admin | | [詳細](./bgms.md#デフォルト-bgm-削除) |
| **Episodes** | - | - | - | - | [episodes.md](./episodes.md) |
| GET | `/api/v1/channels/:channelId/episodes` | エピソード一覧取得 | Public | | [詳細](./episodes.md#エピソード一覧取得公開用) |
| GET | `/api/v1/channels/:channelId/episodes/:episodeId` | エピソード取得 | Public | ✅ | [詳細](./episodes.md#エピソード取得) |
| POST | `/api/v1/channels/:channelId/episodes` | エピソード作成 | Owner | ✅ | [詳細](./episodes.md#エピソード作成) |
| PATCH | `/api/v1/channels/:channelId/episodes/:episodeId` | エピソード更新 | Owner | ✅ | [詳細](./episodes.md#エピソード更新) |
| DELETE | `/api/v1/channels/:channelId/episodes/:episodeId` | エピソード削除 | Owner | ✅ | [詳細](./episodes.md#エピソード削除) |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/publish` | エピソード公開 | Owner | ✅ | [詳細](./episodes.md#エピソード公開) |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/unpublish` | エピソード非公開 | Owner | ✅ | [詳細](./episodes.md#エピソード非公開) |
| PUT | `/api/v1/channels/:channelId/episodes/:episodeId/bgm` | エピソード BGM 設定 | Owner | ✅ | [詳細](./episodes.md#エピソード-bgm-設定) |
| DELETE | `/api/v1/channels/:channelId/episodes/:episodeId/bgm` | エピソード BGM 削除 | Owner | ✅ | [詳細](./episodes.md#エピソード-bgm-削除) |
| GET | `/api/v1/me/channels/:channelId/episodes` | 自分のチャンネルのエピソード一覧 | Owner | ✅ | [詳細](./episodes.md#自分のチャンネルのエピソード一覧取得) |
| GET | `/api/v1/me/channels/:channelId/episodes/:episodeId` | 自分のチャンネルのエピソード取得 | Owner | ✅ | [詳細](./episodes.md#自分のチャンネルのエピソード取得) |
| **Script（台本）** | - | - | - | - | [script.md](./script.md) |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/script/generate-async` | 台本を AI で生成（非同期） | Owner | ✅ | [詳細](./script.md#台本を-ai-で生成非同期) |
| GET | `/api/v1/script-jobs/:jobId` | 台本生成ジョブ取得 | Owner | ✅ | [詳細](./script.md#台本生成ジョブ取得) |
| POST | `/api/v1/script-jobs/:jobId/cancel` | 台本生成ジョブキャンセル | Owner | ✅ | [詳細](./script.md#台本生成ジョブキャンセル) |
| GET | `/api/v1/me/script-jobs` | 自分の台本生成ジョブ一覧 | Owner | ✅ | [詳細](./script.md#自分の台本生成ジョブ一覧) |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/script/import` | 台本テキスト取り込み | Owner | ✅ | [詳細](./script.md#台本テキスト取り込み) |
| GET | `/api/v1/channels/:channelId/episodes/:episodeId/script/export` | 台本テキスト出力 | Owner | ✅ | [詳細](./script.md#台本テキスト出力) |
| GET | `/api/v1/channels/:channelId/episodes/:episodeId/script/lines` | 台本行一覧取得 | Owner | ✅ | [詳細](./script.md#台本行一覧取得) |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/script/lines` | 行追加 | Owner | ✅ | [詳細](./script.md#行追加) |
| PATCH | `/api/v1/channels/:channelId/episodes/:episodeId/script/lines/:lineId` | 行更新 | Owner | ✅ | [詳細](./script.md#行更新) |
| DELETE | `/api/v1/channels/:channelId/episodes/:episodeId/script/lines/:lineId` | 行削除 | Owner | ✅ | [詳細](./script.md#行削除) |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/script/reorder` | 行並び替え | Owner | ✅ | [詳細](./script.md#行並び替え) |
| **Audio（音声生成）** | - | - | - | - | [media.md](./media.md) |
| POST | `/api/v1/channels/:channelId/episodes/:episodeId/audio/generate-async` | 非同期音声生成（BGM ミキシング対応） | Owner | ✅ | [詳細](./media.md#非同期音声生成) |
| GET | `/api/v1/audio-jobs/:jobId` | 音声生成ジョブ取得 | Owner | ✅ | [詳細](./media.md#音声生成ジョブ取得) |
| POST | `/api/v1/audio-jobs/:jobId/cancel` | 音声生成ジョブキャンセル | Owner | ✅ | [詳細](./media.md#音声生成ジョブキャンセル) |
| GET | `/api/v1/me/audio-jobs` | 自分の音声生成ジョブ一覧 | Owner | ✅ | [詳細](./media.md#自分の音声生成ジョブ一覧) |
| POST | `/api/v1/audios` | 音声アップロード | Owner | ✅ | [詳細](./media.md#音声アップロード) |
| **WebSocket** | - | - | - | - | [media.md](./media.md#websocket) |
| WS | `/ws/jobs` | ジョブのリアルタイム通知（音声・台本共通） | Owner | ✅ | [詳細](./media.md#websocket-接続) |
| **Images（画像ファイル）** | - | - | - | - | [media.md](./media.md#images画像ファイル) |
| POST | `/api/v1/images` | 画像アップロード | Owner | ✅ | [詳細](./media.md#画像アップロード) |
| **Recommendations（おすすめ）** | - | - | - | - | [recommendations.md](./recommendations.md) |
| GET | `/api/v1/recommendations/channels` | おすすめチャンネル取得 | Optional | ✅ | [詳細](./recommendations.md#おすすめチャンネル取得) |
| GET | `/api/v1/recommendations/episodes` | おすすめエピソード取得 | Optional | ✅ | [詳細](./recommendations.md#おすすめエピソード取得) |
| **Search（検索）** | - | - | - | - | [engagement.md](./engagement.md) |
| GET | `/api/v1/search/channels` | チャンネル検索 | Public | | [詳細](./engagement.md#チャンネル検索) |
| GET | `/api/v1/search/episodes` | エピソード検索 | Public | | [詳細](./engagement.md#エピソード検索) |
| GET | `/api/v1/search/users` | ユーザー検索 | Public | | [詳細](./engagement.md#ユーザー検索) |
| **Reactions（リアクション）** | - | - | - | - | [engagement.md](./engagement.md#reactionsリアクション) |
| POST | `/api/v1/episodes/:episodeId/reactions` | リアクション登録（like/bad） | Owner | | [詳細](./engagement.md#リアクション登録) |
| DELETE | `/api/v1/episodes/:episodeId/reactions` | リアクション解除 | Owner | | [詳細](./engagement.md#リアクション解除) |
| GET | `/api/v1/me/likes` | 高評価したエピソード一覧 | Owner | | [詳細](./engagement.md#高評価したエピソード一覧) |
| **Playlists（プレイリスト）** | - | - | - | - | [engagement.md](./engagement.md#playlistsプレイリスト) |
| GET | `/api/v1/me/playlists` | プレイリスト一覧取得 | Owner | ✅ | [詳細](./engagement.md#プレイリスト一覧取得) |
| GET | `/api/v1/me/playlists/:playlistId` | プレイリスト詳細取得 | Owner | ✅ | [詳細](./engagement.md#プレイリスト詳細取得) |
| POST | `/api/v1/me/playlists` | プレイリスト作成 | Owner | ✅ | [詳細](./engagement.md#プレイリスト作成) |
| PATCH | `/api/v1/me/playlists/:playlistId` | プレイリスト更新 | Owner | ✅ | [詳細](./engagement.md#プレイリスト更新) |
| DELETE | `/api/v1/me/playlists/:playlistId` | プレイリスト削除 | Owner | ✅ | [詳細](./engagement.md#プレイリスト削除) |
| POST | `/api/v1/me/playlists/:playlistId/items` | プレイリストにアイテム追加 | Owner | ✅ | [詳細](./engagement.md#プレイリストにアイテム追加) |
| DELETE | `/api/v1/me/playlists/:playlistId/items/:itemId` | プレイリストからアイテム削除 | Owner | ✅ | [詳細](./engagement.md#プレイリストからアイテム削除) |
| POST | `/api/v1/me/playlists/:playlistId/items/reorder` | プレイリストアイテム並び替え | Owner | ✅ | [詳細](./engagement.md#プレイリストアイテム並び替え) |
| GET | `/api/v1/me/default-playlist` | 再生リスト一覧取得 | Owner | ✅ | [詳細](./engagement.md#再生リスト一覧取得) |
| POST | `/api/v1/episodes/:episodeId/default-playlist` | 再生リストに追加 | Owner | ✅ | [詳細](./engagement.md#再生リストに追加) |
| DELETE | `/api/v1/episodes/:episodeId/default-playlist` | 再生リストから削除 | Owner | ✅ | [詳細](./engagement.md#再生リストから削除) |
| **Playback History（再生履歴）** | - | - | - | - | [engagement.md](./engagement.md#playback-history再生履歴) |
| PUT | `/api/v1/episodes/:episodeId/playback` | 再生履歴を更新 | Owner | ✅ | [詳細](./engagement.md#再生履歴を更新) |
| DELETE | `/api/v1/episodes/:episodeId/playback` | 再生履歴を削除 | Owner | ✅ | [詳細](./engagement.md#再生履歴を削除) |
| GET | `/api/v1/me/playback-history` | 再生履歴一覧を取得 | Owner | ✅ | [詳細](./engagement.md#再生履歴一覧を取得) |
| **Follows（フォロー）** | - | - | - | - | [engagement.md](./engagement.md#followsフォロー) |
| POST | `/api/v1/users/:userId/follow` | フォロー登録 | Owner | | [詳細](./engagement.md#フォロー登録) |
| DELETE | `/api/v1/users/:userId/follow` | フォロー解除 | Owner | | [詳細](./engagement.md#フォロー解除) |
| GET | `/api/v1/me/follows` | フォロー中のユーザー一覧 | Owner | | [詳細](./engagement.md#フォロー中のユーザー一覧) |
| **Comments（コメント）** | - | - | - | - | [engagement.md](./engagement.md#commentsコメント) |
| POST | `/api/v1/episodes/:episodeId/comments` | コメント投稿 | Owner | | [詳細](./engagement.md#コメント投稿) |
| GET | `/api/v1/episodes/:episodeId/comments` | コメント一覧取得 | Public | | [詳細](./engagement.md#コメント一覧取得) |
| PATCH | `/api/v1/comments/:commentId` | コメント編集 | Owner/Admin | | [詳細](./engagement.md#コメント編集) |
| DELETE | `/api/v1/comments/:commentId` | コメント削除 | Owner/Admin | | [詳細](./engagement.md#コメント削除) |
| GET | `/api/v1/me/comments` | 自分のコメント一覧 | Owner | | [詳細](./engagement.md#自分のコメント一覧) |
| **Voices（ボイス）** | - | - | - | - | [master.md](./master.md) |
| GET | `/api/v1/voices` | ボイス一覧取得 | Public | ✅ | [詳細](./master.md#ボイス一覧取得) |
| GET | `/api/v1/voices/:voiceId` | ボイス取得 | Public | ✅ | [詳細](./master.md#ボイス取得) |
| **Categories（カテゴリ）** | - | - | - | - | [master.md](./master.md#categoriesカテゴリ) |
| GET | `/api/v1/categories` | カテゴリ一覧取得 | Public | ✅ | [詳細](./master.md#カテゴリ一覧取得) |
| **Feedbacks（フィードバック）** | - | - | - | - | [feedbacks.md](./feedbacks.md) |
| POST | `/api/v1/feedbacks` | フィードバック送信 | Owner | ✅ | [詳細](./feedbacks.md#フィードバック送信) |
| **Admin（管理者）** | - | - | - | - | [admin.md](./admin.md) |
| POST | `/admin/cleanup/orphaned-media` | 孤児メディアファイル削除 | Admin | ✅ | [詳細](./admin.md#孤児メディアファイル削除) |

---

## 関連ドキュメント

- [共通仕様](./common.md) - レスポンス形式、ページネーション、権限、公開状態によるアクセス制御
- [エラーコード一覧](./error-codes.md) - API で返却されるエラーコード一覧
