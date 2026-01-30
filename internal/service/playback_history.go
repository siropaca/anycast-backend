package service

import (
	"context"
	"time"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// PlaybackHistoryService は再生履歴関連のビジネスロジックインターフェースを表す
type PlaybackHistoryService interface {
	ListPlaybackHistory(ctx context.Context, userID string, completed *bool, limit, offset int) (*response.PlaybackHistoryListWithPaginationResponse, error)
	UpdatePlayback(ctx context.Context, userID, episodeID string, req request.UpdatePlaybackRequest) (*response.PlaybackDataResponse, error)
	DeletePlayback(ctx context.Context, userID, episodeID string) error
}

type playbackHistoryService struct {
	playbackHistoryRepo repository.PlaybackHistoryRepository
	episodeRepo         repository.EpisodeRepository
	storageClient       storage.Client
}

// NewPlaybackHistoryService は playbackHistoryService を生成して PlaybackHistoryService として返す
func NewPlaybackHistoryService(
	playbackHistoryRepo repository.PlaybackHistoryRepository,
	episodeRepo repository.EpisodeRepository,
	storageClient storage.Client,
) PlaybackHistoryService {
	return &playbackHistoryService{
		playbackHistoryRepo: playbackHistoryRepo,
		episodeRepo:         episodeRepo,
		storageClient:       storageClient,
	}
}

// ListPlaybackHistory は自分の再生履歴一覧を取得する
func (s *playbackHistoryService) ListPlaybackHistory(ctx context.Context, userID string, completed *bool, limit, offset int) (*response.PlaybackHistoryListWithPaginationResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	histories, total, err := s.playbackHistoryRepo.FindByUserID(ctx, uid, completed, limit, offset)
	if err != nil {
		return nil, err
	}

	// レスポンスに変換
	data := make([]response.PlaybackHistoryItemResponse, 0, len(histories))
	for _, history := range histories {
		item := s.toPlaybackHistoryItemResponse(ctx, &history)
		data = append(data, item)
	}

	return &response.PlaybackHistoryListWithPaginationResponse{
		Data: data,
		Pagination: response.PaginationResponse{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}, nil
}

// UpdatePlayback は再生履歴を更新する（なければ作成）
func (s *playbackHistoryService) UpdatePlayback(ctx context.Context, userID, episodeID string, req request.UpdatePlaybackRequest) (*response.PlaybackDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return nil, err
	}

	// エピソードが存在するかチェック
	_, err = s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	// 既存の再生履歴を取得、なければ新規作成
	history, err := s.playbackHistoryRepo.FindByUserIDAndEpisodeID(ctx, uid, eid)
	if err != nil {
		// 見つからない場合は新規作成
		if apperror.IsCode(err, apperror.CodeNotFound) {
			history = &model.PlaybackHistory{
				UserID:     uid,
				EpisodeID:  eid,
				ProgressMs: 0,
				Completed:  false,
				PlayedAt:   time.Now(),
			}
		} else {
			return nil, err
		}
	}

	// リクエストに応じて更新
	if req.ProgressMs != nil {
		history.ProgressMs = *req.ProgressMs
	}
	if req.Completed != nil {
		history.Completed = *req.Completed
	}
	history.PlayedAt = time.Now()

	if err := s.playbackHistoryRepo.Upsert(ctx, history); err != nil {
		return nil, err
	}

	return &response.PlaybackDataResponse{
		Data: response.PlaybackResponse{
			EpisodeID:  history.EpisodeID,
			ProgressMs: history.ProgressMs,
			Completed:  history.Completed,
			PlayedAt:   history.PlayedAt,
		},
	}, nil
}

// DeletePlayback は再生履歴を削除する
func (s *playbackHistoryService) DeletePlayback(ctx context.Context, userID, episodeID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return err
	}

	// 既存の再生履歴を取得
	history, err := s.playbackHistoryRepo.FindByUserIDAndEpisodeID(ctx, uid, eid)
	if err != nil {
		return err
	}

	return s.playbackHistoryRepo.Delete(ctx, history.ID)
}

// toPlaybackHistoryItemResponse は PlaybackHistory を PlaybackHistoryItemResponse に変換する
func (s *playbackHistoryService) toPlaybackHistoryItemResponse(ctx context.Context, history *model.PlaybackHistory) response.PlaybackHistoryItemResponse {
	episode := history.Episode
	channel := episode.Channel

	// チャンネルのアートワーク URL
	var channelArtwork *response.ArtworkResponse
	if channel.Artwork != nil {
		var chArtworkURL string
		if storage.IsExternalURL(channel.Artwork.Path) {
			chArtworkURL = channel.Artwork.Path
		} else {
			var err error
			chArtworkURL, err = s.storageClient.GenerateSignedURL(ctx, channel.Artwork.Path, storage.SignedURLExpirationImage)
			if err != nil {
				chArtworkURL = ""
			}
		}
		if chArtworkURL != "" {
			channelArtwork = &response.ArtworkResponse{
				ID:  channel.Artwork.ID,
				URL: chArtworkURL,
			}
		}
	}

	// エピソードの音声 URL
	var fullAudio *response.AudioResponse
	if episode.FullAudio != nil {
		url, err := s.storageClient.GenerateSignedURL(ctx, episode.FullAudio.Path, storage.SignedURLExpirationAudio)
		if err == nil {
			fullAudio = &response.AudioResponse{
				ID:         episode.FullAudio.ID,
				URL:        url,
				DurationMs: episode.FullAudio.DurationMs,
			}
		}
	}

	return response.PlaybackHistoryItemResponse{
		Episode: response.PlaybackHistoryEpisodeResponse{
			ID:          episode.ID,
			Title:       episode.Title,
			Description: episode.Description,
			FullAudio:   fullAudio,
			Channel: response.PlaybackHistoryChannelResponse{
				ID:      channel.ID,
				Name:    channel.Name,
				Artwork: channelArtwork,
			},
			PublishedAt: episode.PublishedAt,
		},
		ProgressMs: history.ProgressMs,
		Completed:  history.Completed,
		PlayedAt:   history.PlayedAt,
	}
}
