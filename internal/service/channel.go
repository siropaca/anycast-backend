package service

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

const (
	// MinCharactersPerChannel はチャンネルに設定できるキャラクターの最小人数
	MinCharactersPerChannel = 1
	// MaxCharactersPerChannel はチャンネルに設定できるキャラクターの最大人数
	// Google TTS の Multi-speaker API が最大2人までしかサポートしていないため
	MaxCharactersPerChannel = 2
)

// ChannelService はチャンネル関連のビジネスロジックインターフェースを表す
type ChannelService interface {
	GetChannel(ctx context.Context, userID, channelID string) (*response.ChannelDataResponse, error)
	GetMyChannel(ctx context.Context, userID, channelID string) (*response.ChannelDataResponse, error)
	ListMyChannels(ctx context.Context, userID string, filter repository.ChannelFilter) (*response.ChannelListWithPaginationResponse, error)
	CreateChannel(ctx context.Context, userID string, req request.CreateChannelRequest) (*response.ChannelDataResponse, error)
	UpdateChannel(ctx context.Context, userID, channelID string, req request.UpdateChannelRequest) (*response.ChannelDataResponse, error)
	DeleteChannel(ctx context.Context, userID, channelID string) error
	PublishChannel(ctx context.Context, userID, channelID string, publishedAt *string) (*response.ChannelDataResponse, error)
	UnpublishChannel(ctx context.Context, userID, channelID string) (*response.ChannelDataResponse, error)
	SetUserPrompt(ctx context.Context, userID, channelID string, req request.SetUserPromptRequest) (*response.ChannelDataResponse, error)
	SetDefaultBgm(ctx context.Context, userID, channelID string, req request.SetDefaultBgmRequest) (*response.ChannelDataResponse, error)
	DeleteDefaultBgm(ctx context.Context, userID, channelID string) (*response.ChannelDataResponse, error)
	AddChannelCharacter(ctx context.Context, userID, channelID string, req request.AddChannelCharacterRequest) (*response.ChannelDataResponse, error)
	ReplaceChannelCharacter(ctx context.Context, userID, channelID, characterID string, req request.ReplaceChannelCharacterRequest) (*response.ChannelDataResponse, error)
	RemoveChannelCharacter(ctx context.Context, userID, channelID, characterID string) (*response.ChannelDataResponse, error)
}

type channelService struct {
	db                  *gorm.DB
	channelRepo         repository.ChannelRepository
	characterRepo       repository.CharacterRepository
	categoryRepo        repository.CategoryRepository
	imageRepo           repository.ImageRepository
	voiceRepo           repository.VoiceRepository
	episodeRepo         repository.EpisodeRepository
	scriptLineRepo      repository.ScriptLineRepository
	bgmRepo             repository.BgmRepository
	systemBgmRepo       repository.SystemBgmRepository
	playbackHistoryRepo repository.PlaybackHistoryRepository
	storageClient       storage.Client
}

// NewChannelService は channelService を生成して ChannelService として返す
func NewChannelService(
	db *gorm.DB,
	channelRepo repository.ChannelRepository,
	characterRepo repository.CharacterRepository,
	categoryRepo repository.CategoryRepository,
	imageRepo repository.ImageRepository,
	voiceRepo repository.VoiceRepository,
	episodeRepo repository.EpisodeRepository,
	scriptLineRepo repository.ScriptLineRepository,
	bgmRepo repository.BgmRepository,
	systemBgmRepo repository.SystemBgmRepository,
	playbackHistoryRepo repository.PlaybackHistoryRepository,
	storageClient storage.Client,
) ChannelService {
	return &channelService{
		db:                  db,
		channelRepo:         channelRepo,
		characterRepo:       characterRepo,
		categoryRepo:        categoryRepo,
		imageRepo:           imageRepo,
		voiceRepo:           voiceRepo,
		episodeRepo:         episodeRepo,
		scriptLineRepo:      scriptLineRepo,
		bgmRepo:             bgmRepo,
		systemBgmRepo:       systemBgmRepo,
		playbackHistoryRepo: playbackHistoryRepo,
		storageClient:       storageClient,
	}
}

