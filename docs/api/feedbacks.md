# Feedbacks（フィードバック）

ユーザーからのフィードバックを受け付ける API です。

## フィードバック送信

```
POST /feedbacks
```

ログインユーザーがフィードバックを送信します。
スクリーンショット画像を添付することも可能です。

**リクエスト（multipart/form-data）:**

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| content | string | ◯ | フィードバック内容（1〜5000文字） |
| screenshot | file | | スクリーンショット画像（png, jpeg, webp） |
| pageUrl | string | | 現在のページ URL |
| userAgent | string | | ブラウザの User-Agent |

**レスポンス（201 Created）:**
```json
{
  "data": {
    "id": "uuid",
    "content": "ここにフィードバック内容が入ります",
    "screenshot": {
      "id": "uuid",
      "url": "https://storage.example.com/images/xxx.png"
    },
    "pageUrl": "https://app.example.com/channels/123",
    "userAgent": "Mozilla/5.0 ...",
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```

| フィールド | 型 | 説明 |
|------------|-----|------|
| id | uuid | フィードバック ID |
| content | string | フィードバック内容 |
| screenshot | object | スクリーンショット画像（null の場合あり） |
| screenshot.id | uuid | 画像 ID |
| screenshot.url | string | 署名付き URL（有効期限あり） |
| pageUrl | string | ページ URL（null の場合あり） |
| userAgent | string | User-Agent（null の場合あり） |
| createdAt | datetime | 作成日時 |

**処理フロー:**

1. フィードバック内容をバリデーション
2. スクリーンショットがある場合は GCS にアップロード
3. フィードバックを DB に保存
4. Slack 通知（設定されている場合のみ、非同期）
5. レスポンスを返却

**エラー:**

| コード | 説明 |
|--------|------|
| VALIDATION_ERROR | フィードバック内容が空、または5000文字超過 |
| VALIDATION_ERROR | スクリーンショットの形式が不正（png, jpeg, webp 以外） |

---

## Slack 通知

環境変数 `SLACK_WEBHOOK_URL` が設定されている場合、フィードバック受信時に Slack へ通知を送信します。

**通知内容:**

- ユーザー情報（メールアドレス、表示名）
- フィードバック内容
- スクリーンショット URL（設定されている場合）
- ページ URL（設定されている場合）
- User-Agent（設定されている場合）
- 送信日時

通知はバックグラウンドで非同期に実行されるため、Slack への送信に失敗してもフィードバックの保存自体は成功します。
