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

// エピソード関連のビジネスロジックインターフェース
type EpisodeService interface {
	ListMyChannelEpisodes(ctx context.Context, userID, channelID string, filter repository.EpisodeFilter) (*response.EpisodeListWithPaginationResponse, error)
	CreateEpisode(ctx context.Context, userID, channelID, title string, description *string, scriptPrompt string, artworkImageID, bgmAudioID *string) (*response.EpisodeResponse, error)
	UpdateEpisode(ctx context.Context, userID, channelID, episodeID string, req request.UpdateEpisodeRequest) (*response.EpisodeDataResponse, error)
	DeleteEpisode(ctx context.Context, userID, channelID, episodeID string) error
}

type episodeService struct {
	episodeRepo repository.EpisodeRepository
	channelRepo repository.ChannelRepository
}

// EpisodeService の実装を返す
func NewEpisodeService(
	episodeRepo repository.EpisodeRepository,
	channelRepo repository.ChannelRepository,
) EpisodeService {
	return &episodeService{
		episodeRepo: episodeRepo,
		channelRepo: channelRepo,
	}
}

// 自分のチャンネルのエピソード一覧を取得する
func (s *episodeService) ListMyChannelEpisodes(ctx context.Context, userID, channelID string, filter repository.EpisodeFilter) (*response.EpisodeListWithPaginationResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("You do not have permission to access this channel")
	}

	// エピソード一覧を取得
	episodes, total, err := s.episodeRepo.FindByChannelID(ctx, cid, filter)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeListWithPaginationResponse{
		Data:       toEpisodeResponses(episodes),
		Pagination: response.PaginationResponse{Total: total, Limit: filter.Limit, Offset: filter.Offset},
	}, nil
}

// エピソードを作成する
func (s *episodeService) CreateEpisode(ctx context.Context, userID, channelID, title string, description *string, scriptPrompt string, artworkImageID, bgmAudioID *string) (*response.EpisodeResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("You do not have permission to access this channel")
	}

	// エピソードを作成
	episode := &model.Episode{
		ChannelID:    cid,
		Title:        title,
		Description:  description,
		ScriptPrompt: scriptPrompt,
	}

	// アートワークが指定されている場合
	if artworkImageID != nil {
		artworkID, err := uuid.Parse(*artworkImageID)
		if err != nil {
			return nil, err
		}
		episode.ArtworkID = &artworkID
	}

	// BGM が指定されている場合
	if bgmAudioID != nil {
		bgmID, err := uuid.Parse(*bgmAudioID)
		if err != nil {
			return nil, err
		}
		episode.BgmID = &bgmID
	}

	if err := s.episodeRepo.Create(ctx, episode); err != nil {
		return nil, err
	}

	return &response.EpisodeResponse{
		ID:           episode.ID,
		Title:        episode.Title,
		Description:  episode.Description,
		ScriptPrompt: episode.ScriptPrompt,
		PublishedAt:  episode.PublishedAt,
		CreatedAt:    episode.CreatedAt,
		UpdatedAt:    episode.UpdatedAt,
	}, nil
}

// エピソードを更新する
func (s *episodeService) UpdateEpisode(ctx context.Context, userID, channelID, episodeID string, req request.UpdateEpisodeRequest) (*response.EpisodeDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return nil, err
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	if channel.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("You do not have permission to update this episode")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("Episode not found in this channel")
	}

	// 各フィールドを更新（指定されたもののみ）
	if req.Title != nil {
		episode.Title = *req.Title
	}
	if req.Description != nil {
		episode.Description = req.Description
	}
	if req.ScriptPrompt != nil {
		episode.ScriptPrompt = *req.ScriptPrompt
	}

	// アートワークの更新
	if req.ArtworkImageID != nil {
		if *req.ArtworkImageID == "" {
			// 空文字の場合は null に設定
			episode.ArtworkID = nil
		} else {
			artworkID, err := uuid.Parse(*req.ArtworkImageID)
			if err != nil {
				return nil, err
			}
			episode.ArtworkID = &artworkID
		}
	}

	// BGM の更新
	if req.BgmAudioID != nil {
		if *req.BgmAudioID == "" {
			// 空文字の場合は null に設定
			episode.BgmID = nil
		} else {
			bgmID, err := uuid.Parse(*req.BgmAudioID)
			if err != nil {
				return nil, err
			}
			episode.BgmID = &bgmID
		}
	}

	// 公開日時の更新
	if req.PublishedAt != nil {
		if *req.PublishedAt == "" {
			// 空文字の場合は null に設定（非公開化）
			episode.PublishedAt = nil
		} else {
			publishedAt, err := time.Parse(time.RFC3339, *req.PublishedAt)
			if err != nil {
				return nil, apperror.ErrValidation.WithMessage("Invalid publishedAt format. Use RFC3339 format.")
			}
			episode.PublishedAt = &publishedAt
		}
	}

	// エピソードを更新
	if err := s.episodeRepo.Update(ctx, episode); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.episodeRepo.FindByID(ctx, episode.ID)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeDataResponse{
		Data: toEpisodeResponse(updated),
	}, nil
}

// エピソードを削除する
func (s *episodeService) DeleteEpisode(ctx context.Context, userID, channelID, episodeID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return err
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return err
	}

	if channel.UserID != uid {
		return apperror.ErrForbidden.WithMessage("You do not have permission to delete this episode")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return err
	}

	if episode.ChannelID != cid {
		return apperror.ErrNotFound.WithMessage("Episode not found in this channel")
	}

	// エピソードを削除
	return s.episodeRepo.Delete(ctx, eid)
}

// Episode モデルのスライスをレスポンス DTO のスライスに変換する
func toEpisodeResponses(episodes []model.Episode) []response.EpisodeResponse {
	result := make([]response.EpisodeResponse, len(episodes))

	for i, e := range episodes {
		result[i] = toEpisodeResponse(&e)
	}

	return result
}

// Episode モデルをレスポンス DTO に変換する
func toEpisodeResponse(e *model.Episode) response.EpisodeResponse {
	resp := response.EpisodeResponse{
		ID:           e.ID,
		Title:        e.Title,
		Description:  e.Description,
		ScriptPrompt: e.ScriptPrompt,
		PublishedAt:  e.PublishedAt,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}

	if e.Artwork != nil {
		resp.Artwork = &response.ArtworkResponse{
			ID:  e.Artwork.ID,
			URL: e.Artwork.URL,
		}
	}

	if e.FullAudio != nil {
		resp.FullAudio = &response.AudioResponse{
			ID:         e.FullAudio.ID,
			URL:        e.FullAudio.URL,
			DurationMs: e.FullAudio.DurationMs,
		}
	}

	return resp
}
