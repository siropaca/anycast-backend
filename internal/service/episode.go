package service

import (
	"context"
	"time"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/infrastructure/tts"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// EpisodeService はエピソード関連のビジネスロジックインターフェースを表す
type EpisodeService interface {
	GetEpisode(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error)
	GetMyChannelEpisode(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error)
	ListChannelEpisodes(ctx context.Context, userID, channelID string, filter repository.EpisodeFilter) (*response.EpisodeListWithPaginationResponse, error)
	ListMyChannelEpisodes(ctx context.Context, userID, channelID string, filter repository.EpisodeFilter) (*response.EpisodeListWithPaginationResponse, error)
	CreateEpisode(ctx context.Context, userID, channelID, title, description string, artworkImageID *string) (*response.EpisodeResponse, error)
	UpdateEpisode(ctx context.Context, userID, channelID, episodeID string, req request.UpdateEpisodeRequest) (*response.EpisodeDataResponse, error)
	DeleteEpisode(ctx context.Context, userID, channelID, episodeID string) error
	PublishEpisode(ctx context.Context, userID, channelID, episodeID string, publishedAt *string) (*response.EpisodeDataResponse, error)
	UnpublishEpisode(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error)
	SetEpisodeBgm(ctx context.Context, userID, channelID, episodeID string, req request.SetEpisodeBgmRequest) (*response.EpisodeDataResponse, error)
	DeleteEpisodeBgm(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error)
	IncrementPlayCount(ctx context.Context, episodeID string) error
}

type episodeService struct {
	episodeRepo         repository.EpisodeRepository
	channelRepo         repository.ChannelRepository
	scriptLineRepo      repository.ScriptLineRepository
	audioRepo           repository.AudioRepository
	imageRepo           repository.ImageRepository
	bgmRepo             repository.BgmRepository
	systemBgmRepo       repository.SystemBgmRepository
	playbackHistoryRepo repository.PlaybackHistoryRepository
	playlistRepo        repository.PlaylistRepository
	storageClient       storage.Client
	ttsClient           tts.Client
}

// NewEpisodeService は episodeService を生成して EpisodeService として返す
func NewEpisodeService(
	episodeRepo repository.EpisodeRepository,
	channelRepo repository.ChannelRepository,
	scriptLineRepo repository.ScriptLineRepository,
	audioRepo repository.AudioRepository,
	imageRepo repository.ImageRepository,
	bgmRepo repository.BgmRepository,
	systemBgmRepo repository.SystemBgmRepository,
	playbackHistoryRepo repository.PlaybackHistoryRepository,
	playlistRepo repository.PlaylistRepository,
	storageClient storage.Client,
	ttsClient tts.Client,
) EpisodeService {
	return &episodeService{
		episodeRepo:         episodeRepo,
		channelRepo:         channelRepo,
		scriptLineRepo:      scriptLineRepo,
		audioRepo:           audioRepo,
		imageRepo:           imageRepo,
		bgmRepo:             bgmRepo,
		systemBgmRepo:       systemBgmRepo,
		playbackHistoryRepo: playbackHistoryRepo,
		playlistRepo:        playlistRepo,
		storageClient:       storageClient,
		ttsClient:           ttsClient,
	}
}

// GetEpisode は指定されたエピソードを取得する
func (s *episodeService) GetEpisode(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error) {
	var uid uuid.UUID
	if userID != "" {
		var err error
		uid, err = uuid.Parse(userID)
		if err != nil {
			return nil, err
		}
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return nil, err
	}

	// チャンネルの存在確認
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	isOwner := userID != "" && channel.UserID == uid
	isChannelPublished := channel.PublishedAt != nil && !channel.PublishedAt.After(time.Now())

	// オーナーでなく、かつチャンネルが公開されていない場合は 404
	if !isOwner && !isChannelPublished {
		return nil, apperror.ErrNotFound.WithMessage("チャンネルが見つかりません")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("このチャンネルにエピソードが見つかりません")
	}

	// 非オーナーの場合、エピソードの公開状態チェック
	if !isOwner {
		isEpisodePublished := episode.PublishedAt != nil && !episode.PublishedAt.After(time.Now())
		if !isEpisodePublished {
			return nil, apperror.ErrNotFound.WithMessage("エピソードが見つかりません")
		}
	}

	// 認証済みの場合は再生履歴を取得
	var playback *model.PlaybackHistory
	if userID != "" {
		ph, err := s.playbackHistoryRepo.FindByUserIDAndEpisodeID(ctx, uid, eid)
		if err == nil {
			playback = ph
		}
	}

	resp, err := s.toEpisodeResponse(ctx, episode, &channel.User, playback)
	if err != nil {
		return nil, err
	}

	// 認証済みの場合は再生リスト所属情報を取得
	if userID != "" {
		playlistIDs, err := s.playlistRepo.FindPlaylistIDsByUserIDAndEpisodeID(ctx, uid, eid)
		if err == nil {
			if playlistIDs == nil {
				playlistIDs = []uuid.UUID{}
			}
			resp.PlaylistIDs = playlistIDs
		}
	}

	return &response.EpisodeDataResponse{
		Data: resp,
	}, nil
}

// GetMyChannelEpisode は自分のチャンネルのエピソードを取得する
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
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルへのアクセス権限がありません")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("このチャンネルにエピソードが見つかりません")
	}

	// 再生履歴を取得
	var playback *model.PlaybackHistory
	ph, err := s.playbackHistoryRepo.FindByUserIDAndEpisodeID(ctx, uid, eid)
	if err == nil {
		playback = ph
	}

	resp, err := s.toEpisodeResponse(ctx, episode, &channel.User, playback)
	if err != nil {
		return nil, err
	}

	// 再生リスト所属情報を取得
	playlistIDs, err := s.playlistRepo.FindPlaylistIDsByUserIDAndEpisodeID(ctx, uid, eid)
	if err == nil {
		if playlistIDs == nil {
			playlistIDs = []uuid.UUID{}
		}
		resp.PlaylistIDs = playlistIDs
	}

	return &response.EpisodeDataResponse{
		Data: resp,
	}, nil
}

