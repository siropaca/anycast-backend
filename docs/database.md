# Anycast データベース設計

## ER 図

```mermaid
erDiagram
    users ||--o| credentials : has
    users ||--o{ oauth_accounts : has
    users ||--o{ channels : owns
    users ||--o{ likes : has
    users ||--o{ bookmarks : has
    users ||--o{ playback_histories : has
    users ||--o{ follows : has
    users ||--o| images : avatar
    categories ||--o{ channels : has
    channels ||--o{ characters : has
    channels ||--o{ episodes : has
    channels ||--o| images : artwork
    characters ||--|| voices : uses
    episodes ||--o{ script_lines : has
    episodes ||--o{ likes : has
    episodes ||--o{ bookmarks : has
    episodes ||--o{ playback_histories : has
    episodes ||--o{ follows : has
    episodes ||--o| audios : bgm
    episodes ||--o| audios : full_audio
    script_lines ||--o| characters : speaker
    script_lines ||--o| audios : audio
    script_lines ||--o| sound_effects : sfx
    sound_effects ||--|| audios : audio

    likes {
        uuid id PK
        uuid user_id FK
        uuid episode_id FK
        timestamp created_at
    }

    bookmarks {
        uuid id PK
        uuid user_id FK
        uuid episode_id FK
        timestamp created_at
    }

    playback_histories {
        uuid id PK
        uuid user_id FK
        uuid episode_id FK
        integer progress_ms
        boolean completed
        timestamp played_at
        timestamp created_at
        timestamp updated_at
    }

    follows {
        uuid id PK
        uuid user_id FK
        uuid episode_id FK
        timestamp created_at
    }

    users {
        uuid id PK
        varchar email
        varchar username
        varchar display_name
        uuid avatar_id FK
        timestamp created_at
        timestamp updated_at
    }

    credentials {
        uuid id PK
        uuid user_id FK
        varchar password_hash
        timestamp created_at
        timestamp updated_at
    }

    oauth_accounts {
        uuid id PK
        uuid user_id FK
        varchar provider
        varchar provider_user_id
        varchar access_token
        varchar refresh_token
        timestamp expires_at
        timestamp created_at
        timestamp updated_at
    }

    categories {
        uuid id PK
        varchar slug
        varchar name
        integer sort_order
        timestamp created_at
        timestamp updated_at
    }

    channels {
        uuid id PK
        uuid user_id FK
        uuid category_id FK
        varchar name
        text description
        text script_prompt
        uuid artwork_id FK
        timestamp published_at
        timestamp created_at
        timestamp updated_at
    }

    characters {
        uuid id PK
        uuid channel_id FK
        varchar name
        text persona
        uuid voice_id FK
        timestamp created_at
        timestamp updated_at
    }

    voices {
        uuid id PK
        varchar provider
        varchar provider_voice_id
        varchar name
        varchar gender
        boolean is_active
        timestamp created_at
        timestamp updated_at
    }

    episodes {
        uuid id PK
        uuid channel_id FK
        varchar title
        text description
        text script_prompt
        uuid bgm_id FK
        uuid full_audio_id FK
        timestamp published_at
        timestamp created_at
        timestamp updated_at
    }

    script_lines {
        uuid id PK
        uuid episode_id FK
        integer line_order
        varchar line_type
        uuid speaker_id FK
        text text
        text emotion
        integer duration_ms
        uuid sfx_id FK
        decimal volume
        uuid audio_id FK
        timestamp created_at
        timestamp updated_at
    }

    sound_effects {
        uuid id PK
        varchar name
        text description
        uuid audio_id FK
        timestamp created_at
        timestamp updated_at
    }

    audios {
        uuid id PK
        varchar mime_type
        varchar url
        varchar filename
        integer file_size
        integer duration_ms
        timestamp created_at
    }

    images {
        uuid id PK
        varchar mime_type
        varchar url
        varchar filename
        integer file_size
        timestamp created_at
    }
```

---

## テーブル定義

### 認証テーブル

ユーザー認証に関するデータ。

---

#### users

ユーザー情報を管理する。

| カラム名 | 型 | NULLABLE | デフォルト | 説明 |
|----------|-----|:--------:|------------|------|
| id | UUID | | gen_random_uuid() | 主キー |
| email | VARCHAR(255) | | - | メールアドレス |
| username | VARCHAR(20) | | - | ユーザー ID（displayName から自動生成、日本語可） |
| display_name | VARCHAR(20) | | - | 表示名 |
| avatar_id | UUID | ◯ | - | アバター画像（images 参照） |
| created_at | TIMESTAMP | | CURRENT_TIMESTAMP | 作成日時 |
| updated_at | TIMESTAMP | | CURRENT_TIMESTAMP | 更新日時 |

