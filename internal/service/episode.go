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

// エピソード関連のビジネスロジックインターフェース
type EpisodeService interface {
	GetMyChannelEpisode(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error)
	ListMyChannelEpisodes(ctx context.Context, userID, channelID string, filter repository.EpisodeFilter) (*response.EpisodeListWithPaginationResponse, error)
	CreateEpisode(ctx context.Context, userID, channelID, title, description string, artworkImageID *string) (*response.EpisodeResponse, error)
	UpdateEpisode(ctx context.Context, userID, channelID, episodeID string, req request.UpdateEpisodeRequest) (*response.EpisodeDataResponse, error)
	DeleteEpisode(ctx context.Context, userID, channelID, episodeID string) error
	PublishEpisode(ctx context.Context, userID, channelID, episodeID string, publishedAt *string) (*response.EpisodeDataResponse, error)
	UnpublishEpisode(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error)
	SetEpisodeBgm(ctx context.Context, userID, channelID, episodeID, bgmAudioID string) (*response.EpisodeDataResponse, error)
	RemoveEpisodeBgm(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error)
}

type episodeService struct {
	episodeRepo   repository.EpisodeRepository
	channelRepo   repository.ChannelRepository
	storageClient storage.Client
}

// EpisodeService の実装を返す
func NewEpisodeService(
	episodeRepo repository.EpisodeRepository,
	channelRepo repository.ChannelRepository,
	storageClient storage.Client,
) EpisodeService {
	return &episodeService{
		episodeRepo:   episodeRepo,
		channelRepo:   channelRepo,
		storageClient: storageClient,
	}
}

// 自分のチャンネルのエピソードを取得する
func (s *episodeService) GetMyChannelEpisode(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("You do not have permission to access this channel")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("Episode not found in this channel")
	}

	resp, err := s.toEpisodeResponse(ctx, episode)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeDataResponse{
		Data: resp,
	}, nil
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

	responses, err := s.toEpisodeResponses(ctx, episodes)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeListWithPaginationResponse{
		Data:       responses,
		Pagination: response.PaginationResponse{Total: total, Limit: filter.Limit, Offset: filter.Offset},
	}, nil
}

