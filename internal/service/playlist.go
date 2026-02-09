package service

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// DefaultPlaylistName はデフォルト再生リストの名前
const DefaultPlaylistName = "後で聴く"

// PlaylistService は再生リスト関連のビジネスロジックインターフェースを表す
type PlaylistService interface {
	// 再生リスト管理
	ListPlaylists(ctx context.Context, userID string, limit, offset int) (*response.PlaylistListWithPaginationResponse, error)
	GetPlaylist(ctx context.Context, userID, playlistID string) (*response.PlaylistDetailDataResponse, error)
	CreatePlaylist(ctx context.Context, userID string, req request.CreatePlaylistRequest) (*response.PlaylistDataResponse, error)
	UpdatePlaylist(ctx context.Context, userID, playlistID string, req request.UpdatePlaylistRequest) (*response.PlaylistDataResponse, error)
	DeletePlaylist(ctx context.Context, userID, playlistID string) error

	// 再生リストアイテム管理
	ReorderItems(ctx context.Context, userID, playlistID string, req request.ReorderPlaylistItemsRequest) (*response.PlaylistDetailDataResponse, error)

	// デフォルト再生リスト作成（ユーザー登録時に呼び出される）
	CreateDefaultPlaylist(ctx context.Context, userID uuid.UUID) error

	// エピソードの再生リスト所属一括更新
	UpdateEpisodePlaylists(ctx context.Context, userID, episodeID string, req request.UpdateEpisodePlaylistsRequest) (*response.EpisodePlaylistIDsDataResponse, error)
}

type playlistService struct {
	db            *gorm.DB
	playlistRepo  repository.PlaylistRepository
	episodeRepo   repository.EpisodeRepository
	storageClient storage.Client
}

// NewPlaylistService は playlistService を生成して PlaylistService として返す
func NewPlaylistService(
	db *gorm.DB,
	playlistRepo repository.PlaylistRepository,
	episodeRepo repository.EpisodeRepository,
	storageClient storage.Client,
) PlaylistService {
	return &playlistService{
		db:            db,
		playlistRepo:  playlistRepo,
		episodeRepo:   episodeRepo,
		storageClient: storageClient,
	}
}

// ListPlaylists は自分の再生リスト一覧を取得する
func (s *playlistService) ListPlaylists(ctx context.Context, userID string, limit, offset int) (*response.PlaylistListWithPaginationResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	playlists, total, err := s.playlistRepo.FindByUserID(ctx, uid, limit, offset)
	if err != nil {
		return nil, err
	}

	// レスポンスに変換
	data := make([]response.PlaylistResponse, 0, len(playlists))
	for _, playlist := range playlists {
		itemCount, err := s.playlistRepo.CountItemsByPlaylistID(ctx, playlist.ID)
		if err != nil {
			return nil, err
		}

		data = append(data, response.PlaylistResponse{
			ID:          playlist.ID,
			Name:        playlist.Name,
			Description: playlist.Description,
			IsDefault:   playlist.IsDefault,
			ItemCount:   int(itemCount),
			CreatedAt:   playlist.CreatedAt,
			UpdatedAt:   playlist.UpdatedAt,
		})
	}

	return &response.PlaylistListWithPaginationResponse{
		Data: data,
		Pagination: response.PaginationResponse{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}, nil
}

// GetPlaylist は自分の再生リスト詳細を取得する
func (s *playlistService) GetPlaylist(ctx context.Context, userID, playlistID string) (*response.PlaylistDetailDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	pid, err := uuid.Parse(playlistID)
	if err != nil {
		return nil, err
	}

	playlist, err := s.playlistRepo.FindByIDWithItems(ctx, pid)
	if err != nil {
		return nil, err
	}

	// オーナーチェック
	if playlist.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("この再生リストへのアクセス権限がありません")
	}

	return &response.PlaylistDetailDataResponse{
		Data: s.toPlaylistDetailResponse(ctx, playlist),
	}, nil
}

