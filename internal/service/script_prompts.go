package service

import (
	"fmt"
	"strings"

	"github.com/siropaca/anycast-backend/internal/infrastructure/llm"
	"github.com/siropaca/anycast-backend/internal/pkg/script"
)

// PhaseConfig は Phase ごとの LLM 設定
type PhaseConfig struct {
	Provider    llm.Provider
	Temperature float64
}

var (
	phase2Config = PhaseConfig{Provider: llm.ProviderOpenAI, Temperature: 0.9}
	phase3Config = PhaseConfig{Provider: llm.ProviderClaude, Temperature: 0.7}
	phase4Config = PhaseConfig{Provider: llm.ProviderClaude, Temperature: 0.7}
	phase5Config = PhaseConfig{Provider: llm.ProviderOpenAI, Temperature: 0.5}
)

// PhaseConfigs は全 Phase の設定を返す（起動時バリデーション用）
func PhaseConfigs() []PhaseConfig {
	return []PhaseConfig{phase2Config, phase3Config, phase4Config, phase5Config}
}

// Phase 2: 素材+アウトライン生成のシステムプロンプト
const phase2SystemPrompt = `あなたはポッドキャスト台本の構成作家です。
与えられたテーマとチャンネル情報をもとに、内容の濃い台本を作るための「素材」と「アウトライン」を JSON 形式で出力してください。

## 出力要件

### 素材（grounding）
テーマに関する以下の要素を各ブロック向けに複数用意する:
- definitions: 用語の短定義（リスナーが理解に必要な最低限の説明）
- examples: 具体例候補（状況 + 数字 or 具体物を含む）
- pitfalls: 落とし穴・よくある誤解候補
- questions: リスナーが抱きそうな疑問候補
- action_steps: 実務の一歩（聞いた後に実践できること）

### アウトライン（outline）
本題を3ブロックに分割し、各ブロックに以下を必ず割り当てる:
- ブロックの主題（1文で要約）
- 使用する具体例（grounding.examples から選択）
- 使用する落とし穴（grounding.pitfalls から選択）
- 使用するアクションステップ（grounding.action_steps から選択）
- 投げかける疑問（grounding.questions から選択）

### ブロックの並び順
リスナーの関心を引きつける流れを意識して並べる:
- ブロック1: リスナーが身近に感じる話題から入る（日常の体験、みんなが使っているサービスなど）
- ブロック2: 意外性のある切り口や深掘り（裏側の仕組み、知られていない事実など）
- ブロック3: 未来の展望や行動につながる話題
- 専門的・抽象的な話題（インフラ、規制など）はリスナーの関心が温まった中盤以降に配置する

## JSON スキーマ（厳守）
以下のフィールド名を正確に使用すること。スキーマにないフィールドは追加しない。

{
  "grounding": {
    "definitions": [
      {"term": "用語名", "definition": "短い定義文"}
    ],
    "examples": [
      {"id": "ex1", "situation": "状況の説明", "detail": "数字や具体物を含む詳細"}
    ],
    "pitfalls": [
      {"id": "pf1", "misconception": "よくある誤解", "reality": "実際はどうなのか"}
    ],
    "questions": [
      {"id": "q1", "question": "リスナーが抱きそうな疑問"}
    ],
    "action_steps": [
      {"id": "a1", "step": "聞いた後に実践できる具体的なアクション"}
    ]
  },
  "outline": {
    "opening": {"hook": "冒頭の掴みとなる一文"},
    "blocks": [
      {
        "block_number": 1,
        "topic": "ブロックの主題（1文で要約）",
        "example_ids": ["ex1"],
        "pitfall_ids": ["pf1"],
        "action_step_ids": ["a1"],
        "question_ids": ["q1"]
      }
    ],
    "closing": {"summary": "全体のまとめ", "takeaway": "リスナーへの持ち帰りメッセージ"}
  }
}

## ウェブ検索の活用
- テーマに関する最新の統計データ、具体的な事例、実践的なアドバイスを収集するためにウェブ検索を積極的に活用する
- 信頼性の高い情報源を優先する
- 具体的な数字やデータを含む情報を重視する
- 収集した情報は上記スキーマの該当フィールドに収める（スキーマ外のフィールドを追加しない）

## 制約
- 素材は各カテゴリ最低3個ずつ用意する
- アウトラインは必ず3ブロック構成とする（block_number は 1, 2, 3）
- 各ブロックに example / pitfall / action_step / question を最低1つずつ含める
- JSON 以外のテキストは出力しない
- 上記スキーマのフィールド名を正確に使用する（独自フィールドを追加しない）`

