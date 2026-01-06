package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// キャラクター関連のビジネスロジックインターフェース
type CharacterService interface {
	ListMyCharacters(ctx context.Context, userID string, filter repository.CharacterFilter) (*response.CharacterListWithPaginationResponse, error)
}

type characterService struct {
	characterRepo repository.CharacterRepository
}

// CharacterService の実装を返す
func NewCharacterService(characterRepo repository.CharacterRepository) CharacterService {
	return &characterService{
		characterRepo: characterRepo,
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

	responses := s.toCharacterResponses(characters)

	return &response.CharacterListWithPaginationResponse{
		Data:       responses,
		Pagination: response.PaginationResponse{Total: total, Limit: filter.Limit, Offset: filter.Offset},
	}, nil
}

// Character モデルのスライスをレスポンス DTO のスライスに変換する
func (s *characterService) toCharacterResponses(characters []model.Character) []response.CharacterResponse {
	result := make([]response.CharacterResponse, len(characters))

	for i, c := range characters {
		result[i] = response.CharacterResponse{
			ID:      c.ID,
			Name:    c.Name,
			Persona: c.Persona,
			Voice: response.CharacterVoiceResponse{
				ID:       c.Voice.ID,
				Name:     c.Voice.Name,
				Provider: c.Voice.Provider,
				Gender:   string(c.Voice.Gender),
			},
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		}
	}

	return result
}