// CreatePlaylist は再生リストを作成する
func (s *playlistService) CreatePlaylist(ctx context.Context, userID string, req request.CreatePlaylistRequest) (*response.PlaylistDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	// 名前の重複チェック
	exists, err := s.playlistRepo.ExistsByUserIDAndName(ctx, uid, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.ErrDuplicateName.WithMessage("この名前の再生リストは既に存在します")
	}

	description := ""
	if req.Description != nil {
		description = *req.Description
	}

	playlist := &model.Playlist{
		UserID:      uid,
		Name:        req.Name,
		Description: description,
		IsDefault:   false,
	}

	if err := s.playlistRepo.Create(ctx, playlist); err != nil {
		return nil, err
	}

	return &response.PlaylistDataResponse{
		Data: response.PlaylistResponse{
			ID:          playlist.ID,
			Name:        playlist.Name,
			Description: playlist.Description,
			IsDefault:   playlist.IsDefault,
			ItemCount:   0,
			CreatedAt:   playlist.CreatedAt,
			UpdatedAt:   playlist.UpdatedAt,
		},
	}, nil
}

// UpdatePlaylist は再生リストを更新する
func (s *playlistService) UpdatePlaylist(ctx context.Context, userID, playlistID string, req request.UpdatePlaylistRequest) (*response.PlaylistDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	pid, err := uuid.Parse(playlistID)
	if err != nil {
		return nil, err
	}

	playlist, err := s.playlistRepo.FindByID(ctx, pid)
	if err != nil {
		return nil, err
	}

	// オーナーチェック
	if playlist.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("この再生リストへのアクセス権限がありません")
	}

	// デフォルト再生リストの名前変更は不可
	if playlist.IsDefault && req.Name != nil {
		return nil, apperror.ErrDefaultPlaylist.WithMessage("デフォルト再生リストの名前は変更できません")
	}

	// 名前の重複チェック（名前が変更される場合のみ）
	if req.Name != nil && *req.Name != playlist.Name {
		exists, err := s.playlistRepo.ExistsByUserIDAndName(ctx, uid, *req.Name)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperror.ErrDuplicateName.WithMessage("この名前の再生リストは既に存在します")
		}
		playlist.Name = *req.Name
	}

	if req.Description != nil {
		playlist.Description = *req.Description
	}

	playlist.UpdatedAt = time.Now()

	if err := s.playlistRepo.Update(ctx, playlist); err != nil {
		return nil, err
	}

	itemCount, err := s.playlistRepo.CountItemsByPlaylistID(ctx, playlist.ID)
	if err != nil {
		return nil, err
	}

	return &response.PlaylistDataResponse{
		Data: response.PlaylistResponse{
			ID:          playlist.ID,
			Name:        playlist.Name,
			Description: playlist.Description,
			IsDefault:   playlist.IsDefault,
			ItemCount:   int(itemCount),
			CreatedAt:   playlist.CreatedAt,
			UpdatedAt:   playlist.UpdatedAt,
		},
	}, nil
}

// DeletePlaylist は再生リストを削除する
func (s *playlistService) DeletePlaylist(ctx context.Context, userID, playlistID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	pid, err := uuid.Parse(playlistID)
	if err != nil {
		return err
	}

	playlist, err := s.playlistRepo.FindByID(ctx, pid)
	if err != nil {
		return err
	}

	// オーナーチェック
	if playlist.UserID != uid {
		return apperror.ErrForbidden.WithMessage("この再生リストへのアクセス権限がありません")
	}

	// デフォルト再生リストは削除不可
	if playlist.IsDefault {
		return apperror.ErrDefaultPlaylist.WithMessage("デフォルト再生リストは削除できません")
	}

	return s.playlistRepo.Delete(ctx, pid)
}

// ReorderItems は再生リストアイテムを並び替える
func (s *playlistService) ReorderItems(ctx context.Context, userID, playlistID string, req request.ReorderPlaylistItemsRequest) (*response.PlaylistDetailDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	pid, err := uuid.Parse(playlistID)
	if err != nil {
		return nil, err
	}

	playlist, err := s.playlistRepo.FindByID(ctx, pid)
	if err != nil {
		return nil, err
	}

	// オーナーチェック
	if playlist.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("この再生リストへのアクセス権限がありません")
	}

	// アイテム ID を UUID にパース
	itemIDs := make([]uuid.UUID, 0, len(req.ItemIDs))
	for _, idStr := range req.ItemIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		itemIDs = append(itemIDs, id)
	}

	// position を更新
	if err := s.playlistRepo.UpdateItemPositions(ctx, pid, itemIDs); err != nil {
		return nil, err
	}

	// 更新後の再生リストを取得して返す
	playlist, err = s.playlistRepo.FindByIDWithItems(ctx, pid)
	if err != nil {
		return nil, err
	}

	return &response.PlaylistDetailDataResponse{
		Data: s.toPlaylistDetailResponse(ctx, playlist),
	}, nil
}

