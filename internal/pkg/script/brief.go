package script

import (
	"encoding/json"
)

// TalkMode は会話形式の種別
type TalkMode string

const (
	TalkModeDialogue  TalkMode = "dialogue"
	TalkModeMonologue TalkMode = "monologue"
)

// Brief は正規化されたブリーフ情報
type Brief struct {
	Episode     BriefEpisode     `json:"episode"`
	Channel     BriefChannel     `json:"channel"`
	Characters  []BriefCharacter `json:"characters"`
	MasterGuide string           `json:"master_guide"`
	Theme       string           `json:"theme"`
	Constraints BriefConstraints `json:"constraints"`
}

// BriefEpisode はエピソード情報のスロット
type BriefEpisode struct {
	Title           string `json:"title"`
	Description     string `json:"description,omitempty"`
	DurationMinutes int    `json:"duration_minutes"`
	EpisodeNumber   int    `json:"episode_number"`
}

// BriefChannel はチャンネル情報のスロット
type BriefChannel struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category"`
	StyleGuide  string `json:"style_guide,omitempty"`
}

// BriefCharacter はキャラクター情報のスロット
type BriefCharacter struct {
	Name               string `json:"name"`
	Gender             string `json:"gender"`
	Persona            string `json:"persona,omitempty"`
	RoleInConversation string `json:"role_in_conversation,omitempty"`
	InteractionStyle   string `json:"interaction_style,omitempty"`
}

// BriefConstraints は制約条件のスロット
type BriefConstraints struct {
	TalkMode     TalkMode `json:"talk_mode"`
	WithEmotion  bool     `json:"with_emotion"`
	TTSOptimized bool     `json:"tts_optimized"`
	Avoid        []string `json:"avoid"`
}

// BriefInput は NormalizeBrief への入力（model 層に依存しない形）
type BriefInput struct {
	// Episode 情報
	EpisodeTitle       string
	EpisodeDescription string
	DurationMinutes    int
	EpisodeNumber      int

	// Channel 情報
	ChannelName        string
	ChannelDescription string
	ChannelCategory    string
	ChannelStyleGuide  string

	// Characters
	Characters []BriefInputCharacter

	// User
	MasterGuide string

	// Theme（ユーザー入力 prompt）
	Theme string

	// Options
	WithEmotion bool
}

// BriefInputCharacter はキャラクター入力情報
type BriefInputCharacter struct {
	Name               string
	Gender             string
	Persona            string
	RoleInConversation string
	InteractionStyle   string
}

// NormalizeBrief は入力情報を正規化ブリーフに変換する
func NormalizeBrief(input BriefInput) Brief {
	characters := make([]BriefCharacter, len(input.Characters))
	for i, c := range input.Characters {
		characters[i] = BriefCharacter(c)
	}

	return Brief{
		Episode: BriefEpisode{
			Title:           input.EpisodeTitle,
			Description:     input.EpisodeDescription,
			DurationMinutes: input.DurationMinutes,
			EpisodeNumber:   input.EpisodeNumber,
		},
		Channel: BriefChannel{
			Name:        input.ChannelName,
			Description: input.ChannelDescription,
			Category:    input.ChannelCategory,
			StyleGuide:  input.ChannelStyleGuide,
		},
		Characters:  characters,
		MasterGuide: input.MasterGuide,
		Theme:       input.Theme,
		Constraints: BriefConstraints{
			TalkMode:     DetectTalkMode(len(input.Characters)),
			WithEmotion:  input.WithEmotion,
			TTSOptimized: true,
			Avoid:        []string{},
		},
	}
}

// DetectTalkMode はキャラクター数から TalkMode を判定する
func DetectTalkMode(characterCount int) TalkMode {
	// 1人なら monologue、2人以上なら dialogue
	if characterCount <= 1 {
		return TalkModeMonologue
	}
	return TalkModeDialogue
}

// ToJSON は Brief を JSON 文字列に変換する
func (b *Brief) ToJSON() (string, error) {
	data, err := json.Marshal(b)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
