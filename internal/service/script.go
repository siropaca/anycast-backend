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
