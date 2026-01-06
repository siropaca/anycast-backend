package service

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/infrastructure/tts"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/audio"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// 署名 URL の有効期限（1時間）
const signedURLExpiration = 1 * time.Hour

// 台本行関連のビジネスロジックインターフェース
type ScriptLineService interface {
	ListByEpisodeID(ctx context.Context, userID, channelID, episodeID string) (*response.ScriptLineListResponse, error)
	GenerateAudio(ctx context.Context, userID, channelID, episodeID, lineID string) (*response.GenerateAudioResponse, error)
}

type scriptLineService struct {
	db             *gorm.DB
	scriptLineRepo repository.ScriptLineRepository
	episodeRepo    repository.EpisodeRepository
	channelRepo    repository.ChannelRepository
	audioRepo      repository.AudioRepository
	ttsClient      tts.Client
	storageClient  storage.Client
}

// ScriptLineService の実装を返す
func NewScriptLineService(
	db *gorm.DB,
	scriptLineRepo repository.ScriptLineRepository,
	episodeRepo repository.EpisodeRepository,
	channelRepo repository.ChannelRepository,
	audioRepo repository.AudioRepository,
	ttsClient tts.Client,
	storageClient storage.Client,
) ScriptLineService {
	return &scriptLineService{
		db:             db,
		scriptLineRepo: scriptLineRepo,
		episodeRepo:    episodeRepo,
		channelRepo:    channelRepo,
		audioRepo:      audioRepo,
		ttsClient:      ttsClient,
		storageClient:  storageClient,
	}
}

// 指定されたエピソードの台本行一覧を取得する
func (s *scriptLineService) ListByEpisodeID(ctx context.Context, userID, channelID, episodeID string) (*response.ScriptLineListResponse, error) {
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

	// 台本行一覧を取得
	scriptLines, err := s.scriptLineRepo.FindByEpisodeID(ctx, eid)
	if err != nil {
		return nil, err
	}

	// レスポンスに変換（署名 URL を生成）
	responses, err := s.toScriptLineResponses(ctx, scriptLines)
	if err != nil {
		return nil, err
	}

	return &response.ScriptLineListResponse{
		Data: responses,
	}, nil
}

// ScriptLine モデルのスライスをレスポンス DTO のスライスに変換する
func (s *scriptLineService) toScriptLineResponses(ctx context.Context, scriptLines []model.ScriptLine) ([]response.ScriptLineResponse, error) {
	result := make([]response.ScriptLineResponse, len(scriptLines))

	for i, sl := range scriptLines {
		resp, err := s.toScriptLineResponse(ctx, &sl)
		if err != nil {
			return nil, err
		}
		result[i] = resp
	}

	return result, nil
}

// ScriptLine モデルをレスポンス DTO に変換する
func (s *scriptLineService) toScriptLineResponse(ctx context.Context, sl *model.ScriptLine) (response.ScriptLineResponse, error) {
	resp := response.ScriptLineResponse{
		ID:         sl.ID,
		LineOrder:  sl.LineOrder,
		LineType:   string(sl.LineType),
		Text:       sl.Text,
		Emotion:    sl.Emotion,
		DurationMs: sl.DurationMs,
		CreatedAt:  sl.CreatedAt,
		UpdatedAt:  sl.UpdatedAt,
	}

	// decimal.Decimal から float64 に変換
	if sl.Volume != nil {
		v, _ := sl.Volume.Float64()
		resp.Volume = &v
	}

	if sl.Speaker != nil {
		resp.Speaker = &response.SpeakerResponse{
			ID:   sl.Speaker.ID,
			Name: sl.Speaker.Name,
		}
	}

	if sl.Sfx != nil {
		resp.Sfx = &response.SfxResponse{
			ID:   sl.Sfx.ID,
			Name: sl.Sfx.Name,
		}
	}

	if sl.Audio != nil {
		// 署名付き URL を生成
		signedURL, err := s.storageClient.GenerateSignedURL(ctx, sl.Audio.Path, signedURLExpiration)
		if err != nil {
			return response.ScriptLineResponse{}, err
		}
		resp.Audio = &response.AudioResponse{
			ID:         sl.Audio.ID,
			URL:        signedURL,
			MimeType:   sl.Audio.MimeType,
			FileSize:   sl.Audio.FileSize,
			DurationMs: sl.Audio.DurationMs,
		}
	}

	return resp, nil
}

// 台本行の音声を生成する
func (s *scriptLineService) GenerateAudio(ctx context.Context, userID, channelID, episodeID, lineID string) (*response.GenerateAudioResponse, error) {
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

	lid, err := uuid.Parse(lineID)
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

	// 台本行の取得
	scriptLine, err := s.scriptLineRepo.FindByID(ctx, lid)
	if err != nil {
		return nil, err
	}

	if scriptLine.EpisodeID != eid {
		return nil, apperror.ErrNotFound.WithMessage("Script line not found in this episode")
	}

	// speech 行以外はエラー
	if scriptLine.LineType != model.LineTypeSpeech {
		return nil, apperror.ErrValidation.WithMessage("Audio generation is only supported for speech lines")
	}

	// Speaker と Voice の確認
	if scriptLine.Speaker == nil {
		return nil, apperror.ErrValidation.WithMessage("Script line has no speaker")
	}

	if scriptLine.Text == nil || *scriptLine.Text == "" {
		return nil, apperror.ErrValidation.WithMessage("Script line has no text")
	}

	// TTS で音声生成（emotion は Prompt として別途渡す）
	audioData, err := s.ttsClient.Synthesize(
		ctx,
		*scriptLine.Text,
		scriptLine.Emotion,
		scriptLine.Speaker.Voice.ProviderVoiceID,
		scriptLine.Speaker.Voice.Gender,
	)
	if err != nil {
		return nil, err
	}

	// MP3 の duration を取得
	durationMs := audio.GetMP3DurationMs(audioData)

	// Audio モデルを作成（ID を先に生成して GCS パスに使用）
	newAudio := &model.Audio{
		MimeType:   "audio/mpeg",
		Filename:   fmt.Sprintf("%s.mp3", lid.String()),
		FileSize:   len(audioData),
		DurationMs: durationMs,
	}
	// GORM の BeforeCreate で ID が生成されるため、ここで明示的に生成
	newAudio.ID = uuid.New()
	newAudio.Path = s.storageClient.GenerateAudioPath(newAudio.ID.String())

	// GCS にアップロード
	_, err = s.storageClient.Upload(ctx, audioData, newAudio.Path, "audio/mpeg")
	if err != nil {
		return nil, err
	}

	// トランザクションで Audio 作成と ScriptLine 更新
	err = s.db.Transaction(func(tx *gorm.DB) error {
		txAudioRepo := repository.NewAudioRepository(tx)
		txScriptLineRepo := repository.NewScriptLineRepository(tx)

		// Audio を作成
		if err := txAudioRepo.Create(ctx, newAudio); err != nil {
			return err
		}

		// ScriptLine の AudioID を更新
		scriptLine.AudioID = &newAudio.ID
		if err := txScriptLineRepo.Update(ctx, scriptLine); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 署名付き URL を生成
	signedURL, err := s.storageClient.GenerateSignedURL(ctx, newAudio.Path, signedURLExpiration)
	if err != nil {
		return nil, err
	}

	return &response.GenerateAudioResponse{
		Audio: response.AudioResponse{
			ID:         newAudio.ID,
			URL:        signedURL,
			MimeType:   newAudio.MimeType,
			FileSize:   newAudio.FileSize,
			DurationMs: newAudio.DurationMs,
		},
	}, nil
}