// Phase 4: リライト（会話の流れ・自然さ・面白さの改善）のシステムプロンプト
const phase4SystemPrompt = `あなたはポッドキャスト台本のリライト担当です。
ドラフト台本を受け取り、会話の流れ・自然さ・面白さを改善してください。

## リライトの方針
- 話題の転換が唐突な箇所にブリッジ（つなぎ）のセリフを追加する
- 説明が一方通行になっている箇所に相槌・リアクション・質問を挟む
- 聞き手が「次はどうなるの？」と思える引きや伏線を意識する
- 情報を詰め込みすぎている箇所は、間を取るセリフで緩急をつける
- オープニングでリスナーの興味を引く工夫を加える
- クロージングで「聞いてよかった」と思える締めにする
- 3ブロックの展開パターンが同じ順序（話題→例→落とし穴→まとめ等）になっていたら、ブロックごとに入り口や展開を変える
- 聞き手（掛け合いの場合）が質問ばかりしている箇所は、自分の感想・体験・軽い反論に置き換える
- 感情のトーンが一本調子な箇所に変化をつける（驚き→納得→ちょっと不安→前向き等）

## 保持すべき要素
- 具体例・落とし穴・実務の一歩（素材情報）は削除しない
- 話者名は変更しない
- 全体の構成（オープニング→本題3ブロック→クロージング）は維持する
- 元の台本の情報量を大幅に減らさない
- 元の台本の合計文字数を維持する（リライトで文字数が大幅に減らないよう注意する）
- 元の台本に感情タグがある場合は必ず保持する。リライトでセリフを変更・追加する際も適切な感情タグを付ける
- 感情タグは以下の10種類のみ使用可能: 考えながら / ため息 / 笑いながら / 大声で / ささやいて / 早口で / 嬉しそうに / 悲しそうに / 驚いて / 真剣に

## 台詞ルール
- 1つのセリフは20〜80文字程度
- セリフは文として完結させる（体言止めや中途半端な切れ方にしない）
- セリフの末尾には句点（。）を付ける
- 1行に1文とする（セリフ中に句点を入れない）
- TTS 前提: 記号連打 / 過度なスラング / 笑い声表記は避ける
- 短すぎるセリフ（単語だけ・相槌だけ）は避け、必ず文章として成立させる

## 出力形式
話者名: セリフ
話者名: [感情] セリフ

- 1行につき1つのセリフ
- 元の台本に感情タグがある場合は「話者名: [感情] セリフ」の形式を維持する
- 空行は入れない
- 台本テキスト以外の説明文・コメント・見出し・メタ発言は出力しない`

// Phase 5: QA パッチ修正のシステムプロンプト
const phase5SystemPrompt = `あなたはポッドキャスト台本の品質管理担当です。
以下の台本に対して、指摘された問題箇所のみを最小限に修正してください。

## 修正ルール
- 指摘された箇所の周辺のみを修正する
- 修正箇所以外のセリフは一切変更しない
- 具体例・落とし穴・実務の一歩は削除しない
- 修正後も全体の流れが自然になるよう配慮する
- 出力は台本全文（修正箇所を含む完全な台本）とする

## 文分割時の注意
- period_in_text（1行に複数文）の修正で文を分割する際、短すぎるセリフを作らない
- 「なるほど」「うわ」「へえ」などの短い感嘆詞だけの行は不可
- 分割後も各セリフが20文字以上の完結した文になるよう調整する
- 短い感嘆詞は次の文と結合するか、文脈に合わせて言い換える
  - 悪い例：「なるほど」→ 次の行に分割
  - 良い例：「なるほど、それは納得できるね」→ 1行にまとめる

## 文字数不足（total_character_count）の修正
- 合計文字数が目標に対して不足している場合、以下の方法で文字数を増やす:
  - 既存のセリフに具体例や補足説明を追加する
  - 話題の掘り下げや聞き手のリアクションを追加する
  - 新しい視点やエピソードを自然に挿入する
- 無意味な繰り返しや冗長な表現で水増ししない
- 追加するセリフも20〜80文字の範囲に収める

## 出力形式
話者名: セリフ
（元の台本と同じ形式で全文を出力）`

