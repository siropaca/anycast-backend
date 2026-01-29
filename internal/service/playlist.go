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

// DefaultPlaylistName はデフォルトプレイリストの名前
const DefaultPlaylistName = "後で聴く"

// PlaylistService はプレイリスト関連のビジネスロジックインターフェースを表す
type PlaylistService interface {
	// プレイリスト管理
	ListPlaylists(ctx context.Context, userID string, limit, offset int) (*response.PlaylistListWithPaginationResponse, error)
	GetPlaylist(ctx context.Context, userID, playlistID string) (*response.PlaylistDetailDataResponse, error)
	CreatePlaylist(ctx context.Context, userID string, req request.CreatePlaylistRequest) (*response.PlaylistDataResponse, error)
	UpdatePlaylist(ctx context.Context, userID, playlistID string, req request.UpdatePlaylistRequest) (*response.PlaylistDataResponse, error)
	DeletePlaylist(ctx context.Context, userID, playlistID string) error

	// プレイリストアイテム管理
	AddItem(ctx context.Context, userID, playlistID string, req request.AddPlaylistItemRequest) (*response.PlaylistItemDataResponse, error)
	RemoveItem(ctx context.Context, userID, playlistID, itemID string) error
	ReorderItems(ctx context.Context, userID, playlistID string, req request.ReorderPlaylistItemsRequest) (*response.PlaylistDetailDataResponse, error)

	// 「後で聴く」ショートカット
	AddToListenLater(ctx context.Context, userID, episodeID string) (*response.PlaylistItemDataResponse, error)
	RemoveFromListenLater(ctx context.Context, userID, episodeID string) error
	GetListenLater(ctx context.Context, userID string) (*response.PlaylistDetailDataResponse, error)

	// デフォルトプレイリスト作成（ユーザー登録時に呼び出される）
	CreateDefaultPlaylist(ctx context.Context, userID uuid.UUID) error
}

type playlistService struct {
	playlistRepo  repository.PlaylistRepository
	episodeRepo   repository.EpisodeRepository
	storageClient storage.Client
}

// NewPlaylistService は playlistService を生成して PlaylistService として返す
func NewPlaylistService(
	playlistRepo repository.PlaylistRepository,
	episodeRepo repository.EpisodeRepository,
	storageClient storage.Client,
) PlaylistService {
	return &playlistService{
		playlistRepo:  playlistRepo,
		episodeRepo:   episodeRepo,
		storageClient: storageClient,
	}
}

// ListPlaylists は自分のプレイリスト一覧を取得する
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

// GetPlaylist は自分のプレイリスト詳細を取得する
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
		return nil, apperror.ErrForbidden.WithMessage("このプレイリストへのアクセス権限がありません")
	}

	return &response.PlaylistDetailDataResponse{
		Data: s.toPlaylistDetailResponse(ctx, playlist),
	}, nil
}

// CreatePlaylist はプレイリストを作成する
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
		return nil, apperror.ErrDuplicateName.WithMessage("この名前のプレイリストは既に存在します")
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

// UpdatePlaylist はプレイリストを更新する
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
		return nil, apperror.ErrForbidden.WithMessage("このプレイリストへのアクセス権限がありません")
	}

	// デフォルトプレイリストの名前変更は不可
	if playlist.IsDefault && req.Name != nil {
		return nil, apperror.ErrDefaultPlaylist.WithMessage("デフォルトプレイリストの名前は変更できません")
	}

	// 名前の重複チェック（名前が変更される場合のみ）
	if req.Name != nil && *req.Name != playlist.Name {
		exists, err := s.playlistRepo.ExistsByUserIDAndName(ctx, uid, *req.Name)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperror.ErrDuplicateName.WithMessage("この名前のプレイリストは既に存在します")
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

// DeletePlaylist はプレイリストを削除する
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
		return apperror.ErrForbidden.WithMessage("このプレイリストへのアクセス権限がありません")
	}

	// デフォルトプレイリストは削除不可
	if playlist.IsDefault {
		return apperror.ErrDefaultPlaylist.WithMessage("デフォルトプレイリストは削除できません")
	}

	return s.playlistRepo.Delete(ctx, pid)
}