// ListMyChannelEpisodes は自分のチャンネルのエピソード一覧を取得する
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
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルへのアクセス権限がありません")
	}

	// エピソード一覧を取得
	episodes, total, err := s.episodeRepo.FindByChannelID(ctx, cid, filter)
	if err != nil {
		return nil, err
	}

	responses, err := s.toEpisodeResponses(ctx, episodes, &channel.User)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeListWithPaginationResponse{
		Data:       responses,
		Pagination: response.PaginationResponse{Total: total, Limit: filter.Limit, Offset: filter.Offset},
	}, nil
}

// ListChannelEpisodes は指定されたチャンネルのエピソード一覧を取得する
func (s *episodeService) ListChannelEpisodes(ctx context.Context, userID, channelID string, filter repository.EpisodeFilter) (*response.EpisodeListWithPaginationResponse, error) {
	var uid uuid.UUID
	if userID != "" {
		var err error
		uid, err = uuid.Parse(userID)
		if err != nil {
			return nil, err
		}
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}

	// チャンネルの存在確認
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	isOwner := userID != "" && channel.UserID == uid
	isChannelPublished := channel.PublishedAt != nil && !channel.PublishedAt.After(time.Now())

	// オーナーでなく、かつチャンネルが公開されていない場合は 404
	if !isOwner && !isChannelPublished {
		return nil, apperror.ErrNotFound.WithMessage("チャンネルが見つかりません")
	}

	// 非オーナーの場合は公開済みエピソードのみ返す
	if !isOwner {
		published := "published"
		filter.Status = &published
	}

	// エピソード一覧を取得
	episodes, total, err := s.episodeRepo.FindByChannelID(ctx, cid, filter)
	if err != nil {
		return nil, err
	}

	responses, err := s.toEpisodeResponses(ctx, episodes, &channel.User)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeListWithPaginationResponse{
		Data:       responses,
		Pagination: response.PaginationResponse{Total: total, Limit: filter.Limit, Offset: filter.Offset},
	}, nil
}