// GetChannel は指定されたチャンネルを取得する
// オーナーまたは公開中のチャンネルのみ取得可能
func (s *channelService) GetChannel(ctx context.Context, userID, channelID string) (*response.ChannelDataResponse, error) {
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

	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	isOwner := userID != "" && channel.UserID == uid
	isPublished := channel.PublishedAt != nil && !channel.PublishedAt.After(time.Now())

	// オーナーでなく、かつ公開されていない場合は 404
	if !isOwner && !isPublished {
		return nil, apperror.ErrNotFound.WithMessage("チャンネルが見つかりません")
	}

	resp, err := s.toChannelResponse(ctx, channel, false, uid)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: resp,
	}, nil
}

// GetMyChannel は自分のチャンネルを取得する（オーナーのみ取得可能）
func (s *channelService) GetMyChannel(ctx context.Context, userID, channelID string) (*response.ChannelDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}

	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	// オーナーでない場合は 403
	if channel.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルへのアクセス権限がありません")
	}

	resp, err := s.toChannelResponse(ctx, channel, true, uid)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: resp,
	}, nil
}

// ListMyChannels は自分のチャンネル一覧を取得する
func (s *channelService) ListMyChannels(ctx context.Context, userID string, filter repository.ChannelFilter) (*response.ChannelListWithPaginationResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	channels, total, err := s.channelRepo.FindByUserID(ctx, uid, filter)
	if err != nil {
		return nil, err
	}

	responses, err := s.toChannelResponses(ctx, channels, uid)
	if err != nil {
		return nil, err
	}

	return &response.ChannelListWithPaginationResponse{
		Data:       responses,
		Pagination: response.PaginationResponse{Total: total, Limit: filter.Limit, Offset: filter.Offset},
	}, nil
}

// CreateChannel は新しいチャンネルを作成する
func (s *channelService) CreateChannel(ctx context.Context, userID string, req request.CreateChannelRequest) (*response.ChannelDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, err
	}

	// カテゴリの存在確認
	if _, err := s.categoryRepo.FindByID(ctx, categoryID); err != nil {
		return nil, err
	}

	// Artwork 画像の存在確認（指定時のみ）
	var artworkID *uuid.UUID
	if req.ArtworkImageID != nil {
		aid, err := uuid.Parse(*req.ArtworkImageID)
		if err != nil {
			return nil, err
		}
		if _, err := s.imageRepo.FindByID(ctx, aid); err != nil {
			return nil, err
		}
		artworkID = &aid
	}

	// キャラクター数のバリデーション
	if req.Characters.Total() < MinCharactersPerChannel || req.Characters.Total() > MaxCharactersPerChannel {
		return nil, apperror.ErrValidation.WithMessage(fmt.Sprintf("キャラクターは %d〜%d 人で設定してください", MinCharactersPerChannel, MaxCharactersPerChannel))
	}

	var created *model.Channel

	// トランザクションでキャラクター作成・チャンネル作成・紐づけを実行
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// トランザクション内で使うリポジトリを作成
		txCharacterRepo := repository.NewCharacterRepository(tx)
		txChannelRepo := repository.NewChannelRepository(tx)

		// キャラクターの処理（既存 or 新規作成）
		characterIDs, err := s.processCharacterInputs(ctx, uid, req.Characters, txCharacterRepo)
		if err != nil {
			return err
		}

		// チャンネルモデルを作成
		channel := &model.Channel{
			UserID:      uid,
			Name:        req.Name,
			Description: req.Description,
			CategoryID:  categoryID,
			ArtworkID:   artworkID,
		}

		// チャンネルを保存
		if err := txChannelRepo.Create(ctx, channel); err != nil {
			return err
		}

		// キャラクターを紐づけ
		if err := txChannelRepo.ReplaceChannelCharacters(ctx, channel.ID, characterIDs); err != nil {
			return err
		}

		// リレーションをプリロードして取得
		created, err = txChannelRepo.FindByID(ctx, channel.ID)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	resp, err := s.toChannelResponse(ctx, created, true, uuid.Nil)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: resp,
	}, nil
}

