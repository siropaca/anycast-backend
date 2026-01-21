package service

import (
	"context"
	"strings"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// キャラクター関連のビジネスロジックインターフェース
type CharacterService interface {
	ListMyCharacters(ctx context.Context, userID string, filter repository.CharacterFilter) (*response.CharacterListWithPaginationResponse, error)
	GetMyCharacter(ctx context.Context, userID, characterID string) (*response.CharacterDataResponse, error)
	CreateCharacter(ctx context.Context, userID string, req request.CreateCharacterRequest) (*response.CharacterDataResponse, error)
	UpdateCharacter(ctx context.Context, userID, characterID string, req request.UpdateCharacterRequest) (*response.CharacterDataResponse, error)
	DeleteCharacter(ctx context.Context, userID, characterID string) error
}

type characterService struct {
	characterRepo repository.CharacterRepository
	voiceRepo     repository.VoiceRepository
	imageRepo     repository.ImageRepository
	storageClient storage.Client
}

// CharacterService の実装を返す
func NewCharacterService(
	characterRepo repository.CharacterRepository,
	voiceRepo repository.VoiceRepository,
	imageRepo repository.ImageRepository,
	storageClient storage.Client,
) CharacterService {
	return &characterService{
		characterRepo: characterRepo,
		voiceRepo:     voiceRepo,
		imageRepo:     imageRepo,
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

// 自分のキャラクターを取得する
func (s *characterService) GetMyCharacter(ctx context.Context, userID, characterID string) (*response.CharacterDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(characterID)
	if err != nil {
		return nil, err
	}

	character, err := s.characterRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	// 所有者チェック
	if character.UserID != uid {
		return nil, apperror.ErrNotFound.WithMessage("キャラクターが見つかりません")
	}

	res, err := s.toCharacterWithChannelsResponse(ctx, *character)
	if err != nil {
		return nil, err
	}

	return &response.CharacterDataResponse{Data: res}, nil
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
		signedURL, err := s.storageClient.GenerateSignedURL(ctx, c.Avatar.Path, storage.SignedURLExpirationImage)
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

// キャラクターを作成する
func (s *characterService) CreateCharacter(ctx context.Context, userID string, req request.CreateCharacterRequest) (*response.CharacterDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	// 名前が __ で始まる場合は禁止
	if strings.HasPrefix(req.Name, "__") {
		return nil, apperror.ErrValidation.WithMessage("名前は '__' で始めることはできません")
	}

	// 同一ユーザー内で同じ名前のキャラクターが存在するかチェック
	exists, err := s.characterRepo.ExistsByUserIDAndName(ctx, uid, req.Name, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.ErrDuplicateName.WithMessage("同じ名前のキャラクターが既に存在します")
	}

	// ボイスの存在確認（アクティブなボイスのみ）
	voice, err := s.voiceRepo.FindActiveByID(ctx, req.VoiceID)
	if err != nil {
		return nil, err
	}

	// アバター画像の存在確認（指定された場合のみ）
	var avatar *model.Image
	var avatarID *uuid.UUID
	if req.AvatarID != nil && *req.AvatarID != "" {
		aid, err := uuid.Parse(*req.AvatarID)
		if err != nil {
			return nil, err
		}
		avatar, err = s.imageRepo.FindByID(ctx, aid)
		if err != nil {
			return nil, err
		}
		avatarID = &aid
	}

	// キャラクターを作成
	character := &model.Character{
		ID:       uuid.New(),
		UserID:   uid,
		Name:     req.Name,
		Persona:  req.Persona,
		AvatarID: avatarID,
		VoiceID:  voice.ID,
	}

	if err := s.characterRepo.Create(ctx, character); err != nil {
		return nil, err
	}

	// レスポンス用にリレーションを設定
	character.Voice = *voice
	if avatar != nil {
		character.Avatar = avatar
	}

	res, err := s.toCharacterWithChannelsResponse(ctx, *character)
	if err != nil {
		return nil, err
	}

	return &response.CharacterDataResponse{Data: res}, nil
}

// キャラクターを更新する
func (s *characterService) UpdateCharacter(ctx context.Context, userID, characterID string, req request.UpdateCharacterRequest) (*response.CharacterDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(characterID)
	if err != nil {
		return nil, err
	}

	character, err := s.characterRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	// 所有者チェック
	if character.UserID != uid {
		return nil, apperror.ErrNotFound.WithMessage("キャラクターが見つかりません")
	}

	// 名前の更新
	if req.Name != nil {
		// 名前が __ で始まる場合は禁止
		if strings.HasPrefix(*req.Name, "__") {
			return nil, apperror.ErrValidation.WithMessage("名前は '__' で始めることはできません")
		}

		// 同一ユーザー内で同じ名前のキャラクターが存在するかチェック（自分自身は除外）
		exists, err := s.characterRepo.ExistsByUserIDAndName(ctx, uid, *req.Name, &cid)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperror.ErrDuplicateName.WithMessage("同じ名前のキャラクターが既に存在します")
		}

		character.Name = *req.Name
	}

	// ペルソナの更新
	if req.Persona != nil {
		character.Persona = *req.Persona
	}

	// ボイスの更新
	if req.VoiceID != nil {
		voice, err := s.voiceRepo.FindActiveByID(ctx, *req.VoiceID)
		if err != nil {
			return nil, err
		}
		character.VoiceID = voice.ID
		character.Voice = *voice
	}

	// アバター画像の更新
	if req.AvatarID != nil {
		if *req.AvatarID == "" {
			// 空文字の場合はアバターを削除
			character.AvatarID = nil
			character.Avatar = nil
		} else {
			aid, err := uuid.Parse(*req.AvatarID)
			if err != nil {
				return nil, err
			}
			avatar, err := s.imageRepo.FindByID(ctx, aid)
			if err != nil {
				return nil, err
			}
			character.AvatarID = &aid
			character.Avatar = avatar
		}
	}

	if err := s.characterRepo.Update(ctx, character); err != nil {
		return nil, err
	}

	res, err := s.toCharacterWithChannelsResponse(ctx, *character)
	if err != nil {
		return nil, err
	}

	return &response.CharacterDataResponse{Data: res}, nil
}

// キャラクターを削除する
func (s *characterService) DeleteCharacter(ctx context.Context, userID, characterID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	cid, err := uuid.Parse(characterID)
	if err != nil {
		return err
	}

	character, err := s.characterRepo.FindByID(ctx, cid)
	if err != nil {
		return err
	}

	// 所有者チェック
	if character.UserID != uid {
		return apperror.ErrNotFound.WithMessage("キャラクターが見つかりません")
	}

	// いずれかのチャンネルで使用中かチェック
	inUse, err := s.characterRepo.IsUsedInAnyChannel(ctx, cid)
	if err != nil {
		return err
	}
	if inUse {
		return apperror.ErrCharacterInUse.WithMessage("このキャラクターは使用中のため削除できません")
	}

	return s.characterRepo.Delete(ctx, cid)
}
