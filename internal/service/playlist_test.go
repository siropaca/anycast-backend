package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// PlaylistRepository のモック
type mockPlaylistRepository struct {
	mock.Mock
}

func (m *mockPlaylistRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Playlist, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Playlist), args.Error(1)
}

func (m *mockPlaylistRepository) FindByIDWithItems(ctx context.Context, id uuid.UUID) (*model.Playlist, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Playlist), args.Error(1)
}

func (m *mockPlaylistRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Playlist, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]model.Playlist), args.Get(1).(int64), args.Error(2)
}

func (m *mockPlaylistRepository) FindDefaultByUserID(ctx context.Context, userID uuid.UUID) (*model.Playlist, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Playlist), args.Error(1)
}

func (m *mockPlaylistRepository) ExistsByUserIDAndName(ctx context.Context, userID uuid.UUID, name string) (bool, error) {
	args := m.Called(ctx, userID, name)
	return args.Bool(0), args.Error(1)
}

func (m *mockPlaylistRepository) Create(ctx context.Context, playlist *model.Playlist) error {
	args := m.Called(ctx, playlist)
	return args.Error(0)
}

func (m *mockPlaylistRepository) Update(ctx context.Context, playlist *model.Playlist) error {
	args := m.Called(ctx, playlist)
	return args.Error(0)
}

func (m *mockPlaylistRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockPlaylistRepository) CountItemsByPlaylistID(ctx context.Context, playlistID uuid.UUID) (int64, error) {
	args := m.Called(ctx, playlistID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockPlaylistRepository) FindItemByID(ctx context.Context, id uuid.UUID) (*model.PlaylistItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.PlaylistItem), args.Error(1)
}

func (m *mockPlaylistRepository) FindItemsByPlaylistID(ctx context.Context, playlistID uuid.UUID) ([]model.PlaylistItem, error) {
	args := m.Called(ctx, playlistID)
	return args.Get(0).([]model.PlaylistItem), args.Error(1)
}

func (m *mockPlaylistRepository) FindItemByPlaylistIDAndEpisodeID(ctx context.Context, playlistID, episodeID uuid.UUID) (*model.PlaylistItem, error) {
	args := m.Called(ctx, playlistID, episodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.PlaylistItem), args.Error(1)
}

func (m *mockPlaylistRepository) CreateItem(ctx context.Context, item *model.PlaylistItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *mockPlaylistRepository) DeleteItem(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockPlaylistRepository) GetMaxPosition(ctx context.Context, playlistID uuid.UUID) (int, error) {
	args := m.Called(ctx, playlistID)
	return args.Int(0), args.Error(1)
}

func (m *mockPlaylistRepository) UpdateItemPositions(ctx context.Context, playlistID uuid.UUID, itemIDs []uuid.UUID) error {
	args := m.Called(ctx, playlistID, itemIDs)
	return args.Error(0)
}

func (m *mockPlaylistRepository) DecrementPositionsAfter(ctx context.Context, playlistID uuid.UUID, position int) error {
	args := m.Called(ctx, playlistID, position)
	return args.Error(0)
}

func (m *mockPlaylistRepository) FindPlaylistIDsByUserIDAndEpisodeID(ctx context.Context, userID, episodeID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, userID, episodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

// EpisodeRepository のモック（プレイリストテスト用）
type mockEpisodeRepositoryForPlaylist struct {
	mock.Mock
}

func (m *mockEpisodeRepositoryForPlaylist) FindByID(ctx context.Context, id uuid.UUID) (*model.Episode, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Episode), args.Error(1)
}

func (m *mockEpisodeRepositoryForPlaylist) FindByChannelID(ctx context.Context, channelID uuid.UUID, filter repository.EpisodeFilter) ([]model.Episode, int64, error) {
	args := m.Called(ctx, channelID, filter)
	return args.Get(0).([]model.Episode), args.Get(1).(int64), args.Error(2)
}

func (m *mockEpisodeRepositoryForPlaylist) Create(ctx context.Context, episode *model.Episode) error {
	args := m.Called(ctx, episode)
	return args.Error(0)
}

func (m *mockEpisodeRepositoryForPlaylist) Update(ctx context.Context, episode *model.Episode) error {
	args := m.Called(ctx, episode)
	return args.Error(0)
}

