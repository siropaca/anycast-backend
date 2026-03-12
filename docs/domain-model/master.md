# マスタデータ

システムが管理するマスタデータ。ユーザーは参照のみ可能。

## Voice（ボイス）

TTS ボイスの設定情報。

| 属性 | 型 | 必須 | 説明 |
|------|-----|:----:|------|
| id | UUID | ◯ | 識別子 |
| provider | String | ◯ | TTS プロバイダ（google / elevenlabs） |
| providerVoiceId | String | ◯ | プロバイダ側の音声 ID（例: ja-JP-Wavenet-C） |
| name | String | ◯ | 表示名 |
| gender | Gender | ◯ | 性別（male / female / neutral） |
| sampleAudioUrl | String | ◯ | サンプルボイス音声の URL |
| isActive | Boolean | ◯ | 有効フラグ |

### 制約

- provider + providerVoiceId の組み合わせは一意
- isActive = false の Voice は新規キャラクター作成時に選択不可
- 既存キャラクターは isActive = false の Voice を継続利用可能
- 物理削除は行わず、isActive フラグで無効化

### TTS プロバイダ

| プロバイダ | 説明 |
|------------|------|
| google | Gemini TTS（Vertex AI 経由） |
| elevenlabs | ElevenLabs TTS |

---

## Category（カテゴリ）

ポッドキャストのカテゴリ。Apple Podcasts 準拠。

| 属性 | 型 | 必須 | 説明 |
|------|-----|:----:|------|
| id | UUID | ◯ | 識別子 |
| slug | String | ◯ | URL フレンドリーな識別子（例: technology） |
| name | String | ◯ | 表示名（例: テクノロジー） |
| image | Image | | カテゴリ画像 |
| sortOrder | Int | ◯ | 表示順 |
| isActive | Boolean | ◯ | 有効フラグ |

### 制約

- slug は一意
- isActive = false のカテゴリは新規チャンネル作成時に選択不可
- 既存チャンネルは isActive = false のカテゴリを継続利用可能
- 物理削除は行わず、isActive フラグで無効化

---

## SystemBgm（システム BGM）

管理者が提供するシステム BGM。ユーザーは参照のみ可能。

| 属性 | 型 | 必須 | 説明 |
|------|-----|:----:|------|
| id | UUID | ◯ | 識別子 |
| audioId | UUID | ◯ | 音声ファイル（Audio） |
| name | String | ◯ | BGM 名 |
| sortOrder | Int | ◯ | 表示順 |
| isActive | Boolean | ◯ | 有効フラグ |

### 制約

- name はシステム全体で一意
- isActive = false のシステム BGM は新規設定時に選択不可
- 既存エピソードは isActive = false のシステム BGM を継続利用可能
- 物理削除は行わず、isActive フラグで無効化
