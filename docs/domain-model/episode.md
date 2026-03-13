# Episode 集約

## Episode（エピソード）

チャンネル内の各エピソード。台本と BGM を持つ。

| 属性 | 型 | 必須 | 説明 |
|------|-----|:----:|------|
| id | UUID | ◯ | 識別子 |
| channelId | UUID | ◯ | 所属する Channel |
| title | String | ◯ | エピソードタイトル |
| description | String | | エピソードの説明（公開情報） |
| script | ScriptLine[] | | 台本（順序付きの ScriptLine 配列） |
| bgmId | UUID | | ユーザー BGM の参照（Bgm エンティティ） |
| systemBgmId | UUID | | システム BGM の参照（SystemBgm エンティティ） |
| voiceAudio | Audio | | ボイス単体の音声（BGM なし） |
| fullAudio | Audio | | 結合済み音声（全 ScriptLine + BGM） |
| playCount | Int | ◯ | 再生回数（デフォルト: 0） |
| publishedAt | DateTime | | 公開日時（NULL = 下書き） |

### playCount の更新ルール

- 再生開始から 30 秒経過した時点でクライアントが API を呼び出し、`play_count` を +1 する
- 同一ユーザーが同じエピソードを複数回再生した場合も毎回カウントする
- 公開中（`publishedAt <= NOW()`）のエピソードのみカウント対象

### 入力形式

エピソード作成時の入力は以下のいずれか：

| 入力形式 | 説明 |
|----------|------|
| テキスト入力 | テーマやシナリオを入力して LLM が台本生成。URL が含まれていれば RAG で内容を取得 |
| ファイル入力 | 台本フォーマットのファイルをインポート |
| 音声ファイルアップロード | 録音・編集済みの音声ファイルを直接アップロード（voiceAudio / fullAudio に同一ファイルを設定） |

### BGM

- bgmId（ユーザー BGM）と systemBgmId（システム BGM）は排他的（同時に設定不可）
- BGM は音声生成時に指定する（`type=full` / `type=remix`）
- 最後に使用した BGM をエピソードに記録（次回生成時の復元表示用）
- 音声ファイル形式: mp3

---

## ScriptLine（台本行）

台本の各行（セリフ）。

| 属性 | 型 | 必須 | 説明 |
|------|-----|:----:|------|
| id | UUID | ◯ | 識別子（= lineId）。並び替え・編集後も追跡可能 |
| episodeId | UUID | ◯ | 所属する Episode |
| lineOrder | Int | ◯ | 行の順序（0 始まり） |
| speaker | Character | ◯ | 話者 |
| text | String | ◯ | セリフ（空禁止） |
| emotion | String | | 感情・喋り方。例: excited, laughing, curious |

### emotion の反映

emotion は TTS 生成時に text の先頭にカッコで付けて表現する：

```
[laughing] こんにちは。
[empathetic] さようなら...
```

### 制約

- speaker は同じ Channel に属する Character のみ指定可能
- lineOrder は Episode 内で一意

### 台本の編集操作

- 行の追加
- 行の全削除（エピソード内の全行を削除）
- 行の削除
- 行の並び替え
- 行の更新

---

## 台本テキストフォーマット

台本の取り込み・出力時に使用するテキスト形式。

### 基本ルール

- 区切りは半角 `:` のみ
- 1 行 = 1 セリフ
- 複数行のセリフは同キャラを連続して記述

### 構文

```
<speakerName>: <text>
```

### 例

```
太郎: こんにちは
花子: やあ、元気？
太郎: うん、元気だよ
花子: それは良かった
```

---

## 取り込みバリデーション

### 入力

- 台本テキスト
- allowedSpeakers（Channel の登場人物名リスト）

### 出力

- 成功: ScriptLine 配列
- 失敗: エラー（行番号 + 理由）

### バリデーションルール

| 条件 | 結果 |
|------|------|
| `:` が無い行 | エラー |
| lhs / rhs を trim 後に処理 | - |
| lhs が allowedSpeakers に存在しない | エラー |
| rhs（セリフ）が空 | エラー |

---

## 音声生成