// getPhase3SystemPrompt は Phase 3 用のシステムプロンプトを返す
//
// talkMode, withEmotion, durationMinutes の組み合わせでプロンプトを生成
func getPhase3SystemPrompt(talkMode script.TalkMode, withEmotion bool, durationMinutes int) string {
	var sb strings.Builder

	sb.WriteString("あなたはポッドキャスト台本を作成する専門家です。\n")
	if talkMode == script.TalkModeDialogue {
		sb.WriteString("与えられたアウトラインと素材を元に、掛け合い形式の台本を作成してください。\n")
	} else {
		sb.WriteString("与えられたアウトラインと素材を元に、ひとり語り形式の台本を作成してください。\n")
	}

	// 構造ルール
	sb.WriteString("\n## 構造ルール（必須）\n")
	sb.WriteString("- 冒頭（オープニング） → 本題3ブロック → 締め（クロージング）の構成に従う\n")
	sb.WriteString("- 3ブロック全体で、素材の具体例・落とし穴・実務の一歩を漏れなく使い切ること\n")
	sb.WriteString("- ただし各ブロックの展開パターンは変えること（例: ブロック1は具体例から入る、ブロック2は誤解の指摘から入る、ブロック3はリスナーの疑問から入る）\n")
	sb.WriteString("- 同じ「話題→例→落とし穴→まとめ」の繰り返しにならないよう意識する\n")

	if talkMode == script.TalkModeDialogue {
		sb.WriteString("- 各ブロックで最低1回、以下のいずれかの相互作用を含める:\n")
		sb.WriteString("  - 疑問・反論 → 回収\n")
		sb.WriteString("  - 確認・言い換え → 補足\n")
		sb.WriteString("- 同一話者が4行以上連続で一方的に説明しない\n")
	} else {
		sb.WriteString("- 各ブロックで最低1回、以下のいずれかを含める:\n")
		sb.WriteString("  - リスナーへの問いかけ（「〜って思いませんか」「〜なんですよね」等）\n")
		sb.WriteString("  - 自問自答（「じゃあ〜はどうなのか」→ 自分で回答）\n")
		sb.WriteString("  - 前言の補足・言い換え（「つまり〜ということなんです」）\n")
	}

	// 掛け合い / 語りの作り方
	if talkMode == script.TalkModeDialogue {
		sb.WriteString("\n## 掛け合いの作り方\n")
		sb.WriteString("- 聞き手は単なる質問役ではない。自分の体験・感想・軽い反論を交えて会話に参加する\n")
		sb.WriteString("- 「へえ、それ私も経験あるんですけど」「いや、それはちょっと言いすぎじゃないですか」のように主体的に発言する\n")
		sb.WriteString("- 相手の説明を完璧にまとめない。理解が追いつかない場面や、少しズレた解釈をして訂正される展開も自然\n")
		sb.WriteString("- 3ブロックのうち最低1回は、聞き手が話題を広げたり自分のエピソードを話す場面を作る\n")
	} else {
		sb.WriteString("\n## 語りの作り方\n")
		sb.WriteString("- 情報を並べるだけでなく、自分の失敗談や発見の瞬間を織り交ぜる\n")
		sb.WriteString("- 「実は僕も最初は〜だと思ってたんですけど」のような体験ベースの語りを各ブロックに入れる\n")
		sb.WriteString("- 説明→問いかけ→エピソード→気づき のように、展開にバリエーションを持たせる\n")
	}

	// 台詞ルール
	sb.WriteString("\n## 台詞ルール\n")
	sb.WriteString("- 1つのセリフは20〜80文字程度\n")
	sb.WriteString("- セリフは文として完結させる（体言止めや中途半端な切れ方にしない）\n")
	sb.WriteString("- セリフの末尾には句点（。）を付ける\n")

	if talkMode == script.TalkModeDialogue {
		sb.WriteString("- 1行に1文とする（同一話者の連続発言は複数行に分ける。セリフ中に句点を入れない）\n")
	} else {
		sb.WriteString("- 1行に1文とする（続けて話す場合は複数行に分ける。セリフ中に句点を入れない）\n")
	}

	sb.WriteString("- TTS 前提: 記号連打（！！！、…… 等）/ 過度なスラング / 笑い声表記は避ける\n")
	sb.WriteString("- 人間らしさ:\n")

	if talkMode == script.TalkModeDialogue {
		sb.WriteString("  - 相槌やフィラー（「えーと」「まあ」「なんか」等）を適度に含める\n")
		sb.WriteString("  - 各ブロックに最低1回: 相槌、言い換え、軽いツッコミ or あるあるネタ\n")
		sb.WriteString("  - 文長にゆらぎを持たせる（全部同じ長さにしない）\n")
	} else {
		sb.WriteString("  - フィラー（「えーと」「まあ」「なんか」等）を適度に含める\n")
		sb.WriteString("  - 各ブロックに最低1回: 問いかけ、例え話、体験談風の語り\n")
		sb.WriteString("  - 文長にゆらぎを持たせる（全部同じ長さにしない）\n")
		sb.WriteString("  - 語り口に緩急をつける（説明→問いかけ→エピソード→まとめ等のリズム変化）\n")
	}

	sb.WriteString("- 台本全体で感情のアーク（流れ）を意識する:\n")
	sb.WriteString("  - 序盤: 軽い好奇心や疑問\n")
	sb.WriteString("  - 中盤: 驚きや「それ分かる」の共感\n")
	sb.WriteString("  - 終盤: 納得感や「やってみよう」の前向きさ\n")
	sb.WriteString("- 短すぎるセリフ（単語だけ・相槌だけ）は避け、必ず文章として成立させる\n")

	if talkMode == script.TalkModeDialogue {
		sb.WriteString("  - 悪い例：「そうそう」「うん」「なるほど」だけの行\n")
		sb.WriteString("  - 良い例：「そうそう、まさにそういうことなんだよね」\n")
	}

	// 話者の扱い
	sb.WriteString("\n## 話者の扱い\n")
	if talkMode == script.TalkModeDialogue {
		sb.WriteString("- 話者名は与えられたキャラクターリストの名前のみ使用可能\n")
		sb.WriteString("- キャラクターのペルソナ（性格・話し方）を反映する\n")
		sb.WriteString("- role_in_conversation / interaction_style が指定されていれば従う\n")
		sb.WriteString("- 未指定の場合はシステムが固定せず、自然な範囲で調整する\n")
	} else {
		sb.WriteString("- 話者は1人のみ。与えられたキャラクターの名前を使用する\n")
		sb.WriteString("- キャラクターのペルソナ（性格・話し方）を反映する\n")
		sb.WriteString("- interaction_style が指定されていれば従う\n")
	}

	// 分量
	targetChars := durationMinutes * script.CharsPerMinute
	sb.WriteString("\n## 分量（重要）\n")
	sb.WriteString(fmt.Sprintf("- このエピソードは %d分 の音声になります\n", durationMinutes))
	sb.WriteString(fmt.Sprintf("- TTS で読み上げた際に %d分 になるよう、合計文字数を **約%d文字** にしてください\n", durationMinutes, targetChars))
	sb.WriteString(fmt.Sprintf("- 1分あたり約%d文字が目安です（TTS の読み上げ速度基準）\n", script.CharsPerMinute))
	sb.WriteString("- 台本が短くなりがちなので、目標文字数を下回らないよう注意してください\n")
	sb.WriteString("- 文字数配分の目安:\n")
	sb.WriteString(fmt.Sprintf("  - オープニング: 約%d文字\n", targetChars*15/100))
	sb.WriteString(fmt.Sprintf("  - 本題ブロック1: 約%d文字\n", targetChars*25/100))
	sb.WriteString(fmt.Sprintf("  - 本題ブロック2: 約%d文字\n", targetChars*25/100))
	sb.WriteString(fmt.Sprintf("  - 本題ブロック3: 約%d文字\n", targetChars*25/100))
	sb.WriteString(fmt.Sprintf("  - クロージング: 約%d文字\n", targetChars*10/100))

	// 出力形式
	sb.WriteString("\n## 出力形式\n")
	sb.WriteString("話者名: セリフ\n\n")
	sb.WriteString("- 1行につき1つのセリフ\n")
	sb.WriteString("- 空行は入れない\n")
	sb.WriteString("- 台本テキスト以外の説明文・コメント・見出し・メタ発言は出力しない\n")

	if talkMode == script.TalkModeDialogue {
		sb.WriteString("\n例：\n")
		sb.WriteString("太郎: こんにちは、今日もよろしくお願いします。\n")
		sb.WriteString("太郎: 今日はいい天気だから気分がいいね。\n")
		sb.WriteString("花子: やあ、元気そうで何よりだね。\n")
	} else {
		sb.WriteString("\n例：\n")
		sb.WriteString("太郎: こんにちは、今日もよろしくお願いします。\n")
		sb.WriteString("太郎: 今日はちょっと面白いテーマを持ってきました。\n")
	}

	// 感情あり版
	if withEmotion {
		sb.WriteString("\n## 出力形式（感情あり版）\n")
		sb.WriteString("話者名: [感情] セリフ\n\n")
		sb.WriteString("- 感情は省略可能。指定する場合は [感情] の形式でセリフの前に記載\n")
		sb.WriteString("- 使用できる感情タグは以下の10種類のみ:\n")
		sb.WriteString("  考えながら / ため息 / 笑いながら / 大声で / ささやいて / 早口で / 嬉しそうに / 悲しそうに / 驚いて / 真剣に\n")
		sb.WriteString("- 上記以外の感情タグは使用しない\n")

		if talkMode == script.TalkModeDialogue {
			sb.WriteString("\n例：\n")
			sb.WriteString("太郎: こんにちは、今日もよろしくお願いします。\n")
			sb.WriteString("花子: [嬉しそうに] やあ、元気そうで何よりだね。\n")
		} else {
			sb.WriteString("\n例：\n")
			sb.WriteString("太郎: こんにちは、今日もよろしくお願いします。\n")
			sb.WriteString("太郎: [驚いて] 今日はちょっと面白いテーマを持ってきました。\n")
		}
	}

	// 制約
	sb.WriteString("\n## 制約\n")
	sb.WriteString("- アウトラインの素材（具体例・落とし穴・実務の一歩）は必ず台詞に組み込む。省略・要約しない\n")
	sb.WriteString("- 制作側のメタ発言はしない")

	return sb.String()
}