**インデックス:**
- PRIMARY KEY (id)
- UNIQUE (email)
- UNIQUE (username)

**外部キー:**
- avatar_id → images(id) ON DELETE SET NULL

---

#### credentials

パスワード認証情報を管理する。1 ユーザーにつき 1 レコード。

| カラム名 | 型 | NULLABLE | デフォルト | 説明 |
|----------|-----|:--------:|------------|------|
| id | UUID | | gen_random_uuid() | 主キー |
| user_id | UUID | | - | ユーザー（users 参照） |
| password_hash | VARCHAR(255) | | - | パスワードハッシュ（bcrypt） |
| created_at | TIMESTAMP | | CURRENT_TIMESTAMP | 作成日時 |
| updated_at | TIMESTAMP | | CURRENT_TIMESTAMP | 更新日時 |

**インデックス:**
- PRIMARY KEY (id)
- UNIQUE (user_id)

**外部キー:**
- user_id → users(id) ON DELETE CASCADE

---

#### oauth_accounts

OAuth 認証情報を管理する。1 ユーザーに複数の OAuth プロバイダを紐付け可能。

| カラム名 | 型 | NULLABLE | デフォルト | 説明 |
|----------|-----|:--------:|------------|------|
| id | UUID | | gen_random_uuid() | 主キー |
| user_id | UUID | | - | ユーザー（users 参照） |
| provider | oauth_provider | | - | プロバイダ: `google` |
| provider_user_id | VARCHAR(255) | | - | プロバイダ側のユーザー ID |
| access_token | VARCHAR(1024) | ◯ | - | アクセストークン |
| refresh_token | VARCHAR(1024) | ◯ | - | リフレッシュトークン |
| expires_at | TIMESTAMP | ◯ | - | トークン有効期限 |
| created_at | TIMESTAMP | | CURRENT_TIMESTAMP | 作成日時 |
| updated_at | TIMESTAMP | | CURRENT_TIMESTAMP | 更新日時 |

**インデックス:**
- PRIMARY KEY (id)
- UNIQUE (provider, provider_user_id)
- INDEX (user_id)

**外部キー:**
- user_id → users(id) ON DELETE CASCADE

---

### ユーザーデータテーブル

ユーザーが作成・管理するデータ。

---

#### channels

チャンネル情報を管理する。

| カラム名 | 型 | NULLABLE | デフォルト | 説明 |
|----------|-----|:--------:|------------|------|
| id | UUID | | gen_random_uuid() | 主キー |
| user_id | UUID | | - | オーナー（users 参照） |
| category_id | UUID | | - | カテゴリ（categories 参照） |
| name | VARCHAR(255) | | - | チャンネル名 |
| description | TEXT | | - | チャンネルの説明（公開情報） |
| script_prompt | TEXT | ◯ | - | 台本生成の全体方針（AI への指示、内部管理用） |
| artwork_id | UUID | ◯ | - | カバー画像（images 参照） |
| published_at | TIMESTAMP | ◯ | - | 公開日時（NULL = 下書き） |
| created_at | TIMESTAMP | | CURRENT_TIMESTAMP | 作成日時 |
| updated_at | TIMESTAMP | | CURRENT_TIMESTAMP | 更新日時 |

**インデックス:**
- PRIMARY KEY (id)
- INDEX (user_id)
- INDEX (category_id)
- INDEX (published_at)

**外部キー:**
- user_id → users(id) ON DELETE CASCADE
- category_id → categories(id) ON DELETE RESTRICT
- artwork_id → images(id) ON DELETE SET NULL

---

#### characters

チャンネルに登場するキャラクター情報を管理する。

| カラム名 | 型 | NULLABLE | デフォルト | 説明 |
|----------|-----|:--------:|------------|------|
| id | UUID | | gen_random_uuid() | 主キー |
| channel_id | UUID | | - | 所属チャンネル |
| name | VARCHAR(255) | | - | キャラクター名 |
| persona | TEXT | | - | キャラクター設定 |
| voice_id | UUID | | - | ボイス（voices 参照） |
| created_at | TIMESTAMP | | CURRENT_TIMESTAMP | 作成日時 |
| updated_at | TIMESTAMP | | CURRENT_TIMESTAMP | 更新日時 |

**インデックス:**
- PRIMARY KEY (id)
- UNIQUE (channel_id, name)
- INDEX (channel_id)