// AddItem はプレイリストにアイテムを追加する
func (s *playlistService) AddItem(ctx context.Context, userID, playlistID string, req request.AddPlaylistItemRequest) (*response.PlaylistItemDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	pid, err := uuid.Parse(playlistID)
	if err != nil {
		return nil, err
	}

	eid, err := uuid.Parse(req.EpisodeID)
	if err != nil {
		return nil, err
	}

	playlist, err := s.playlistRepo.FindByID(ctx, pid)
	if err != nil {
		return nil, err
	}

	// オーナーチェック
	if playlist.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("このプレイリストへのアクセス権限がありません")
	}

	// エピソードの存在確認
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	// 重複チェック
	_, err = s.playlistRepo.FindItemByPlaylistIDAndEpisodeID(ctx, pid, eid)
	if err == nil {
		return nil, apperror.ErrAlreadyInPlaylist.WithMessage("このエピソードは既にプレイリストに追加されています")
	}
	if !apperror.IsCode(err, apperror.CodeNotFound) {
		return nil, err
	}

	// 最大 position を取得
	maxPosition, err := s.playlistRepo.GetMaxPosition(ctx, pid)
	if err != nil {
		return nil, err
	}

	item := &model.PlaylistItem{
		PlaylistID: pid,
		EpisodeID:  eid,
		Position:   maxPosition + 1,
		Episode:    *episode,
	}

	if err := s.playlistRepo.CreateItem(ctx, item); err != nil {
		return nil, err
	}

	// アイテムを再取得してリレーションをプリロード
	item, err = s.playlistRepo.FindItemByID(ctx, item.ID)
	if err != nil {
		return nil, err
	}

	return &response.PlaylistItemDataResponse{
		Data: s.toPlaylistItemResponse(ctx, item),
	}, nil
}

// RemoveItem はプレイリストからアイテムを削除する
func (s *playlistService) RemoveItem(ctx context.Context, userID, playlistID, itemID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	pid, err := uuid.Parse(playlistID)
	if err != nil {
		return err
	}

	iid, err := uuid.Parse(itemID)
	if err != nil {
		return err
	}

	playlist, err := s.playlistRepo.FindByID(ctx, pid)
	if err != nil {
		return err
	}

	// オーナーチェック
	if playlist.UserID != uid {
		return apperror.ErrForbidden.WithMessage("このプレイリストへのアクセス権限がありません")
	}

	// アイテムの存在確認とプレイリストの一致チェック
	item, err := s.playlistRepo.FindItemByID(ctx, iid)
	if err != nil {
		return err
	}

	if item.PlaylistID != pid {
		return apperror.ErrNotFound.WithMessage("プレイリストアイテムが見つかりません")
	}

	// アイテムを削除
	if err := s.playlistRepo.DeleteItem(ctx, iid); err != nil {
		return err
	}

	// 削除されたアイテムより後のアイテムの position を調整
	return s.playlistRepo.DecrementPositionsAfter(ctx, pid, item.Position)
}

// ReorderItems はプレイリストアイテムを並び替える
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
		return nil, apperror.ErrForbidden.WithMessage("このプレイリストへのアクセス権限がありません")
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

	// 更新後のプレイリストを取得して返す
	playlist, err = s.playlistRepo.FindByIDWithItems(ctx, pid)
	if err != nil {
		return nil, err
	}

	return &response.PlaylistDetailDataResponse{
		Data: s.toPlaylistDetailResponse(ctx, playlist),
	}, nil
}

// AddToListenLater は「後で聴く」プレイリストにエピソードを追加する
func (s *playlistService) AddToListenLater(ctx context.Context, userID, episodeID string) (*response.PlaylistItemDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	// デフォルトプレイリストを取得
	playlist, err := s.playlistRepo.FindDefaultByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return s.AddItem(ctx, userID, playlist.ID.String(), request.AddPlaylistItemRequest{
		EpisodeID: episodeID,
	})
}

// RemoveFromListenLater は「後で聴く」プレイリストからエピソードを削除する
func (s *playlistService) RemoveFromListenLater(ctx context.Context, userID, episodeID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return err
	}

	// デフォルトプレイリストを取得
	playlist, err := s.playlistRepo.FindDefaultByUserID(ctx, uid)
	if err != nil {
		return err
	}

	// エピソードに対応するアイテムを検索
	item, err := s.playlistRepo.FindItemByPlaylistIDAndEpisodeID(ctx, playlist.ID, eid)
	if err != nil {
		return err
	}

	// アイテムを削除
	if err := s.playlistRepo.DeleteItem(ctx, item.ID); err != nil {
		return err
	}

	// 削除されたアイテムより後のアイテムの position を調整
	return s.playlistRepo.DecrementPositionsAfter(ctx, playlist.ID, item.Position)
}

// GetListenLater は「後で聴く」プレイリストを取得する
func (s *playlistService) GetListenLater(ctx context.Context, userID string) (*response.PlaylistDetailDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	// デフォルトプレイリストを取得
	playlist, err := s.playlistRepo.FindDefaultByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return s.GetPlaylist(ctx, userID, playlist.ID.String())
}

// CreateDefaultPlaylist はユーザーのデフォルトプレイリストを作成する
func (s *playlistService) CreateDefaultPlaylist(ctx context.Context, userID uuid.UUID) error {
	playlist := &model.Playlist{
		UserID:      userID,
		Name:        DefaultPlaylistName,
		Description: "",
		IsDefault:   true,
	}

	return s.playlistRepo.Create(ctx, playlist)
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
