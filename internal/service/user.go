package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// UserService はユーザー関連のビジネスロジックインターフェースを表す
type UserService interface {
	GetUser(ctx context.Context, username string) (*response.PublicUserDataResponse, error)
}

type userService struct {
	userRepo      repository.UserRepository
	channelRepo   repository.ChannelRepository
	storageClient storage.Client
}

// NewUserService は userService を生成して UserService として返す
func NewUserService(
	userRepo repository.UserRepository,
	channelRepo repository.ChannelRepository,
	storageClient storage.Client,
) UserService {
	return &userService{
		userRepo:      userRepo,
		channelRepo:   channelRepo,
		storageClient: storageClient,
	}
}

// GetUser は指定されたユーザー名のユーザーの公開プロフィールを取得する
func (s *userService) GetUser(ctx context.Context, username string) (*response.PublicUserDataResponse, error) {
	user, err := s.userRepo.FindByUsernameWithAvatar(ctx, username)
	if err != nil {
		return nil, err
	}

	channels, err := s.channelRepo.FindPublishedByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// アバターの署名付き URL を生成
	var avatar *response.AvatarResponse
	if user.Avatar != nil {
		avatarURL, err := s.generateImageURL(ctx, user.Avatar)
		if err != nil {
			return nil, err
		}
		avatar = &response.AvatarResponse{
			ID:  user.Avatar.ID,
			URL: avatarURL,
		}
	}

	// チャンネル一覧のレスポンスを構築
	channelResponses, err := s.toPublicUserChannelResponses(ctx, channels)
	if err != nil {
		return nil, err
	}

	return &response.PublicUserDataResponse{
		Data: response.PublicUserResponse{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Avatar:      avatar,
			Channels:    channelResponses,
			CreatedAt:   user.CreatedAt,
		},
	}, nil
}

// toPublicUserChannelResponses は Channel のスライスを公開ユーザー用チャンネルレスポンスのスライスに変換する
func (s *userService) toPublicUserChannelResponses(ctx context.Context, channels []model.Channel) ([]response.PublicUserChannelResponse, error) {
	result := make([]response.PublicUserChannelResponse, len(channels))

	for i, c := range channels {
		resp := response.PublicUserChannelResponse{
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
			artworkURL, err := s.generateImageURL(ctx, c.Artwork)
			if err != nil {
				return nil, err
			}
			resp.Artwork = &response.ArtworkResponse{
				ID:  c.Artwork.ID,
				URL: artworkURL,
			}
		}

		result[i] = resp
	}

	return result, nil
}

// generateImageURL は画像の署名付き URL を生成する
func (s *userService) generateImageURL(ctx context.Context, image *model.Image) (string, error) {
	if storage.IsExternalURL(image.Path) {
		return image.Path, nil
	}
	return s.storageClient.GenerateSignedURL(ctx, image.Path, storage.SignedURLExpirationImage)
}
