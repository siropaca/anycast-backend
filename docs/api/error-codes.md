# エラーコード一覧

| コード | HTTP Status | 説明 |
|--------|-------------|------|
| VALIDATION_ERROR | 400 | バリデーションエラー |
| SCRIPT_PARSE_ERROR | 400 | 台本のパースに失敗 |
| UNAUTHORIZED | 401 | 認証が必要 |
| INVALID_CREDENTIALS | 401 | メールアドレスまたはパスワードが正しくない |
| INVALID_REFRESH_TOKEN | 401 | リフレッシュトークンが無効または期限切れ |
| FORBIDDEN | 403 | アクセス権限がない |
| NOT_FOUND | 404 | リソースが見つからない |
| DUPLICATE_EMAIL | 409 | メールアドレスが既に登録済み |
| DUPLICATE_USERNAME | 409 | ユーザー名が既に使用されている |
| DUPLICATE_NAME | 409 | 名前が重複している |
| ALREADY_LIKED | 409 | 既にお気に入り済み |
| ALREADY_IN_PLAYLIST | 409 | 既にプレイリストに追加済み |
| ALREADY_FOLLOWED | 409 | 既にフォロー済み |
| DEFAULT_PLAYLIST | 409 | デフォルトプレイリストは変更不可 |
| SELF_FOLLOW_NOT_ALLOWED | 400 | 自分自身はフォロー不可 |
| CHARACTER_IN_USE | 409 | キャラクターが使用中のため削除不可 |
| BGM_IN_USE | 409 | BGM が使用中のため削除不可 |
| CANCELED | 499 | ジョブがキャンセルされた |
| INTERNAL_ERROR | 500 | サーバー内部エラー |
| GENERATION_FAILED | 500 | 音声/台本の生成に失敗 |
| MEDIA_UPLOAD_FAILED | 500 | メディアアップロードに失敗 |
