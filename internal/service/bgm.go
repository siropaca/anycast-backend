package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// BGM 関連のビジネスロジックインターフェース
type BgmService interface {
	ListMyBgms(ctx context.Context, userID string, req request.ListMyBgmsRequest) (*response.BgmListWithPaginationResponse, error)
}

type bgmService struct {
	bgmRepo        repository.BgmRepository
	defaultBgmRepo repository.DefaultBgmRepository
	storageClient  storage.Client
}

// BgmService の実装を返す
func NewBgmService(
	bgmRepo repository.BgmRepository,
	defaultBgmRepo repository.DefaultBgmRepository,
	storageClient storage.Client,
) BgmService {
	return &bgmService{
		bgmRepo:        bgmRepo,
		defaultBgmRepo: defaultBgmRepo,
		storageClient:  storageClient,
	}
}

// 自分の BGM 一覧を取得する
func (s *bgmService) ListMyBgms(ctx context.Context, userID string, req request.ListMyBgmsRequest) (*response.BgmListWithPaginationResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	if req.IncludeDefault {
		return s.listBgmsWithDefault(ctx, uid, req)
	}

	return s.listUserBgmsOnly(ctx, uid, req)
}

// ユーザー BGM のみを取得する
func (s *bgmService) listUserBgmsOnly(ctx context.Context, userID uuid.UUID, req request.ListMyBgmsRequest) (*response.BgmListWithPaginationResponse, error) {
	filter := repository.BgmFilter{
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	bgms, total, err := s.bgmRepo.FindByUserID(ctx, userID, filter)
	if err != nil {
		return nil, err
	}

	responses, err := s.toBgmResponses(ctx, bgms)
	if err != nil {
		return nil, err
	}

	return &response.BgmListWithPaginationResponse{
		Data:       responses,
		Pagination: response.PaginationResponse{Total: total, Limit: req.Limit, Offset: req.Offset},
	}, nil
}

// デフォルト BGM とユーザー BGM を結合して取得する
func (s *bgmService) listBgmsWithDefault(ctx context.Context, userID uuid.UUID, req request.ListMyBgmsRequest) (*response.BgmListWithPaginationResponse, error) {
	// デフォルト BGM の総件数を取得
	defaultTotal, err := s.defaultBgmRepo.CountActive(ctx)
	if err != nil {
		return nil, err
	}

	// ユーザー BGM の総件数を取得
	userBgms, userTotal, err := s.bgmRepo.FindByUserID(ctx, userID, repository.BgmFilter{Limit: 0, Offset: 0})
	if err != nil {
		return nil, err
	}
	_ = userBgms // 総件数のみ使用

	total := defaultTotal + userTotal

	var responses []response.BgmResponse

	// オフセットがデフォルト BGM の範囲内の場合
	if int64(req.Offset) < defaultTotal {
		// デフォルト BGM を取得
		defaultFilter := repository.DefaultBgmFilter{
			Limit:  req.Limit,
			Offset: req.Offset,
		}
		defaultBgms, _, err := s.defaultBgmRepo.FindActive(ctx, defaultFilter)
		if err != nil {
			return nil, err
		}

		defaultResponses, err := s.toDefaultBgmResponses(ctx, defaultBgms)
		if err != nil {
			return nil, err
		}
		responses = append(responses, defaultResponses...)

		// デフォルト BGM だけで limit を満たさない場合、ユーザー BGM も取得
		remaining := req.Limit - len(responses)
		if remaining > 0 {
			userFilter := repository.BgmFilter{
				Limit:  remaining,
				Offset: 0,
			}
			userBgms, _, err := s.bgmRepo.FindByUserID(ctx, userID, userFilter)
			if err != nil {
				return nil, err
			}

			userResponses, err := s.toBgmResponses(ctx, userBgms)
			if err != nil {
				return nil, err
			}
			responses = append(responses, userResponses...)
		}
	} else {
		// オフセットがデフォルト BGM の範囲外の場合、ユーザー BGM のみ取得
		userOffset := req.Offset - int(defaultTotal)
		userFilter := repository.BgmFilter{
			Limit:  req.Limit,
			Offset: userOffset,
		}
		userBgms, _, err := s.bgmRepo.FindByUserID(ctx, userID, userFilter)
		if err != nil {
			return nil, err
		}

		userResponses, err := s.toBgmResponses(ctx, userBgms)
		if err != nil {
			return nil, err
		}
		responses = append(responses, userResponses...)
	}

	return &response.BgmListWithPaginationResponse{
		Data:       responses,
		Pagination: response.PaginationResponse{Total: total, Limit: req.Limit, Offset: req.Offset},
	}, nil
}

// Bgm モデルのスライスをレスポンス DTO のスライスに変換する
func (s *bgmService) toBgmResponses(ctx context.Context, bgms []model.Bgm) ([]response.BgmResponse, error) {
	result := make([]response.BgmResponse, len(bgms))

	for i, b := range bgms {
		res, err := s.toBgmResponse(ctx, b)
		if err != nil {
			return nil, err
		}
		result[i] = res
	}

	return result, nil
}

// Bgm モデルをレスポンス DTO に変換する
func (s *bgmService) toBgmResponse(ctx context.Context, b model.Bgm) (response.BgmResponse, error) {
	audioResponse, err := s.toAudioResponse(ctx, b.Audio)
	if err != nil {
		return response.BgmResponse{}, err
	}

	return response.BgmResponse{
		ID:        b.ID,
		Name:      b.Name,
		IsDefault: false,
		Audio:     audioResponse,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}, nil
}

// DefaultBgm モデルのスライスをレスポンス DTO のスライスに変換する
func (s *bgmService) toDefaultBgmResponses(ctx context.Context, bgms []model.DefaultBgm) ([]response.BgmResponse, error) {
	result := make([]response.BgmResponse, len(bgms))

	for i, b := range bgms {
		res, err := s.toDefaultBgmResponse(ctx, b)
		if err != nil {
			return nil, err
		}
		result[i] = res
	}

	return result, nil
}

// DefaultBgm モデルをレスポンス DTO に変換する
func (s *bgmService) toDefaultBgmResponse(ctx context.Context, b model.DefaultBgm) (response.BgmResponse, error) {
	audioResponse, err := s.toAudioResponse(ctx, b.Audio)
	if err != nil {
		return response.BgmResponse{}, err
	}

	return response.BgmResponse{
		ID:        b.ID,
		Name:      b.Name,
		IsDefault: true,
		Audio:     audioResponse,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}, nil
}

// Audio モデルをレスポンス DTO に変換する
func (s *bgmService) toAudioResponse(ctx context.Context, audio model.Audio) (response.BgmAudioResponse, error) {
	var url string
	if s.storageClient != nil {
		signedURL, err := s.storageClient.GenerateSignedURL(ctx, audio.Path, storage.SignedURLExpirationAudio)
		if err == nil {
			url = signedURL
		}
		// URL 生成に失敗した場合は空文字のまま
	}

	return response.BgmAudioResponse{
		ID:         audio.ID,
		URL:        url,
		DurationMs: audio.DurationMs,
	}, nil
}
