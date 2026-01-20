package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// 台本行関連のビジネスロジックインターフェース
type ScriptLineService interface {
	ListByEpisodeID(ctx context.Context, userID, channelID, episodeID string) (*response.ScriptLineListResponse, error)
	Update(ctx context.Context, userID, channelID, episodeID, lineID string, req request.UpdateScriptLineRequest) (*response.ScriptLineResponse, error)
	Delete(ctx context.Context, userID, channelID, episodeID, lineID string) error
}

type scriptLineService struct {
	scriptLineRepo repository.ScriptLineRepository
	episodeRepo    repository.EpisodeRepository
	channelRepo    repository.ChannelRepository
}

// ScriptLineService の実装を返す
func NewScriptLineService(
	scriptLineRepo repository.ScriptLineRepository,
	episodeRepo repository.EpisodeRepository,
	channelRepo repository.ChannelRepository,
) ScriptLineService {
	return &scriptLineService{
		scriptLineRepo: scriptLineRepo,
		episodeRepo:    episodeRepo,
		channelRepo:    channelRepo,
	}
}

// 指定されたエピソードの台本行一覧を取得する
func (s *scriptLineService) ListByEpisodeID(ctx context.Context, userID, channelID, episodeID string) (*response.ScriptLineListResponse, error) {
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

	// 台本行一覧を取得
	scriptLines, err := s.scriptLineRepo.FindByEpisodeID(ctx, eid)
	if err != nil {
		return nil, err
	}

	// レスポンスに変換
	responses := s.toScriptLineResponses(scriptLines)

	return &response.ScriptLineListResponse{
		Data: responses,
	}, nil
}

// 指定された台本行を更新する
func (s *scriptLineService) Update(ctx context.Context, userID, channelID, episodeID, lineID string, req request.UpdateScriptLineRequest) (*response.ScriptLineResponse, error) {
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

	lid, err := uuid.Parse(lineID)
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

	// 台本行の存在確認とエピソードの一致チェック
	scriptLine, err := s.scriptLineRepo.FindByID(ctx, lid)
	if err != nil {
		return nil, err
	}

	if scriptLine.EpisodeID != eid {
		return nil, apperror.ErrNotFound.WithMessage("Script line not found in this episode")
	}

	// フィールドを更新
	if req.Text != nil {
		scriptLine.Text = *req.Text
	}

	if req.Emotion != nil {
		if *req.Emotion == "" {
			scriptLine.Emotion = nil
		} else {
			scriptLine.Emotion = req.Emotion
		}
	}

	if err := s.scriptLineRepo.Update(ctx, scriptLine); err != nil {
		return nil, err
	}

	// レスポンスに変換
	resp := s.toScriptLineResponse(scriptLine)

	return &resp, nil
}

// 指定された台本行を削除する
func (s *scriptLineService) Delete(ctx context.Context, userID, channelID, episodeID, lineID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return err
	}

	lid, err := uuid.Parse(lineID)
	if err != nil {
		return err
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return err
	}

	if channel.UserID != uid {
		return apperror.ErrForbidden.WithMessage("You do not have permission to access this channel")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return err
	}

	if episode.ChannelID != cid {
		return apperror.ErrNotFound.WithMessage("Episode not found in this channel")
	}

	// 台本行の存在確認とエピソードの一致チェック
	scriptLine, err := s.scriptLineRepo.FindByID(ctx, lid)
	if err != nil {
		return err
	}

	if scriptLine.EpisodeID != eid {
		return apperror.ErrNotFound.WithMessage("Script line not found in this episode")
	}

	// 台本行を削除
	return s.scriptLineRepo.Delete(ctx, lid)
}

// ScriptLine モデルのスライスをレスポンス DTO のスライスに変換する
func (s *scriptLineService) toScriptLineResponses(scriptLines []model.ScriptLine) []response.ScriptLineResponse {
	result := make([]response.ScriptLineResponse, len(scriptLines))

	for i, sl := range scriptLines {
		result[i] = s.toScriptLineResponse(&sl)
	}

	return result
}

// ScriptLine モデルをレスポンス DTO に変換する
func (s *scriptLineService) toScriptLineResponse(sl *model.ScriptLine) response.ScriptLineResponse {
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
	}
}
