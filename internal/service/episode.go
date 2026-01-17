package service

import (
	"context"
	"time"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/infrastructure/tts"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/audio"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// エピソード関連のビジネスロジックインターフェース
type EpisodeService interface {
	GetMyChannelEpisode(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error)
	ListMyChannelEpisodes(ctx context.Context, userID, channelID string, filter repository.EpisodeFilter) (*response.EpisodeListWithPaginationResponse, error)
	CreateEpisode(ctx context.Context, userID, channelID, title, description string, artworkImageID *string) (*response.EpisodeResponse, error)
	UpdateEpisode(ctx context.Context, userID, channelID, episodeID string, req request.UpdateEpisodeRequest) (*response.EpisodeDataResponse, error)
	DeleteEpisode(ctx context.Context, userID, channelID, episodeID string) error
	PublishEpisode(ctx context.Context, userID, channelID, episodeID string, publishedAt *string) (*response.EpisodeDataResponse, error)
	UnpublishEpisode(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error)
	SetEpisodeBgm(ctx context.Context, userID, channelID, episodeID, bgmAudioID string) (*response.EpisodeDataResponse, error)
	RemoveEpisodeBgm(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error)
	GenerateAudio(ctx context.Context, userID, channelID, episodeID string) (*response.GenerateAudioResponse, error)
}

type episodeService struct {
	episodeRepo    repository.EpisodeRepository
	channelRepo    repository.ChannelRepository
	scriptLineRepo repository.ScriptLineRepository
	audioRepo      repository.AudioRepository
	imageRepo      repository.ImageRepository
	storageClient  storage.Client
	ttsClient      tts.Client
}

// EpisodeService の実装を返す
func NewEpisodeService(
	episodeRepo repository.EpisodeRepository,
	channelRepo repository.ChannelRepository,
	scriptLineRepo repository.ScriptLineRepository,
	audioRepo repository.AudioRepository,
	imageRepo repository.ImageRepository,
	storageClient storage.Client,
	ttsClient tts.Client,
) EpisodeService {
	return &episodeService{
		episodeRepo:    episodeRepo,
		channelRepo:    channelRepo,
		scriptLineRepo: scriptLineRepo,
		audioRepo:      audioRepo,
		imageRepo:      imageRepo,
		storageClient:  storageClient,
		ttsClient:      ttsClient,
	}
}

// 自分のチャンネルのエピソードを取得する
func (s *episodeService) GetMyChannelEpisode(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return nil, err
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	if channel.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("You do not have permission to access this channel")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("Episode not found in this channel")
	}

	resp, err := s.toEpisodeResponse(ctx, episode)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeDataResponse{
		Data: resp,
	}, nil
}

// 自分のチャンネルのエピソード一覧を取得する
func (s *episodeService) ListMyChannelEpisodes(ctx context.Context, userID, channelID string, filter repository.EpisodeFilter) (*response.EpisodeListWithPaginationResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("You do not have permission to access this channel")
	}

	// エピソード一覧を取得
	episodes, total, err := s.episodeRepo.FindByChannelID(ctx, cid, filter)
	if err != nil {
		return nil, err
	}

	responses, err := s.toEpisodeResponses(ctx, episodes)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeListWithPaginationResponse{
		Data:       responses,
		Pagination: response.PaginationResponse{Total: total, Limit: filter.Limit, Offset: filter.Offset},
	}, nil
}

