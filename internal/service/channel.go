package service

import (
	"context"
	"time"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// チャンネル関連のビジネスロジックインターフェース
type ChannelService interface {
	GetChannel(ctx context.Context, userID, channelID string) (*response.ChannelDataResponse, error)
	ListMyChannels(ctx context.Context, userID string, filter repository.ChannelFilter) (*response.ChannelListWithPaginationResponse, error)
	CreateChannel(ctx context.Context, userID string, req request.CreateChannelRequest) (*response.ChannelDataResponse, error)
	DeleteChannel(ctx context.Context, userID, channelID string) error
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

// チャンネルを取得する
// オーナーまたは公開中のチャンネルのみ取得可能
func (s *channelService) GetChannel(ctx context.Context, userID, channelID string) (*response.ChannelDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}

	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	isOwner := channel.UserID == uid
	isPublished := channel.PublishedAt != nil && !channel.PublishedAt.After(time.Now())

	// オーナーでなく、かつ公開されていない場合は 404
	if !isOwner && !isPublished {
		return nil, apperror.ErrNotFound.WithMessage("Channel not found")
	}

	return &response.ChannelDataResponse{
		Data: toChannelResponse(channel, isOwner),
	}, nil
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
		Data: toChannelResponse(created, true),
	}, nil
}

// チャンネルを削除する
// オーナーのみ削除可能
func (s *channelService) DeleteChannel(ctx context.Context, userID, channelID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return err
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return err
	}

	if channel.UserID != uid {
		return apperror.ErrForbidden.WithMessage("You do not have permission to delete this channel")
	}

	return s.channelRepo.Delete(ctx, cid)
}

// Channel モデルのスライスをレスポンス DTO のスライスに変換する
// ListMyChannels で使用するため、常にオーナーとして扱う
func toChannelResponses(channels []model.Channel) []response.ChannelResponse {
	result := make([]response.ChannelResponse, len(channels))
	for i, c := range channels {
		result[i] = toChannelResponse(&c, true)
	}
	return result
}

// Channel モデルをレスポンス DTO に変換する
// isOwner が false の場合、scriptPrompt は空文字になる
func toChannelResponse(c *model.Channel, isOwner bool) response.ChannelResponse {
	scriptPrompt := ""
	if isOwner {
		scriptPrompt = c.ScriptPrompt
	}

	resp := response.ChannelResponse{
		ID:           c.ID,
		Name:         c.Name,
		Description:  c.Description,
		ScriptPrompt: scriptPrompt,
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