**外部キー:**
- channel_id → channels(id) ON DELETE CASCADE
- voice_id → voices(id) ON DELETE RESTRICT

---

#### episodes

エピソード情報を管理する。

| カラム名 | 型 | NULLABLE | デフォルト | 説明 |
|----------|-----|:--------:|------------|------|
| id | UUID | | gen_random_uuid() | 主キー |
| channel_id | UUID | | - | 所属チャンネル |
| title | VARCHAR(255) | | - | エピソードタイトル |
| description | TEXT | | - | エピソードの説明（公開情報） |
| script_prompt | TEXT | ◯ | - | エピソード固有の台本生成設定（内部管理用） |
| bgm_id | UUID | ◯ | - | BGM（audios 参照） |
| full_audio_id | UUID | ◯ | - | 結合済み音声（audios 参照） |
| published_at | TIMESTAMP | ◯ | - | 公開日時（NULL = 下書き） |
| created_at | TIMESTAMP | | CURRENT_TIMESTAMP | 作成日時 |
| updated_at | TIMESTAMP | | CURRENT_TIMESTAMP | 更新日時 |

**インデックス:**
- PRIMARY KEY (id)
- INDEX (channel_id)
- INDEX (published_at)

**外部キー:**
- channel_id → channels(id) ON DELETE CASCADE
- bgm_id → audios(id) ON DELETE SET NULL
- full_audio_id → audios(id) ON DELETE SET NULL

---

#### likes

エピソードへのいいねを管理する。

| カラム名 | 型 | NULLABLE | デフォルト | 説明 |
|----------|-----|:--------:|------------|------|
| id | UUID | | gen_random_uuid() | 主キー |
| user_id | UUID | | - | ユーザー（users 参照） |
| episode_id | UUID | | - | エピソード（episodes 参照） |
| created_at | TIMESTAMP | | CURRENT_TIMESTAMP | いいね登録日時 |

**インデックス:**
- PRIMARY KEY (id)
- UNIQUE (user_id, episode_id)
- INDEX (user_id)
- INDEX (episode_id)

**外部キー:**
- user_id → users(id) ON DELETE CASCADE
- episode_id → episodes(id) ON DELETE CASCADE

---

#### bookmarks

エピソードへの「後で見る」を管理する。

| カラム名 | 型 | NULLABLE | デフォルト | 説明 |
|----------|-----|:--------:|------------|------|
| id | UUID | | gen_random_uuid() | 主キー |
| user_id | UUID | | - | ユーザー（users 参照） |
| episode_id | UUID | | - | エピソード（episodes 参照） |
| created_at | TIMESTAMP | | CURRENT_TIMESTAMP | ブックマーク登録日時 |

**インデックス:**
- PRIMARY KEY (id)
- UNIQUE (user_id, episode_id)
- INDEX (user_id)
- INDEX (episode_id)

**外部キー:**
- user_id → users(id) ON DELETE CASCADE
- episode_id → episodes(id) ON DELETE CASCADE

---

#### playback_histories

エピソードの再生履歴を管理する。

| カラム名 | 型 | NULLABLE | デフォルト | 説明 |
|----------|-----|:--------:|------------|------|
| id | UUID | | gen_random_uuid() | 主キー |
| user_id | UUID | | - | ユーザー（users 参照） |
| episode_id | UUID | | - | エピソード（episodes 参照） |
| progress_ms | INTEGER | | 0 | 再生位置（ミリ秒） |
| completed | BOOLEAN | | false | 再生完了フラグ |
| played_at | TIMESTAMP | | CURRENT_TIMESTAMP | 最終再生日時 |
| created_at | TIMESTAMP | | CURRENT_TIMESTAMP | 初回再生日時 |
| updated_at | TIMESTAMP | | CURRENT_TIMESTAMP | 更新日時 |

**インデックス:**
- PRIMARY KEY (id)
- UNIQUE (user_id, episode_id)
- INDEX (user_id)
- INDEX (episode_id)
- INDEX (user_id, played_at)

**外部キー:**
- user_id → users(id) ON DELETE CASCADE
- episode_id → episodes(id) ON DELETE CASCADE

---

#### follows

エピソードへのフォローを管理する。自分のチャンネルのエピソードはフォロー不可。

| カラム名 | 型 | NULLABLE | デフォルト | 説明 |
|----------|-----|:--------:|------------|------|
| id | UUID | | gen_random_uuid() | 主キー |
| user_id | UUID | | - | ユーザー（users 参照） |
| episode_id | UUID | | - | エピソード（episodes 参照） |
| created_at | TIMESTAMP | | CURRENT_TIMESTAMP | フォロー登録日時 |

