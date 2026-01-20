# Admin API

管理者専用の API エンドポイント。

---

## Cleanup（クリーンアップ）

システムのメンテナンス用エンドポイント。

---

### 孤児メディアファイル削除

どのテーブルからも参照されていない `audios` / `images` レコードを検出し、GCS ファイルと DB レコードを削除する。

```
POST /admin/cleanup/orphaned-media
```

**権限:** Admin

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| dry_run | boolean | false | `true` の場合、削除対象の一覧を返すのみで実際の削除は行わない |

**対象となる孤児レコード:**

audios（以下すべてに該当しないもの）:
- `episodes.bgm_id`
- `episodes.full_audio_id`

images（以下すべてに該当しないもの）:
- `users.avatar_id`
- `channels.artwork_id`
- `episodes.artwork_id`

**対象条件:**
- `created_at` から 1 時間以上経過したレコードのみ

**削除順序:**
1. GCS ファイル削除
2. DB レコード削除

**レスポンス:**

```json
{
  "data": {
    "dryRun": true,
    "orphanedAudios": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "url": "https://storage.googleapis.com/bucket/audios/xxx.mp3",
        "filename": "xxx.mp3",
        "fileSize": 1024000,
        "createdAt": "2024-01-01T12:00:00Z"
      }
    ],
    "orphanedImages": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440001",
        "url": "https://storage.googleapis.com/bucket/images/yyy.png",
        "filename": "yyy.png",
        "fileSize": 512000,
        "createdAt": "2024-01-01T12:00:00Z"
      }
    ],
    "deletedAudioCount": 0,
    "deletedImageCount": 0,
    "failedAudioCount": 0,
    "failedImageCount": 0
  }
}
```

**レスポンスフィールド:**

| フィールド | 型 | 説明 |
|------------|-----|------|
| dryRun | boolean | dry-run モードかどうか |
| orphanedAudios | array | 孤児 audio レコードの一覧 |
| orphanedImages | array | 孤児 image レコードの一覧 |
| deletedAudioCount | int | 削除した audio の件数（dry-run 時は 0） |
| deletedImageCount | int | 削除した image の件数（dry-run 時は 0） |
| failedAudioCount | int | 削除に失敗した audio の件数 |
| failedImageCount | int | 削除に失敗した image の件数 |

**エラー:**

| コード | HTTP Status | 説明 |
|--------|-------------|------|
| UNAUTHORIZED | 401 | 認証が必要 |
| FORBIDDEN | 403 | Admin 権限が必要 |
| INTERNAL_ERROR | 500 | サーバー内部エラー |

**使用例:**

```bash
# dry-run（削除対象を確認）
curl -X POST "http://localhost:8080/admin/cleanup/orphaned-media?dry_run=true" \
  -H "Authorization: Bearer <admin-token>"

# 実行（実際に削除）
curl -X POST "http://localhost:8080/admin/cleanup/orphaned-media" \
  -H "Authorization: Bearer <admin-token>"
```