// CreateEpisode は新しいエピソードを作成する
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
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルへのアクセス権限がありません")
	}

	// エピソードを作成
	episode := &model.Episode{
		ChannelID:   cid,
		Title:       title,
		Description: description,
	}

	// チャンネルのデフォルト BGM を継承
	if channel.DefaultBgmID != nil {
		episode.BgmID = channel.DefaultBgmID
	} else if channel.DefaultSystemBgmID != nil {
		episode.SystemBgmID = channel.DefaultSystemBgmID
	}

	// アートワークが指定されている場合
	if artworkImageID != nil {
		artworkID, err := uuid.Parse(*artworkImageID)
		if err != nil {
			return nil, err
		}
		// 画像の存在確認
		if _, err := s.imageRepo.FindByID(ctx, artworkID); err != nil {
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
		VoiceStyle:  episode.VoiceStyle,
		PublishedAt: episode.PublishedAt,
		CreatedAt:   episode.CreatedAt,
		UpdatedAt:   episode.UpdatedAt,
	}, nil
}

// UpdateEpisode は指定されたエピソードを更新する
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
		return nil, apperror.ErrForbidden.WithMessage("このエピソードの更新権限がありません")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("このチャンネルにエピソードが見つかりません")
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
			// 画像の存在確認
			if _, err := s.imageRepo.FindByID(ctx, artworkID); err != nil {
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

	resp, err := s.toEpisodeResponse(ctx, updated, &channel.User, nil)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeDataResponse{
		Data: resp,
	}, nil
}

// DeleteEpisode は指定されたエピソードを削除する
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
		return apperror.ErrForbidden.WithMessage("このエピソードの削除権限がありません")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return err
	}

	if episode.ChannelID != cid {
		return apperror.ErrNotFound.WithMessage("このチャンネルにエピソードが見つかりません")
	}

	// 削除前に GCS ファイルのパスを収集
	var filesToDelete []string
	if episode.Artwork != nil {
		filesToDelete = append(filesToDelete, episode.Artwork.Path)
	}
	if episode.FullAudio != nil {
		filesToDelete = append(filesToDelete, episode.FullAudio.Path)
	}
	if episode.Bgm != nil && episode.Bgm.Audio.ID != uuid.Nil {
		filesToDelete = append(filesToDelete, episode.Bgm.Audio.Path)
	}

	// エピソードを削除
	if err := s.episodeRepo.Delete(ctx, eid); err != nil {
		return err
	}

	// GCS からファイルを削除（失敗してもログを出すだけで続行）
	for _, path := range filesToDelete {
		if err := s.storageClient.Delete(ctx, path); err != nil {
			logger.FromContext(ctx).Warn("failed to delete file from storage", "path", path, "error", err)
		}
	}

	return nil
}

// PublishEpisode は指定されたエピソードを公開する
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
		return nil, apperror.ErrForbidden.WithMessage("このエピソードの公開権限がありません")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("このチャンネルにエピソードが見つかりません")
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
			return nil, apperror.ErrValidation.WithMessage("公開日時の形式が無効です。RFC3339 形式で指定してください")
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

	resp, err := s.toEpisodeResponse(ctx, updated, &channel.User, nil)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeDataResponse{
		Data: resp,
	}, nil
}

// UnpublishEpisode は指定されたエピソードを非公開にする
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
		return nil, apperror.ErrForbidden.WithMessage("このエピソードの非公開権限がありません")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("このチャンネルにエピソードが見つかりません")
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

	resp, err := s.toEpisodeResponse(ctx, updated, &channel.User, nil)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeDataResponse{
		Data: resp,
	}, nil
}

// IncrementPlayCount は指定されたエピソードの再生回数をインクリメントする
func (s *episodeService) IncrementPlayCount(ctx context.Context, episodeID string) error {
	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return err
	}

	// エピソードの存在確認と公開状態チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return err
	}

	// 公開中のエピソードのみカウント対象
	if episode.PublishedAt == nil || episode.PublishedAt.After(time.Now()) {
		return apperror.ErrValidation.WithMessage("公開中のエピソードのみ再生回数をカウントできます")
	}

	return s.episodeRepo.IncrementPlayCount(ctx, eid)
}

