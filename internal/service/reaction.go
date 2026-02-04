package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// ReactionService はリアクション関連のビジネスロジックインターフェースを表す
type ReactionService interface {
	ListLikes(ctx context.Context, userID string, limit, offset int) (*response.LikeListWithPaginationResponse, error)
	GetReactionStatus(ctx context.Context, userID, episodeID string) (*response.ReactionStatusDataResponse, error)
	CreateOrUpdateReaction(ctx context.Context, userID, episodeID, reactionType string) (*response.ReactionDataResponse, bool, error)
	DeleteReaction(ctx context.Context, userID, episodeID string) error
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

// GetReactionStatus は指定エピソードへのリアクション状態を返す
func (s *reactionService) GetReactionStatus(ctx context.Context, userID, episodeID string) (*response.ReactionStatusDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return nil, err
	}

	reaction, err := s.reactionRepo.FindByUserIDAndEpisodeID(ctx, uid, eid)
	if err != nil {
		if apperror.IsCode(err, apperror.CodeNotFound) {
			return &response.ReactionStatusDataResponse{
				Data: response.ReactionStatusResponse{
					ReactionType: nil,
				},
			}, nil
		}
		return nil, err
	}

	reactionType := string(reaction.ReactionType)
	return &response.ReactionStatusDataResponse{
		Data: response.ReactionStatusResponse{
			ReactionType: &reactionType,
		},
	}, nil
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

// CreateOrUpdateReaction はリアクションを登録または更新する
func (s *reactionService) CreateOrUpdateReaction(ctx context.Context, userID, episodeID, reactionType string) (*response.ReactionDataResponse, bool, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, false, err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return nil, false, err
	}

	reaction := &model.Reaction{
		UserID:       uid,
		EpisodeID:    eid,
		ReactionType: model.ReactionType(reactionType),
	}

	created, err := s.reactionRepo.Upsert(ctx, reaction)
	if err != nil {
		return nil, false, err
	}

	return &response.ReactionDataResponse{
		Data: response.ReactionResponse{
			ID:           reaction.ID,
			EpisodeID:    reaction.EpisodeID,
			ReactionType: string(reaction.ReactionType),
			CreatedAt:    reaction.CreatedAt,
		},
	}, created, nil
}

// DeleteReaction はリアクションを削除する
func (s *reactionService) DeleteReaction(ctx context.Context, userID, episodeID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return err
	}

	return s.reactionRepo.DeleteByUserIDAndEpisodeID(ctx, uid, eid)
}