**インデックス:**
- PRIMARY KEY (id)
- UNIQUE (user_id, episode_id)
- INDEX (user_id)
- INDEX (episode_id)

**外部キー:**
- user_id → users(id) ON DELETE CASCADE
- episode_id → episodes(id) ON DELETE CASCADE

**制約:**
- 自分が所有するチャンネルのエピソードはフォロー不可（アプリケーション層で検証）

---

#### script_lines

台本の各行（イベント）を管理する。

| カラム名 | 型 | NULLABLE | デフォルト | 説明 |
|----------|-----|:--------:|------------|------|
| id | UUID | | gen_random_uuid() | 主キー（= lineId） |
| episode_id | UUID | | - | 所属エピソード |
| line_order | INTEGER | | - | 行の順序（0 始まり） |
| line_type | line_type | | - | 行種別: `speech` / `silence` / `sfx` |
| speaker_id | UUID | ◯ | - | 話者（speech 時のみ、characters 参照） |
| text | TEXT | ◯ | - | セリフ（speech 時のみ） |
| emotion | TEXT | ◯ | - | 感情・喋り方（speech 時のみ）例: 嬉しい、悲しい、笑いながら |
| duration_ms | INTEGER | ◯ | - | 無音時間 ms（silence 時のみ） |
| sfx_id | UUID | ◯ | - | 効果音（sfx 時のみ、sound_effects 参照） |
| volume | DECIMAL(3,2) | ◯ | 1.00 | 音量 0.00〜1.00（sfx 時のみ） |
| audio_id | UUID | ◯ | - | 生成済み音声（audios 参照） |
| created_at | TIMESTAMP | | CURRENT_TIMESTAMP | 作成日時 |
| updated_at | TIMESTAMP | | CURRENT_TIMESTAMP | 更新日時 |

**インデックス:**
- PRIMARY KEY (id)
- INDEX (episode_id)
- UNIQUE (episode_id, line_order)

**外部キー:**
- episode_id → episodes(id) ON DELETE CASCADE
- speaker_id → characters(id) ON DELETE RESTRICT
- sfx_id → sound_effects(id) ON DELETE RESTRICT
- audio_id → audios(id) ON DELETE SET NULL

**制約:**
- line_type が `speech` の場合: speaker_id, text は NOT NULL
- line_type が `silence` の場合: duration_ms は NOT NULL、0〜10000 の範囲
- line_type が `sfx` の場合: sfx_id は NOT NULL
- volume は 0.00〜1.00 の範囲
- speaker_id は同じ Channel に属する Character のみ指定可能
- （アプリケーション層で検証）

---

#### audios

音声ファイルを管理する。

| カラム名 | 型 | NULLABLE | デフォルト | 説明 |
|----------|-----|:--------:|------------|------|
| id | UUID | | gen_random_uuid() | 主キー |
| mime_type | VARCHAR(100) | | - | MIME タイプ（audio/mpeg など） |
| url | VARCHAR(1024) | | - | ストレージ URL |
| filename | VARCHAR(255) | | - | 元ファイル名 |
| file_size | INTEGER | | - | ファイルサイズ（バイト） |
| duration_ms | INTEGER | | - | 再生時間（ms） |
| created_at | TIMESTAMP | | CURRENT_TIMESTAMP | 作成日時 |

**インデックス:**
- PRIMARY KEY (id)

---

#### images

画像ファイルを管理する。

| カラム名 | 型 | NULLABLE | デフォルト | 説明 |
|----------|-----|:--------:|------------|------|
| id | UUID | | gen_random_uuid() | 主キー |
| mime_type | VARCHAR(100) | | - | MIME タイプ（image/png, image/jpeg など） |
| url | VARCHAR(1024) | | - | ストレージ URL |
| filename | VARCHAR(255) | | - | 元ファイル名 |
| file_size | INTEGER | | - | ファイルサイズ（バイト） |
| created_at | TIMESTAMP | | CURRENT_TIMESTAMP | 作成日時 |

**インデックス:**
- PRIMARY KEY (id)

---

### マスタテーブル

システムが管理するマスタデータ。ユーザーは参照のみ可能。

---

#### categories

ポッドキャストのカテゴリマスタを管理する。

| カラム名 | 型 | NULLABLE | デフォルト | 説明 |
|----------|-----|:--------:|------------|------|
| id | UUID | | gen_random_uuid() | 主キー |
| slug | VARCHAR(50) | | - | 一意識別子（例: technology） |
| name | VARCHAR(100) | | - | 表示名（例: テクノロジー） |
| sort_order | INTEGER | | 0 | 表示順 |
| created_at | TIMESTAMP | | CURRENT_TIMESTAMP | 作成日時 |
| updated_at | TIMESTAMP | | CURRENT_TIMESTAMP | 更新日時 |