// UpdateChannel は指定されたチャンネルを更新する（オーナーのみ更新可能）
func (s *channelService) UpdateChannel(ctx context.Context, userID, channelID string, req request.UpdateChannelRequest) (*response.ChannelDataResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルの更新権限がありません")
	}

	// 各フィールドを更新
	channel.Name = req.Name
	channel.Description = req.Description

	// カテゴリの更新
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, err
	}
	if _, err := s.categoryRepo.FindByID(ctx, categoryID); err != nil {
		return nil, err
	}
	channel.CategoryID = categoryID

	// アートワークの更新
	if req.ArtworkImageID.IsSet {
		if req.ArtworkImageID.Value == nil {
			// null の場合は削除
			channel.ArtworkID = nil
		} else {
			artworkID, err := uuid.Parse(*req.ArtworkImageID.Value)
			if err != nil {
				return nil, apperror.ErrValidation.WithMessage("artworkImageId は有効な UUID である必要があります")
			}
			if _, err := s.imageRepo.FindByID(ctx, artworkID); err != nil {
				return nil, err
			}
			channel.ArtworkID = &artworkID
		}
		channel.Artwork = nil
	}

	// チャンネルを更新
	if err := s.channelRepo.Update(ctx, channel); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.channelRepo.FindByID(ctx, channel.ID)
	if err != nil {
		return nil, err
	}

	resp, err := s.toChannelResponse(ctx, updated, true, uuid.Nil)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: resp,
	}, nil
}

// DeleteChannel は指定されたチャンネルを削除する（オーナーのみ削除可能）
func (s *channelService) DeleteChannel(ctx context.Context, userID, channelID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return err
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return err
	}

	if channel.UserID != uid {
		return apperror.ErrForbidden.WithMessage("このチャンネルの削除権限がありません")
	}

	// 削除前に GCS ファイルのパスを収集
	var filesToDelete []string

	// チャンネルのアートワーク
	if channel.Artwork != nil {
		filesToDelete = append(filesToDelete, channel.Artwork.Path)
	}

	// 関連する全エピソードのファイルを収集
	episodes, _, err := s.episodeRepo.FindByChannelID(ctx, cid, repository.EpisodeFilter{Limit: 10000})
	if err != nil {
		return err
	}

	for _, episode := range episodes {
		if episode.Artwork != nil {
			filesToDelete = append(filesToDelete, episode.Artwork.Path)
		}
		if episode.FullAudio != nil {
			filesToDelete = append(filesToDelete, episode.FullAudio.Path)
		}
		if episode.Bgm != nil && episode.Bgm.Audio.ID != uuid.Nil {
			filesToDelete = append(filesToDelete, episode.Bgm.Audio.Path)
		}
	}

	// チャンネルを削除（カスケードでエピソード等も削除される）
	if err := s.channelRepo.Delete(ctx, cid); err != nil {
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

// PublishChannel は指定されたチャンネルを公開する
func (s *channelService) PublishChannel(ctx context.Context, userID, channelID string, publishedAt *string) (*response.ChannelDataResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルの公開権限がありません")
	}

	// 公開日時を設定
	if publishedAt == nil || *publishedAt == "" {
		// 省略時は現在時刻で即時公開
		now := time.Now()
		channel.PublishedAt = &now
	} else {
		// 指定された日時でパース
		parsedTime, err := time.Parse(time.RFC3339, *publishedAt)
		if err != nil {
			return nil, apperror.ErrValidation.WithMessage("公開日時の形式が無効です。RFC3339 形式で指定してください")
		}
		channel.PublishedAt = &parsedTime
	}

	// チャンネルを更新
	if err := s.channelRepo.Update(ctx, channel); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.channelRepo.FindByID(ctx, channel.ID)
	if err != nil {
		return nil, err
	}

	resp, err := s.toChannelResponse(ctx, updated, true, uuid.Nil)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: resp,
	}, nil
}

// UnpublishChannel は指定されたチャンネルを非公開にする
func (s *channelService) UnpublishChannel(ctx context.Context, userID, channelID string) (*response.ChannelDataResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルの非公開権限がありません")
	}

	// 公開日時を null に設定（非公開化）
	channel.PublishedAt = nil

	// チャンネルを更新
	if err := s.channelRepo.Update(ctx, channel); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.channelRepo.FindByID(ctx, channel.ID)
	if err != nil {
		return nil, err
	}

	resp, err := s.toChannelResponse(ctx, updated, true, uuid.Nil)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: resp,
	}, nil
}