// CreateDefaultPlaylist はユーザーのデフォルト再生リストを作成する
func (s *playlistService) CreateDefaultPlaylist(ctx context.Context, userID uuid.UUID) error {
	playlist := &model.Playlist{
		UserID:      userID,
		Name:        DefaultPlaylistName,
		Description: "",
		IsDefault:   true,
	}

	return s.playlistRepo.Create(ctx, playlist)
}

// UpdateEpisodePlaylists はエピソードの再生リスト所属を一括更新する
func (s *playlistService) UpdateEpisodePlaylists(ctx context.Context, userID, episodeID string, req request.UpdateEpisodePlaylistsRequest) (*response.EpisodePlaylistIDsDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return nil, err
	}

	// エピソードの存在確認
	if _, err := s.episodeRepo.FindByID(ctx, eid); err != nil {
		return nil, err
	}

	// リクエストの playlistIDs を UUID にパース＋重複排除
	seen := make(map[uuid.UUID]bool)
	var requestedIDs []uuid.UUID
	for _, idStr := range req.PlaylistIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		if !seen[id] {
			seen[id] = true
			requestedIDs = append(requestedIDs, id)
		}
	}

	// 指定再生リストのオーナーチェック（全てが認証ユーザーのものか確認）
	for _, pid := range requestedIDs {
		playlist, err := s.playlistRepo.FindByID(ctx, pid)
		if err != nil {
			return nil, err
		}
		if playlist.UserID != uid {
			return nil, apperror.ErrForbidden.WithMessage("この再生リストへのアクセス権限がありません")
		}
	}

	// 現在の所属再生リスト ID 取得
	currentIDs, err := s.playlistRepo.FindPlaylistIDsByUserIDAndEpisodeID(ctx, uid, eid)
	if err != nil {
		return nil, err
	}

	// 差分計算
	currentSet := make(map[uuid.UUID]bool)
	for _, id := range currentIDs {
		currentSet[id] = true
	}

	requestedSet := make(map[uuid.UUID]bool)
	for _, id := range requestedIDs {
		requestedSet[id] = true
	}

	var toAdd []uuid.UUID
	for _, id := range requestedIDs {
		if !currentSet[id] {
			toAdd = append(toAdd, id)
		}
	}

	var toRemove []uuid.UUID
	for _, id := range currentIDs {
		if !requestedSet[id] {
			toRemove = append(toRemove, id)
		}
	}

	// 差分がない場合は早期リターン
	if len(toAdd) == 0 && len(toRemove) == 0 {
		resultIDs := requestedIDs
		if resultIDs == nil {
			resultIDs = []uuid.UUID{}
		}
		return &response.EpisodePlaylistIDsDataResponse{
			Data: response.EpisodePlaylistIDsResponse{
				PlaylistIDs: resultIDs,
			},
		}, nil
	}

	// トランザクション内で追加・削除を実行
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 削除: 各再生リストからエピソードを外す
		for _, pid := range toRemove {
			item, err := s.playlistRepo.FindItemByPlaylistIDAndEpisodeID(ctx, pid, eid)
			if err != nil {
				return err
			}
			if err := s.playlistRepo.DeleteItem(ctx, item.ID); err != nil {
				return err
			}
			if err := s.playlistRepo.DecrementPositionsAfter(ctx, pid, item.Position); err != nil {
				return err
			}
		}

		// 追加: 各再生リストにエピソードを追加
		for _, pid := range toAdd {
			maxPosition, err := s.playlistRepo.GetMaxPosition(ctx, pid)
			if err != nil {
				return err
			}
			item := &model.PlaylistItem{
				PlaylistID: pid,
				EpisodeID:  eid,
				Position:   maxPosition + 1,
			}
			if err := s.playlistRepo.CreateItem(ctx, item); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// 更新後の所属 ID を取得して返却
	updatedIDs, err := s.playlistRepo.FindPlaylistIDsByUserIDAndEpisodeID(ctx, uid, eid)
	if err != nil {
		return nil, err
	}
	if updatedIDs == nil {
		updatedIDs = []uuid.UUID{}
	}

	return &response.EpisodePlaylistIDsDataResponse{
		Data: response.EpisodePlaylistIDsResponse{
			PlaylistIDs: updatedIDs,
		},
	}, nil
}

