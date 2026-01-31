package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// ReactionService はリアクション関連のビジネスロジックインターフェースを表す
type ReactionService interface {
	ListLikes(ctx context.Context, userID string, limit, offset int) (*response.LikeListWithPaginationResponse, error)
}

type reactionService struct {
	reactionRepo  repository.ReactionRepository
	storageClient storage.Client
}

// NewReactionService は reactionService を生成して ReactionService として返す
func NewReactionService(
	reactionRepo repository.ReactionRepository,
	storageClient storage.Client,
) ReactionService {
	return &reactionService{
		reactionRepo:  reactionRepo,
		storageClient: storageClient,
	}
}

// ListLikes は高評価したエピソード一覧を取得する
func (s *reactionService) ListLikes(ctx context.Context, userID string, limit, offset int) (*response.LikeListWithPaginationResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	reactions, total, err := s.reactionRepo.FindLikesByUserID(ctx, uid, limit, offset)
	if err != nil {
		return nil, err
	}

	data := make([]response.LikeItemResponse, 0, len(reactions))
	for _, reaction := range reactions {
		item := s.toLikeItemResponse(ctx, &reaction)
		data = append(data, item)
	}

	return &response.LikeListWithPaginationResponse{
		Data: data,
		Pagination: response.PaginationResponse{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}, nil
}

// toLikeItemResponse は Reaction を LikeItemResponse に変換する
func (s *reactionService) toLikeItemResponse(ctx context.Context, reaction *model.Reaction) response.LikeItemResponse {
	episode := reaction.Episode
	channel := episode.Channel

	// チャンネルのアートワーク URL
	var channelArtwork *response.ArtworkResponse
	if channel.Artwork != nil {
		var artworkURL string
		if storage.IsExternalURL(channel.Artwork.Path) {
			artworkURL = channel.Artwork.Path
		} else {
			var err error
			artworkURL, err = s.storageClient.GenerateSignedURL(ctx, channel.Artwork.Path, storage.SignedURLExpirationImage)
			if err != nil {
				artworkURL = ""
			}
		}
		if artworkURL != "" {
			channelArtwork = &response.ArtworkResponse{
				ID:  channel.Artwork.ID,
				URL: artworkURL,
			}
		}
	}

	return response.LikeItemResponse{
		Episode: response.LikeEpisodeResponse{
			ID:          episode.ID,
			Title:       episode.Title,
			Description: episode.Description,
			Channel: response.LikeChannelResponse{
				ID:      channel.ID,
				Name:    channel.Name,
				Artwork: channelArtwork,
			},
			PublishedAt: episode.PublishedAt,
		},
		LikedAt: reaction.CreatedAt,
	}
}
