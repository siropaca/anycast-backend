# Gemini TTS マークアップタグ仕様

Gemini TTS（gemini-2.5-pro-tts）は SSML ではなく、角括弧形式のマークアップタグを使用します。

## 非音声音 (Non-speech sounds)

| タグ | 効果 |
|------|------|
| `[sigh]` | ため息 |
| `[laughing]` | 笑い |
| `[uhm]` | 躊躇音 |

## スタイル修飾子 (Style modifiers)

| タグ | 効果 |
|------|------|
| `[sarcasm]` | 皮肉なトーン |
| `[robotic]` | ロボット的な声 |
| `[shouting]` | 大声 |
| `[whispering]` | ささやく |
| `[extremely fast]` | 非常に速い |

## 発声されるマークアップ (Spoken markup)

| タグ | 効果 |
|------|------|
| `[scared]` | 怖い |
| `[curious]` | 好奇心的 |
| `[bored]` | つまらなそう |

## ペースと休止 (Pacing and pauses)

| タグ | 長さ |
|------|------|
| `[short pause]` | 約250ms（コンマ程度） |
| `[medium pause]` | 約500ms（文の区切り程度） |
| `[long pause]` | 1000ms以上（ドラマティック効果用） |

標準的な句読点（コンマ、ピリオド、セミコロン）も自然なポーズを生成します。

## 感情タグ

上記以外にも、自然言語で感情を指定できます：

- `[excited]` - 興奮した
- `[sad]` - 悲しい
- `[angry]` - 怒った
- `[笑いながら]` - 日本語での指定も可能

## 使用例

```
[whispering] これは秘密だよ [long pause] 誰にも言わないでね
[laughing] それは面白いね！
[scared] 何か聞こえた...
```

## 注意事項

- SSML（`<break time="1s"/>` など）はサポートされていません
- タグは自然な位置に配置しないと無視される場合があります
- 複数のポーズタグを組み合わせて長い休止を作ることができます

## ElevenLabs との互換性

ElevenLabs の Text-to-Dialogue API も同じ角括弧形式（`[cheerfully]`, `[laughing]` 等）をサポートしている。そのため、台本生成で付与された感情タグはプロバイダを問わずそのまま使用できる。

## 参考

- [Gemini-TTS | Cloud Text-to-Speech | Google Cloud Documentation](https://docs.cloud.google.com/text-to-speech/docs/gemini-tts)
- [ElevenLabs Text to Dialogue](https://elevenlabs.io/docs/eleven-api/guides/cookbooks/text-to-dialogue)
