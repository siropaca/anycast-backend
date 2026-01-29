package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// BgmService は BGM 関連のビジネスロジックインターフェースを表す
type BgmService interface {
	ListMyBgms(ctx context.Context, userID string, req request.ListMyBgmsRequest) (*response.BgmListWithPaginationResponse, error)
	GetMyBgm(ctx context.Context, userID string, bgmID string) (*response.BgmDataResponse, error)
	CreateBgm(ctx context.Context, userID string, req request.CreateBgmRequest) (*response.BgmDataResponse, error)
	UpdateMyBgm(ctx context.Context, userID string, bgmID string, req request.UpdateBgmRequest) (*response.BgmDataResponse, error)
	DeleteMyBgm(ctx context.Context, userID string, bgmID string) error
}

type bgmService struct {
	bgmRepo       repository.BgmRepository
	systemBgmRepo repository.SystemBgmRepository
	audioRepo     repository.AudioRepository
	storageClient storage.Client
}

// NewBgmService は bgmService を生成して BgmService として返す
func NewBgmService(
	bgmRepo repository.BgmRepository,
	systemBgmRepo repository.SystemBgmRepository,
	audioRepo repository.AudioRepository,
	storageClient storage.Client,
) BgmService {
	return &bgmService{
		bgmRepo:       bgmRepo,
		systemBgmRepo: systemBgmRepo,
		audioRepo:     audioRepo,
		storageClient: storageClient,
	}
}

// ListMyBgms は自分の BGM 一覧を取得する
func (s *bgmService) ListMyBgms(ctx context.Context, userID string, req request.ListMyBgmsRequest) (*response.BgmListWithPaginationResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	if req.IncludeSystem {
		return s.listBgmsWithSystem(ctx, uid, req)
	}

	return s.listUserBgmsOnly(ctx, uid, req)
}

