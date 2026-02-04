package service

import (
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
	phase3Config = PhaseConfig{Provider: llm.ProviderOpenAI, Temperature: 0.7}
	phase4Config = PhaseConfig{Provider: llm.ProviderOpenAI, Temperature: 0.5}
)

// PhaseConfigs は全 Phase の設定を返す（起動時バリデーション用）
func PhaseConfigs() []PhaseConfig {
	return []PhaseConfig{phase2Config, phase3Config, phase4Config}
}

// Phase 2: 素材+アウトライン生成のシステムプロンプト
const phase2SystemPrompt = `あなたはポッドキャスト台本の構成作家です。
与えられたテーマとチャンネル情報をもとに、内容の濃い台本を作るための
「素材」と「アウトライン」を JSON 形式で出力してください。

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

## 制約
- 素材は各カテゴリ最低3個ずつ用意する
- アウトラインは必ず3ブロック構成とする
- 各ブロックに example / pitfall / action_step / question を最低1つずつ含める
- JSON 以外のテキストは出力しない`

// Phase 4: QA パッチ修正のシステムプロンプト
const phase4SystemPrompt = `あなたはポッドキャスト台本の品質管理担当です。
以下の台本に対して、指摘された問題箇所のみを最小限に修正してください。

## 修正ルール
- 指摘された箇所の周辺のみを修正する
- 修正箇所以外のセリフは一切変更しない
- 具体例・落とし穴・実務の一歩は削除しない
- 修正後も全体の流れが自然になるよう配慮する
- 出力は台本全文（修正箇所を含む完全な台本）とする

## 出力形式
話者名: セリフ
（元の台本と同じ形式で全文を出力）`

// getPhase3SystemPrompt は Phase 3 用のシステムプロンプトを返す
//
// talkMode と withEmotion の組み合わせで4パターンのプロンプトを生成
func getPhase3SystemPrompt(talkMode script.TalkMode, withEmotion bool) string {
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
	sb.WriteString("- 各ブロックで以下を台詞として必ず含める:\n")
	sb.WriteString("  - 具体例（状況 + 数字 or 具体物）\n")
	sb.WriteString("  - 落とし穴・よくある誤解\n")
	sb.WriteString("  - 実務の一歩（リスナーが実践できること）\n")

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

	// 台詞ルール
	sb.WriteString("\n## 台詞ルール\n")
	sb.WriteString("- 1つのセリフは20〜80文字程度\n")
	sb.WriteString("- セリフは文として完結させる（体言止めや中途半端な切れ方にしない）\n")
	sb.WriteString("- セリフの末尾に句点（。）は付けない\n")

	if talkMode == script.TalkModeDialogue {
		sb.WriteString("- セリフ中に句点を入れない（1行に1文。同一話者の連続発言は複数行に分ける）\n")
	} else {
		sb.WriteString("- セリフ中に句点を入れない（1行に1文。続けて話す場合は複数行に分ける）\n")
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
	sb.WriteString("\n## 分量\n")
	sb.WriteString("- 1分あたり約300文字を目安に、指定されたエピソード長に合わせる\n")

	// 出力形式
	sb.WriteString("\n## 出力形式\n")
	sb.WriteString("話者名: セリフ\n\n")
	sb.WriteString("- 1行につき1つのセリフ\n")
	sb.WriteString("- 空行は入れない\n")
	sb.WriteString("- 台本テキスト以外の説明文・コメント・見出し・メタ発言は出力しない\n")

	if talkMode == script.TalkModeDialogue {
		sb.WriteString("\n例：\n")
		sb.WriteString("太郎: こんにちは、今日もよろしくお願いします\n")
		sb.WriteString("太郎: 今日はいい天気だから気分がいいね\n")
		sb.WriteString("花子: やあ、元気そうで何よりだね\n")
	} else {
		sb.WriteString("\n例：\n")
		sb.WriteString("太郎: こんにちは、今日もよろしくお願いします\n")
		sb.WriteString("太郎: 今日はちょっと面白いテーマを持ってきました\n")
	}

	// 感情あり版
	if withEmotion {
		sb.WriteString("\n## 出力形式（感情あり版）\n")
		sb.WriteString("話者名: [感情] セリフ\n\n")
		sb.WriteString("- 感情は省略可能。指定する場合は [感情] の形式でセリフの前に記載\n")

		if talkMode == script.TalkModeDialogue {
			sb.WriteString("\n例：\n")
			sb.WriteString("太郎: こんにちは、今日もよろしくお願いします\n")
			sb.WriteString("花子: [嬉しそうに] やあ、元気そうで何よりだね\n")
		} else {
			sb.WriteString("\n例：\n")
			sb.WriteString("太郎: こんにちは、今日もよろしくお願いします\n")
			sb.WriteString("太郎: [ワクワクした様子で] 今日はちょっと面白いテーマを持ってきました\n")
		}
	}

	// 制約
	sb.WriteString("\n## 制約\n")
	sb.WriteString("- アウトラインの素材（具体例・落とし穴・実務の一歩）は必ず台詞に組み込む。省略・要約しない\n")
	sb.WriteString("- 制作側のメタ発言はしない")

	return sb.String()
}
