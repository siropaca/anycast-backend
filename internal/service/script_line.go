package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// 台本行関連のビジネスロジックインターフェース
type ScriptLineService interface {
	ListByEpisodeID(ctx context.Context, userID, channelID, episodeID string) (*response.ScriptLineListResponse, error)
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

	return &response.ScriptLineListResponse{
		Data: toScriptLineResponses(scriptLines),
	}, nil
}

// ScriptLine モデルのスライスをレスポンス DTO のスライスに変換する
func toScriptLineResponses(scriptLines []model.ScriptLine) []response.ScriptLineResponse {
	result := make([]response.ScriptLineResponse, len(scriptLines))

	for i, sl := range scriptLines {
		result[i] = toScriptLineResponse(&sl)
	}

	return result
}

// ScriptLine モデルをレスポンス DTO に変換する
func toScriptLineResponse(sl *model.ScriptLine) response.ScriptLineResponse {
	resp := response.ScriptLineResponse{
		ID:         sl.ID,
		LineOrder:  sl.LineOrder,
		LineType:   string(sl.LineType),
		Text:       sl.Text,
		Emotion:    sl.Emotion,
		DurationMs: sl.DurationMs,
		CreatedAt:  sl.CreatedAt,
		UpdatedAt:  sl.UpdatedAt,
	}

	// decimal.Decimal から float64 に変換
	if sl.Volume != nil {
		v, _ := sl.Volume.Float64()
		resp.Volume = &v
	}

	if sl.Speaker != nil {
		resp.Speaker = &response.SpeakerResponse{
			ID:   sl.Speaker.ID,
			Name: sl.Speaker.Name,
		}
	}

	if sl.Sfx != nil {
		resp.Sfx = &response.SfxResponse{
			ID:   sl.Sfx.ID,
			Name: sl.Sfx.Name,
		}
	}

	if sl.Audio != nil {
		resp.Audio = &response.AudioResponse{
			ID:         sl.Audio.ID,
			URL:        sl.Audio.URL,
			DurationMs: sl.Audio.DurationMs,
		}
	}

	return resp
}
