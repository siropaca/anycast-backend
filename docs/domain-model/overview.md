# 公開状態・アクセス制御・値オブジェクト定義

## 公開状態

Channel と Episode は公開状態（`publishedAt`）を持つ。

| 状態 | 条件 | 説明 |
|------|------|------|
| 下書き | `publishedAt IS NULL` | 非公開。オーナーのみ閲覧可能 |
| 公開中 | `publishedAt <= NOW()` | 公開済み。全ユーザーが閲覧可能 |
| 予約公開 | `publishedAt > NOW()` | 指定日時に自動公開（将来対応） |

### アクセス制御

- **他ユーザー**: 公開中のチャンネル・エピソードのみ閲覧可能
- **オーナー**: 自分のチャンネル・エピソードは全て閲覧可能（下書き含む）
- **管理者**: 全ユーザーのチャンネル・エピソードを閲覧・編集・削除可能
- **検索**: 公開中のコンテンツのみ対象

### ロールと権限

| ロール | 権限 |
|--------|------|
| user | 自分のコンテンツのみ管理可能 |
| admin | マスタデータ（Voice, Category）の CRUD、全ユーザーのコンテンツ管理 |

---

## 値オブジェクト定義

値オブジェクトは **ドメイン固有のルール・バリデーション** を持つ型。同値性で比較され、不変（immutable）として扱う。

| 値オブジェクト | 型 | ルール |
|---------------|-----|--------|
| Email | String | メールアドレス形式、255文字以内、小文字正規化 |
| Username | String | 20文字以内、一意、`__` 始まり禁止、日本語可 |
| OAuthProvider | Enum | `google`（将来的に `apple`, `github` など追加可能） |
| Gender | Enum | `male` / `female` / `neutral` |
| MimeType | String | 有効な MIME タイプ形式（例: `audio/mpeg`, `image/png`） |
| Role | Enum | `user` / `admin` |
| ReactionType | Enum | `like` / `bad` |
| ContactCategory | Enum | `general` / `bug_report` / `feature_request` / `other` |

### 値オブジェクトにしない例

以下は特別なルールを持たないため、単純な `String` として扱う：

- `displayName` - 文字数制限のみ（20文字以内）
- `bio` - 文字数制限のみ（200文字以内）
- `description` - 自由入力テキスト
- `persona` - キャラクター設定（自由記述）
- `name`（Channel, Character など）- 一意性は DB 制約で担保
