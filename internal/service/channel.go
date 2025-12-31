package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// チャンネル関連のビジネスロジックインターフェース
type ChannelService interface {
	ListMyChannels(ctx context.Context, userID string, filter repository.ChannelFilter) (*response.ChannelListWithPaginationResponse, error)
}

type channelService struct {
	channelRepo repository.ChannelRepository
}

// ChannelService の実装を返す
func NewChannelService(channelRepo repository.ChannelRepository) ChannelService {
	return &channelService{channelRepo: channelRepo}
}

// 自分のチャンネル一覧を取得する
func (s *channelService) ListMyChannels(ctx context.Context, userID string, filter repository.ChannelFilter) (*response.ChannelListWithPaginationResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	channels, total, err := s.channelRepo.FindByUserID(ctx, uid, filter)
	if err != nil {
		return nil, err
	}

	return &response.ChannelListWithPaginationResponse{
		Data:       toChannelResponses(channels),
		Pagination: response.PaginationResponse{Total: total, Limit: filter.Limit, Offset: filter.Offset},
	}, nil
}

// Channel モデルのスライスをレスポンス DTO のスライスに変換する
func toChannelResponses(channels []model.Channel) []response.ChannelResponse {
	result := make([]response.ChannelResponse, len(channels))
	for i, c := range channels {
		result[i] = toChannelResponse(&c)
	}
	return result
}

// Channel モデルをレスポンス DTO に変換する
func toChannelResponse(c *model.Channel) response.ChannelResponse {
	resp := response.ChannelResponse{
		ID:           c.ID,
		Name:         c.Name,
		Description:  c.Description,
		ScriptPrompt: c.ScriptPrompt,
		Category: response.CategoryResponse{
			ID:   c.Category.ID,
			Slug: c.Category.Slug,
			Name: c.Category.Name,
		},
		PublishedAt: c.PublishedAt,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}

	if c.Artwork != nil {
		resp.Artwork = &response.ArtworkResponse{
			ID:  c.Artwork.ID,
			URL: c.Artwork.URL,
		}
	}

	return resp
}