// エピソードを作成する
func (s *episodeService) CreateEpisode(ctx context.Context, userID, channelID, title, description string, artworkImageID *string) (*response.EpisodeResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("You do not have permission to access this channel")
	}

	// エピソードを作成
	episode := &model.Episode{
		ChannelID:   cid,
		Title:       title,
		Description: description,
	}

	// アートワークが指定されている場合
	if artworkImageID != nil {
		artworkID, err := uuid.Parse(*artworkImageID)
		if err != nil {
			return nil, err
		}
		// 画像の存在確認
		if _, err := s.imageRepo.FindByID(ctx, artworkID); err != nil {
			return nil, err
		}
		episode.ArtworkID = &artworkID
	}

	if err := s.episodeRepo.Create(ctx, episode); err != nil {
		return nil, err
	}

	return &response.EpisodeResponse{
		ID:          episode.ID,
		Title:       episode.Title,
		Description: episode.Description,
		UserPrompt:  episode.UserPrompt,
		PublishedAt: episode.PublishedAt,
		CreatedAt:   episode.CreatedAt,
		UpdatedAt:   episode.UpdatedAt,
	}, nil
}

// エピソードを更新する
func (s *episodeService) UpdateEpisode(ctx context.Context, userID, channelID, episodeID string, req request.UpdateEpisodeRequest) (*response.EpisodeDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return nil, err
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	if channel.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("You do not have permission to update this episode")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("Episode not found in this channel")
	}

	// 各フィールドを更新
	episode.Title = req.Title
	episode.Description = req.Description

	// アートワークの更新
	if req.ArtworkImageID != nil {
		if *req.ArtworkImageID == "" {
			// 空文字の場合は null に設定
			episode.ArtworkID = nil
		} else {
			artworkID, err := uuid.Parse(*req.ArtworkImageID)
			if err != nil {
				return nil, err
			}
			// 画像の存在確認
			if _, err := s.imageRepo.FindByID(ctx, artworkID); err != nil {
				return nil, err
			}
			episode.ArtworkID = &artworkID
		}
		episode.Artwork = nil
	}

	// エピソードを更新
	if err := s.episodeRepo.Update(ctx, episode); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.episodeRepo.FindByID(ctx, episode.ID)
	if err != nil {
		return nil, err
	}

	resp, err := s.toEpisodeResponse(ctx, updated)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeDataResponse{
		Data: resp,
	}, nil
}

// エピソードを削除する
func (s *episodeService) DeleteEpisode(ctx context.Context, userID, channelID, episodeID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return err
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return err
	}

	if channel.UserID != uid {
		return apperror.ErrForbidden.WithMessage("You do not have permission to delete this episode")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return err
	}

	if episode.ChannelID != cid {
		return apperror.ErrNotFound.WithMessage("Episode not found in this channel")
	}

	// エピソードを削除
	return s.episodeRepo.Delete(ctx, eid)
}

// エピソードを公開する
func (s *episodeService) PublishEpisode(ctx context.Context, userID, channelID, episodeID string, publishedAt *string) (*response.EpisodeDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return nil, err
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	if channel.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("You do not have permission to publish this episode")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("Episode not found in this channel")
	}

	// 公開日時を設定
	if publishedAt == nil || *publishedAt == "" {
		// 省略時は現在時刻で即時公開
		now := time.Now()
		episode.PublishedAt = &now
	} else {
		// 指定された日時でパース
		parsedTime, err := time.Parse(time.RFC3339, *publishedAt)
		if err != nil {
			return nil, apperror.ErrValidation.WithMessage("Invalid publishedAt format. Use RFC3339 format.")
		}
		episode.PublishedAt = &parsedTime
	}

	// エピソードを更新
	if err := s.episodeRepo.Update(ctx, episode); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.episodeRepo.FindByID(ctx, episode.ID)
	if err != nil {
		return nil, err
	}

	resp, err := s.toEpisodeResponse(ctx, updated)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeDataResponse{
		Data: resp,
	}, nil
}

// エピソードを非公開にする
func (s *episodeService) UnpublishEpisode(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return nil, err
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	if channel.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("You do not have permission to unpublish this episode")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("Episode not found in this channel")
	}

	// 公開日時を null に設定（非公開化）
	episode.PublishedAt = nil

	// エピソードを更新
	if err := s.episodeRepo.Update(ctx, episode); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.episodeRepo.FindByID(ctx, episode.ID)
	if err != nil {
		return nil, err
	}

	resp, err := s.toEpisodeResponse(ctx, updated)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeDataResponse{
		Data: resp,
	}, nil
}

