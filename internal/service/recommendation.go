package service

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// RecommendationService はおすすめ関連のビジネスロジックインターフェースを表す
type RecommendationService interface {
	GetRecommendedChannels(ctx context.Context, userID *string, req request.RecommendChannelsRequest) (*response.RecommendedChannelListResponse, error)
	GetRecommendedEpisodes(ctx context.Context, userID *string, req request.RecommendEpisodesRequest) (*response.RecommendedEpisodeListResponse, error)
}

type recommendationService struct {
	recommendationRepo repository.RecommendationRepository
	storageClient      storage.Client
}

// NewRecommendationService は recommendationService を生成して RecommendationService として返す
func NewRecommendationService(
	recommendationRepo repository.RecommendationRepository,
	storageClient storage.Client,
) RecommendationService {
	return &recommendationService{
		recommendationRepo: recommendationRepo,
		storageClient:      storageClient,
	}
}

// スコア計算で使用する重み定数
const (
	playCountWeight = 0.4
	recencyWeight   = 0.6
	recencyHalfLife = 30.0 // 日数ベースの半減期

	categoryBonusWeight    = 50.0
	listenLaterBonusWeight = 30.0
	playedChannelPenalty   = -40.0
	maxConsecutiveCategory = 3
	maxRecommendationLimit = 50

	// エピソード用定数
	episodeRecencyHalfLife     = 14.0 // エピソードは 2 週間の半減期
	episodeCategoryBonusWeight = 30.0
	maxConsecutiveChannel      = 2
	maxInProgressEpisodes      = 3
	maxListenLaterEpisodes     = 3
)

// scoredChannel はスコア付きチャンネルを表す
type scoredChannel struct {
	channel repository.RecommendedChannel
	score   float64
}

// scoredEpisode はスコア付きエピソードを表す
type scoredEpisode struct {
	episode *model.Episode
	score   float64
}

// GetRecommendedChannels はおすすめチャンネル一覧を取得する
func (s *recommendationService) GetRecommendedChannels(ctx context.Context, userID *string, req request.RecommendChannelsRequest) (*response.RecommendedChannelListResponse, error) {
	// limit の上限を 50 に制限
	limit := req.Limit
	if limit > maxRecommendationLimit {
		limit = maxRecommendationLimit
	}

	var categoryID *uuid.UUID
	if req.CategoryID != nil {
		parsed, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			return nil, err
		}
		categoryID = &parsed
	}

	// 全件取得してスコア計算後にページネーション
	params := repository.RecommendChannelParams{
		CategoryID: categoryID,
		Limit:      1000, // スコア計算用に多めに取得
		Offset:     0,
	}

	channels, total, err := s.recommendationRepo.FindRecommendedChannels(ctx, params)
	if err != nil {
		return nil, err
	}

	// スコア計算
	scored := s.calculateScores(ctx, channels, userID)

	// 多様性フィルタを適用
	scored = s.applyDiversityFilter(scored)

	// ページネーション
	filteredTotal := int64(len(scored))
	start := req.Offset
	if start > len(scored) {
		start = len(scored)
	}
	end := start + limit
	if end > len(scored) {
		end = len(scored)
	}
	paged := scored[start:end]

	// レスポンスに変換
	data := make([]response.RecommendedChannelResponse, 0, len(paged))
	for _, sc := range paged {
		item := s.toRecommendedChannelResponse(ctx, &sc.channel)
		data = append(data, item)
	}

	// total はフィルタ前のDB上の総件数を使用（カテゴリフィルタ時はフィルタ済み件数）
	// ただし多様性フィルタ適用後の件数をページネーションに反映
	_ = total

	return &response.RecommendedChannelListResponse{
		Data: data,
		Pagination: response.PaginationResponse{
			Total:  filteredTotal,
			Limit:  limit,
			Offset: req.Offset,
		},
	}, nil
}

// calculateScores はチャンネルにスコアを計算する
func (s *recommendationService) calculateScores(ctx context.Context, channels []repository.RecommendedChannel, userID *string) []scoredChannel {
	now := time.Now()

	// ベーススコア計算
	scored := make([]scoredChannel, 0, len(channels))
	for _, ch := range channels {
		baseScore := s.calculateBaseScore(ch, now)
		scored = append(scored, scoredChannel{
			channel: ch,
			score:   baseScore,
		})
	}

	// ログイン時の追加スコア
	if userID != nil {
		uid, err := uuid.Parse(*userID)
		if err == nil {
			scored = s.applyPersonalizedScores(ctx, scored, uid)
		}
	}

	// スコア順にソート
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	return scored
}

