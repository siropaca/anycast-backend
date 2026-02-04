package script

import (
	"encoding/json"
	"fmt"
)

// Phase2Output は Phase 2（素材+アウトライン生成）の出力
type Phase2Output struct {
	Grounding Grounding `json:"grounding"`
	Outline   Outline   `json:"outline"`
}

// Grounding は素材情報
type Grounding struct {
	Definitions []Definition `json:"definitions"`
	Examples    []Example    `json:"examples"`
	Pitfalls    []Pitfall    `json:"pitfalls"`
	Questions   []Question   `json:"questions"`
	ActionSteps []ActionStep `json:"action_steps"`
}

// Definition は用語の短定義
type Definition struct {
	Term       string `json:"term"`
	Definition string `json:"definition"`
}

// Example は具体例候補
type Example struct {
	ID        string `json:"id"`
	Situation string `json:"situation"`
	Detail    string `json:"detail"`
}

// Pitfall は落とし穴・よくある誤解候補
type Pitfall struct {
	ID            string `json:"id"`
	Misconception string `json:"misconception"`
	Reality       string `json:"reality"`
}

// Question はリスナーが抱きそうな疑問候補
type Question struct {
	ID       string `json:"id"`
	Question string `json:"question"`
}

// ActionStep は実務の一歩
type ActionStep struct {
	ID   string `json:"id"`
	Step string `json:"step"`
}

// Outline はアウトライン情報
type Outline struct {
	Opening Opening        `json:"opening"`
	Blocks  []OutlineBlock `json:"blocks"`
	Closing Closing        `json:"closing"`
}

// UnmarshalJSON は LLM が outline を配列で返した場合にも対応する
//
// LLM が {"outline": [...]} のように blocks 配列を直接返すケースに対応。
// 通常の {"outline": {"opening":..., "blocks":..., "closing":...}} もそのまま処理。
func (o *Outline) UnmarshalJSON(data []byte) error {
	// 配列の場合: blocks のみの配列として扱う
	var blocks []OutlineBlock
	if err := json.Unmarshal(data, &blocks); err == nil {
		o.Blocks = blocks
		return nil
	}

	// オブジェクトの場合: 通常のパース
	type outlineAlias Outline
	var alias outlineAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*o = Outline(alias)
	return nil
}

// Opening は冒頭の掴み
type Opening struct {
	Hook string `json:"hook"`
}

// OutlineBlock はアウトラインの1ブロック
type OutlineBlock struct {
	BlockNumber   int      `json:"block_number"`
	Topic         string   `json:"topic"`
	ExampleIDs    []string `json:"example_ids"`
	PitfallIDs    []string `json:"pitfall_ids"`
	ActionStepIDs []string `json:"action_step_ids"`
	QuestionIDs   []string `json:"question_ids"`
}

// Closing はまとめ
type Closing struct {
	Summary  string `json:"summary"`
	Takeaway string `json:"takeaway"`
}

// ParsePhase2Output は LLM 出力テキストから Phase2Output をパースする
func ParsePhase2Output(text string) (*Phase2Output, error) {
	// ExtractJSON で JSON 部分を抽出し、Unmarshal + 基本バリデーション
	jsonStr, err := ExtractJSON(text)
	if err != nil {
		return nil, fmt.Errorf("Phase 2 出力から JSON を抽出できません: %w", err)
	}

	var output Phase2Output
	if err := json.Unmarshal([]byte(jsonStr), &output); err != nil {
		return nil, fmt.Errorf("Phase 2 出力の JSON パースに失敗: %w", err)
	}

	// 基本バリデーション
	if len(output.Outline.Blocks) != 3 {
		return nil, fmt.Errorf("アウトラインのブロック数が3ではありません: %d", len(output.Outline.Blocks))
	}

	if len(output.Grounding.Examples) == 0 {
		return nil, fmt.Errorf("素材の具体例が空です")
	}

	if len(output.Grounding.Pitfalls) == 0 {
		return nil, fmt.Errorf("素材の落とし穴が空です")
	}

	if len(output.Grounding.ActionSteps) == 0 {
		return nil, fmt.Errorf("素材のアクションステップが空です")
	}

	return &output, nil
}
