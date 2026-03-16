# 音声生成パイプライン

エピソードの台本から完成形の音声ファイルを生成するパイプラインの詳細設計。  
API 仕様・ジョブ管理・WebSocket については [audio-generate-async-api.md](audio-generate-async-api.md) を参照。

## 概要

音声生成は `type` パラメータで 3 種類のフローに分岐する。

| type | 説明 | TTS | BGM |
|------|------|:---:|:---:|
| `voice` | ボイス音声のみ生成 | ◯ | - |
| `full` | TTS + BGM ミキシング | ◯ | ◯ |
| `remix` | 既存ボイスに BGM を差し替え | - | ◯ |

`voice` / `full` は台本から TTS で音声を合成する。
話者が複数いる場合は **マルチスピーカー再アセンブル方式** で合成する。
これが本パイプラインのコア技術であり、以下のセクションで詳細を説明する。

---

## 全体フロー

```
台本（ScriptLine[]）
  │
  ├─ 話者が 1 人 → シングルスピーカー合成
  │                  全テキストを連結して一括 TTS
  │
  └─ 話者が 2 人以上 → マルチスピーカー再アセンブル
                         話者別に並列 TTS → STT で行分割 → 元の順序に再結合
  │
  ▼
ボイス音声（PCM → MP3）
  │
  ├─ type=voice → ボイス音声を保存して完了
  │
  └─ type=full → BGM ミキシング → 最終音声を保存
```

---

## マルチスピーカー再アセンブル

Gemini TTS のマルチスピーカーモードは長い台本で不安定になるため、話者ごとにシングルスピーカーで合成し、行単位に分割して元の台本順に組み立て直す方式を採用している。
技術選定の背景は [ADR-019](../adr/019-stt-timestamp-audio-segmentation.md) を参照。

### 処理フロー

```
Phase 1: 話者グループ化
  台本行を話者ごとにグループ化し、各グループの末尾にダミー行（"以上です。"）を追加
  ↓
Phase 2: 話者別 TTS 合成（並列）
  各話者のテキストを連結してシングルスピーカー TTS で一括合成
  出力: 話者ごとの PCM 音声
  ↓
Phase 3: STT アライメント + セグメント分割（並列）
  各話者の音声について:
    a. STT で単語レベルタイムスタンプを取得
    b. DP アライメントで行境界を特定
    c. silencedetect で無音区間を検出
    d. 行境界を最寄りの無音区間にスナップ
    e. タイムスタンプで PCM を行単位に分割
  出力: 行単位の PCM セグメント配列
  ↓
Phase 4: 再アセンブル
  全セグメントを元の台本順にソートし、セグメント間に 200ms 無音を挿入して連結
  出力: 完成形の PCM 音声
```

### Phase 1: 話者グループ化

台本行を話者でグループ化する。
各グループの末尾に **ダミー行**（`"以上です。"`）を追加する。

ダミー行の目的:
- TTS は末尾のテキストを早口で読み上げたり省略する傾向がある
- ダミー行を犠牲にすることで、実際の最終セリフの品質を保つ

### Phase 2: 話者別 TTS 合成（並列実行）

各話者について `errgroup` で並列に TTS を実行する。

- 話者のテキストを改行で連結
- Gemini TTS の場合、`DefaultVoiceStyle` プロンプトを先頭に追加
- TTS API で一括合成（リトライ最大 2 回）
- 出力: PCM（24kHz, mono, s16le）

### Phase 3: STT アライメント + セグメント分割（並列実行）

各話者の PCM 音声を台本行単位に分割する。
この Phase もまた `errgroup` で話者ごとに並列実行する。

#### Step a: STT タイムスタンプ取得

Google Cloud Speech-to-Text v2 で音声を文字起こしし、**単語レベルのタイムスタンプ**を取得する。

```
PCM 音声 → STT API → [{Word: "こんにちは", Start: 0.2s, End: 0.8s}, ...]
```

- API 制限（60 秒）に対応するため、55 秒チャンク + 5 秒オーバーラップで分割送信
- チャンク境界の単語重複はオーバーラップ中間点で切り替えて解消

#### Step b: DP アライメントで行境界を特定

元の台本テキストと STT の文字起こしテキストを **Needleman-Wunsch アルゴリズム**（文字レベル DP）で最適対応させ、各行の終了位置に対応する STT 単語のタイムスタンプから行境界を特定する。

```
元テキスト:  "こんにちはお元気ですか" | "今日はいい天気ですね"
STT テキスト: "こんにちはお元気ですか今日はいい天気ですね"
                        ↑ 行境界                ↑ 行境界
```

- 正規化: 句読点・記号・空白を除去してからアライメント
- TTS がテキストをスキップ・追加した場合も gap として局所的に吸収
- 乖離率が 50% を超える場合はエラー

#### Step c: 無音区間の検出

FFmpeg `silencedetect` で音声中の無音区間を検出する。

```
ffmpeg -f s16le -ar 24000 -ac 1 -i input.pcm -af silencedetect=noise=-30dB:d=0.2 -f null -
```

