# User 集約

## User（ユーザー）

サービスの利用者。複数のチャンネルとキャラクターを所有できる。

| 属性 | 型 | 必須 | 説明 |
|------|-----|:----:|------|
| id | UUID | ◯ | 識別子 |
| email | Email | ◯ | メールアドレス（一意） |
| username | Username | ◯ | ユーザー ID（20文字以内、一意、日本語可） |
| displayName | String | ◯ | 表示名（20文字以内） |
| role | Role | ◯ | ロール（デフォルト: user） |
| avatar | Image | | アバター画像 |
| headerImage | Image | | プロフィールのヘッダー画像 |
| bio | String | | 自己紹介文（200文字以内） |
| userPrompt | String | | 台本生成の基本方針（全チャンネル・エピソードに適用、内部管理用） |

### userPrompt の適用

台本生成時、プロンプトは以下の順序で結合（追記）される：

1. **User.userPrompt** - ユーザーの基本方針
2. **Channel.userPrompt** - チャンネル固有の方針

### username の自動生成

- 新規登録時、`displayName` から自動生成される
- スペースはアンダースコアに変換
- 重複時はランダムな番号をサフィックスとして付与
- 後からアカウント設定画面で変更可能

---

## Credential（パスワード認証）

メール/パスワード認証用の認証情報。

| 属性 | 型 | 必須 | 説明 |
|------|-----|:----:|------|
| id | UUID | ◯ | 識別子 |
| userId | UUID | ◯ | 所属する User |
| passwordHash | String | ◯ | パスワードハッシュ（bcrypt） |

### 制約

- User と 1:1 の関係（1 ユーザーにつき 1 つ）
- OAuth のみで登録したユーザーは Credential を持たない

---

## OAuthAccount（OAuth 認証）

OAuth プロバイダとの連携情報。

| 属性 | 型 | 必須 | 説明 |
|------|-----|:----:|------|
| id | UUID | ◯ | 識別子 |
| userId | UUID | ◯ | 所属する User |
| provider | OAuthProvider | ◯ | プロバイダ（google） |
| providerUserId | String | ◯ | プロバイダ側のユーザー ID |
| accessToken | String | | アクセストークン |
| refreshToken | String | | リフレッシュトークン |
| expiresAt | DateTime | | トークン有効期限 |

### 制約

- User と 1:N の関係（複数プロバイダ連携可能）
- provider + providerUserId の組み合わせは一意

---

## RefreshToken（リフレッシュトークン）

アクセストークンを再発行するためのリフレッシュトークン。DB に保存してサーバー側で無効化を管理する。

| 属性 | 型 | 必須 | 説明 |
|------|-----|:----:|------|
| id | UUID | ◯ | 識別子 |
| userId | UUID | ◯ | 所属する User |
| token | String | ◯ | トークン文字列（ランダム生成、一意） |
| expiresAt | DateTime | ◯ | 有効期限 |

### 制約

- User と 1:N の関係（複数デバイスからの同時ログインに対応）
- token は一意
- トークンリフレッシュ時はトークンローテーションを行う（旧トークンを削除し、新しいトークンを発行）
- ログアウト時にトークンを削除して無効化
- 有効期限: 30日

---

## ApiKey（API キー）

JWT Bearer トークンの代替認証手段。セキュリティ上、平文は作成時に 1 度だけ表示し、DB には SHA-256 ハッシュのみ保存する。

| 属性 | 型 | 必須 | 説明 |
|------|-----|:----:|------|
| id | UUID | ◯ | 識別子 |
| userId | UUID | ◯ | 所属する User |
| name | String | ◯ | 管理名（100文字以内） |
| keyHash | String | ◯ | API Key の SHA-256 ハッシュ |
| prefix | String | ◯ | 表示用プレフィックス（例: ak_a1b2c3...） |
| lastUsedAt | DateTime | | 最終使用日時 |

### 制約

- User と 1:N の関係（複数の API Key を所持可能）
- 同一 User 内で name は一意
- keyHash は全体で一意
- 認証方式: `X-API-Key` ヘッダーまたは `Authorization: Bearer ak_...` をサポート
- 平文キーは `ak_` プレフィックス + 32 バイトのランダム hex 文字列

---

## Character（キャラクター）

チャンネルに登場させるキャラクター。ユーザーが作成・管理し、複数のチャンネルで使い回すことができる。

| 属性 | 型 | 必須 | 説明 |
|------|-----|:----:|------|
| id | UUID | ◯ | 識別子 |
| userId | UUID | ◯ | 所有する User |
| name | String | ◯ | キャラクター名 |
| persona | String | | キャラクター設定（性格・話し方など） |
| avatar | Image | | アバター画像 |
| voice | Voice | ◯ | TTS ボイス |

### 制約

- **名前の一意性**: 同一 User 内で重複禁止
- **ボイス選択**: is_active = true の Voice のみ選択可能
- **削除制限**: いずれかの Channel で使用中のキャラクターは削除不可