// Episode モデルのスライスをレスポンス DTO のスライスに変換する
func (s *episodeService) toEpisodeResponses(ctx context.Context, episodes []model.Episode) ([]response.EpisodeResponse, error) {
	result := make([]response.EpisodeResponse, len(episodes))

	for i, e := range episodes {
		resp, err := s.toEpisodeResponse(ctx, &e)
		if err != nil {
			return nil, err
		}
		result[i] = resp
	}

	return result, nil
}

// Episode モデルをレスポンス DTO に変換する
func (s *episodeService) toEpisodeResponse(ctx context.Context, e *model.Episode) (response.EpisodeResponse, error) {
	resp := response.EpisodeResponse{
		ID:          e.ID,
		Title:       e.Title,
		Description: e.Description,
		UserPrompt:  e.UserPrompt,
		PublishedAt: e.PublishedAt,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}

	if e.Artwork != nil {
		signedURL, err := s.storageClient.GenerateSignedURL(ctx, e.Artwork.Path, storage.SignedURLExpirationImage)
		if err != nil {
			return response.EpisodeResponse{}, err
		}
		resp.Artwork = &response.ArtworkResponse{
			ID:  e.Artwork.ID,
			URL: signedURL,
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

	if e.Bgm != nil {
		signedURL, err := s.storageClient.GenerateSignedURL(ctx, e.Bgm.Path, storage.SignedURLExpirationAudio)
		if err != nil {
			return response.EpisodeResponse{}, err
		}
		resp.Bgm = &response.AudioResponse{
			ID:         e.Bgm.ID,
			URL:        signedURL,
			MimeType:   e.Bgm.MimeType,
			FileSize:   e.Bgm.FileSize,
			DurationMs: e.Bgm.DurationMs,
		}
	}

	return resp, nil
}

// エピソードに BGM を設定する
func (s *episodeService) SetEpisodeBgm(ctx context.Context, userID, channelID, episodeID, bgmAudioID string) (*response.EpisodeDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return nil, err
	}

	bgmID, err := uuid.Parse(bgmAudioID)
	if err != nil {
		return nil, err
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	if channel.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("You do not have permission to update this episode")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("Episode not found in this channel")
	}

	// BGM 音声の存在確認
	if _, err := s.audioRepo.FindByID(ctx, bgmID); err != nil {
		return nil, err
	}

	// BGM を設定
	episode.BgmID = &bgmID
	episode.Bgm = nil

	// エピソードを更新
	if err := s.episodeRepo.Update(ctx, episode); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.episodeRepo.FindByID(ctx, episode.ID)
	if err != nil {
		return nil, err
	}

	resp, err := s.toEpisodeResponse(ctx, updated)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeDataResponse{
		Data: resp,
	}, nil
}

// エピソードの BGM を削除する
func (s *episodeService) RemoveEpisodeBgm(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return nil, err
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	if channel.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("You do not have permission to update this episode")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("Episode not found in this channel")
	}

	// BGM を削除
	episode.BgmID = nil
	episode.Bgm = nil

	// エピソードを更新
	if err := s.episodeRepo.Update(ctx, episode); err != nil {
		return nil, err
	}

	// リレーションをプリロードして取得
	updated, err := s.episodeRepo.FindByID(ctx, episode.ID)
	if err != nil {
		return nil, err
	}

	resp, err := s.toEpisodeResponse(ctx, updated)
	if err != nil {
		return nil, err
	}

	return &response.EpisodeDataResponse{
		Data: resp,
	}, nil
}

