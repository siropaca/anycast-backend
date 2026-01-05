package service

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/llm"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/script"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

const (
	// デフォルトのエピソード長さ（分）
	defaultDurationMinutes = 10
)

// 台本生成のシステムプロンプト
const systemPrompt = `
あなたはポッドキャスト台本を作成する専門家です。

## 基本ルール
- 自然な会話のテンポを意識する
- 登場人物それぞれの個性（ペルソナ）を反映したセリフにする
- 長すぎるセリフは避け、1つのセリフは50〜100文字程度を目安にする
- 聞き手が理解しやすいよう、適度に相槌や確認を入れる
- 人間らしい自然な会話感を出すために、適度にフィラー（「えーと」「あのー」「まあ」「なんか」「うーん」など）を入れる
- 指定されたエピソードの長さ（分）に合わせた台本を作成する
  - 目安として、1分あたり約300文字程度のセリフ量になるよう調整する
  - 10分のエピソードなら約3000文字、30分なら約9000文字が目安

## 出力形式
以下のテキスト形式で出力してください：

話者名: [感情] セリフ

- 感情は省略可能。指定する場合は [感情] の形式でセリフの前に記載
- 1行につき1つのセリフ
- 空行は入れない

例：
太郎: こんにちは！
花子: [嬉しそうに] やあ、元気？
太郎: 元気だよ！

## 制約
- 話者名は与えられた登場人物リストの名前のみ使用可能
- 台本テキスト以外の説明文やコメントは出力しない`

// 台本関連のビジネスロジックインターフェース
type ScriptService interface {
	GenerateScript(ctx context.Context, userID, channelID, episodeID, prompt string, durationMinutes *int) (*response.ScriptLineListResponse, error)
}

type scriptService struct {
	db             *gorm.DB
	channelRepo    repository.ChannelRepository
	episodeRepo    repository.EpisodeRepository
	scriptLineRepo repository.ScriptLineRepository
	llmClient      llm.Client
}

// ScriptService の実装を返す
func NewScriptService(
	db *gorm.DB,
	channelRepo repository.ChannelRepository,
	episodeRepo repository.EpisodeRepository,
	scriptLineRepo repository.ScriptLineRepository,
	llmClient llm.Client,
) ScriptService {
	return &scriptService{
		db:             db,
		channelRepo:    channelRepo,
		episodeRepo:    episodeRepo,
		scriptLineRepo: scriptLineRepo,
		llmClient:      llmClient,
	}
}

// AI を使って台本を生成する
func (s *scriptService) GenerateScript(ctx context.Context, userID, channelID, episodeID, prompt string, durationMinutes *int) (*response.ScriptLineListResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return nil, err
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	if channel.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("You do not have permission to access this channel")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("Episode not found in this channel")
	}

	// durationMinutes のデフォルト値設定
	duration := defaultDurationMinutes
	if durationMinutes != nil {
		duration = *durationMinutes
	}

	// LLM 用ユーザープロンプトを構築
	userPrompt := s.buildUserPrompt(channel, episode, prompt, duration)

	// LLM で台本生成
	generatedText, err := s.llmClient.GenerateScript(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	// 許可された話者名のリストを作成
	allowedSpeakers := make([]string, len(channel.Characters))
	speakerMap := make(map[string]*model.Character, len(channel.Characters))
	for i, c := range channel.Characters {
		allowedSpeakers[i] = c.Name
		speakerMap[c.Name] = &channel.Characters[i]
	}

	// 生成されたテキストをパース
	parseResult := script.Parse(generatedText, allowedSpeakers)

	// パースエラーがある場合（全行失敗の場合のみエラー）
	if len(parseResult.Lines) == 0 && parseResult.HasErrors() {
		return nil, apperror.ErrGenerationFailed.WithMessage("Failed to parse generated script")
	}

	// ScriptLine モデルに変換
	scriptLines := make([]model.ScriptLine, len(parseResult.Lines))
	for i, line := range parseResult.Lines {
		speaker := speakerMap[line.SpeakerName]
		scriptLines[i] = model.ScriptLine{
			EpisodeID: eid,
			LineOrder: i,
			LineType:  model.LineTypeSpeech,
			SpeakerID: &speaker.ID,
			Text:      &line.Text,
			Emotion:   line.Emotion,
		}
	}

	// トランザクションで既存行削除・新規作成・エピソード更新を実行
	var createdLines []model.ScriptLine
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// トランザクション内で使うリポジトリを作成
		txScriptLineRepo := repository.NewScriptLineRepository(tx)
		txEpisodeRepo := repository.NewEpisodeRepository(tx)

		// 既存の台本行を削除
		if err := txScriptLineRepo.DeleteByEpisodeID(ctx, eid); err != nil {
			return err
		}

		// 新しい台本行を一括作成
		created, err := txScriptLineRepo.CreateBatch(ctx, scriptLines)
		if err != nil {
			return err
		}
		createdLines = created

		// episode.userPrompt を更新
		episode.UserPrompt = &prompt
		if err := txEpisodeRepo.Update(ctx, episode); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// レスポンス DTO に変換
	return &response.ScriptLineListResponse{
		Data: toScriptLineResponses(createdLines),
	}, nil
}

// LLM 用のユーザープロンプトを構築する
func (s *scriptService) buildUserPrompt(channel *model.Channel, episode *model.Episode, prompt string, durationMinutes int) string {
	var sb strings.Builder

	// チャンネル情報
	sb.WriteString("## チャンネル情報\n")
	sb.WriteString(fmt.Sprintf("チャンネル名: %s\n", channel.Name))
	if channel.Description != "" {
		sb.WriteString(fmt.Sprintf("説明: %s\n", channel.Description))
	}
	sb.WriteString("\n")

	// チャンネル設定
	if channel.UserPrompt != "" {
		sb.WriteString("## チャンネル設定\n")
		sb.WriteString(channel.UserPrompt)
		sb.WriteString("\n\n")
	}

	// 登場人物
	sb.WriteString("## 登場人物\n")
	for _, c := range channel.Characters {
		if c.Persona != "" {
			sb.WriteString(fmt.Sprintf("- %s（%s）: %s\n", c.Name, c.Voice.Gender, c.Persona))
		} else {
			sb.WriteString(fmt.Sprintf("- %s（%s）\n", c.Name, c.Voice.Gender))
		}
	}
	sb.WriteString("\n")

	// エピソード情報
	sb.WriteString("## エピソード情報\n")
	sb.WriteString(fmt.Sprintf("タイトル: %s\n", episode.Title))
	if episode.Description != nil && *episode.Description != "" {
		sb.WriteString(fmt.Sprintf("説明: %s\n", *episode.Description))
	}
	sb.WriteString("\n")

	// エピソードの長さ
	sb.WriteString(fmt.Sprintf("## エピソードの長さ\n%d分\n\n", durationMinutes))

	// 今回のテーマ
	sb.WriteString("## 今回のテーマ\n")
	sb.WriteString(prompt)

	return sb.String()
}
