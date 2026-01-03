package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// エピソード関連のビジネスロジックインターフェース
type EpisodeService interface {
	ListMyChannelEpisodes(ctx context.Context, userID, channelID string, filter repository.EpisodeFilter) (*response.EpisodeListWithPaginationResponse, error)
}

type episodeService struct {
	episodeRepo repository.EpisodeRepository
	channelRepo repository.ChannelRepository
}

// EpisodeService の実装を返す
func NewEpisodeService(
	episodeRepo repository.EpisodeRepository,
	channelRepo repository.ChannelRepository,
) EpisodeService {
	return &episodeService{
		episodeRepo: episodeRepo,
		channelRepo: channelRepo,
	}
}

// 自分のチャンネルのエピソード一覧を取得する
func (s *episodeService) ListMyChannelEpisodes(ctx context.Context, userID, channelID string, filter repository.EpisodeFilter) (*response.EpisodeListWithPaginationResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(channelID)
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

	// エピソード一覧を取得
	episodes, total, err := s.episodeRepo.FindByChannelID(ctx, cid, filter)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeListWithPaginationResponse{
		Data:       toEpisodeResponses(episodes),
		Pagination: response.PaginationResponse{Total: total, Limit: filter.Limit, Offset: filter.Offset},
	}, nil
}

// Episode モデルのスライスをレスポンス DTO のスライスに変換する
func toEpisodeResponses(episodes []model.Episode) []response.EpisodeResponse {
	result := make([]response.EpisodeResponse, len(episodes))

	for i, e := range episodes {
		result[i] = toEpisodeResponse(&e)
	}

	return result
}

// Episode モデルをレスポンス DTO に変換する
func toEpisodeResponse(e *model.Episode) response.EpisodeResponse {
	resp := response.EpisodeResponse{
		ID:           e.ID,
		Title:        e.Title,
		Description:  e.Description,
		ScriptPrompt: e.ScriptPrompt,
		PublishedAt:  e.PublishedAt,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}

	if e.FullAudio != nil {
		resp.FullAudio = &response.AudioResponse{
			ID:         e.FullAudio.ID,
			URL:        e.FullAudio.URL,
			DurationMs: e.FullAudio.DurationMs,
		}
	}

	return resp
}