// エピソードの音声を生成する
func (s *episodeService) GenerateAudio(ctx context.Context, userID, channelID, episodeID string) (*response.GenerateAudioResponse, error) {
	log := logger.FromContext(ctx)

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return nil, err
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	if channel.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("You do not have permission to generate audio for this episode")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("Episode not found in this channel")
	}

	// 台本行を取得（Voice 情報を含む）
	scriptLines, err := s.scriptLineRepo.FindByEpisodeIDWithVoice(ctx, eid)
	if err != nil {
		return nil, err
	}

	if len(scriptLines) == 0 {
		return nil, apperror.ErrValidation.WithMessage("No script lines found for this episode")
	}

	// speech 行から turns と voiceConfigs を構築
	var turns []tts.SpeakerTurn
	voiceConfigMap := make(map[string]string) // speakerName -> voiceID

	for _, line := range scriptLines {
		if line.LineType != model.LineTypeSpeech {
			continue
		}

		if line.Text == nil || *line.Text == "" {
			continue
		}

		if line.Speaker == nil {
			log.Warn("speech line has no speaker", "line_id", line.ID)
			continue
		}

		// Turn を追加
		turns = append(turns, tts.SpeakerTurn{
			Speaker: line.Speaker.Name,
			Text:    *line.Text,
		})

		// VoiceConfig を収集（重複しないように）
		if _, exists := voiceConfigMap[line.Speaker.Name]; !exists {
			voiceConfigMap[line.Speaker.Name] = line.Speaker.Voice.ProviderVoiceID
		}
	}

	if len(turns) == 0 {
		return nil, apperror.ErrValidation.WithMessage("No speech lines found for audio generation")
	}

	// voiceConfigs を構築
	voiceConfigs := make([]tts.SpeakerVoiceConfig, 0, len(voiceConfigMap))
	for speakerName, voiceID := range voiceConfigMap {
		voiceConfigs = append(voiceConfigs, tts.SpeakerVoiceConfig{
			SpeakerAlias: speakerName,
			VoiceID:      voiceID,
		})
	}

	// Multi-speaker TTS で音声を生成
	combinedAudio, err := s.ttsClient.SynthesizeMultiSpeaker(ctx, turns, voiceConfigs)
	if err != nil {
		log.Error("failed to synthesize multi-speaker audio", "error", err)
		return nil, apperror.ErrGenerationFailed.WithMessage("Failed to generate audio").WithError(err)
	}

	// 新しい Audio ID を生成してパスを作成
	audioID := uuid.New()
	audioPath := storage.GenerateAudioPath(audioID.String())

	// GCS にアップロード
	if _, err := s.storageClient.Upload(ctx, combinedAudio, audioPath, "audio/mpeg"); err != nil {
		log.Error("failed to upload audio", "error", err)
		return nil, apperror.ErrInternal.WithMessage("Failed to upload audio").WithError(err)
	}

	// Audio レコードを作成
	audioRecord := &model.Audio{
		ID:         audioID,
		MimeType:   "audio/mpeg",
		Path:       audioPath,
		Filename:   audioID.String() + ".mp3",
		FileSize:   len(combinedAudio),
		DurationMs: audio.GetMP3DurationMs(combinedAudio),
	}

	if err := s.audioRepo.Create(ctx, audioRecord); err != nil {
		log.Error("failed to create audio record", "error", err)
		return nil, apperror.ErrInternal.WithMessage("Failed to save audio record").WithError(err)
	}

	// エピソードの FullAudioID を更新
	episode.FullAudioID = &audioID
	episode.FullAudio = nil

	if err := s.episodeRepo.Update(ctx, episode); err != nil {
		return nil, err
	}

	// 署名付き URL を生成
	signedURL, err := s.storageClient.GenerateSignedURL(ctx, audioPath, storage.SignedURLExpirationAudio)
	if err != nil {
		log.Error("failed to generate signed URL for audio", "error", err)
		return nil, apperror.ErrInternal.WithMessage("Failed to generate audio URL").WithError(err)
	}

	return &response.GenerateAudioResponse{
		Data: response.AudioResponse{
			ID:         audioID,
			URL:        signedURL,
			MimeType:   "audio/mpeg",
			FileSize:   len(combinedAudio),
			DurationMs: audio.GetMP3DurationMs(combinedAudio),
		},
	}, nil
}
