# サポート

## Contact（お問い合わせ）

ユーザーからのお問い合わせ。認証任意で、未ログインユーザーからも受け付け可能。

| 属性 | 型 | 必須 | 説明 |
|------|-----|:----:|------|
| id | UUID | ◯ | 識別子 |
| userId | UUID | | 送信ユーザー（ログイン済みの場合のみ） |
| category | ContactCategory | ◯ | カテゴリ |
| email | Email | ◯ | メールアドレス |
| name | String | ◯ | 名前（100文字以内） |
| content | String | ◯ | お問い合わせ内容（1〜5000文字） |
| userAgent | String | | ブラウザの User-Agent |
| createdAt | DateTime | ◯ | 作成日時 |

### 制約

- 認証任意: ログイン済みの場合は `userId` が自動的にセットされる
- `userId` が設定されている場合、ユーザー削除時に SET NULL（お問い合わせ自体は残る）
- content は 1〜5000 文字