// SetUserPrompt は指定されたチャンネルに台本プロンプトを設定する
func (s *channelService) SetUserPrompt(ctx context.Context, userID, channelID string, req request.SetUserPromptRequest) (*response.ChannelDataResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルの台本プロンプト設定権限がありません")
	}

	// 台本プロンプトを設定
	channel.UserPrompt = req.UserPrompt

	// チャンネルを更新
	if err := s.channelRepo.Update(ctx, channel); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.channelRepo.FindByID(ctx, channel.ID)
	if err != nil {
		return nil, err
	}

	resp, err := s.toChannelResponse(ctx, updated, true, uuid.Nil)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: resp,
	}, nil
}

// SetDefaultBgm は指定されたチャンネルにデフォルト BGM を設定する
func (s *channelService) SetDefaultBgm(ctx context.Context, userID, channelID string, req request.SetDefaultBgmRequest) (*response.ChannelDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(channelID)
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
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルのデフォルト BGM 設定権限がありません")
	}

	// 前の BGM 設定をクリア
	channel.DefaultBgmID = nil
	channel.DefaultSystemBgmID = nil
	channel.DefaultBgm = nil
	channel.DefaultSystemBgm = nil

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

		channel.DefaultBgmID = &bgmID
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

		channel.DefaultSystemBgmID = &systemBgmID
	}

	// チャンネルを更新
	if err := s.channelRepo.Update(ctx, channel); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.channelRepo.FindByID(ctx, channel.ID)
	if err != nil {
		return nil, err
	}

	resp, err := s.toChannelResponse(ctx, updated, true, uuid.Nil)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: resp,
	}, nil
}

// DeleteDefaultBgm は指定されたチャンネルのデフォルト BGM を削除する
func (s *channelService) DeleteDefaultBgm(ctx context.Context, userID, channelID string) (*response.ChannelDataResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルのデフォルト BGM 削除権限がありません")
	}

	// デフォルト BGM をクリア
	channel.DefaultBgmID = nil
	channel.DefaultSystemBgmID = nil
	channel.DefaultBgm = nil
	channel.DefaultSystemBgm = nil

	// チャンネルを更新
	if err := s.channelRepo.Update(ctx, channel); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.channelRepo.FindByID(ctx, channel.ID)
	if err != nil {
		return nil, err
	}

	resp, err := s.toChannelResponse(ctx, updated, true, uuid.Nil)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: resp,
	}, nil
}

// AddChannelCharacter はチャンネルにキャラクターを1人追加する
func (s *channelService) AddChannelCharacter(ctx context.Context, userID, channelID string, req request.AddChannelCharacterRequest) (*response.ChannelDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}

	// connect / create のどちらか一方のみ指定可能
	if (req.Connect == nil) == (req.Create == nil) {
		return nil, apperror.ErrValidation.WithMessage("connect または create のいずれか一方を指定してください")
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	if channel.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルのキャラクター変更権限がありません")
	}

	// キャラクター数の上限チェック
	count, err := s.channelRepo.CountChannelCharacters(ctx, cid)
	if err != nil {
		return nil, err
	}

	if count >= MaxCharactersPerChannel {
		return nil, apperror.ErrValidation.WithMessage(fmt.Sprintf("キャラクターは最大 %d 人までです", MaxCharactersPerChannel))
	}

	// キャラクター ID を解決
	characterID, newProvider, err := s.resolveCharacterID(ctx, uid, req.Connect, req.Create)
	if err != nil {
		return nil, err
	}

	// プロバイダーの整合性チェック
	existingProvider := getChannelProvider(channel.ChannelCharacters, nil)
	if existingProvider != "" && existingProvider != newProvider {
		return nil, apperror.ErrValidation.WithMessage("同一チャンネルのキャラクターは同じボイスプロバイダーを使用してください")
	}

	// 既にチャンネルに紐づいていないことを確認
	exists, err := s.channelRepo.HasChannelCharacter(ctx, cid, characterID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.ErrValidation.WithMessage("このキャラクターは既にチャンネルに追加されています")
	}

	// キャラクターを追加
	if err := s.channelRepo.AddChannelCharacter(ctx, cid, characterID); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	resp, err := s.toChannelResponse(ctx, updated, true, uuid.Nil)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: resp,
	}, nil
}

