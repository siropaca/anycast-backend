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
	SearchEpisodes(ctx context.Context, filter repository.SearchEpisodeFilter) (*response.SearchEpisodeListResponse, error)
	SearchUsers(ctx context.Context, filter repository.SearchUserFilter) (*response.SearchUserListResponse, error)
}

type searchService struct {
	channelRepo   repository.ChannelRepository
	episodeRepo   repository.EpisodeRepository
	userRepo      repository.UserRepository
	storageClient storage.Client
}

// NewSearchService は searchService を生成して SearchService として返す
func NewSearchService(
	channelRepo repository.ChannelRepository,
	episodeRepo repository.EpisodeRepository,
	userRepo repository.UserRepository,
	storageClient storage.Client,
) SearchService {
	return &searchService{
		channelRepo:   channelRepo,
		episodeRepo:   episodeRepo,
		userRepo:      userRepo,
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

// SearchEpisodes は公開エピソードをキーワードで検索する
func (s *searchService) SearchEpisodes(ctx context.Context, filter repository.SearchEpisodeFilter) (*response.SearchEpisodeListResponse, error) {
	episodes, total, err := s.episodeRepo.Search(ctx, filter)
	if err != nil {
		return nil, err
	}

	data := make([]response.SearchEpisodeResponse, len(episodes))
	for i, e := range episodes {
		data[i] = toSearchEpisodeResponse(&e)
	}

	return &response.SearchEpisodeListResponse{
		Data:       data,
		Pagination: response.PaginationResponse{Total: total, Limit: filter.Limit, Offset: filter.Offset},
	}, nil
}

// toSearchEpisodeResponse は Episode を検索用レスポンス DTO に変換する
func toSearchEpisodeResponse(e *model.Episode) response.SearchEpisodeResponse {
	return response.SearchEpisodeResponse{
		ID:          e.ID,
		Title:       e.Title,
		Description: e.Description,
		Channel: response.SearchEpisodeChannelResponse{
			ID:   e.Channel.ID,
			Name: e.Channel.Name,
		},
		PublishedAt: e.PublishedAt,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

// SearchUsers はユーザーをキーワードで検索する
func (s *searchService) SearchUsers(ctx context.Context, filter repository.SearchUserFilter) (*response.SearchUserListResponse, error) {
	users, total, err := s.userRepo.Search(ctx, filter)
	if err != nil {
		return nil, err
	}

	data := make([]response.SearchUserResponse, len(users))
	for i, u := range users {
		resp, err := s.toSearchUserResponse(ctx, &u)
		if err != nil {
			return nil, err
		}
		data[i] = resp
	}

	return &response.SearchUserListResponse{
		Data:       data,
		Pagination: response.PaginationResponse{Total: total, Limit: filter.Limit, Offset: filter.Offset},
	}, nil
}

// toSearchUserResponse は User を検索用レスポンス DTO に変換する
func (s *searchService) toSearchUserResponse(ctx context.Context, u *model.User) (response.SearchUserResponse, error) {
	resp := response.SearchUserResponse{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		CreatedAt:   u.CreatedAt,
	}

	if u.Avatar != nil {
		var avatarURL string
		if storage.IsExternalURL(u.Avatar.Path) {
			avatarURL = u.Avatar.Path
		} else {
			var err error
			avatarURL, err = s.storageClient.GenerateSignedURL(ctx, u.Avatar.Path, storage.SignedURLExpirationImage)
			if err != nil {
				return response.SearchUserResponse{}, err
			}
		}
		resp.Avatar = &response.ArtworkResponse{
			ID:  u.Avatar.ID,
			URL: avatarURL,
		}
	}

	return resp, nil
}
