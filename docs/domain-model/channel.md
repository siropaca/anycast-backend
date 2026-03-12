# Channel 集約

## Channel（チャンネル）

ポッドキャストのチャンネル。複数のエピソードを管理し、User が所有するキャラクターを紐づける。

| 属性 | 型 | 必須 | 説明 |
|------|-----|:----:|------|
| id | UUID | ◯ | 識別子 |
| userId | UUID | ◯ | オーナー |
| name | String | ◯ | チャンネル名 |
| description | String | | チャンネルの概要・説明（公開情報） |
| userPrompt | String | | 台本生成の全体方針（AI への指示、内部管理用） |
| category | Category | ◯ | カテゴリ |
| artwork | Image | | カバー画像（ポッドキャストのアートワーク） |
| characters | Character[] | ◯ | 登場人物（1〜2 人、User が所有するキャラクターへの参照） |
| defaultBgm | Bgm | | デフォルト BGM（ユーザー所有） |
| defaultSystemBgm | SystemBgm | | デフォルト BGM（システム提供） |
| publishedAt | DateTime | | 公開日時（NULL = 下書き） |

### 不変条件

- characters は 1〜2 人（最小 1 人、最大 2 人）
- characters は同一 User が所有するキャラクターのみ紐づけ可能
- characters は全員同一のボイスプロバイダー（Voice.provider）を使用すること
- defaultBgm と defaultSystemBgm は同時に設定不可（排他的）
- 公開中のチャンネルは削除不可（先に非公開化が必要、将来的に検討）