// ReplaceChannelCharacter はチャンネル内の既存キャラクターを別のキャラクターに差し替える
func (s *channelService) ReplaceChannelCharacter(ctx context.Context, userID, channelID, characterID string, req request.ReplaceChannelCharacterRequest) (*response.ChannelDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}

	oldCharacterID, err := uuid.Parse(characterID)
	if err != nil {
		return nil, err
	}

	// connect / create のどちらか一方のみ指定可能
	if (req.Connect == nil) == (req.Create == nil) {
		return nil, apperror.ErrValidation.WithMessage("connect または create のいずれか一方を指定してください")
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	if channel.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルのキャラクター変更権限がありません")
	}

	// 置換元キャラクターがチャンネルに紐づいていることを確認
	exists, err := s.channelRepo.HasChannelCharacter(ctx, cid, oldCharacterID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, apperror.ErrNotFound.WithMessage("置換元のキャラクターがチャンネルに紐づいていません")
	}

	// 新しいキャラクター ID を解決
	newCharacterID, newProvider, err := s.resolveCharacterID(ctx, uid, req.Connect, req.Create)
	if err != nil {
		return nil, err
	}

	// プロバイダーの整合性チェック（置換元を除外して残りのキャラクターで判定）
	remainingProvider := getChannelProvider(channel.ChannelCharacters, &oldCharacterID)
	if remainingProvider != "" && remainingProvider != newProvider {
		return nil, apperror.ErrValidation.WithMessage("同一チャンネルのキャラクターは同じボイスプロバイダーを使用してください")
	}

	// 置換先が既にチャンネルに紐づいていないことを確認
	exists, err = s.channelRepo.HasChannelCharacter(ctx, cid, newCharacterID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.ErrValidation.WithMessage("置換先のキャラクターは既にチャンネルに追加されています")
	}

	// キャラクターを置換
	if err := s.channelRepo.ReplaceChannelCharacter(ctx, cid, oldCharacterID, newCharacterID); err != nil {
		return nil, err
	}

	// 既存台本の speaker_id を新キャラクターに一括更新
	if err := s.scriptLineRepo.UpdateSpeakerIDByChannelID(ctx, cid, oldCharacterID, newCharacterID); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	resp, err := s.toChannelResponse(ctx, updated, true, uuid.Nil)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: resp,
	}, nil
}

// RemoveChannelCharacter はチャンネルからキャラクターの紐づけを解除する
func (s *channelService) RemoveChannelCharacter(ctx context.Context, userID, channelID, characterID string) (*response.ChannelDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}

	targetCharacterID, err := uuid.Parse(characterID)
	if err != nil {
		return nil, err
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	if channel.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルのキャラクター変更権限がありません")
	}

	// キャラクター数の下限チェック
	count, err := s.channelRepo.CountChannelCharacters(ctx, cid)
	if err != nil {
		return nil, err
	}

	if count <= MinCharactersPerChannel {
		return nil, apperror.ErrValidation.WithMessage(fmt.Sprintf("キャラクターは最低 %d 人必要です", MinCharactersPerChannel))
	}

	// 台本で使用中のキャラクターは削除不可
	used, err := s.scriptLineRepo.ExistsBySpeakerIDAndChannelID(ctx, targetCharacterID, cid)
	if err != nil {
		return nil, err
	}
	if used {
		return nil, apperror.ErrValidation.WithMessage("このキャラクターは台本で使用されているため削除できません")
	}

	// キャラクターを削除
	if err := s.channelRepo.RemoveChannelCharacter(ctx, cid, targetCharacterID); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	resp, err := s.toChannelResponse(ctx, updated, true, uuid.Nil)
	if err != nil {
		return nil, err
	}

	return &response.ChannelDataResponse{
		Data: resp,
	}, nil
}