エピソード単位で完成形の音声を生成する。TTS API の制約により、行単位ではなくエピソード全体を一括で生成する。

### ユースケース

音声生成は統合エンドポイント（`generate-async`）の `type` パラメータで 3 種類を切り替える。

| type | 説明 | TTS | BGM |
|------|------|:---:|:---:|
| `voice` | ボイス音声のみ生成（BGM なし） | ◯ | - |
| `full` | TTS 音声生成 + BGM ミキシング | ◯ | ◯ |
| `remix` | 既存ボイス音声を再利用して BGM のみ差し替え | - | ◯ |

### 生成フロー（type=voice / type=full）

1. 台本の全行を取得
2. 各行のテキストを結合（emotion は `[感情]` 形式で先頭に付与）
3. TTS API で音声を一括生成
4. ボイス音声を `Episode.voiceAudio` に保存
5. `type=full` の場合、リクエストで指定された BGM とミックス
6. 結果を `Episode.fullAudio` に保存
7. `type=full` / `type=remix` の場合、`Episode.bgmId` / `systemBgmId` を更新

### 生成フロー（type=remix）

1. `Episode.voiceAudio` が存在することを確認
2. GCS から `voiceAudio` をダウンロード
3. リクエストで指定された BGM とミキシング
4. 結果を `Episode.fullAudio` に保存
5. `Episode.bgmId` / `systemBgmId` を更新

### 出力形式

- 形式: mp3
- 保存先: `Episode.fullAudio`

### 再生成

- 台本を編集した場合は音声の再生成が必要
- 再生成時は既存の `fullAudio` を削除して新規生成

---

## ScriptJob（台本生成ジョブ）

台本の非同期生成ジョブを管理する。

| 属性 | 型 | 必須 | 説明 |
|------|-----|:----:|------|
| id | UUID | ◯ | 識別子 |
| episodeId | UUID | ◯ | 対象エピソード |
| userId | UUID | ◯ | ジョブ作成者 |
| status | ScriptJobStatus | ◯ | ステータス |
| progress | Int | ◯ | 進捗（0-100） |
| prompt | String | ◯ | 台本のテーマ・内容の指示 |
| durationMinutes | Int | ◯ | エピソードの長さ（分） |
| withEmotion | Boolean | ◯ | 感情タグを付与するか |
| errorMessage | String | | エラーメッセージ |
| errorCode | String | | エラーコード |
| startedAt | DateTime | | 処理開始日時 |
| completedAt | DateTime | | 処理完了日時 |

### ステータス遷移

```
pending → processing → completed / failed
pending → canceled
processing → canceling → canceled / failed
```

---

## AudioJob（音声生成ジョブ）

音声の非同期生成ジョブを管理する。

| 属性 | 型 | 必須 | 説明 |
|------|-----|:----:|------|
| id | UUID | ◯ | 識別子 |
| episodeId | UUID | ◯ | 対象エピソード |
| userId | UUID | ◯ | ジョブ作成者 |
| status | AudioJobStatus | ◯ | ステータス |
| jobType | AudioJobType | ◯ | ジョブ種別（voice / full / remix） |
| progress | Int | ◯ | 進捗（0-100） |
| bgmId | UUID | | ユーザー BGM |
| systemBgmId | UUID | | システム BGM |
| bgmVolumeDb | Decimal | ◯ | BGM 音量（dB）、デフォルト: -20.0 |
| fadeOutMs | Int | ◯ | フェードアウト時間（ms）、デフォルト: 3000 |
| paddingStartMs | Int | ◯ | 音声開始前の余白（ms）、デフォルト: 1000 |
| paddingEndMs | Int | ◯ | 音声終了後の余白（ms）、デフォルト: 3000 |
| resultAudioId | UUID | | 生成された音声 |
| errorMessage | String | | エラーメッセージ |
| errorCode | String | | エラーコード |
| startedAt | DateTime | | 処理開始日時 |
| completedAt | DateTime | | 処理完了日時 |

### ステータス遷移

```
pending → processing → completed / failed
pending → canceled
processing → canceling → canceled / failed
```
