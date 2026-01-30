package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// RecommendationRepository のモック
type mockRecommendationRepository struct {
	mock.Mock
}

func (m *mockRecommendationRepository) FindRecommendedChannels(ctx context.Context, params repository.RecommendChannelParams) ([]repository.RecommendedChannel, int64, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]repository.RecommendedChannel), args.Get(1).(int64), args.Error(2)
}

func (m *mockRecommendationRepository) FindUserCategoryPreferences(ctx context.Context, userID uuid.UUID) ([]repository.CategoryPreference, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.CategoryPreference), args.Error(1)
}

func (m *mockRecommendationRepository) FindUserPlayedChannelIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (m *mockRecommendationRepository) FindUserDefaultPlaylistCategoryPreferences(ctx context.Context, userID uuid.UUID) ([]repository.CategoryPreference, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.CategoryPreference), args.Error(1)
}

func (m *mockRecommendationRepository) FindUserChannelIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (m *mockRecommendationRepository) FindPublishedEpisodes(ctx context.Context, params repository.RecommendEpisodeParams) ([]model.Episode, int64, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.Episode), args.Get(1).(int64), args.Error(2)
}

func (m *mockRecommendationRepository) FindUserPlaybackHistories(ctx context.Context, userID uuid.UUID) ([]model.PlaybackHistory, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.PlaybackHistory), args.Error(1)
}

func (m *mockRecommendationRepository) FindUserDefaultPlaylistEpisodeIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

// テスト用チャンネルデータを生成するヘルパー
func createTestChannel(id, categoryID uuid.UUID, playCount int, latestEpisodeAt *time.Time) repository.RecommendedChannel {
	return repository.RecommendedChannel{
		Channel: model.Channel{
			ID:         id,
			Name:       "Channel " + id.String()[:8],
			CategoryID: categoryID,
			Category: model.Category{
				ID:   categoryID,
				Slug: "cat-" + categoryID.String()[:8],
				Name: "Category " + categoryID.String()[:8],
			},
		},
		EpisodeCount:    3,
		TotalPlayCount:  playCount,
		LatestEpisodeAt: latestEpisodeAt,
	}
}

