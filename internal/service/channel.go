package service

import (
	"context"
	"strings"
	"time"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
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
	channelRepo   repository.ChannelRepository
	characterRepo repository.CharacterRepository
	categoryRepo  repository.CategoryRepository
	imageRepo     repository.ImageRepository
	voiceRepo     repository.VoiceRepository
	storageClient storage.Client
}

// ChannelService の実装を返す
func NewChannelService(
	channelRepo repository.ChannelRepository,
	characterRepo repository.CharacterRepository,
	categoryRepo repository.CategoryRepository,
	imageRepo repository.ImageRepository,
	voiceRepo repository.VoiceRepository,
	storageClient storage.Client,
) ChannelService {
	return &channelService{
		channelRepo:   channelRepo,
		characterRepo: characterRepo,
		categoryRepo:  categoryRepo,
		imageRepo:     imageRepo,
		voiceRepo:     voiceRepo,
		storageClient: storageClient,
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

	resp, err := s.toChannelResponse(ctx, channel, isOwner)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: resp,
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

	resp, err := s.toChannelResponse(ctx, channel, true)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: resp,
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

	responses, err := s.toChannelResponses(ctx, channels)
	if err != nil {
		return nil, err
	}

	return &response.ChannelListWithPaginationResponse{
		Data:       responses,
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

	// キャラクター数のバリデーション（1〜2件）
	if req.Characters.Total() < 1 || req.Characters.Total() > 2 {
		return nil, apperror.ErrValidation.WithMessage("Characters must have 1 to 2 items")
	}

	// キャラクターの処理（既存 or 新規作成）
	characterIDs, err := s.processCharacterInputs(ctx, uid, req.Characters)
	if err != nil {
		return nil, err
	}

	// チャンネルモデルを作成
	channel := &model.Channel{
		UserID:      uid,
		Name:        req.Name,
		Description: req.Description,
		UserPrompt:  req.UserPrompt,
		CategoryID:  categoryID,
		ArtworkID:   artworkID,
	}

	// チャンネルを保存
	if err := s.channelRepo.Create(ctx, channel); err != nil {
		return nil, err
	}

	// キャラクターを紐づけ
	if err := s.channelRepo.ReplaceChannelCharacters(ctx, channel.ID, characterIDs); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	created, err := s.channelRepo.FindByID(ctx, channel.ID)
	if err != nil {
		return nil, err
	}

	resp, err := s.toChannelResponse(ctx, created, true)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: resp,
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

	// 各フィールドを更新
	channel.Name = req.Name
	channel.Description = req.Description
	channel.UserPrompt = req.UserPrompt

	// カテゴリの更新
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, err
	}
	if _, err := s.categoryRepo.FindByID(ctx, categoryID); err != nil {
		return nil, err
	}
	channel.CategoryID = categoryID

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

	resp, err := s.toChannelResponse(ctx, updated, true)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: resp,
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

	resp, err := s.toChannelResponse(ctx, updated, true)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: resp,
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

	resp, err := s.toChannelResponse(ctx, updated, true)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: resp,
	}, nil
}

// キャラクター入力を処理して、キャラクター ID のスライスを返す
// connect は既存キャラクターの紐づけ、create は新規キャラクターの作成
func (s *channelService) processCharacterInputs(ctx context.Context, userID uuid.UUID, input request.ChannelCharactersInput) ([]uuid.UUID, error) {
	characterIDs := make([]uuid.UUID, 0, input.Total())

	// 既存キャラクターの紐づけ処理
	for _, connect := range input.Connect {
		cid, err := uuid.Parse(connect.ID)
		if err != nil {
			return nil, err
		}

		// キャラクターの存在確認とオーナーチェック
		character, err := s.characterRepo.FindByID(ctx, cid)
		if err != nil {
			return nil, err
		}
		if character.UserID != userID {
			return nil, apperror.ErrForbidden.WithMessage("You do not own the specified character")
		}

		characterIDs = append(characterIDs, cid)
	}

	// 新規キャラクターの作成処理
	for _, create := range input.Create {
		// 予約語チェック
		if strings.HasPrefix(create.Name, "__") {
			return nil, apperror.ErrReservedName.WithMessage("Character name cannot start with '__'")
		}

		// 同一ユーザー内での名前重複チェック
		exists, err := s.characterRepo.ExistsByUserIDAndName(ctx, userID, create.Name, nil)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperror.ErrDuplicateName.WithMessage("Character with this name already exists")
		}

		voiceID, err := uuid.Parse(create.VoiceID)
		if err != nil {
			return nil, err
		}

		// ボイスの存在確認（アクティブなもののみ）
		if _, err := s.voiceRepo.FindActiveByID(ctx, create.VoiceID); err != nil {
			return nil, err
		}

		// アバター画像の存在確認（指定時のみ）
		var avatarID *uuid.UUID
		if create.AvatarID != nil {
			aid, err := uuid.Parse(*create.AvatarID)
			if err != nil {
				return nil, err
			}
			if _, err := s.imageRepo.FindByID(ctx, aid); err != nil {
				return nil, err
			}
			avatarID = &aid
		}

		character := &model.Character{
			UserID:   userID,
			Name:     create.Name,
			Persona:  create.Persona,
			AvatarID: avatarID,
			VoiceID:  voiceID,
		}

		if err := s.characterRepo.Create(ctx, character); err != nil {
			return nil, err
		}

		characterIDs = append(characterIDs, character.ID)
	}

	return characterIDs, nil
}

// Channel モデルのスライスをレスポンス DTO のスライスに変換する
// ListMyChannels で使用するため、常にオーナーとして扱う
func (s *channelService) toChannelResponses(ctx context.Context, channels []model.Channel) ([]response.ChannelResponse, error) {
	result := make([]response.ChannelResponse, len(channels))

	for i, c := range channels {
		resp, err := s.toChannelResponse(ctx, &c, true)
		if err != nil {
			return nil, err
		}
		result[i] = resp
	}

	return result, nil
}

// Channel モデルをレスポンス DTO に変換する
// isOwner が false の場合、userPrompt は空文字になる
func (s *channelService) toChannelResponse(ctx context.Context, c *model.Channel, isOwner bool) (response.ChannelResponse, error) {
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
		Characters:  s.toCharacterResponsesFromChannelCharacters(c.ChannelCharacters),
		PublishedAt: c.PublishedAt,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}

	if c.Artwork != nil {
		signedURL, err := s.storageClient.GenerateSignedURL(ctx, c.Artwork.Path, signedURLExpiration)
		if err != nil {
			return response.ChannelResponse{}, err
		}
		resp.Artwork = &response.ArtworkResponse{
			ID:  c.Artwork.ID,
			URL: signedURL,
		}
	}

	return resp, nil
}

// ChannelCharacter のスライスからレスポンス DTO のスライスに変換する
func (s *channelService) toCharacterResponsesFromChannelCharacters(channelCharacters []model.ChannelCharacter) []response.CharacterResponse {
	result := make([]response.CharacterResponse, len(channelCharacters))

	for i, cc := range channelCharacters {
		result[i] = response.CharacterResponse{
			ID:      cc.Character.ID,
			Name:    cc.Character.Name,
			Persona: cc.Character.Persona,
			Voice: response.CharacterVoiceResponse{
				ID:       cc.Character.Voice.ID,
				Name:     cc.Character.Voice.Name,
				Provider: cc.Character.Voice.Provider,
				Gender:   string(cc.Character.Voice.Gender),
			},
			CreatedAt: cc.Character.CreatedAt,
			UpdatedAt: cc.Character.UpdatedAt,
		}
	}

	return result
}
