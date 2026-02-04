package service

import (
	"context"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/script"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// デフォルトのエピソード長さ（分）
const defaultDurationMinutes = 10

// 台本生成のシステムプロンプト（感情なし）
const systemPromptWithoutEmotion = `
あなたはポッドキャスト台本を作成する専門家です。

## 基本ルール
- 自然な会話のテンポを意識する
- 登場人物それぞれの個性（ペルソナ）を反映したセリフにする
- 1つのセリフは20〜80文字程度を目安にする
  - 短すぎるセリフ（単語だけ、相槌だけ）は避け、必ず文章として成立させる
  - 相槌を入れる場合は、その後に続く内容と組み合わせて1行にまとめる
  - 悪い例：「そうそう」「うん」「なるほど」だけの行
  - 良い例：「そうそう、まさにそういうことなんだよね。」「なるほど、それで納得しました。」
- セリフは必ず句点（。）で終わる
- 聞き手が理解しやすいよう、適度に相槌や確認を入れる
- 人間らしい自然な会話感を出すために、適度にフィラー（「えーと」「あのー」「まあ」「なんか」「うーん」など）を入れる
- 指定されたエピソードの長さ（分）に合わせた台本を作成する
  - 目安として、1分あたり約300文字程度のセリフ量になるよう調整する
  - 10分のエピソードなら約3000文字、30分なら約9000文字が目安
- 制作側のメタ発言はしない
- 1行に複数の句点（。）を入れない。一人が続けて話す場合は複数行に分ける
  - 同じキャラクターが連続して発言するのは問題ない

## 出力形式
以下のテキスト形式で出力してください：

話者名: セリフ

- 1行につき1つのセリフ
- 空行は入れない

例：
太郎: こんにちは、今日もよろしくお願いします。
太郎: 今日はいい天気だから気分がいいね。
花子: やあ、元気そうで何よりだね。
太郎: おかげさまで絶好調だよ。

## 制約
- 話者名は与えられた登場人物リストの名前のみ使用可能
- 台本テキスト以外の説明文やコメントは出力しない`

// 台本生成のシステムプロンプト（感情あり）
const systemPromptWithEmotion = `
あなたはポッドキャスト台本を作成する専門家です。

## 基本ルール
- 自然な会話のテンポを意識する
- 登場人物それぞれの個性（ペルソナ）を反映したセリフにする
- 1つのセリフは20〜80文字程度を目安にする
  - 短すぎるセリフ（単語だけ、相槌だけ）は避け、必ず文章として成立させる
  - 相槌を入れる場合は、その後に続く内容と組み合わせて1行にまとめる
  - 悪い例：「そうそう」「うん」「なるほど」だけの行
  - 良い例：「そうそう、まさにそういうことなんだよね。」「なるほど、それで納得しました。」
- セリフは必ず句点（。）で終わる
- 聞き手が理解しやすいよう、適度に相槌や確認を入れる
- 人間らしい自然な会話感を出すために、適度にフィラー（「えーと」「あのー」「まあ」「なんか」「うーん」など）を入れる
- 指定されたエピソードの長さ（分）に合わせた台本を作成する
  - 目安として、1分あたり約300文字程度のセリフ量になるよう調整する
  - 10分のエピソードなら約3000文字、30分なら約9000文字が目安
- 制作側のメタ発言はしない
- 1行に複数の句点（。）を入れない。一人が続けて話す場合は複数行に分ける
  - 同じキャラクターが連続して発言するのは問題ない

## 出力形式
以下のテキスト形式で出力してください：

話者名: [感情] セリフ

- 感情は省略可能。指定する場合は [感情] の形式でセリフの前に記載
- 1行につき1つのセリフ
- 空行は入れない

例：
太郎: こんにちは、今日もよろしくお願いします。
太郎: 今日はいい天気だから気分がいいね。
花子: [嬉しそうに] やあ、元気そうで何よりだね。
太郎: おかげさまで絶好調だよ。

## 制約
- 話者名は与えられた登場人物リストの名前のみ使用可能
- 台本テキスト以外の説明文やコメントは出力しない`

// ExportScriptResult は台本エクスポート結果を表す
type ExportScriptResult struct {
	EpisodeTitle string // エピソードタイトル
	Text         string // 台本テキスト
}

// ScriptService は台本関連のビジネスロジックインターフェースを表す
type ScriptService interface {
	ImportScript(ctx context.Context, userID, channelID, episodeID, text string) (*response.ScriptLineListResponse, error)
	ExportScript(ctx context.Context, userID, channelID, episodeID string) (*ExportScriptResult, error)
}

type scriptService struct {
	db             *gorm.DB
	channelRepo    repository.ChannelRepository
	episodeRepo    repository.EpisodeRepository
	scriptLineRepo repository.ScriptLineRepository
	storageClient  storage.Client
}

// NewScriptService は scriptService を生成して ScriptService として返す
func NewScriptService(
	db *gorm.DB,
	channelRepo repository.ChannelRepository,
	episodeRepo repository.EpisodeRepository,
	scriptLineRepo repository.ScriptLineRepository,
	storageClient storage.Client,
) ScriptService {
	return &scriptService{
		db:             db,
		channelRepo:    channelRepo,
		episodeRepo:    episodeRepo,
		scriptLineRepo: scriptLineRepo,
		storageClient:  storageClient,
	}
}

// toScriptLineResponses は ScriptLine のスライスをレスポンス DTO のスライスに変換する
func (s *scriptService) toScriptLineResponses(scriptLines []model.ScriptLine) ([]response.ScriptLineResponse, error) {
	result := make([]response.ScriptLineResponse, len(scriptLines))

	for i, sl := range scriptLines {
		resp, err := s.toScriptLineResponse(&sl)
		if err != nil {
			return nil, err
		}
		result[i] = resp
	}

	return result, nil
}

// toScriptLineResponse は ScriptLine をレスポンス DTO に変換する
func (s *scriptService) toScriptLineResponse(sl *model.ScriptLine) (response.ScriptLineResponse, error) {
	return response.ScriptLineResponse{
		ID:        sl.ID,
		LineOrder: sl.LineOrder,
		Speaker: response.SpeakerResponse{
			ID:      sl.Speaker.ID,
			Name:    sl.Speaker.Name,
			Persona: sl.Speaker.Persona,
			Voice: response.CharacterVoiceResponse{
				ID:       sl.Speaker.Voice.ID,
				Name:     sl.Speaker.Voice.Name,
				Provider: sl.Speaker.Voice.Provider,
				Gender:   string(sl.Speaker.Voice.Gender),
			},
		},
		Text:      sl.Text,
		Emotion:   sl.Emotion,
		CreatedAt: sl.CreatedAt,
		UpdatedAt: sl.UpdatedAt,
	}, nil
}

// ImportScript はテキスト形式の台本をインポートする
func (s *scriptService) ImportScript(ctx context.Context, userID, channelID, episodeID, text string) (*response.ScriptLineListResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルへのアクセス権限がありません")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("このチャンネルにエピソードが見つかりません")
	}

	// 許可された話者名のリストを作成
	allowedSpeakers := make([]string, len(channel.ChannelCharacters))
	speakerMap := make(map[string]*model.Character, len(channel.ChannelCharacters))
	for i, cc := range channel.ChannelCharacters {
		allowedSpeakers[i] = cc.Character.Name
		speakerMap[cc.Character.Name] = &channel.ChannelCharacters[i].Character
	}

	// テキストをパース
	parseResult := script.Parse(text, allowedSpeakers)

	// パースエラーがある場合
	if parseResult.HasErrors() {
		details := make([]map[string]any, len(parseResult.Errors))
		for i, e := range parseResult.Errors {
			details[i] = map[string]any{
				"line":   e.Line,
				"reason": e.Reason,
			}
		}
		return nil, apperror.ErrScriptParse.WithMessage("台本のパースに失敗しました").WithDetails(details)
	}

	// ScriptLine モデルに変換
	scriptLines := make([]model.ScriptLine, len(parseResult.Lines))
	for i, line := range parseResult.Lines {
		speaker := speakerMap[line.SpeakerName]
		scriptLines[i] = model.ScriptLine{
			EpisodeID: eid,
			LineOrder: i,
			SpeakerID: speaker.ID,
			Text:      line.Text,
			Emotion:   line.Emotion,
		}
	}

	// トランザクションで既存行削除・新規作成を実行
	var createdLines []model.ScriptLine
	err = s.db.Transaction(func(tx *gorm.DB) error {
		txScriptLineRepo := repository.NewScriptLineRepository(tx)

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

		return nil
	})

	if err != nil {
		return nil, err
	}

	// レスポンス DTO に変換
	responses, err := s.toScriptLineResponses(createdLines)
	if err != nil {
		return nil, err
	}

	return &response.ScriptLineListResponse{
		Data: responses,
	}, nil
}

// ExportScript は台本をテキスト形式でエクスポートする
func (s *scriptService) ExportScript(ctx context.Context, userID, channelID, episodeID string) (*ExportScriptResult, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルへのアクセス権限がありません")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("このチャンネルにエピソードが見つかりません")
	}

	// 台本行を取得
	scriptLines, err := s.scriptLineRepo.FindByEpisodeID(ctx, eid)
	if err != nil {
		return nil, err
	}

	// テキスト形式に変換
	formatLines := make([]script.FormatLine, len(scriptLines))
	for i, sl := range scriptLines {
		formatLines[i] = script.FormatLine{
			SpeakerName: sl.Speaker.Name,
			Text:        sl.Text,
			Emotion:     sl.Emotion,
		}
	}

	text := script.Format(formatLines)

	return &ExportScriptResult{
		EpisodeTitle: episode.Title,
		Text:         text,
	}, nil
}