| パラメータ | 値 | 説明 |
|-----------|-----|------|
| noise | -30 dB | 無音判定しきい値 |
| d | 0.2 秒 | 最小無音継続時間 |

#### Step d: 行境界を無音区間にスナップ

STT で特定した行境界を、最寄りの無音区間の**中間点**にスナップする。
これにより、発話途中ではなく自然な無音部分でカットできる。

```
STT 境界 ──── maxSnapDistance (500ms) 以内 ───→ 無音区間の中間点
```

- 距離判定は無音区間の**最寄りの端**で計算（区間内なら距離 0）
- 複数の無音に該当する場合は**最長の無音**を選択
- 同じ無音に複数カットがスナップしないよう追跡
- 適切な無音区間がなければ STT 境界をそのまま使用（フォールバック）

#### Step e: PCM を行単位に分割

スナップ後の行境界タイムスタンプで PCM バイト列を分割する。

- 時刻 → バイトオフセット変換: `offset = time × sampleRate × channels × bytesPerSample`
- ブロックアライメント（2 バイト）に調整
- ダミー行のセグメントは除外

### Phase 4: 再アセンブル

全話者のセグメントを元の台本順（`originalIndex`）でソートし、連結する。
セグメント間に **200ms の無音パディング**を挿入して自然な間を確保する。

---

## フォーマット変換

TTS プロバイダによって出力形式が異なる。

| プロバイダ | TTS 出力 | 変換 |
|-----------|----------|------|
| Gemini TTS | PCM（24kHz, mono, s16le） | FFmpeg で MP3 に変換 |
| ElevenLabs | MP3（44.1kHz, 128kbps） | 変換不要 |

MP3 変換コマンド:
```
ffmpeg -f s16le -ar 24000 -ac 1 -i input.pcm -c:a libmp3lame -b:a 192k output.mp3
```

---

## BGM ミキシング

`type=full` / `type=remix` の場合、ボイス音声と BGM を FFmpeg でミキシングする。

### パラメータ

| パラメータ | デフォルト | 範囲 | 説明 |
|-----------|-----------|------|------|
| bgmVolumeDb | -25 dB | -60 〜 0 | BGM の音量 |
| fadeOutMs | 3000 ms | 0 〜 30000 | BGM のフェードアウト時間 |
| paddingStartMs | 1000 ms | 0 〜 10000 | ボイス開始前の BGM のみ区間 |
| paddingEndMs | 3000 ms | 0 〜 10000 | ボイス終了後の BGM フェードアウト区間 |

### FFmpeg フィルタグラフ

```
[BGM]  → aloop（無限ループ）→ volume（音量調整）→ afade（フェードアウト）→ atrim（長さ調整）→ [bgm]
[Voice] → adelay（開始余白）→ [voice]
[bgm][voice] → amix（ミックス）→ [out]
```

出力時間 = paddingStartMs + voiceDurationMs + paddingEndMs

### 出力

- 形式: MP3（192kbps, libmp3lame）
- 保存先: GCS `audios/{audioId}.mp3`

---

## TTS プロバイダ

キャラクターの Voice に紐づく Provider から動的にプロバイダを選択する。

### Gemini TTS

| 設定 | 値 |
|------|------|
| モデル | gemini-2.5-pro-tts |
| 言語 | ja-JP |
| 出力形式 | PCM 16bit 24kHz mono |

### ElevenLabs

| 設定 | 値 |
|------|------|
| シングルスピーカー API | `/v1/text-to-speech/{voice_id}` |
| マルチスピーカー API | `/v1/text-to-dialogue` |
| モデル | eleven_v3（マルチ）/ eleven_multilingual_v2（シングル） |
| 出力形式 | MP3 44.1kHz 128kbps |

---

## 関連ドキュメント

| ドキュメント | 説明 |
|-------------|------|
| [audio-generate-async-api.md](audio-generate-async-api.md) | API 仕様・ジョブ管理・WebSocket・進捗 |
| [ADR-019](../adr/019-stt-timestamp-audio-segmentation.md) | STT タイムスタンプ分割の技術選定・方式の発展経緯 |
| [system.md](system.md) | TTS タイムアウト等のシステム設定 |

## 関連ファイル

| ファイル | 説明 |
|---------|------|
| internal/service/audio_job.go | ジョブ実行・マルチスピーカー再アセンブル |
| internal/service/ffmpeg.go | FFmpeg ミキシング・変換処理 |
| internal/infrastructure/tts/gemini_client.go | Gemini TTS クライアント |
| internal/infrastructure/tts/elevenlabs_client.go | ElevenLabs TTS クライアント |
| internal/infrastructure/stt/client.go | Google Cloud STT クライアント |
| internal/pkg/audio/align.go | DP アライメント・境界スナップ |
| internal/pkg/audio/split.go | silencedetect 無音検出・PCM 分割 |
| internal/pkg/audio/concat.go | MP3 連結処理 |
| internal/pkg/audio/pcm.go | PCM 連結・無音生成 |