func TestRecommendationService_GetRecommendedChannels(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	categoryID := uuid.New()

	t.Run("未ログインでおすすめチャンネル一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockRecommendationRepository)
		mockStorage := new(mockStorageClient)

		channels := []repository.RecommendedChannel{
			createTestChannel(uuid.New(), categoryID, 100, &now),
			createTestChannel(uuid.New(), categoryID, 50, &now),
		}
		mockRepo.On("FindRecommendedChannels", mock.Anything, mock.AnythingOfType("repository.RecommendChannelParams")).Return(channels, int64(2), nil)

		svc := NewRecommendationService(mockRepo, mockStorage)

		req := request.RecommendChannelsRequest{
			PaginationRequest: request.PaginationRequest{Limit: 20, Offset: 0},
		}
		result, err := svc.GetRecommendedChannels(ctx, nil, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, int64(2), result.Pagination.Total)
		assert.Equal(t, 20, result.Pagination.Limit)
		assert.Equal(t, 0, result.Pagination.Offset)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ログイン時はパーソナライズされた結果を返す", func(t *testing.T) {
		mockRepo := new(mockRecommendationRepository)
		mockStorage := new(mockStorageClient)

		userID := uuid.New()
		userIDStr := userID.String()
		ownChannelID := uuid.New()
		otherCategoryID := uuid.New()

		channels := []repository.RecommendedChannel{
			createTestChannel(uuid.New(), categoryID, 100, &now),
			createTestChannel(ownChannelID, categoryID, 200, &now), // 自分のチャンネル
			createTestChannel(uuid.New(), otherCategoryID, 50, &now),
		}

		mockRepo.On("FindRecommendedChannels", mock.Anything, mock.AnythingOfType("repository.RecommendChannelParams")).Return(channels, int64(3), nil)
		mockRepo.On("FindUserCategoryPreferences", mock.Anything, userID).Return([]repository.CategoryPreference{
			{CategoryID: categoryID, PlayCount: 10},
		}, nil)
		mockRepo.On("FindUserDefaultPlaylistCategoryPreferences", mock.Anything, userID).Return([]repository.CategoryPreference{}, nil)
		mockRepo.On("FindUserPlayedChannelIDs", mock.Anything, userID).Return([]uuid.UUID{}, nil)
		mockRepo.On("FindUserChannelIDs", mock.Anything, userID).Return([]uuid.UUID{ownChannelID}, nil)

		svc := NewRecommendationService(mockRepo, mockStorage)

		req := request.RecommendChannelsRequest{
			PaginationRequest: request.PaginationRequest{Limit: 20, Offset: 0},
		}
		result, err := svc.GetRecommendedChannels(ctx, &userIDStr, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		// 自分のチャンネルが除外されているので 2 件
		assert.Len(t, result.Data, 2)
		for _, ch := range result.Data {
			assert.NotEqual(t, ownChannelID, ch.ID)
		}
		mockRepo.AssertExpectations(t)
	})

	t.Run("limit が 50 を超える場合は 50 に制限される", func(t *testing.T) {
		mockRepo := new(mockRecommendationRepository)
		mockStorage := new(mockStorageClient)

		mockRepo.On("FindRecommendedChannels", mock.Anything, mock.AnythingOfType("repository.RecommendChannelParams")).Return([]repository.RecommendedChannel{}, int64(0), nil)

		svc := NewRecommendationService(mockRepo, mockStorage)

		req := request.RecommendChannelsRequest{
			PaginationRequest: request.PaginationRequest{Limit: 100, Offset: 0},
		}
		result, err := svc.GetRecommendedChannels(ctx, nil, req)

		assert.NoError(t, err)
		assert.Equal(t, 50, result.Pagination.Limit)
		mockRepo.AssertExpectations(t)
	})

	t.Run("空のチャンネル一覧でもエラーにならない", func(t *testing.T) {
		mockRepo := new(mockRecommendationRepository)
		mockStorage := new(mockStorageClient)

		mockRepo.On("FindRecommendedChannels", mock.Anything, mock.AnythingOfType("repository.RecommendChannelParams")).Return([]repository.RecommendedChannel{}, int64(0), nil)

		svc := NewRecommendationService(mockRepo, mockStorage)

		req := request.RecommendChannelsRequest{
			PaginationRequest: request.PaginationRequest{Limit: 20, Offset: 0},
		}
		result, err := svc.GetRecommendedChannels(ctx, nil, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Data)
		assert.Equal(t, int64(0), result.Pagination.Total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("カテゴリ ID でフィルタできる", func(t *testing.T) {
		mockRepo := new(mockRecommendationRepository)
		mockStorage := new(mockStorageClient)

		catIDStr := categoryID.String()
		channels := []repository.RecommendedChannel{
			createTestChannel(uuid.New(), categoryID, 100, &now),
		}
		mockRepo.On("FindRecommendedChannels", mock.Anything, mock.MatchedBy(func(params repository.RecommendChannelParams) bool {
			return params.CategoryID != nil && *params.CategoryID == categoryID
		})).Return(channels, int64(1), nil)

		svc := NewRecommendationService(mockRepo, mockStorage)

		req := request.RecommendChannelsRequest{
			PaginationRequest: request.PaginationRequest{Limit: 20, Offset: 0},
			CategoryID:        &catIDStr,
		}
		result, err := svc.GetRecommendedChannels(ctx, nil, req)

		assert.NoError(t, err)
		assert.Len(t, result.Data, 1)
		mockRepo.AssertExpectations(t)
	})

	t.Run("リポジトリエラー時はエラーを返す", func(t *testing.T) {
		mockRepo := new(mockRecommendationRepository)
		mockStorage := new(mockStorageClient)

		mockRepo.On("FindRecommendedChannels", mock.Anything, mock.AnythingOfType("repository.RecommendChannelParams")).Return(nil, int64(0), assert.AnError)

		svc := NewRecommendationService(mockRepo, mockStorage)

		req := request.RecommendChannelsRequest{
			PaginationRequest: request.PaginationRequest{Limit: 20, Offset: 0},
		}
		result, err := svc.GetRecommendedChannels(ctx, nil, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestRecommendationService_calculateBaseScore(t *testing.T) {
	svc := &recommendationService{}
	now := time.Now()

	t.Run("再生回数が多いほどスコアが高い", func(t *testing.T) {
		ch1 := repository.RecommendedChannel{TotalPlayCount: 10, LatestEpisodeAt: &now}
		ch2 := repository.RecommendedChannel{TotalPlayCount: 1000, LatestEpisodeAt: &now}

		score1 := svc.calculateBaseScore(ch1, now)
		score2 := svc.calculateBaseScore(ch2, now)

		assert.Greater(t, score2, score1)
	})

	t.Run("最新エピソードが新しいほどスコアが高い", func(t *testing.T) {
		recent := now.Add(-1 * 24 * time.Hour) // 1日前
		old := now.Add(-90 * 24 * time.Hour)   // 90日前

		ch1 := repository.RecommendedChannel{TotalPlayCount: 100, LatestEpisodeAt: &old}
		ch2 := repository.RecommendedChannel{TotalPlayCount: 100, LatestEpisodeAt: &recent}

		score1 := svc.calculateBaseScore(ch1, now)
		score2 := svc.calculateBaseScore(ch2, now)

		assert.Greater(t, score2, score1)
	})

	t.Run("最新エピソードがない場合は新着度スコアが 0", func(t *testing.T) {
		ch := repository.RecommendedChannel{TotalPlayCount: 100, LatestEpisodeAt: nil}

		score := svc.calculateBaseScore(ch, now)

		// 再生回数分のスコアのみ
		assert.Greater(t, score, 0.0)
	})

	t.Run("再生回数 0 で最新エピソードがない場合のスコアは 0", func(t *testing.T) {
		ch := repository.RecommendedChannel{TotalPlayCount: 0, LatestEpisodeAt: nil}

		score := svc.calculateBaseScore(ch, now)

		assert.Equal(t, 0.0, score)
	})
}

func TestRecommendationService_applyDiversityFilter(t *testing.T) {
	svc := &recommendationService{}

	t.Run("空のスライスではエラーにならない", func(t *testing.T) {
		result := svc.applyDiversityFilter([]scoredChannel{})
		assert.Empty(t, result)
	})

	t.Run("同一カテゴリが 3 件以上連続しない", func(t *testing.T) {
		catA := uuid.New()
		catB := uuid.New()

		scored := []scoredChannel{
			{channel: repository.RecommendedChannel{Channel: model.Channel{ID: uuid.New(), CategoryID: catA}}, score: 100},
			{channel: repository.RecommendedChannel{Channel: model.Channel{ID: uuid.New(), CategoryID: catA}}, score: 90},
			{channel: repository.RecommendedChannel{Channel: model.Channel{ID: uuid.New(), CategoryID: catA}}, score: 80},
			{channel: repository.RecommendedChannel{Channel: model.Channel{ID: uuid.New(), CategoryID: catA}}, score: 70},
			{channel: repository.RecommendedChannel{Channel: model.Channel{ID: uuid.New(), CategoryID: catB}}, score: 60},
		}

		result := svc.applyDiversityFilter(scored)

		assert.Len(t, result, 5)

		// 同一カテゴリが 4 件連続しないことを確認
		for i := 3; i < len(result); i++ {
			if result[i].channel.CategoryID == result[i-1].channel.CategoryID &&
				result[i-1].channel.CategoryID == result[i-2].channel.CategoryID &&
				result[i-2].channel.CategoryID == result[i-3].channel.CategoryID {
				t.Error("同一カテゴリが 4 件以上連続している")
			}
		}
	})

	t.Run("全て同じカテゴリでも全件が結果に含まれる", func(t *testing.T) {
		catA := uuid.New()

		scored := []scoredChannel{
			{channel: repository.RecommendedChannel{Channel: model.Channel{ID: uuid.New(), CategoryID: catA}}, score: 100},
			{channel: repository.RecommendedChannel{Channel: model.Channel{ID: uuid.New(), CategoryID: catA}}, score: 90},
			{channel: repository.RecommendedChannel{Channel: model.Channel{ID: uuid.New(), CategoryID: catA}}, score: 80},
			{channel: repository.RecommendedChannel{Channel: model.Channel{ID: uuid.New(), CategoryID: catA}}, score: 70},
		}

		result := svc.applyDiversityFilter(scored)

		assert.Len(t, result, 4)
	})

	t.Run("異なるカテゴリの場合はスコア順のまま", func(t *testing.T) {
		catA := uuid.New()
		catB := uuid.New()
		catC := uuid.New()

		scored := []scoredChannel{
			{channel: repository.RecommendedChannel{Channel: model.Channel{ID: uuid.New(), CategoryID: catA}}, score: 100},
			{channel: repository.RecommendedChannel{Channel: model.Channel{ID: uuid.New(), CategoryID: catB}}, score: 90},
			{channel: repository.RecommendedChannel{Channel: model.Channel{ID: uuid.New(), CategoryID: catC}}, score: 80},
		}

		result := svc.applyDiversityFilter(scored)

		assert.Len(t, result, 3)
		assert.Equal(t, 100.0, result[0].score)
		assert.Equal(t, 90.0, result[1].score)
		assert.Equal(t, 80.0, result[2].score)
	})
}

func TestRecommendationService_applyPersonalizedScores(t *testing.T) {
	ctx := context.Background()

	t.Run("自分のチャンネルは除外される", func(t *testing.T) {
		mockRepo := new(mockRecommendationRepository)
		svc := &recommendationService{recommendationRepo: mockRepo}

		userID := uuid.New()
		ownChannelID := uuid.New()
		otherChannelID := uuid.New()
		catID := uuid.New()

		scored := []scoredChannel{
			{channel: repository.RecommendedChannel{Channel: model.Channel{ID: ownChannelID, CategoryID: catID}}, score: 100},
			{channel: repository.RecommendedChannel{Channel: model.Channel{ID: otherChannelID, CategoryID: catID}}, score: 50},
		}

		mockRepo.On("FindUserCategoryPreferences", mock.Anything, userID).Return([]repository.CategoryPreference{}, nil)
		mockRepo.On("FindUserDefaultPlaylistCategoryPreferences", mock.Anything, userID).Return([]repository.CategoryPreference{}, nil)
		mockRepo.On("FindUserPlayedChannelIDs", mock.Anything, userID).Return([]uuid.UUID{}, nil)
		mockRepo.On("FindUserChannelIDs", mock.Anything, userID).Return([]uuid.UUID{ownChannelID}, nil)

		result := svc.applyPersonalizedScores(ctx, scored, userID)

		assert.Len(t, result, 1)
		assert.Equal(t, otherChannelID, result[0].channel.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("カテゴリボーナスが適用される", func(t *testing.T) {
		mockRepo := new(mockRecommendationRepository)
		svc := &recommendationService{recommendationRepo: mockRepo}

		userID := uuid.New()
		favCatID := uuid.New()
		otherCatID := uuid.New()

		scored := []scoredChannel{
			{channel: repository.RecommendedChannel{Channel: model.Channel{ID: uuid.New(), CategoryID: favCatID}}, score: 50},
			{channel: repository.RecommendedChannel{Channel: model.Channel{ID: uuid.New(), CategoryID: otherCatID}}, score: 50},
		}

		mockRepo.On("FindUserCategoryPreferences", mock.Anything, userID).Return([]repository.CategoryPreference{
			{CategoryID: favCatID, PlayCount: 20},
		}, nil)
		mockRepo.On("FindUserDefaultPlaylistCategoryPreferences", mock.Anything, userID).Return([]repository.CategoryPreference{}, nil)
		mockRepo.On("FindUserPlayedChannelIDs", mock.Anything, userID).Return([]uuid.UUID{}, nil)
		mockRepo.On("FindUserChannelIDs", mock.Anything, userID).Return([]uuid.UUID{}, nil)

		result := svc.applyPersonalizedScores(ctx, scored, userID)

		assert.Len(t, result, 2)
		// 高評価カテゴリのチャンネルのスコアが高い
		var favScore, otherScore float64
		for _, sc := range result {
			if sc.channel.CategoryID == favCatID {
				favScore = sc.score
			} else {
				otherScore = sc.score
			}
		}
		assert.Greater(t, favScore, otherScore)
		mockRepo.AssertExpectations(t)
	})

	t.Run("既再生チャンネルにペナルティが適用される", func(t *testing.T) {
		mockRepo := new(mockRecommendationRepository)
		svc := &recommendationService{recommendationRepo: mockRepo}

		userID := uuid.New()
		playedChannelID := uuid.New()
		newChannelID := uuid.New()
		catID := uuid.New()

		scored := []scoredChannel{
			{channel: repository.RecommendedChannel{Channel: model.Channel{ID: playedChannelID, CategoryID: catID}}, score: 100},
			{channel: repository.RecommendedChannel{Channel: model.Channel{ID: newChannelID, CategoryID: catID}}, score: 100},
		}

		mockRepo.On("FindUserCategoryPreferences", mock.Anything, userID).Return([]repository.CategoryPreference{}, nil)
		mockRepo.On("FindUserDefaultPlaylistCategoryPreferences", mock.Anything, userID).Return([]repository.CategoryPreference{}, nil)
		mockRepo.On("FindUserPlayedChannelIDs", mock.Anything, userID).Return([]uuid.UUID{playedChannelID}, nil)
		mockRepo.On("FindUserChannelIDs", mock.Anything, userID).Return([]uuid.UUID{}, nil)

		result := svc.applyPersonalizedScores(ctx, scored, userID)

		assert.Len(t, result, 2)
		var playedScore, newScore float64
		for _, sc := range result {
			if sc.channel.ID == playedChannelID {
				playedScore = sc.score
			} else {
				newScore = sc.score
			}
		}
		assert.Greater(t, newScore, playedScore)
		mockRepo.AssertExpectations(t)
	})

	t.Run("カテゴリ傾向取得エラー時はベーススコアのまま返す", func(t *testing.T) {
		mockRepo := new(mockRecommendationRepository)
		svc := &recommendationService{recommendationRepo: mockRepo}

		userID := uuid.New()
		catID := uuid.New()

		scored := []scoredChannel{
			{channel: repository.RecommendedChannel{Channel: model.Channel{ID: uuid.New(), CategoryID: catID}}, score: 100},
		}

		mockRepo.On("FindUserCategoryPreferences", mock.Anything, userID).Return(nil, assert.AnError)

		result := svc.applyPersonalizedScores(ctx, scored, userID)

		assert.Len(t, result, 1)
		assert.Equal(t, 100.0, result[0].score)
		mockRepo.AssertExpectations(t)
	})
}

// テスト用エピソードを生成するヘルパー
func createTestEpisode(id, channelID, categoryID uuid.UUID, playCount int, publishedAt *time.Time) model.Episode {
	return model.Episode{
		ID:          id,
		ChannelID:   channelID,
		Title:       "Episode " + id.String()[:8],
		Description: "Test episode",
		PlayCount:   playCount,
		PublishedAt: publishedAt,
		Channel: model.Channel{
			ID:         channelID,
			CategoryID: categoryID,
			Name:       "Channel " + channelID.String()[:8],
			Category: model.Category{
				ID:   categoryID,
				Slug: "cat-" + categoryID.String()[:8],
				Name: "Category " + categoryID.String()[:8],
			},
		},
	}
}

func TestRecommendationService_GetRecommendedEpisodes(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	categoryID := uuid.New()
	channelID := uuid.New()

	t.Run("未ログインでおすすめエピソード一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockRecommendationRepository)
		mockStorage := new(mockStorageClient)

		episodes := []model.Episode{
			createTestEpisode(uuid.New(), channelID, categoryID, 100, &now),
			createTestEpisode(uuid.New(), channelID, categoryID, 50, &now),
		}
		mockRepo.On("FindPublishedEpisodes", mock.Anything, mock.AnythingOfType("repository.RecommendEpisodeParams")).Return(episodes, int64(2), nil)

		svc := NewRecommendationService(mockRepo, mockStorage)

		req := request.RecommendEpisodesRequest{
			PaginationRequest: request.PaginationRequest{Limit: 20, Offset: 0},
		}
		result, err := svc.GetRecommendedEpisodes(ctx, nil, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, int64(2), result.Pagination.Total)
		// 全て playbackProgress=nil, inDefaultPlaylist=false
		for _, ep := range result.Data {
			assert.Nil(t, ep.PlaybackProgress)
			assert.False(t, ep.InDefaultPlaylist)
		}
		mockRepo.AssertExpectations(t)
	})

	t.Run("ログイン時は途中再生→再生リスト→パーソナライズの順で返す", func(t *testing.T) {
		mockRepo := new(mockRecommendationRepository)
		mockStorage := new(mockStorageClient)

		userID := uuid.New()
		userIDStr := userID.String()
		channelA := uuid.New()
		channelB := uuid.New()
		catA := uuid.New()
		catB := uuid.New()

		inProgressEpID := uuid.New()
		defaultPlaylistEpID := uuid.New()
		normalEpID := uuid.New()

		episodes := []model.Episode{
			createTestEpisode(inProgressEpID, channelA, catA, 200, &now),
			createTestEpisode(defaultPlaylistEpID, channelB, catB, 100, &now),
			createTestEpisode(normalEpID, channelA, catA, 50, &now),
		}

		mockRepo.On("FindPublishedEpisodes", mock.Anything, mock.AnythingOfType("repository.RecommendEpisodeParams")).Return(episodes, int64(3), nil)
		mockRepo.On("FindUserPlaybackHistories", mock.Anything, userID).Return([]model.PlaybackHistory{
			{EpisodeID: inProgressEpID, ProgressMs: 30000, Completed: false, PlayedAt: now},
		}, nil)
		mockRepo.On("FindUserDefaultPlaylistEpisodeIDs", mock.Anything, userID).Return([]uuid.UUID{defaultPlaylistEpID}, nil)
		mockRepo.On("FindUserChannelIDs", mock.Anything, userID).Return([]uuid.UUID{}, nil)
		mockRepo.On("FindUserCategoryPreferences", mock.Anything, userID).Return([]repository.CategoryPreference{}, nil)

		svc := NewRecommendationService(mockRepo, mockStorage)

		req := request.RecommendEpisodesRequest{
			PaginationRequest: request.PaginationRequest{Limit: 20, Offset: 0},
		}
		result, err := svc.GetRecommendedEpisodes(ctx, &userIDStr, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 3)
		// 途中再生が最初
		assert.Equal(t, inProgressEpID, result.Data[0].ID)
		// 再生リストが次
		assert.Equal(t, defaultPlaylistEpID, result.Data[1].ID)
		assert.True(t, result.Data[1].InDefaultPlaylist)
		// 途中再生のエピソードには再生進捗がある
		assert.NotNil(t, result.Data[0].PlaybackProgress)
		assert.Equal(t, 30000, result.Data[0].PlaybackProgress.ProgressMs)
		mockRepo.AssertExpectations(t)
	})

	t.Run("完了済みエピソードはパーソナライズから除外される", func(t *testing.T) {
		mockRepo := new(mockRecommendationRepository)
		mockStorage := new(mockStorageClient)

		userID := uuid.New()
		userIDStr := userID.String()
		completedEpID := uuid.New()
		normalEpID := uuid.New()

		episodes := []model.Episode{
			createTestEpisode(completedEpID, channelID, categoryID, 200, &now),
			createTestEpisode(normalEpID, channelID, categoryID, 50, &now),
		}

		mockRepo.On("FindPublishedEpisodes", mock.Anything, mock.AnythingOfType("repository.RecommendEpisodeParams")).Return(episodes, int64(2), nil)
		mockRepo.On("FindUserPlaybackHistories", mock.Anything, userID).Return([]model.PlaybackHistory{
			{EpisodeID: completedEpID, ProgressMs: 180000, Completed: true, PlayedAt: now},
		}, nil)
		mockRepo.On("FindUserDefaultPlaylistEpisodeIDs", mock.Anything, userID).Return([]uuid.UUID{}, nil)
		mockRepo.On("FindUserChannelIDs", mock.Anything, userID).Return([]uuid.UUID{}, nil)
		mockRepo.On("FindUserCategoryPreferences", mock.Anything, userID).Return([]repository.CategoryPreference{}, nil)

		svc := NewRecommendationService(mockRepo, mockStorage)

		req := request.RecommendEpisodesRequest{
			PaginationRequest: request.PaginationRequest{Limit: 20, Offset: 0},
		}
		result, err := svc.GetRecommendedEpisodes(ctx, &userIDStr, req)

		assert.NoError(t, err)
		assert.Len(t, result.Data, 1)
		assert.Equal(t, normalEpID, result.Data[0].ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("リポジトリエラー時はエラーを返す", func(t *testing.T) {
		mockRepo := new(mockRecommendationRepository)
		mockStorage := new(mockStorageClient)

		mockRepo.On("FindPublishedEpisodes", mock.Anything, mock.AnythingOfType("repository.RecommendEpisodeParams")).Return(nil, int64(0), assert.AnError)

		svc := NewRecommendationService(mockRepo, mockStorage)

		req := request.RecommendEpisodesRequest{
			PaginationRequest: request.PaginationRequest{Limit: 20, Offset: 0},
		}
		result, err := svc.GetRecommendedEpisodes(ctx, nil, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestRecommendationService_calculateEpisodeBaseScore(t *testing.T) {
	svc := &recommendationService{}
	now := time.Now()

	t.Run("再生回数が多いほどスコアが高い", func(t *testing.T) {
		ep1 := &model.Episode{PlayCount: 10, PublishedAt: &now}
		ep2 := &model.Episode{PlayCount: 1000, PublishedAt: &now}

		score1 := svc.calculateEpisodeBaseScore(ep1, now)
		score2 := svc.calculateEpisodeBaseScore(ep2, now)

		assert.Greater(t, score2, score1)
	})

	t.Run("新しいエピソードほどスコアが高い", func(t *testing.T) {
		recent := now.Add(-1 * 24 * time.Hour)
		old := now.Add(-60 * 24 * time.Hour)

		ep1 := &model.Episode{PlayCount: 100, PublishedAt: &old}
		ep2 := &model.Episode{PlayCount: 100, PublishedAt: &recent}

		score1 := svc.calculateEpisodeBaseScore(ep1, now)
		score2 := svc.calculateEpisodeBaseScore(ep2, now)

		assert.Greater(t, score2, score1)
	})
}

func TestRecommendationService_applyEpisodeDiversityFilter(t *testing.T) {
	svc := &recommendationService{}

	t.Run("空のスライスではエラーにならない", func(t *testing.T) {
		result := svc.applyEpisodeDiversityFilter([]scoredEpisode{})
		assert.Empty(t, result)
	})

	t.Run("同一チャンネルが 2 件以上連続しない", func(t *testing.T) {
		chA := uuid.New()
		chB := uuid.New()

		scored := []scoredEpisode{
			{episode: &model.Episode{ID: uuid.New(), ChannelID: chA}, score: 100},
			{episode: &model.Episode{ID: uuid.New(), ChannelID: chA}, score: 90},
			{episode: &model.Episode{ID: uuid.New(), ChannelID: chA}, score: 80},
			{episode: &model.Episode{ID: uuid.New(), ChannelID: chB}, score: 70},
		}

		result := svc.applyEpisodeDiversityFilter(scored)

		assert.Len(t, result, 4)
		// 同一チャンネルが 3 件連続しないことを確認
		for i := 2; i < len(result); i++ {
			if result[i].episode.ChannelID == result[i-1].episode.ChannelID &&
				result[i-1].episode.ChannelID == result[i-2].episode.ChannelID {
				t.Error("同一チャンネルが 3 件以上連続している")
			}
		}
	})

	t.Run("異なるチャンネルの場合はスコア順のまま", func(t *testing.T) {
		chA := uuid.New()
		chB := uuid.New()
		chC := uuid.New()

		scored := []scoredEpisode{
			{episode: &model.Episode{ID: uuid.New(), ChannelID: chA}, score: 100},
			{episode: &model.Episode{ID: uuid.New(), ChannelID: chB}, score: 90},
			{episode: &model.Episode{ID: uuid.New(), ChannelID: chC}, score: 80},
		}

		result := svc.applyEpisodeDiversityFilter(scored)

		assert.Len(t, result, 3)
		assert.Equal(t, 100.0, result[0].score)
		assert.Equal(t, 90.0, result[1].score)
		assert.Equal(t, 80.0, result[2].score)
	})
}