// toPlaylistDetailResponse は model.Playlist を response.PlaylistDetailResponse に変換する
func (s *playlistService) toPlaylistDetailResponse(ctx context.Context, playlist *model.Playlist) response.PlaylistDetailResponse {
	items := make([]response.PlaylistItemResponse, 0, len(playlist.Items))
	for _, item := range playlist.Items {
		items = append(items, s.toPlaylistItemResponse(ctx, &item))
	}

	return response.PlaylistDetailResponse{
		ID:          playlist.ID,
		Name:        playlist.Name,
		Description: playlist.Description,
		IsDefault:   playlist.IsDefault,
		Items:       items,
		CreatedAt:   playlist.CreatedAt,
		UpdatedAt:   playlist.UpdatedAt,
	}
}

// toPlaylistItemResponse は model.PlaylistItem を response.PlaylistItemResponse に変換する
func (s *playlistService) toPlaylistItemResponse(ctx context.Context, item *model.PlaylistItem) response.PlaylistItemResponse {
	episode := item.Episode

	// Artwork の URL を生成
	var artwork *response.ArtworkResponse
	if episode.Artwork != nil {
		if storage.IsExternalURL(episode.Artwork.Path) {
			artwork = &response.ArtworkResponse{
				ID:  episode.Artwork.ID,
				URL: episode.Artwork.Path,
			}
		} else {
			signedURL, err := s.storageClient.GenerateSignedURL(ctx, episode.Artwork.Path, storage.SignedURLExpirationImage)
			if err == nil {
				artwork = &response.ArtworkResponse{
					ID:  episode.Artwork.ID,
					URL: signedURL,
				}
			}
		}
	}

	// FullAudio の URL を生成
	var fullAudio *response.AudioResponse
	if episode.FullAudio != nil {
		signedURL, err := s.storageClient.GenerateSignedURL(ctx, episode.FullAudio.Path, storage.SignedURLExpirationAudio)
		if err == nil {
			fullAudio = &response.AudioResponse{
				ID:         episode.FullAudio.ID,
				URL:        signedURL,
				DurationMs: episode.FullAudio.DurationMs,
			}
		}
	}

	// Channel の Artwork の URL を生成
	var channelArtwork *response.ArtworkResponse
	if episode.Channel.Artwork != nil {
		if storage.IsExternalURL(episode.Channel.Artwork.Path) {
			channelArtwork = &response.ArtworkResponse{
				ID:  episode.Channel.Artwork.ID,
				URL: episode.Channel.Artwork.Path,
			}
		} else {
			signedURL, err := s.storageClient.GenerateSignedURL(ctx, episode.Channel.Artwork.Path, storage.SignedURLExpirationImage)
			if err == nil {
				channelArtwork = &response.ArtworkResponse{
					ID:  episode.Channel.Artwork.ID,
					URL: signedURL,
				}
			}
		}
	}

	return response.PlaylistItemResponse{
		ID:       item.ID,
		Position: item.Position,
		Episode: response.PlaylistEpisodeResponse{
			ID:          episode.ID,
			Title:       episode.Title,
			Description: episode.Description,
			Artwork:     artwork,
			FullAudio:   fullAudio,
			PlayCount:   episode.PlayCount,
			PublishedAt: episode.PublishedAt,
			Channel: response.PlaylistChannelResponse{
				ID:      episode.Channel.ID,
				Name:    episode.Channel.Name,
				Artwork: channelArtwork,
			},
		},
		AddedAt: item.AddedAt,
	}
}
