# Contacts（お問い合わせ）

ユーザーからのお問い合わせを受け付ける API です。認証は任意で、未ログインでもアクセス可能です。

## お問い合わせ送信

```
POST /contacts
```

お問い合わせを送信します。ログイン済みの場合は `user_id` が自動的に紐づけられます。

**リクエスト（JSON）:**

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| category | string | ◯ | カテゴリ（`general`, `bug_report`, `feature_request`, `other`） |
| email | string | ◯ | メールアドレス |
| name | string | ◯ | 名前（1〜100文字） |
| content | string | ◯ | お問い合わせ内容（1〜5000文字） |
| userAgent | string | | ブラウザの User-Agent |

```json
{
  "category": "bug_report",
  "email": "user@example.com",
  "name": "山田太郎",
  "content": "ログイン画面でエラーが発生します。",
  "userAgent": "Mozilla/5.0 ..."
}
```

**レスポンス（201 Created）:**

```json
{
  "data": {
    "id": "uuid",
    "category": "bug_report",
    "email": "user@example.com",
    "name": "山田太郎",
    "content": "ログイン画面でエラーが発生します。",
    "userAgent": "Mozilla/5.0 ...",
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```

| フィールド | 型 | 説明 |
|------------|-----|------|
| id | uuid | お問い合わせ ID |
| category | string | カテゴリ |
| email | string | メールアドレス |
| name | string | 名前 |
| content | string | お問い合わせ内容 |
| userAgent | string | User-Agent（null の場合あり） |
| createdAt | datetime | 作成日時 |

**カテゴリ一覧:**

| 値 | 説明 |
|----|------|
| `general` | 一般的なお問い合わせ |
| `bug_report` | 不具合の報告 |
| `feature_request` | 機能リクエスト |
| `other` | その他 |

**処理フロー:**

1. リクエストをバリデーション
2. ログイン済みの場合は `user_id` をセット
3. お問い合わせを DB に保存
4. Slack 通知（設定されている場合のみ、非同期）
5. レスポンスを返却

**エラー:**

| コード | 説明 |
|--------|------|
| VALIDATION_ERROR | 必須項目が未入力、カテゴリが不正、メール形式が不正、内容が5000文字超過 |

---

## Slack 通知

環境変数 `SLACK_CONTACT_WEBHOOK_URL` が設定されている場合、お問い合わせ受信時に Slack へ通知を送信します。

**通知内容:**

- カテゴリ（日本語ラベル）
- 名前、メールアドレス
- お問い合わせ内容
- ユーザー ID（ログイン済みの場合）
- User-Agent（設定されている場合）
- 送信日時

通知はバックグラウンドで非同期に実行されるため、Slack への送信に失敗してもお問い合わせの保存自体は成功します。