// resolveCharacterID は connect/create の入力を処理して characterID とボイスプロバイダーを返す
func (s *channelService) resolveCharacterID(ctx context.Context, userID uuid.UUID, connect *request.ConnectCharacterInput, create *request.CreateCharacterInput) (characterID uuid.UUID, voiceProvider string, err error) {
	if connect != nil {
		cid, err := uuid.Parse(connect.ID)
		if err != nil {
			return uuid.Nil, "", err
		}

		// キャラクターの存在確認とオーナーチェック
		character, err := s.characterRepo.FindByID(ctx, cid)
		if err != nil {
			return uuid.Nil, "", err
		}
		if character.UserID != userID {
			return uuid.Nil, "", apperror.ErrForbidden.WithMessage("指定されたキャラクターの所有権がありません")
		}

		return cid, character.Voice.Provider, nil
	}

	// create の場合
	// 同一ユーザー内での名前重複チェック
	exists, err := s.characterRepo.ExistsByUserIDAndName(ctx, userID, create.Name, nil)
	if err != nil {
		return uuid.Nil, "", err
	}
	if exists {
		return uuid.Nil, "", apperror.ErrDuplicateName.WithMessage("同じ名前のキャラクターが既に存在します")
	}

	voiceID, err := uuid.Parse(create.VoiceID)
	if err != nil {
		return uuid.Nil, "", err
	}

	// ボイスの存在確認（アクティブなもののみ）
	voice, err := s.voiceRepo.FindActiveByID(ctx, create.VoiceID)
	if err != nil {
		return uuid.Nil, "", err
	}

	// アバター画像の存在確認（指定時のみ）
	var avatarID *uuid.UUID
	if create.AvatarID != nil {
		aid, err := uuid.Parse(*create.AvatarID)
		if err != nil {
			return uuid.Nil, "", err
		}
		if _, err := s.imageRepo.FindByID(ctx, aid); err != nil {
			return uuid.Nil, "", err
		}
		avatarID = &aid
	}

	character := &model.Character{
		UserID:   userID,
		Name:     create.Name,
		Persona:  create.Persona,
		AvatarID: avatarID,
		VoiceID:  voiceID,
	}

	if err := s.characterRepo.Create(ctx, character); err != nil {
		return uuid.Nil, "", err
	}

	return character.ID, voice.Provider, nil
}

// getChannelProvider はチャンネルに紐づくキャラクターのボイスプロバイダーを返す
// excludeCharacterID が指定された場合、そのキャラクターを除外して判定する（Replace用）
// キャラクターが0人の場合は空文字を返す
func getChannelProvider(channelCharacters []model.ChannelCharacter, excludeCharacterID *uuid.UUID) string {
	for _, cc := range channelCharacters {
		if excludeCharacterID != nil && cc.CharacterID == *excludeCharacterID {
			continue
		}
		return cc.Character.Voice.Provider
	}
	return ""
}

// processCharacterInputs はキャラクター入力を処理してキャラクター ID のスライスを返す
// connect は既存キャラクターの紐づけ、create は新規キャラクターの作成
// 全キャラクターのボイスプロバイダーが統一されていることも検証する
func (s *channelService) processCharacterInputs(ctx context.Context, userID uuid.UUID, input request.ChannelCharactersInput, characterRepo repository.CharacterRepository) ([]uuid.UUID, error) {
	characterIDs := make([]uuid.UUID, 0, input.Total())
	providers := make([]string, 0, input.Total())

	// 既存キャラクターの紐づけ処理
	for _, connect := range input.Connect {
		cid, err := uuid.Parse(connect.ID)
		if err != nil {
			return nil, err
		}

		// キャラクターの存在確認とオーナーチェック
		character, err := characterRepo.FindByID(ctx, cid)
		if err != nil {
			return nil, err
		}
		if character.UserID != userID {
			return nil, apperror.ErrForbidden.WithMessage("指定されたキャラクターの所有権がありません")
		}

		characterIDs = append(characterIDs, cid)
		providers = append(providers, character.Voice.Provider)
	}

	// 新規キャラクターの作成処理
	for _, create := range input.Create {
		// 同一ユーザー内での名前重複チェック
		exists, err := characterRepo.ExistsByUserIDAndName(ctx, userID, create.Name, nil)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperror.ErrDuplicateName.WithMessage("同じ名前のキャラクターが既に存在します")
		}

		voiceID, err := uuid.Parse(create.VoiceID)
		if err != nil {
			return nil, err
		}

		// ボイスの存在確認（アクティブなもののみ）
		voice, err := s.voiceRepo.FindActiveByID(ctx, create.VoiceID)
		if err != nil {
			return nil, err
		}

		// アバター画像の存在確認（指定時のみ）
		var avatarID *uuid.UUID
		if create.AvatarID != nil {
			aid, err := uuid.Parse(*create.AvatarID)
			if err != nil {
				return nil, err
			}
			if _, err := s.imageRepo.FindByID(ctx, aid); err != nil {
				return nil, err
			}
			avatarID = &aid
		}

		character := &model.Character{
			UserID:   userID,
			Name:     create.Name,
			Persona:  create.Persona,
			AvatarID: avatarID,
			VoiceID:  voiceID,
		}

		if err := characterRepo.Create(ctx, character); err != nil {
			return nil, err
		}

		characterIDs = append(characterIDs, character.ID)
		providers = append(providers, voice.Provider)
	}

	// プロバイダーの整合性チェック
	if len(providers) > 1 {
		first := providers[0]
		for _, p := range providers[1:] {
			if p != first {
				return nil, apperror.ErrValidation.WithMessage("同一チャンネルのキャラクターは同じボイスプロバイダーを使用してください")
			}
		}
	}

	return characterIDs, nil
}