// calculateBaseScore は未ログイン時のベーススコアを計算する
func (s *recommendationService) calculateBaseScore(ch repository.RecommendedChannel, now time.Time) float64 {
	// 再生回数スコア（対数スケール）
	playScore := math.Log1p(float64(ch.TotalPlayCount)) * playCountWeight

	// 新着度スコア（最新エピソード公開日からの日数に基づく減衰）
	var recencyScore float64
	if ch.LatestEpisodeAt != nil {
		daysSince := now.Sub(*ch.LatestEpisodeAt).Hours() / 24
		recencyScore = math.Exp(-daysSince/recencyHalfLife) * recencyWeight * 100
	}

	return playScore + recencyScore
}

// applyPersonalizedScores はログインユーザーに対するパーソナライズドスコアを適用する
func (s *recommendationService) applyPersonalizedScores(ctx context.Context, scored []scoredChannel, userID uuid.UUID) []scoredChannel {
	// カテゴリ傾向を取得
	categoryPrefs, err := s.recommendationRepo.FindUserCategoryPreferences(ctx, userID)
	if err != nil {
		return scored // エラー時はベーススコアのまま
	}

	categoryBonus := make(map[uuid.UUID]float64)
	for _, pref := range categoryPrefs {
		categoryBonus[pref.CategoryID] = math.Log1p(float64(pref.PlayCount)) * categoryBonusWeight
	}

	// 「後で聴く」カテゴリ傾向
	listenLaterPrefs, err := s.recommendationRepo.FindUserListenLaterCategoryPreferences(ctx, userID)
	if err == nil {
		for _, pref := range listenLaterPrefs {
			categoryBonus[pref.CategoryID] += math.Log1p(float64(pref.PlayCount)) * listenLaterBonusWeight
		}
	}

	// 既に再生したチャンネル ID
	playedChannelIDs, err := s.recommendationRepo.FindUserPlayedChannelIDs(ctx, userID)
	if err != nil {
		playedChannelIDs = nil
	}
	playedSet := make(map[uuid.UUID]bool, len(playedChannelIDs))
	for _, id := range playedChannelIDs {
		playedSet[id] = true
	}

	// 自分のチャンネル ID
	ownChannelIDs, err := s.recommendationRepo.FindUserChannelIDs(ctx, userID)
	if err != nil {
		ownChannelIDs = nil
	}
	ownSet := make(map[uuid.UUID]bool, len(ownChannelIDs))
	for _, id := range ownChannelIDs {
		ownSet[id] = true
	}

	// スコアを調整（自分のチャンネルは除外）
	filtered := make([]scoredChannel, 0, len(scored))
	for _, sc := range scored {
		if ownSet[sc.channel.ID] {
			continue
		}

		// カテゴリボーナス
		if bonus, ok := categoryBonus[sc.channel.CategoryID]; ok {
			sc.score += bonus
		}

		// 既再生チャンネルペナルティ
		if playedSet[sc.channel.ID] {
			sc.score += playedChannelPenalty
		}

		filtered = append(filtered, sc)
	}

	return filtered
}

// applyDiversityFilter は同一カテゴリが連続しすぎないようにフィルタする
func (s *recommendationService) applyDiversityFilter(scored []scoredChannel) []scoredChannel {
	if len(scored) == 0 {
		return scored
	}

	result := make([]scoredChannel, 0, len(scored))
	used := make(map[int]bool)

	for len(result) < len(scored) {
		added := false
		for i, sc := range scored {
			if used[i] {
				continue
			}

			// 同一カテゴリの連続数をチェック
			consecutiveCount := 0
			for j := len(result) - 1; j >= 0 && j >= len(result)-maxConsecutiveCategory; j-- {
				if result[j].channel.CategoryID == sc.channel.CategoryID {
					consecutiveCount++
				} else {
					break
				}
			}

			if consecutiveCount >= maxConsecutiveCategory {
				continue
			}

			result = append(result, sc)
			used[i] = true
			added = true
			break
		}

		// 全て連続制限に引っかかった場合は残りをそのまま追加
		if !added {
			for i, sc := range scored {
				if !used[i] {
					result = append(result, sc)
					used[i] = true
				}
			}
			break
		}
	}

	return result
}