// toChannelOwnerResponse は User からチャンネルオーナーレスポンスを生成する
func (s *episodeService) toChannelOwnerResponse(ctx context.Context, user *model.User) (response.ChannelOwnerResponse, error) {
	ownerResp := response.ChannelOwnerResponse{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
	}

	if user.Avatar != nil {
		var avatarURL string
		if storage.IsExternalURL(user.Avatar.Path) {
			avatarURL = user.Avatar.Path
		} else {
			var err error
			avatarURL, err = s.storageClient.GenerateSignedURL(ctx, user.Avatar.Path, storage.SignedURLExpirationImage)
			if err != nil {
				return response.ChannelOwnerResponse{}, err
			}
		}
		ownerResp.Avatar = &response.AvatarResponse{
			ID:  user.Avatar.ID,
			URL: avatarURL,
		}
	}

	return ownerResp, nil
}

// toEpisodeResponses は Episode のスライスをレスポンス DTO のスライスに変換する
func (s *episodeService) toEpisodeResponses(ctx context.Context, episodes []model.Episode, owner *model.User) ([]response.EpisodeResponse, error) {
	result := make([]response.EpisodeResponse, len(episodes))

	for i, e := range episodes {
		resp, err := s.toEpisodeResponse(ctx, &e, owner, nil)
		if err != nil {
			return nil, err
		}
		result[i] = resp
	}

	return result, nil
}