**インデックス:**
- PRIMARY KEY (id)
- UNIQUE (slug)
- INDEX (sort_order)

---

#### voices

TTS ボイスのマスタデータを管理する。システム管理テーブルのため、ユーザーは参照のみ可能。

| カラム名 | 型 | NULLABLE | デフォルト | 説明 |
|----------|-----|:--------:|------------|------|
| id | UUID | | gen_random_uuid() | 主キー |
| provider | VARCHAR(50) | | - | TTS プロバイダ: `google` / `azure` / etc |
| provider_voice_id | VARCHAR(100) | | - | プロバイダの音声 ID（例: ja-JP-Wavenet-C） |
| name | VARCHAR(100) | | - | 表示名（デフォルトは provider_voice_id） |
| gender | gender | | - | 性別: `male` / `female` / `neutral` |
| is_active | BOOLEAN | | true | 有効フラグ（false で新規選択不可） |
| created_at | TIMESTAMP | | CURRENT_TIMESTAMP | 作成日時 |
| updated_at | TIMESTAMP | | CURRENT_TIMESTAMP | 更新日時 |

**インデックス:**
- PRIMARY KEY (id)
- UNIQUE (provider, provider_voice_id)
- INDEX (provider)
- INDEX (is_active)

---

#### sound_effects

効果音のマスタデータを管理する。

| カラム名 | 型 | NULLABLE | デフォルト | 説明 |
|----------|-----|:--------:|------------|------|
| id | UUID | | gen_random_uuid() | 主キー |
| name | VARCHAR(100) | | - | 効果音の識別名（例: chime, applause） |
| description | TEXT | | - | 効果音の説明 |
| audio_id | UUID | | - | 音声ファイル（audios 参照） |
| created_at | TIMESTAMP | | CURRENT_TIMESTAMP | 作成日時 |
| updated_at | TIMESTAMP | | CURRENT_TIMESTAMP | 更新日時 |

**インデックス:**
- PRIMARY KEY (id)
- UNIQUE (name)

**外部キー:**
- audio_id → audios(id) ON DELETE RESTRICT

---

## 補足

### Enum 型

PostgreSQL の enum 型を使用して、値の制約を DB レベルで保証する。

| 型名 | 値 | 用途 |
|------|-----|------|
| oauth_provider | `google` | OAuth プロバイダ |
| line_type | `speech`, `silence`, `sfx` | 台本行の種別 |
| gender | `male`, `female`, `neutral` | ボイスの性別 |

### UUID について

- 全テーブルで UUID を主キーとして使用
- PostgreSQL の `gen_random_uuid()` で自動生成
- script_lines の id は仕様上の `lineId` として機能し、行の追跡に使用

### カスケード削除

- Channel 削除時: 関連する Characters, Episodes が削除
- Episode 削除時: 関連する ScriptLines が削除
- Audio / Image 削除時: 参照元は SET NULL（ファイルが消えても親レコードは残る）
- Voice / SoundEffect 削除時: 使用中の場合は RESTRICT（削除不可）

### メディアファイルの管理

- audios テーブルで音声ファイル、images テーブルで画像ファイルを管理
- url には GCS（Google Cloud Storage）の URL を保存
- 同一ファイルを複数箇所から参照可能（BGM の使い回しなど）
- 未使用ファイルのクリーンアップはアプリケーション層で実施

### 効果音の管理

- sound_effects テーブルで効果音マスタを管理
- name は一意で、台本テキストの `__SFX__: <name>` で参照
- 効果音の実体は audios テーブルで管理

### ボイスの管理

- voices テーブルで TTS ボイスのマスタを管理（システム管理、ユーザーは参照のみ）
- provider + provider_voice_id の組み合わせで一意
- name はデフォルトで provider_voice_id と同じ値を設定
- is_active = false のボイスは新規キャラクター作成時に選択不可（既存キャラクターは継続利用可）
- 初期データとして各プロバイダのボイス一覧をシードで投入
- 物理削除は行わず、is_active フラグで無効化

### カテゴリの管理

- categories テーブルでポッドキャストのカテゴリマスタを管理（システム管理、ユーザーは参照のみ）
- slug は一意で、API やフィルタリングで使用
- sort_order で表示順を制御
- チャンネルは 1 つのカテゴリを持つ（任意）
- 初期データとして Apple Podcasts 準拠のカテゴリをシードで投入