// GetRecommendedEpisodes はおすすめエピソード一覧を取得する
func (s *recommendationService) GetRecommendedEpisodes(ctx context.Context, userID *string, req request.RecommendEpisodesRequest) (*response.RecommendedEpisodeListResponse, error) {
	limit := req.Limit
	if limit > maxRecommendationLimit {
		limit = maxRecommendationLimit
	}

	var categoryID *uuid.UUID
	if req.CategoryID != nil {
		parsed, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			return nil, err
		}
		categoryID = &parsed
	}

	// スコア計算用に多めに取得
	params := repository.RecommendEpisodeParams{
		CategoryID: categoryID,
		Limit:      1000,
		Offset:     0,
	}

	episodes, _, err := s.recommendationRepo.FindPublishedEpisodes(ctx, params)
	if err != nil {
		return nil, err
	}

	// エピソードマップを構築
	episodeMap := make(map[uuid.UUID]*model.Episode, len(episodes))
	for i := range episodes {
		episodeMap[episodes[i].ID] = &episodes[i]
	}

	var scored []scoredEpisode
	var progressMap map[uuid.UUID]*model.PlaybackHistory
	var listenLaterSet map[uuid.UUID]bool

	if userID != nil {
		uid, err := uuid.Parse(*userID)
		if err != nil {
			return nil, err
		}
		scored, progressMap, listenLaterSet = s.buildPersonalizedEpisodeList(ctx, episodes, episodeMap, uid)
	} else {
		scored = s.buildBaseEpisodeList(episodes)
		progressMap = make(map[uuid.UUID]*model.PlaybackHistory)
		listenLaterSet = make(map[uuid.UUID]bool)
	}

	// 多様性フィルタを適用
	scored = s.applyEpisodeDiversityFilter(scored)

	// ページネーション
	filteredTotal := int64(len(scored))
	start := req.Offset
	if start > len(scored) {
		start = len(scored)
	}
	end := start + limit
	if end > len(scored) {
		end = len(scored)
	}
	paged := scored[start:end]

	// レスポンスに変換
	data := make([]response.RecommendedEpisodeResponse, 0, len(paged))
	for _, se := range paged {
		item := s.toRecommendedEpisodeResponse(ctx, se.episode, progressMap, listenLaterSet)
		data = append(data, item)
	}

	return &response.RecommendedEpisodeListResponse{
		Data: data,
		Pagination: response.PaginationResponse{
			Total:  filteredTotal,
			Limit:  limit,
			Offset: req.Offset,
		},
	}, nil
}

