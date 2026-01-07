package service

import (
	"context"
	"time"

	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// 署名付き URL の有効期限
const signedURLExpirationCharacter = 1 * time.Hour

// キャラクター関連のビジネスロジックインターフェース
type CharacterService interface {
	ListMyCharacters(ctx context.Context, userID string, filter repository.CharacterFilter) (*response.CharacterListWithPaginationResponse, error)
}

type characterService struct {
	characterRepo repository.CharacterRepository
	storageClient storage.Client
}

// CharacterService の実装を返す
func NewCharacterService(characterRepo repository.CharacterRepository, storageClient storage.Client) CharacterService {
	return &characterService{
		characterRepo: characterRepo,
		storageClient: storageClient,
	}
}

// 自分のキャラクター一覧を取得する
func (s *characterService) ListMyCharacters(ctx context.Context, userID string, filter repository.CharacterFilter) (*response.CharacterListWithPaginationResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	characters, total, err := s.characterRepo.FindByUserID(ctx, uid, filter)
	if err != nil {
		return nil, err
	}

	responses, err := s.toCharacterWithChannelsResponses(ctx, characters)
	if err != nil {
		return nil, err
	}

	return &response.CharacterListWithPaginationResponse{
		Data:       responses,
		Pagination: response.PaginationResponse{Total: total, Limit: filter.Limit, Offset: filter.Offset},
	}, nil
}

// Character モデルのスライスをチャンネル情報付きレスポンス DTO のスライスに変換する
func (s *characterService) toCharacterWithChannelsResponses(ctx context.Context, characters []model.Character) ([]response.CharacterWithChannelsResponse, error) {
	result := make([]response.CharacterWithChannelsResponse, len(characters))

	for i, c := range characters {
		res, err := s.toCharacterWithChannelsResponse(ctx, c)
		if err != nil {
			return nil, err
		}
		result[i] = res
	}

	return result, nil
}

// Character モデルをチャンネル情報付きレスポンス DTO に変換する
func (s *characterService) toCharacterWithChannelsResponse(ctx context.Context, c model.Character) (response.CharacterWithChannelsResponse, error) {
	channels := make([]response.CharacterChannelResponse, len(c.ChannelCharacters))
	for i, cc := range c.ChannelCharacters {
		channels[i] = response.CharacterChannelResponse{
			ID:   cc.Channel.ID,
			Name: cc.Channel.Name,
		}
	}

	// アバター画像の署名付き URL を生成
	var avatar *response.AvatarResponse
	if c.Avatar != nil && s.storageClient != nil {
		signedURL, err := s.storageClient.GenerateSignedURL(ctx, c.Avatar.Path, signedURLExpirationCharacter)
		if err == nil {
			avatar = &response.AvatarResponse{
				ID:  c.Avatar.ID,
				URL: signedURL,
			}
		}
		// URL 生成に失敗した場合はエラーにせず nil のまま
	}

	return response.CharacterWithChannelsResponse{
		ID:      c.ID,
		Name:    c.Name,
		Persona: c.Persona,
		Avatar:  avatar,
		Voice: response.CharacterVoiceResponse{
			ID:       c.Voice.ID,
			Name:     c.Voice.Name,
			Provider: c.Voice.Provider,
			Gender:   string(c.Voice.Gender),
		},
		Channels:  channels,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}, nil
}
