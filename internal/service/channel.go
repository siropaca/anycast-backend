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
	GetMyChannel(ctx context.Context, userID, channelID string) (*response.ChannelDataResponse, error)
	ListMyChannels(ctx context.Context, userID string, filter repository.ChannelFilter) (*response.ChannelListWithPaginationResponse, error)
	CreateChannel(ctx context.Context, userID string, req request.CreateChannelRequest) (*response.ChannelDataResponse, error)
	UpdateChannel(ctx context.Context, userID, channelID string, req request.UpdateChannelRequest) (*response.ChannelDataResponse, error)
	DeleteChannel(ctx context.Context, userID, channelID string) error
	PublishChannel(ctx context.Context, userID, channelID string, publishedAt *string) (*response.ChannelDataResponse, error)
	UnpublishChannel(ctx context.Context, userID, channelID string) (*response.ChannelDataResponse, error)
}

type channelService struct {
	channelRepo  repository.ChannelRepository
	categoryRepo repository.CategoryRepository
	imageRepo    repository.ImageRepository
	voiceRepo    repository.VoiceRepository
}

// ChannelService の実装を返す
func NewChannelService(
	channelRepo repository.ChannelRepository,
	categoryRepo repository.CategoryRepository,
	imageRepo repository.ImageRepository,
	voiceRepo repository.VoiceRepository,
) ChannelService {
	return &channelService{
		channelRepo:  channelRepo,
		categoryRepo: categoryRepo,
		imageRepo:    imageRepo,
		voiceRepo:    voiceRepo,
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

// 自分のチャンネルを取得する（オーナーのみ取得可能）
func (s *channelService) GetMyChannel(ctx context.Context, userID, channelID string) (*response.ChannelDataResponse, error) {
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

	// オーナーでない場合は 403
	if channel.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("You do not have permission to access this channel")
	}

	return &response.ChannelDataResponse{
		Data: toChannelResponse(channel, true),
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

	// キャラクターのバリデーションと構築
	characters := make([]model.Character, len(req.Characters))
	for i, charReq := range req.Characters {
		voiceID, err := uuid.Parse(charReq.VoiceID)
		if err != nil {
			return nil, err
		}

		// ボイスの存在確認（アクティブなもののみ）
		if _, err := s.voiceRepo.FindActiveByID(ctx, charReq.VoiceID); err != nil {
			return nil, err
		}

		characters[i] = model.Character{
			Name:    charReq.Name,
			Persona: charReq.Persona,
			VoiceID: voiceID,
		}
	}

	// チャンネルモデルを作成
	channel := &model.Channel{
		UserID:      uid,
		Name:        req.Name,
		Description: req.Description,
		UserPrompt:  req.UserPrompt,
		CategoryID:  categoryID,
		ArtworkID:   artworkID,
		Characters:  characters,
	}

	// チャンネルを保存（キャラクターも一緒に保存される）
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

// チャンネルを更新する（オーナーのみ更新可能）
func (s *channelService) UpdateChannel(ctx context.Context, userID, channelID string, req request.UpdateChannelRequest) (*response.ChannelDataResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("You do not have permission to update this channel")
	}

	// 各フィールドを更新（指定されたもののみ）
	if req.Name != nil {
		channel.Name = *req.Name
	}
	if req.Description != nil {
		channel.Description = *req.Description
	}
	if req.UserPrompt != nil {
		channel.UserPrompt = *req.UserPrompt
	}

	// カテゴリの更新
	if req.CategoryID != nil {
		categoryID, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			return nil, err
		}
		if _, err := s.categoryRepo.FindByID(ctx, categoryID); err != nil {
			return nil, err
		}
		channel.CategoryID = categoryID
	}

	// アートワークの更新
	if req.ArtworkImageID != nil {
		if *req.ArtworkImageID == "" {
			// 空文字の場合は null に設定
			channel.ArtworkID = nil
		} else {
			artworkID, err := uuid.Parse(*req.ArtworkImageID)
			if err != nil {
				return nil, err
			}
			if _, err := s.imageRepo.FindByID(ctx, artworkID); err != nil {
				return nil, err
			}
			channel.ArtworkID = &artworkID
		}
	}

	// チャンネルを更新
	if err := s.channelRepo.Update(ctx, channel); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.channelRepo.FindByID(ctx, channel.ID)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: toChannelResponse(updated, true),
	}, nil
}

// チャンネルを削除する（オーナーのみ削除可能）
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

// チャンネルを公開する
func (s *channelService) PublishChannel(ctx context.Context, userID, channelID string, publishedAt *string) (*response.ChannelDataResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("You do not have permission to publish this channel")
	}

	// 公開日時を設定
	if publishedAt == nil || *publishedAt == "" {
		// 省略時は現在時刻で即時公開
		now := time.Now()
		channel.PublishedAt = &now
	} else {
		// 指定された日時でパース
		parsedTime, err := time.Parse(time.RFC3339, *publishedAt)
		if err != nil {
			return nil, apperror.ErrValidation.WithMessage("Invalid publishedAt format. Use RFC3339 format.")
		}
		channel.PublishedAt = &parsedTime
	}

	// チャンネルを更新
	if err := s.channelRepo.Update(ctx, channel); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.channelRepo.FindByID(ctx, channel.ID)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: toChannelResponse(updated, true),
	}, nil
}

// チャンネルを非公開にする
func (s *channelService) UnpublishChannel(ctx context.Context, userID, channelID string) (*response.ChannelDataResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("You do not have permission to unpublish this channel")
	}

	// 公開日時を null に設定（非公開化）
	channel.PublishedAt = nil

	// チャンネルを更新
	if err := s.channelRepo.Update(ctx, channel); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.channelRepo.FindByID(ctx, channel.ID)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: toChannelResponse(updated, true),
	}, nil
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
// isOwner が false の場合、userPrompt は空文字になる
func toChannelResponse(c *model.Channel, isOwner bool) response.ChannelResponse {
	userPrompt := ""
	if isOwner {
		userPrompt = c.UserPrompt
	}

	resp := response.ChannelResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		UserPrompt:  userPrompt,
		Category: response.CategoryResponse{
			ID:        c.Category.ID,
			Slug:      c.Category.Slug,
			Name:      c.Category.Name,
			SortOrder: c.Category.SortOrder,
			IsActive:  c.Category.IsActive,
		},
		Characters:  toCharacterResponses(c.Characters),
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

// Character モデルのスライスをレスポンス DTO のスライスに変換する
func toCharacterResponses(characters []model.Character) []response.CharacterResponse {
	result := make([]response.CharacterResponse, len(characters))

	for i, c := range characters {
		result[i] = response.CharacterResponse{
			ID:      c.ID,
			Name:    c.Name,
			Persona: c.Persona,
			Voice: response.CharacterVoiceResponse{
				ID:     c.Voice.ID,
				Name:   c.Voice.Name,
				Gender: string(c.Voice.Gender),
			},
		}
	}

	return result
}