// buildBaseEpisodeList は未ログイン時のエピソードリストを構築する
func (s *recommendationService) buildBaseEpisodeList(episodes []model.Episode) []scoredEpisode {
	now := time.Now()
	scored := make([]scoredEpisode, 0, len(episodes))
	for i := range episodes {
		score := s.calculateEpisodeBaseScore(&episodes[i], now)
		scored = append(scored, scoredEpisode{
			episode: &episodes[i],
			score:   score,
		})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	return scored
}

// buildPersonalizedEpisodeList はログイン時のパーソナライズされたエピソードリストを構築する
func (s *recommendationService) buildPersonalizedEpisodeList(
	ctx context.Context,
	episodes []model.Episode,
	episodeMap map[uuid.UUID]*model.Episode,
	userID uuid.UUID,
) (scored []scoredEpisode, progressMap map[uuid.UUID]*model.PlaybackHistory, listenLaterSet map[uuid.UUID]bool) {
	now := time.Now()

	// 再生履歴を取得
	histories, err := s.recommendationRepo.FindUserPlaybackHistories(ctx, userID)
	if err != nil {
		histories = nil
	}

	progressMap = make(map[uuid.UUID]*model.PlaybackHistory, len(histories))
	completedSet := make(map[uuid.UUID]bool)
	var inProgressHistories []model.PlaybackHistory
	for i := range histories {
		progressMap[histories[i].EpisodeID] = &histories[i]
		if histories[i].Completed {
			completedSet[histories[i].EpisodeID] = true
		} else {
			inProgressHistories = append(inProgressHistories, histories[i])
		}
	}

	// 途中再生は played_at DESC でソート済み（リポジトリで ORDER BY played_at DESC）

	// 「後で聴く」エピソード ID を取得
	listenLaterIDs, err := s.recommendationRepo.FindUserListenLaterEpisodeIDs(ctx, userID)
	if err != nil {
		listenLaterIDs = nil
	}
	listenLaterSet = make(map[uuid.UUID]bool, len(listenLaterIDs))
	for _, id := range listenLaterIDs {
		listenLaterSet[id] = true
	}

	// 自分のチャンネル ID
	ownChannelIDs, err := s.recommendationRepo.FindUserChannelIDs(ctx, userID)
	if err != nil {
		ownChannelIDs = nil
	}
	ownChannelSet := make(map[uuid.UUID]bool, len(ownChannelIDs))
	for _, id := range ownChannelIDs {
		ownChannelSet[id] = true
	}

	// カテゴリ傾向
	categoryPrefs, err := s.recommendationRepo.FindUserCategoryPreferences(ctx, userID)
	if err != nil {
		categoryPrefs = nil
	}
	categoryBonus := make(map[uuid.UUID]float64)
	for _, pref := range categoryPrefs {
		categoryBonus[pref.CategoryID] = math.Log1p(float64(pref.PlayCount)) * episodeCategoryBonusWeight
	}

	included := make(map[uuid.UUID]bool)

	// 1. 途中再生中のエピソード（最大 3 件）
	count := 0
	for _, h := range inProgressHistories {
		if count >= maxInProgressEpisodes {
			break
		}
		ep, ok := episodeMap[h.EpisodeID]
		if !ok || ownChannelSet[ep.ChannelID] {
			continue
		}
		scored = append(scored, scoredEpisode{episode: ep, score: 10000 - float64(count)})
		included[ep.ID] = true
		count++
	}

	// 2. 「後で聴く」の未再生エピソード（最大 3 件）
	count = 0
	for i := range episodes {
		if count >= maxListenLaterEpisodes {
			break
		}
		ep := &episodes[i]
		if !listenLaterSet[ep.ID] || included[ep.ID] || completedSet[ep.ID] || ownChannelSet[ep.ChannelID] {
			continue
		}
		// 再生履歴がないエピソードのみ（途中再生は上のグループで処理済み）
		if _, hasHistory := progressMap[ep.ID]; hasHistory {
			continue
		}
		scored = append(scored, scoredEpisode{episode: ep, score: 5000 - float64(count)})
		included[ep.ID] = true
		count++
	}

	// 3. パーソナライズされたエピソード（残り）
	var personalized []scoredEpisode
	for i := range episodes {
		ep := &episodes[i]
		if included[ep.ID] || completedSet[ep.ID] || ownChannelSet[ep.ChannelID] {
			continue
		}
		score := s.calculateEpisodeBaseScore(ep, now)
		if bonus, ok := categoryBonus[ep.Channel.CategoryID]; ok {
			score += bonus
		}
		personalized = append(personalized, scoredEpisode{episode: ep, score: score})
	}

	sort.Slice(personalized, func(i, j int) bool {
		return personalized[i].score > personalized[j].score
	})

	scored = append(scored, personalized...)
	return
}

// calculateEpisodeBaseScore はエピソードのベーススコアを計算する
func (s *recommendationService) calculateEpisodeBaseScore(ep *model.Episode, now time.Time) float64 {
	playScore := math.Log1p(float64(ep.PlayCount)) * playCountWeight

	var recencyScore float64
	if ep.PublishedAt != nil {
		daysSince := now.Sub(*ep.PublishedAt).Hours() / 24
		recencyScore = math.Exp(-daysSince/episodeRecencyHalfLife) * recencyWeight * 100
	}

	return playScore + recencyScore
}

// applyEpisodeDiversityFilter は同一チャンネルが連続しすぎないようにフィルタする
func (s *recommendationService) applyEpisodeDiversityFilter(scored []scoredEpisode) []scoredEpisode {
	if len(scored) == 0 {
		return scored
	}

	result := make([]scoredEpisode, 0, len(scored))
	used := make(map[int]bool)

	for len(result) < len(scored) {
		added := false
		for i, se := range scored {
			if used[i] {
				continue
			}

			consecutiveCount := 0
			for j := len(result) - 1; j >= 0 && j >= len(result)-maxConsecutiveChannel; j-- {
				if result[j].episode.ChannelID == se.episode.ChannelID {
					consecutiveCount++
				} else {
					break
				}
			}

			if consecutiveCount >= maxConsecutiveChannel {
				continue
			}

			result = append(result, se)
			used[i] = true
			added = true
			break
		}

		if !added {
			for i, se := range scored {
				if !used[i] {
					result = append(result, se)
					used[i] = true
				}
			}
			break
		}
	}

	return result
}

// toRecommendedEpisodeResponse は Episode をおすすめエピソードレスポンスに変換する
func (s *recommendationService) toRecommendedEpisodeResponse(
	ctx context.Context,
	ep *model.Episode,
	progressMap map[uuid.UUID]*model.PlaybackHistory,
	listenLaterSet map[uuid.UUID]bool,
) response.RecommendedEpisodeResponse {
	// エピソードのアートワーク
	var artwork *response.ArtworkResponse
	if ep.Artwork != nil {
		url, err := s.storageClient.GenerateSignedURL(ctx, ep.Artwork.Path, storage.SignedURLExpirationImage)
		if err == nil {
			artwork = &response.ArtworkResponse{
				ID:  ep.Artwork.ID,
				URL: url,
			}
		}
	}

	// 音声ファイル
	var fullAudio *response.AudioResponse
	if ep.FullAudio != nil {
		url, err := s.storageClient.GenerateSignedURL(ctx, ep.FullAudio.Path, storage.SignedURLExpirationAudio)
		if err == nil {
			fullAudio = &response.AudioResponse{
				ID:         ep.FullAudio.ID,
				URL:        url,
				DurationMs: ep.FullAudio.DurationMs,
			}
		}
	}

	// チャンネルのアートワーク
	channel := ep.Channel
	var channelArtwork *response.ArtworkResponse
	if channel.Artwork != nil {
		url, err := s.storageClient.GenerateSignedURL(ctx, channel.Artwork.Path, storage.SignedURLExpirationImage)
		if err == nil {
			channelArtwork = &response.ArtworkResponse{
				ID:  channel.Artwork.ID,
				URL: url,
			}
		}
	}

	// 再生進捗
	var playbackProgress *response.PlaybackProgressResponse
	if h, ok := progressMap[ep.ID]; ok {
		playbackProgress = &response.PlaybackProgressResponse{
			ProgressMs: h.ProgressMs,
			Completed:  h.Completed,
		}
	}

	return response.RecommendedEpisodeResponse{
		ID:          ep.ID,
		Title:       ep.Title,
		Description: ep.Description,
		Artwork:     artwork,
		FullAudio:   fullAudio,
		PlayCount:   ep.PlayCount,
		PublishedAt: ep.PublishedAt,
		Channel: response.RecommendedEpisodeChannelResponse{
			ID:      channel.ID,
			Name:    channel.Name,
			Artwork: channelArtwork,
			Category: response.CategoryResponse{
				ID:        channel.Category.ID,
				Slug:      channel.Category.Slug,
				Name:      channel.Category.Name,
				SortOrder: channel.Category.SortOrder,
				IsActive:  channel.Category.IsActive,
			},
		},
		PlaybackProgress: playbackProgress,
		InListenLater:    listenLaterSet[ep.ID],
	}
}

// toRecommendedChannelResponse は RecommendedChannel を RecommendedChannelResponse に変換する
func (s *recommendationService) toRecommendedChannelResponse(ctx context.Context, ch *repository.RecommendedChannel) response.RecommendedChannelResponse {
	var artwork *response.ArtworkResponse
	if ch.Artwork != nil {
		url, err := s.storageClient.GenerateSignedURL(ctx, ch.Artwork.Path, storage.SignedURLExpirationImage)
		if err != nil {
			return response.RecommendedChannelResponse{}
		}
		artwork = &response.ArtworkResponse{
			ID:  ch.Artwork.ID,
			URL: url,
		}
	}

	return response.RecommendedChannelResponse{
		ID:          ch.ID,
		Name:        ch.Name,
		Description: ch.Description,
		Artwork:     artwork,
		Category: response.CategoryResponse{
			ID:        ch.Category.ID,
			Slug:      ch.Category.Slug,
			Name:      ch.Category.Name,
			SortOrder: ch.Category.SortOrder,
			IsActive:  ch.Category.IsActive,
		},
		EpisodeCount:    ch.EpisodeCount,
		TotalPlayCount:  ch.TotalPlayCount,
		LatestEpisodeAt: ch.LatestEpisodeAt,
	}
}
