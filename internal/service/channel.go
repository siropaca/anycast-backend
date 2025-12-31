package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// チャンネル関連のビジネスロジックインターフェース
type ChannelService interface {
	ListMyChannels(ctx context.Context, userID string, filter repository.ChannelFilter) (*response.ChannelListWithPaginationResponse, error)
	CreateChannel(ctx context.Context, userID string, req request.CreateChannelRequest) (*response.ChannelDataResponse, error)
}

type channelService struct {
	channelRepo  repository.ChannelRepository
	categoryRepo repository.CategoryRepository
	imageRepo    repository.ImageRepository
}

// ChannelService の実装を返す
func NewChannelService(
	channelRepo repository.ChannelRepository,
	categoryRepo repository.CategoryRepository,
	imageRepo repository.ImageRepository,
) ChannelService {
	return &channelService{
		channelRepo:  channelRepo,
		categoryRepo: categoryRepo,
		imageRepo:    imageRepo,
	}
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

// 新しいチャンネルを作成する
func (s *channelService) CreateChannel(ctx context.Context, userID string, req request.CreateChannelRequest) (*response.ChannelDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, err
	}

	// カテゴリの存在確認
	if _, err := s.categoryRepo.FindByID(ctx, categoryID); err != nil {
		return nil, err
	}

	// Artwork 画像の存在確認（指定時のみ）
	var artworkID *uuid.UUID
	if req.ArtworkImageID != nil {
		aid, err := uuid.Parse(*req.ArtworkImageID)
		if err != nil {
			return nil, err
		}
		if _, err := s.imageRepo.FindByID(ctx, aid); err != nil {
			return nil, err
		}
		artworkID = &aid
	}

	// チャンネルモデルを作成
	channel := &model.Channel{
		UserID:       uid,
		Name:         req.Name,
		Description:  req.Description,
		ScriptPrompt: req.ScriptPrompt,
		CategoryID:   categoryID,
		ArtworkID:    artworkID,
	}

	// チャンネルを保存
	if err := s.channelRepo.Create(ctx, channel); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	created, err := s.channelRepo.FindByID(ctx, channel.ID)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: toChannelResponse(created),
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