// エピソードを作成する
func (s *episodeService) CreateEpisode(ctx context.Context, userID, channelID, title, description string, artworkImageID *string) (*response.EpisodeResponse, error) {
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
		ChannelID:   cid,
		Title:       title,
		Description: description,
	}

	// アートワークが指定されている場合
	if artworkImageID != nil {
		artworkID, err := uuid.Parse(*artworkImageID)
		if err != nil {
			return nil, err
		}
		episode.ArtworkID = &artworkID
	}

	if err := s.episodeRepo.Create(ctx, episode); err != nil {
		return nil, err
	}

	return &response.EpisodeResponse{
		ID:          episode.ID,
		Title:       episode.Title,
		Description: episode.Description,
		UserPrompt:  episode.UserPrompt,
		PublishedAt: episode.PublishedAt,
		CreatedAt:   episode.CreatedAt,
		UpdatedAt:   episode.UpdatedAt,
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

	// 各フィールドを更新
	episode.Title = req.Title
	episode.Description = req.Description

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
		episode.Artwork = nil
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

	resp, err := s.toEpisodeResponse(ctx, updated)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeDataResponse{
		Data: resp,
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

// エピソードを公開する
func (s *episodeService) PublishEpisode(ctx context.Context, userID, channelID, episodeID string, publishedAt *string) (*response.EpisodeDataResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("You do not have permission to publish this episode")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("Episode not found in this channel")
	}

	// 公開日時を設定
	if publishedAt == nil || *publishedAt == "" {
		// 省略時は現在時刻で即時公開
		now := time.Now()
		episode.PublishedAt = &now
	} else {
		// 指定された日時でパース
		parsedTime, err := time.Parse(time.RFC3339, *publishedAt)
		if err != nil {
			return nil, apperror.ErrValidation.WithMessage("Invalid publishedAt format. Use RFC3339 format.")
		}
		episode.PublishedAt = &parsedTime
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

	resp, err := s.toEpisodeResponse(ctx, updated)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeDataResponse{
		Data: resp,
	}, nil
}

// エピソードを非公開にする
func (s *episodeService) UnpublishEpisode(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("You do not have permission to unpublish this episode")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("Episode not found in this channel")
	}

	// 公開日時を null に設定（非公開化）
	episode.PublishedAt = nil

	// エピソードを更新
	if err := s.episodeRepo.Update(ctx, episode); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.episodeRepo.FindByID(ctx, episode.ID)
	if err != nil {
		return nil, err
	}

	resp, err := s.toEpisodeResponse(ctx, updated)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeDataResponse{
		Data: resp,
	}, nil
}

// Episode モデルのスライスをレスポンス DTO のスライスに変換する
func (s *episodeService) toEpisodeResponses(ctx context.Context, episodes []model.Episode) ([]response.EpisodeResponse, error) {
	result := make([]response.EpisodeResponse, len(episodes))

	for i, e := range episodes {
		resp, err := s.toEpisodeResponse(ctx, &e)
		if err != nil {
			return nil, err
		}
		result[i] = resp
	}

	return result, nil
}

// Episode モデルをレスポンス DTO に変換する
func (s *episodeService) toEpisodeResponse(ctx context.Context, e *model.Episode) (response.EpisodeResponse, error) {
	resp := response.EpisodeResponse{
		ID:          e.ID,
		Title:       e.Title,
		Description: e.Description,
		UserPrompt:  e.UserPrompt,
		PublishedAt: e.PublishedAt,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}

	if e.Artwork != nil {
		signedURL, err := s.storageClient.GenerateSignedURL(ctx, e.Artwork.Path, signedURLExpiration)
		if err != nil {
			return response.EpisodeResponse{}, err
		}
		resp.Artwork = &response.ArtworkResponse{
			ID:  e.Artwork.ID,
			URL: signedURL,
		}
	}

	if e.FullAudio != nil {
		signedURL, err := s.storageClient.GenerateSignedURL(ctx, e.FullAudio.Path, signedURLExpiration)
		if err != nil {
			return response.EpisodeResponse{}, err
		}
		resp.FullAudio = &response.AudioResponse{
			ID:         e.FullAudio.ID,
			URL:        signedURL,
			MimeType:   e.FullAudio.MimeType,
			FileSize:   e.FullAudio.FileSize,
			DurationMs: e.FullAudio.DurationMs,
		}
	}

	return resp, nil
}

// エピソードに BGM を設定する
func (s *episodeService) SetEpisodeBgm(ctx context.Context, userID, channelID, episodeID, bgmAudioID string) (*response.EpisodeDataResponse, error) {
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

	bgmID, err := uuid.Parse(bgmAudioID)
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

	// BGM を設定
	episode.BgmID = &bgmID
	episode.Bgm = nil

	// エピソードを更新
	if err := s.episodeRepo.Update(ctx, episode); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.episodeRepo.FindByID(ctx, episode.ID)
	if err != nil {
		return nil, err
	}

	resp, err := s.toEpisodeResponse(ctx, updated)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeDataResponse{
		Data: resp,
	}, nil
}

// エピソードの BGM を削除する
func (s *episodeService) RemoveEpisodeBgm(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error) {
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

	// BGM を削除
	episode.BgmID = nil
	episode.Bgm = nil

	// エピソードを更新
	if err := s.episodeRepo.Update(ctx, episode); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.episodeRepo.FindByID(ctx, episode.ID)
	if err != nil {
		return nil, err
	}

	resp, err := s.toEpisodeResponse(ctx, updated)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeDataResponse{
		Data: resp,
	}, nil
}