func (m *mockEpisodeRepositoryForPlaylist) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockEpisodeRepositoryForPlaylist) IncrementPlayCount(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestUpdateEpisodePlaylists(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	episodeID := uuid.New()
	playlistID1 := uuid.New()

	episode := &model.Episode{
		ID:        episodeID,
		Title:     "Test Episode",
		CreatedAt: now,
		UpdatedAt: now,
	}

	t.Run("差分なし: 変更がない場合はそのまま返す", func(t *testing.T) {
		mockPlaylist := new(mockPlaylistRepository)
		mockEpisode := new(mockEpisodeRepositoryForPlaylist)
		mockStorage := new(mockStorageClient)

		svc := &playlistService{
			playlistRepo:  mockPlaylist,
			episodeRepo:   mockEpisode,
			storageClient: mockStorage,
		}

		mockEpisode.On("FindByID", mock.Anything, episodeID).Return(episode, nil)
		mockPlaylist.On("FindByID", mock.Anything, playlistID1).Return(&model.Playlist{ID: playlistID1, UserID: userID}, nil)
		mockPlaylist.On("FindPlaylistIDsByUserIDAndEpisodeID", mock.Anything, userID, episodeID).Return([]uuid.UUID{playlistID1}, nil)

		req := request.UpdateEpisodePlaylistsRequest{
			PlaylistIDs: []string{playlistID1.String()},
		}

		result, err := svc.UpdateEpisodePlaylists(context.Background(), userID.String(), episodeID.String(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data.PlaylistIDs, 1)
		assert.Equal(t, playlistID1, result.Data.PlaylistIDs[0])
		mockPlaylist.AssertExpectations(t)
		mockEpisode.AssertExpectations(t)
	})

	t.Run("空配列で差分なし: 現在も空の場合はそのまま返す", func(t *testing.T) {
		mockPlaylist := new(mockPlaylistRepository)
		mockEpisode := new(mockEpisodeRepositoryForPlaylist)
		mockStorage := new(mockStorageClient)

		svc := &playlistService{
			playlistRepo:  mockPlaylist,
			episodeRepo:   mockEpisode,
			storageClient: mockStorage,
		}

		mockEpisode.On("FindByID", mock.Anything, episodeID).Return(episode, nil)
		mockPlaylist.On("FindPlaylistIDsByUserIDAndEpisodeID", mock.Anything, userID, episodeID).Return([]uuid.UUID{}, nil)

		req := request.UpdateEpisodePlaylistsRequest{
			PlaylistIDs: []string{},
		}

		result, err := svc.UpdateEpisodePlaylists(context.Background(), userID.String(), episodeID.String(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data.PlaylistIDs, 0)
		mockPlaylist.AssertExpectations(t)
		mockEpisode.AssertExpectations(t)
	})

	t.Run("他ユーザーのプレイリスト指定でエラー", func(t *testing.T) {
		mockPlaylist := new(mockPlaylistRepository)
		mockEpisode := new(mockEpisodeRepositoryForPlaylist)
		mockStorage := new(mockStorageClient)

		svc := &playlistService{
			playlistRepo:  mockPlaylist,
			episodeRepo:   mockEpisode,
			storageClient: mockStorage,
		}

		otherUserID := uuid.New()

		mockEpisode.On("FindByID", mock.Anything, episodeID).Return(episode, nil)
		mockPlaylist.On("FindByID", mock.Anything, playlistID1).Return(&model.Playlist{ID: playlistID1, UserID: otherUserID}, nil)

		req := request.UpdateEpisodePlaylistsRequest{
			PlaylistIDs: []string{playlistID1.String()},
		}

		result, err := svc.UpdateEpisodePlaylists(context.Background(), userID.String(), episodeID.String(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperror.IsCode(err, apperror.CodeForbidden))
		mockPlaylist.AssertExpectations(t)
		mockEpisode.AssertExpectations(t)
	})

	t.Run("存在しないエピソードでエラー", func(t *testing.T) {
		mockPlaylist := new(mockPlaylistRepository)
		mockEpisode := new(mockEpisodeRepositoryForPlaylist)
		mockStorage := new(mockStorageClient)

		svc := &playlistService{
			playlistRepo:  mockPlaylist,
			episodeRepo:   mockEpisode,
			storageClient: mockStorage,
		}

		mockEpisode.On("FindByID", mock.Anything, episodeID).Return(nil, apperror.ErrNotFound.WithMessage("エピソードが見つかりません"))

		req := request.UpdateEpisodePlaylistsRequest{
			PlaylistIDs: []string{playlistID1.String()},
		}

		result, err := svc.UpdateEpisodePlaylists(context.Background(), userID.String(), episodeID.String(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperror.IsCode(err, apperror.CodeNotFound))
		mockEpisode.AssertExpectations(t)
	})

	t.Run("無効な UUID でエラー", func(t *testing.T) {
		mockPlaylist := new(mockPlaylistRepository)
		mockEpisode := new(mockEpisodeRepositoryForPlaylist)
		mockStorage := new(mockStorageClient)

		svc := &playlistService{
			playlistRepo:  mockPlaylist,
			episodeRepo:   mockEpisode,
			storageClient: mockStorage,
		}

		req := request.UpdateEpisodePlaylistsRequest{
			PlaylistIDs: []string{},
		}

		result, err := svc.UpdateEpisodePlaylists(context.Background(), "invalid-uuid", episodeID.String(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("重複する playlistIDs は排除される", func(t *testing.T) {
		mockPlaylist := new(mockPlaylistRepository)
		mockEpisode := new(mockEpisodeRepositoryForPlaylist)
		mockStorage := new(mockStorageClient)

		svc := &playlistService{
			playlistRepo:  mockPlaylist,
			episodeRepo:   mockEpisode,
			storageClient: mockStorage,
		}

		mockEpisode.On("FindByID", mock.Anything, episodeID).Return(episode, nil)
		// 重複排除後は1回だけ FindByID が呼ばれる
		mockPlaylist.On("FindByID", mock.Anything, playlistID1).Return(&model.Playlist{ID: playlistID1, UserID: userID}, nil)
		mockPlaylist.On("FindPlaylistIDsByUserIDAndEpisodeID", mock.Anything, userID, episodeID).Return([]uuid.UUID{playlistID1}, nil)

		req := request.UpdateEpisodePlaylistsRequest{
			PlaylistIDs: []string{playlistID1.String(), playlistID1.String()},
		}

		result, err := svc.UpdateEpisodePlaylists(context.Background(), userID.String(), episodeID.String(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data.PlaylistIDs, 1)
		mockPlaylist.AssertExpectations(t)
		mockEpisode.AssertExpectations(t)
	})
}