// toChannelOwnerResponse は User からチャンネルオーナーレスポンスを生成する
func (s *channelService) toChannelOwnerResponse(ctx context.Context, user *model.User) (response.ChannelOwnerResponse, error) {
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

// toChannelResponses は Channel のスライスをレスポンス DTO のスライスに変換する
// ListMyChannels で使用するため、常にオーナーとして扱う
func (s *channelService) toChannelResponses(ctx context.Context, channels []model.Channel, userID uuid.UUID) ([]response.ChannelResponse, error) {
	result := make([]response.ChannelResponse, len(channels))

	for i, c := range channels {
		resp, err := s.toChannelResponse(ctx, &c, true, userID)
		if err != nil {
			return nil, err
		}
		result[i] = resp
	}

	return result, nil
}

// toChannelResponse は Channel をレスポンス DTO に変換する
// isOwner が false の場合、userPrompt は空文字になる
func (s *channelService) toChannelResponse(ctx context.Context, c *model.Channel, isOwner bool, userID uuid.UUID) (response.ChannelResponse, error) {
	userPrompt := ""
	if isOwner {
		userPrompt = c.UserPrompt
	}

	// オーナー情報を生成
	ownerResp, err := s.toChannelOwnerResponse(ctx, &c.User)
	if err != nil {
		return response.ChannelResponse{}, err
	}

	// エピソード一覧を取得（公開ページではオーナーでない場合、公開済みのみ）
	episodeFilter := repository.EpisodeFilter{Limit: 10000}
	if !isOwner {
		published := "published"
		episodeFilter.Status = &published
	}
	episodes, _, err := s.episodeRepo.FindByChannelID(ctx, c.ID, episodeFilter)
	if err != nil {
		return response.ChannelResponse{}, err
	}

	episodeResponses, err := s.toEpisodeResponses(ctx, episodes, &c.User, userID)
	if err != nil {
		return response.ChannelResponse{}, err
	}

	resp := response.ChannelResponse{
		ID:          c.ID,
		Owner:       ownerResp,
		Name:        c.Name,
		Description: c.Description,
		UserPrompt:  userPrompt,
		Category: response.CategoryResponse{
			ID:        c.Category.ID,
			Slug:      c.Category.Slug,
			Name:      c.Category.Name,
			SortOrder: c.Category.SortOrder,
			IsActive:  c.Category.IsActive,
		},
		Characters:  s.toCharacterResponsesFromChannelCharacters(ctx, c.ChannelCharacters),
		Episodes:    episodeResponses,
		PublishedAt: c.PublishedAt,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}

	if c.Artwork != nil {
		var artworkURL string
		if storage.IsExternalURL(c.Artwork.Path) {
			artworkURL = c.Artwork.Path
		} else {
			var err error
			artworkURL, err = s.storageClient.GenerateSignedURL(ctx, c.Artwork.Path, storage.SignedURLExpirationImage)
			if err != nil {
				return response.ChannelResponse{}, err
			}
		}
		resp.Artwork = &response.ArtworkResponse{
			ID:  c.Artwork.ID,
			URL: artworkURL,
		}
	}

	// デフォルト BGM のレスポンス生成
	if c.DefaultBgm != nil && c.DefaultBgm.Audio.ID != uuid.Nil {
		signedURL, err := s.storageClient.GenerateSignedURL(ctx, c.DefaultBgm.Audio.Path, storage.SignedURLExpirationAudio)
		if err != nil {
			return response.ChannelResponse{}, err
		}
		resp.DefaultBgm = &response.ChannelDefaultBgmResponse{
			ID:       c.DefaultBgm.ID,
			Name:     c.DefaultBgm.Name,
			IsSystem: false,
			Audio: response.BgmAudioResponse{
				ID:         c.DefaultBgm.Audio.ID,
				URL:        signedURL,
				DurationMs: c.DefaultBgm.Audio.DurationMs,
			},
		}
	} else if c.DefaultSystemBgm != nil && c.DefaultSystemBgm.Audio.ID != uuid.Nil {
		signedURL, err := s.storageClient.GenerateSignedURL(ctx, c.DefaultSystemBgm.Audio.Path, storage.SignedURLExpirationAudio)
		if err != nil {
			return response.ChannelResponse{}, err
		}
		resp.DefaultBgm = &response.ChannelDefaultBgmResponse{
			ID:       c.DefaultSystemBgm.ID,
			Name:     c.DefaultSystemBgm.Name,
			IsSystem: true,
			Audio: response.BgmAudioResponse{
				ID:         c.DefaultSystemBgm.Audio.ID,
				URL:        signedURL,
				DurationMs: c.DefaultSystemBgm.Audio.DurationMs,
			},
		}
	}

	return resp, nil
}

// toCharacterResponsesFromChannelCharacters は ChannelCharacter のスライスをレスポンス DTO のスライスに変換する
func (s *channelService) toCharacterResponsesFromChannelCharacters(ctx context.Context, channelCharacters []model.ChannelCharacter) []response.CharacterResponse {
	result := make([]response.CharacterResponse, len(channelCharacters))

	for i, cc := range channelCharacters {
		var avatar *response.AvatarResponse
		if cc.Character.Avatar != nil && s.storageClient != nil {
			if storage.IsExternalURL(cc.Character.Avatar.Path) {
				avatar = &response.AvatarResponse{
					ID:  cc.Character.Avatar.ID,
					URL: cc.Character.Avatar.Path,
				}
			} else {
				signedURL, err := s.storageClient.GenerateSignedURL(ctx, cc.Character.Avatar.Path, storage.SignedURLExpirationImage)
				if err == nil {
					avatar = &response.AvatarResponse{
						ID:  cc.Character.Avatar.ID,
						URL: signedURL,
					}
				}
			}
		}

		result[i] = response.CharacterResponse{
			ID:      cc.Character.ID,
			Name:    cc.Character.Name,
			Persona: cc.Character.Persona,
			Avatar:  avatar,
			Voice: response.CharacterVoiceResponse{
				ID:       cc.Character.Voice.ID,
				Name:     cc.Character.Voice.Name,
				Provider: cc.Character.Voice.Provider,
				Gender:   string(cc.Character.Voice.Gender),
			},
			CreatedAt: cc.Character.CreatedAt,
			UpdatedAt: cc.Character.UpdatedAt,
		}
	}

	return result
}

// toEpisodeResponses は Episode のスライスをレスポンス DTO のスライスに変換する
func (s *channelService) toEpisodeResponses(ctx context.Context, episodes []model.Episode, owner *model.User, userID uuid.UUID) ([]response.EpisodeResponse, error) {
	// 認証済みの場合は再生履歴を一括取得して map に変換
	playbackMap := make(map[uuid.UUID]*model.PlaybackHistory)
	if userID != uuid.Nil && len(episodes) > 0 {
		episodeIDs := make([]uuid.UUID, len(episodes))
		for i, e := range episodes {
			episodeIDs[i] = e.ID
		}
		histories, err := s.playbackHistoryRepo.FindByUserIDAndEpisodeIDs(ctx, userID, episodeIDs)
		if err == nil {
			for i := range histories {
				playbackMap[histories[i].EpisodeID] = &histories[i]
			}
		}
	}

	result := make([]response.EpisodeResponse, len(episodes))
	for i, e := range episodes {
		resp, err := s.toEpisodeResponse(ctx, &e, owner, playbackMap[e.ID])
		if err != nil {
			return nil, err
		}
		result[i] = resp
	}

	return result, nil
}

// toEpisodeResponse は Episode をレスポンス DTO に変換する
func (s *channelService) toEpisodeResponse(ctx context.Context, e *model.Episode, owner *model.User, playback *model.PlaybackHistory) (response.EpisodeResponse, error) {
	ownerResp, err := s.toChannelOwnerResponse(ctx, owner)
	if err != nil {
		return response.EpisodeResponse{}, err
	}

	resp := response.EpisodeResponse{
		ID:          e.ID,
		Owner:       ownerResp,
		Title:       e.Title,
		Description: e.Description,
		PlayCount:   e.PlayCount,
		PublishedAt: e.PublishedAt,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
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
