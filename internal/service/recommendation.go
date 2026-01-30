package service

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// RecommendationService はおすすめチャンネル関連のビジネスロジックインターフェースを表す
type RecommendationService interface {
	GetRecommendedChannels(ctx context.Context, userID *string, req request.RecommendChannelsRequest) (*response.RecommendedChannelListResponse, error)
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
)

// scoredChannel はスコア付きチャンネルを表す
type scoredChannel struct {
	channel repository.RecommendedChannel
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
