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

func (m *mockRecommendationRepository) FindUserListenLaterCategoryPreferences(ctx context.Context, userID uuid.UUID) ([]repository.CategoryPreference, error) {
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
		mockRepo.On("FindUserListenLaterCategoryPreferences", mock.Anything, userID).Return([]repository.CategoryPreference{}, nil)
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
		mockRepo.On("FindUserListenLaterCategoryPreferences", mock.Anything, userID).Return([]repository.CategoryPreference{}, nil)
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
		mockRepo.On("FindUserListenLaterCategoryPreferences", mock.Anything, userID).Return([]repository.CategoryPreference{}, nil)
		mockRepo.On("FindUserPlayedChannelIDs", mock.Anything, userID).Return([]uuid.UUID{}, nil)
		mockRepo.On("FindUserChannelIDs", mock.Anything, userID).Return([]uuid.UUID{}, nil)

		result := svc.applyPersonalizedScores(ctx, scored, userID)

		assert.Len(t, result, 2)
		// お気に入りカテゴリのチャンネルのスコアが高い
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
		mockRepo.On("FindUserListenLaterCategoryPreferences", mock.Anything, userID).Return([]repository.CategoryPreference{}, nil)
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
