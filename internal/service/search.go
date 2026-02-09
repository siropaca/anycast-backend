package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// SearchService は検索関連のビジネスロジックインターフェースを表す
type SearchService interface {
	SearchChannels(ctx context.Context, filter repository.SearchChannelFilter) (*response.SearchChannelListResponse, error)
}

type searchService struct {
	channelRepo   repository.ChannelRepository
	storageClient storage.Client
}

// NewSearchService は searchService を生成して SearchService として返す
func NewSearchService(
	channelRepo repository.ChannelRepository,
	storageClient storage.Client,
) SearchService {
	return &searchService{
		channelRepo:   channelRepo,
		storageClient: storageClient,
	}
}

// SearchChannels は公開チャンネルをキーワードで検索する
func (s *searchService) SearchChannels(ctx context.Context, filter repository.SearchChannelFilter) (*response.SearchChannelListResponse, error) {
	channels, total, err := s.channelRepo.Search(ctx, filter)
	if err != nil {
		return nil, err
	}

	data := make([]response.SearchChannelResponse, len(channels))
	for i, c := range channels {
		resp, err := s.toSearchChannelResponse(ctx, &c)
		if err != nil {
			return nil, err
		}
		data[i] = resp
	}

	return &response.SearchChannelListResponse{
		Data:       data,
		Pagination: response.PaginationResponse{Total: total, Limit: filter.Limit, Offset: filter.Offset},
	}, nil
}

// toSearchChannelResponse は Channel を検索用レスポンス DTO に変換する
func (s *searchService) toSearchChannelResponse(ctx context.Context, c *model.Channel) (response.SearchChannelResponse, error) {
	resp := response.SearchChannelResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Category: response.CategoryResponse{
			ID:        c.Category.ID,
			Slug:      c.Category.Slug,
			Name:      c.Category.Name,
			SortOrder: c.Category.SortOrder,
			IsActive:  c.Category.IsActive,
		},
		PublishedAt: c.PublishedAt,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}

	if c.Artwork != nil {
		var artworkURL string
		if storage.IsExternalURL(c.Artwork.Path) {
			artworkURL = c.Artwork.Path
		} else {
			var err error
			artworkURL, err = s.storageClient.GenerateSignedURL(ctx, c.Artwork.Path, storage.SignedURLExpirationImage)
			if err != nil {
				return response.SearchChannelResponse{}, err
			}
		}
		resp.Artwork = &response.ArtworkResponse{
			ID:  c.Artwork.ID,
			URL: artworkURL,
		}
	}

	return resp, nil
}