// listUserBgmsOnly はユーザー BGM のみを取得する
func (s *bgmService) listUserBgmsOnly(ctx context.Context, userID uuid.UUID, req request.ListMyBgmsRequest) (*response.BgmListWithPaginationResponse, error) {
	filter := repository.BgmFilter{
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	bgms, total, err := s.bgmRepo.FindByUserID(ctx, userID, filter)
	if err != nil {
		return nil, err
	}

	responses, err := s.toBgmWithEpisodesResponses(ctx, bgms)
	if err != nil {
		return nil, err
	}

	return &response.BgmListWithPaginationResponse{
		Data:       responses,
		Pagination: response.PaginationResponse{Total: total, Limit: req.Limit, Offset: req.Offset},
	}, nil
}

// listBgmsWithSystem はシステム BGM とユーザー BGM を結合して取得する
func (s *bgmService) listBgmsWithSystem(ctx context.Context, userID uuid.UUID, req request.ListMyBgmsRequest) (*response.BgmListWithPaginationResponse, error) {
	// システム BGM の総件数を取得
	systemTotal, err := s.systemBgmRepo.CountActive(ctx)
	if err != nil {
		return nil, err
	}

	// ユーザー BGM の総件数を取得
	userBgms, userTotal, err := s.bgmRepo.FindByUserID(ctx, userID, repository.BgmFilter{Limit: 0, Offset: 0})
	if err != nil {
		return nil, err
	}
	_ = userBgms // 総件数のみ使用

	total := systemTotal + userTotal

	responses := make([]response.BgmWithEpisodesResponse, 0)

	// オフセットがシステム BGM の範囲内の場合
	if int64(req.Offset) < systemTotal {
		// システム BGM を取得
		systemFilter := repository.SystemBgmFilter{
			Limit:  req.Limit,
			Offset: req.Offset,
		}
		systemBgms, _, err := s.systemBgmRepo.FindActive(ctx, systemFilter)
		if err != nil {
			return nil, err
		}

		systemResponses, err := s.toSystemBgmWithEpisodesResponses(ctx, systemBgms)
		if err != nil {
			return nil, err
		}
		responses = append(responses, systemResponses...)

		// システム BGM だけで limit を満たさない場合、ユーザー BGM も取得
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

			userResponses, err := s.toBgmWithEpisodesResponses(ctx, userBgms)
			if err != nil {
				return nil, err
			}
			responses = append(responses, userResponses...)
		}
	} else {
		// オフセットがシステム BGM の範囲外の場合、ユーザー BGM のみ取得
		userOffset := req.Offset - int(systemTotal)
		userFilter := repository.BgmFilter{
			Limit:  req.Limit,
			Offset: userOffset,
		}
		userBgms, _, err := s.bgmRepo.FindByUserID(ctx, userID, userFilter)
		if err != nil {
			return nil, err
		}

		userResponses, err := s.toBgmWithEpisodesResponses(ctx, userBgms)
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

// toBgmWithEpisodesResponses は Bgm のスライスをエピソード情報付きレスポンス DTO のスライスに変換する
func (s *bgmService) toBgmWithEpisodesResponses(ctx context.Context, bgms []model.Bgm) ([]response.BgmWithEpisodesResponse, error) {
	result := make([]response.BgmWithEpisodesResponse, len(bgms))

	for i, b := range bgms {
		res, err := s.toBgmWithEpisodesResponse(ctx, b)
		if err != nil {
			return nil, err
		}
		result[i] = res
	}

	return result, nil
}

// toBgmWithEpisodesResponse は Bgm をエピソード・チャンネル情報付きレスポンス DTO に変換する
func (s *bgmService) toBgmWithEpisodesResponse(ctx context.Context, b model.Bgm) (response.BgmWithEpisodesResponse, error) {
	audioResponse, err := s.toAudioResponse(ctx, b.Audio)
	if err != nil {
		return response.BgmWithEpisodesResponse{}, err
	}

	episodes := make([]response.BgmEpisodeResponse, len(b.Episodes))
	for i, e := range b.Episodes {
		episodes[i] = response.BgmEpisodeResponse{
			ID:    e.ID,
			Title: e.Title,
			Channel: response.BgmEpisodeChannelResponse{
				ID:   e.Channel.ID,
				Name: e.Channel.Name,
			},
		}
	}

	channels := make([]response.BgmChannelResponse, len(b.Channels))
	for i, c := range b.Channels {
		channels[i] = response.BgmChannelResponse{
			ID:   c.ID,
			Name: c.Name,
		}
	}

	return response.BgmWithEpisodesResponse{
		ID:        b.ID,
		Name:      b.Name,
		IsSystem: false,
		Audio:     audioResponse,
		Episodes:  episodes,
		Channels:  channels,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}, nil
}

// toSystemBgmWithEpisodesResponses は SystemBgm のスライスをエピソード情報付きレスポンス DTO のスライスに変換する
func (s *bgmService) toSystemBgmWithEpisodesResponses(ctx context.Context, bgms []model.SystemBgm) ([]response.BgmWithEpisodesResponse, error) {
	result := make([]response.BgmWithEpisodesResponse, len(bgms))

	for i, b := range bgms {
		res, err := s.toSystemBgmWithEpisodesResponse(ctx, b)
		if err != nil {
			return nil, err
		}
		result[i] = res
	}

	return result, nil
}

// toSystemBgmWithEpisodesResponse は SystemBgm をエピソード・チャンネル情報付きレスポンス DTO に変換する
// SystemBgm はシステム提供の BGM のため、episodes と channels は常に空配列を返す
func (s *bgmService) toSystemBgmWithEpisodesResponse(ctx context.Context, b model.SystemBgm) (response.BgmWithEpisodesResponse, error) {
	audioResponse, err := s.toAudioResponse(ctx, b.Audio)
	if err != nil {
		return response.BgmWithEpisodesResponse{}, err
	}

	return response.BgmWithEpisodesResponse{
		ID:        b.ID,
		Name:      b.Name,
		IsSystem: true,
		Audio:     audioResponse,
		Episodes:  []response.BgmEpisodeResponse{},
		Channels:  []response.BgmChannelResponse{},
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}, nil
}

// toAudioResponse は Audio をレスポンス DTO に変換する
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

// CreateBgm は新しい BGM を作成する
func (s *bgmService) CreateBgm(ctx context.Context, userID string, req request.CreateBgmRequest) (*response.BgmDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	// 同一ユーザー内で同じ名前の BGM が存在するかチェック
	exists, err := s.bgmRepo.ExistsByUserIDAndName(ctx, uid, req.Name, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.ErrDuplicateName.WithMessage("同じ名前の BGM が既に存在します")
	}

	// 音声ファイルの存在確認
	audioID, err := uuid.Parse(req.AudioID)
	if err != nil {
		return nil, err
	}
	audio, err := s.audioRepo.FindByID(ctx, audioID)
	if err != nil {
		return nil, err
	}

	// BGM を作成
	bgm := &model.Bgm{
		ID:      uuid.New(),
		UserID:  uid,
		AudioID: audioID,
		Name:    req.Name,
	}

	if err := s.bgmRepo.Create(ctx, bgm); err != nil {
		return nil, err
	}

	// レスポンス用にリレーションを設定
	bgm.Audio = *audio
	bgm.Episodes = []model.Episode{} // 新規作成時は空
	bgm.Channels = []model.Channel{} // 新規作成時は空

	res, err := s.toBgmWithEpisodesResponse(ctx, *bgm)
	if err != nil {
		return nil, err
	}

	return &response.BgmDataResponse{Data: res}, nil
}

// GetMyBgm は自分の BGM を取得する
func (s *bgmService) GetMyBgm(ctx context.Context, userID, bgmID string) (*response.BgmDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	bid, err := uuid.Parse(bgmID)
	if err != nil {
		return nil, err
	}

	bgm, err := s.bgmRepo.FindByID(ctx, bid)
	if err != nil {
		return nil, err
	}

	// 所有者チェック
	if bgm.UserID != uid {
		return nil, apperror.ErrNotFound.WithMessage("BGM が見つかりません")
	}

	res, err := s.toBgmWithEpisodesResponse(ctx, *bgm)
	if err != nil {
		return nil, err
	}

	return &response.BgmDataResponse{Data: res}, nil
}

// UpdateMyBgm は自分の BGM を更新する
func (s *bgmService) UpdateMyBgm(ctx context.Context, userID, bgmID string, req request.UpdateBgmRequest) (*response.BgmDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	bid, err := uuid.Parse(bgmID)
	if err != nil {
		return nil, err
	}

	bgm, err := s.bgmRepo.FindByID(ctx, bid)
	if err != nil {
		return nil, err
	}

	// 所有者チェック
	if bgm.UserID != uid {
		return nil, apperror.ErrNotFound.WithMessage("BGM が見つかりません")
	}

	// 名前の更新
	if req.Name != nil {
		// 同一ユーザー内で同じ名前の BGM が存在するかチェック（自分自身を除く）
		exists, err := s.bgmRepo.ExistsByUserIDAndName(ctx, uid, *req.Name, &bid)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperror.ErrDuplicateName.WithMessage("同じ名前の BGM が既に存在します")
		}
		bgm.Name = *req.Name
	}

	if err := s.bgmRepo.Update(ctx, bgm); err != nil {
		return nil, err
	}

	res, err := s.toBgmWithEpisodesResponse(ctx, *bgm)
	if err != nil {
		return nil, err
	}

	return &response.BgmDataResponse{Data: res}, nil
}

// DeleteMyBgm は自分の BGM を削除する
func (s *bgmService) DeleteMyBgm(ctx context.Context, userID, bgmID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	bid, err := uuid.Parse(bgmID)
	if err != nil {
		return err
	}

	bgm, err := s.bgmRepo.FindByID(ctx, bid)
	if err != nil {
		return err
	}

	// 所有者チェック
	if bgm.UserID != uid {
		return apperror.ErrNotFound.WithMessage("BGM が見つかりません")
	}

	// 使用中チェック
	isUsed, err := s.bgmRepo.IsUsedInAnyEpisode(ctx, bid)
	if err != nil {
		return err
	}
	if isUsed {
		return apperror.ErrBgmInUse.WithMessage("この BGM はエピソードで使用中のため削除できません")
	}

	return s.bgmRepo.Delete(ctx, bid)
}