// toEpisodeResponse は Episode をレスポンス DTO に変換する
func (s *episodeService) toEpisodeResponse(ctx context.Context, e *model.Episode, owner *model.User, playback *model.PlaybackHistory) (response.EpisodeResponse, error) {
	ownerResp, err := s.toChannelOwnerResponse(ctx, owner)
	if err != nil {
		return response.EpisodeResponse{}, err
	}

	resp := response.EpisodeResponse{
		ID:            e.ID,
		Owner:         ownerResp,
		Title:         e.Title,
		Description:   e.Description,
		VoiceStyle:    e.VoiceStyle,
		AudioOutdated: e.AudioOutdated,
		PlayCount:     e.PlayCount,
		PublishedAt:   e.PublishedAt,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}

	if e.Artwork != nil {
		var artworkURL string
		if storage.IsExternalURL(e.Artwork.Path) {
			artworkURL = e.Artwork.Path
		} else {
			var err error
			artworkURL, err = s.storageClient.GenerateSignedURL(ctx, e.Artwork.Path, storage.SignedURLExpirationImage)
			if err != nil {
				return response.EpisodeResponse{}, err
			}
		}
		resp.Artwork = &response.ArtworkResponse{
			ID:  e.Artwork.ID,
			URL: artworkURL,
		}
	}

	if e.FullAudio != nil {
		signedURL, err := s.storageClient.GenerateSignedURL(ctx, e.FullAudio.Path, storage.SignedURLExpirationAudio)
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

	// Bgm または SystemBgm からレスポンスを構築
	if e.Bgm != nil && e.Bgm.Audio.ID != uuid.Nil {
		signedURL, err := s.storageClient.GenerateSignedURL(ctx, e.Bgm.Audio.Path, storage.SignedURLExpirationAudio)
		if err != nil {
			return response.EpisodeResponse{}, err
		}
		resp.Bgm = &response.EpisodeBgmResponse{
			ID:       e.Bgm.ID,
			Name:     e.Bgm.Name,
			IsSystem: false,
			Audio: response.BgmAudioResponse{
				ID:         e.Bgm.Audio.ID,
				URL:        signedURL,
				DurationMs: e.Bgm.Audio.DurationMs,
			},
		}
	} else if e.SystemBgm != nil && e.SystemBgm.Audio.ID != uuid.Nil {
		signedURL, err := s.storageClient.GenerateSignedURL(ctx, e.SystemBgm.Audio.Path, storage.SignedURLExpirationAudio)
		if err != nil {
			return response.EpisodeResponse{}, err
		}
		resp.Bgm = &response.EpisodeBgmResponse{
			ID:       e.SystemBgm.ID,
			Name:     e.SystemBgm.Name,
			IsSystem: true,
			Audio: response.BgmAudioResponse{
				ID:         e.SystemBgm.Audio.ID,
				URL:        signedURL,
				DurationMs: e.SystemBgm.Audio.DurationMs,
			},
		}
	}

	if playback != nil {
		resp.Playback = &response.EpisodePlaybackResponse{
			ProgressMs: playback.ProgressMs,
			Completed:  playback.Completed,
			PlayedAt:   playback.PlayedAt,
		}
	}

	return resp, nil
}

// SetEpisodeBgm は指定されたエピソードに BGM を設定する
func (s *episodeService) SetEpisodeBgm(ctx context.Context, userID, channelID, episodeID string, req request.SetEpisodeBgmRequest) (*response.EpisodeDataResponse, error) {
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

	// bgmId と systemBgmId の同時指定チェック
	if req.BgmID != nil && req.SystemBgmID != nil {
		return nil, apperror.ErrValidation.WithMessage("bgmId と systemBgmId は同時に指定できません")
	}

	// どちらも指定されていない場合
	if req.BgmID == nil && req.SystemBgmID == nil {
		return nil, apperror.ErrValidation.WithMessage("bgmId または systemBgmId のいずれかを指定してください")
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	if channel.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("このエピソードの BGM 設定権限がありません")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("このチャンネルにエピソードが見つかりません")
	}

	// 前の BGM 設定をクリア
	episode.BgmID = nil
	episode.SystemBgmID = nil
	episode.Bgm = nil
	episode.SystemBgm = nil

	// ユーザー BGM を設定
	if req.BgmID != nil {
		bgmID, err := uuid.Parse(*req.BgmID)
		if err != nil {
			return nil, apperror.ErrValidation.WithMessage("無効な bgmId です")
		}

		// BGM の存在確認とオーナーチェック
		bgm, err := s.bgmRepo.FindByID(ctx, bgmID)
		if err != nil {
			return nil, err
		}

		if bgm.UserID != uid {
			return nil, apperror.ErrForbidden.WithMessage("この BGM へのアクセス権限がありません")
		}

		episode.BgmID = &bgmID
	}

	// システム BGM を設定
	if req.SystemBgmID != nil {
		systemBgmID, err := uuid.Parse(*req.SystemBgmID)
		if err != nil {
			return nil, apperror.ErrValidation.WithMessage("無効な systemBgmId です")
		}

		// システム BGM の存在確認とアクティブチェック
		systemBgm, err := s.systemBgmRepo.FindByID(ctx, systemBgmID)
		if err != nil {
			return nil, err
		}

		if !systemBgm.IsActive {
			return nil, apperror.ErrNotFound.WithMessage("このシステム BGM は利用できません")
		}

		episode.SystemBgmID = &systemBgmID
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

	resp, err := s.toEpisodeResponse(ctx, updated, &channel.User, nil)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeDataResponse{
		Data: resp,
	}, nil
}

// DeleteEpisodeBgm は指定されたエピソードの BGM を削除する
func (s *episodeService) DeleteEpisodeBgm(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("このエピソードの BGM 削除権限がありません")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("このチャンネルにエピソードが見つかりません")
	}

	// BGM 設定をクリア
	episode.BgmID = nil
	episode.SystemBgmID = nil
	episode.Bgm = nil
	episode.SystemBgm = nil

	// エピソードを更新
	if err := s.episodeRepo.Update(ctx, episode); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.episodeRepo.FindByID(ctx, episode.ID)
	if err != nil {
		return nil, err
	}

	resp, err := s.toEpisodeResponse(ctx, updated, &channel.User, nil)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeDataResponse{
		Data: resp,
	}, nil
}
