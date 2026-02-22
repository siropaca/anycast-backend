package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/cloudtasks"
	"github.com/siropaca/anycast-backend/internal/infrastructure/slack"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/infrastructure/stt"
	"github.com/siropaca/anycast-backend/internal/infrastructure/tts"
	"github.com/siropaca/anycast-backend/internal/infrastructure/websocket"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/audio"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// 音声生成のデフォルトパラメータ
const (
	defaultBgmVolumeDB    = -20.0
	defaultFadeOutMs      = 3000
	defaultPaddingStartMs = 1000
	defaultPaddingEndMs   = 3000
)

// AudioJobService は非同期音声生成ジョブを管理するインターフェースを表す
type AudioJobService interface {
	CreateJob(ctx context.Context, userID, channelID, episodeID string, req request.GenerateAudioAsyncRequest) (*response.AudioJobResponse, error)
	GetJob(ctx context.Context, userID, jobID string) (*response.AudioJobResponse, error)
	ListMyJobs(ctx context.Context, userID string, filter repository.AudioJobFilter) (*response.AudioJobListResponse, error)
	ExecuteJob(ctx context.Context, jobID string) error
	CancelJob(ctx context.Context, userID, jobID string) error
}

type audioJobService struct {
	audioJobRepo   repository.AudioJobRepository
	episodeRepo    repository.EpisodeRepository
	channelRepo    repository.ChannelRepository
	scriptLineRepo repository.ScriptLineRepository
	audioRepo      repository.AudioRepository
	bgmRepo        repository.BgmRepository
	systemBgmRepo  repository.SystemBgmRepository
	storageClient  storage.Client
	ttsRegistry    *tts.Registry
	sttClient      stt.Client
	ffmpegService  FFmpegService
	tasksClient    cloudtasks.Client
	wsHub          *websocket.Hub
	slackClient    slack.Client
}

// NewAudioJobService は audioJobService を生成して AudioJobService として返す
func NewAudioJobService(
	audioJobRepo repository.AudioJobRepository,
	episodeRepo repository.EpisodeRepository,
	channelRepo repository.ChannelRepository,
	scriptLineRepo repository.ScriptLineRepository,
	audioRepo repository.AudioRepository,
	bgmRepo repository.BgmRepository,
	systemBgmRepo repository.SystemBgmRepository,
	storageClient storage.Client,
	ttsRegistry *tts.Registry,
	sttClient stt.Client,
	ffmpegService FFmpegService,
	tasksClient cloudtasks.Client,
	wsHub *websocket.Hub,
	slackClient slack.Client,
) AudioJobService {
	return &audioJobService{
		audioJobRepo:   audioJobRepo,
		episodeRepo:    episodeRepo,
		channelRepo:    channelRepo,
		scriptLineRepo: scriptLineRepo,
		audioRepo:      audioRepo,
		bgmRepo:        bgmRepo,
		systemBgmRepo:  systemBgmRepo,
		storageClient:  storageClient,
		ttsRegistry:    ttsRegistry,
		sttClient:      sttClient,
		ffmpegService:  ffmpegService,
		tasksClient:    tasksClient,
		wsHub:          wsHub,
		slackClient:    slackClient,
	}
}

// CreateJob は非同期音声生成ジョブを作成して返す
func (s *audioJobService) CreateJob(ctx context.Context, userID, channelID, episodeID string, req request.GenerateAudioAsyncRequest) (*response.AudioJobResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("このエピソードの音声生成権限がありません")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("このチャンネルにエピソードが見つかりません")
	}

	// 既存の処理中ジョブを確認
	pendingJob, err := s.audioJobRepo.FindPendingByEpisodeID(ctx, eid)
	if err != nil {
		return nil, err
	}
	if pendingJob != nil {
		return nil, apperror.ErrValidation.WithMessage("このエピソードは既に音声生成中です")
	}

	// type に応じたバリデーション
	jobType := model.AudioJobType(req.Type)

	switch jobType {
	case model.AudioJobTypeVoice, model.AudioJobTypeFull:
		// 台本行の存在確認
		scriptLines, err := s.scriptLineRepo.FindByEpisodeIDWithVoice(ctx, eid)
		if err != nil {
			return nil, err
		}
		if len(scriptLines) == 0 {
			return nil, apperror.ErrValidation.WithMessage("このエピソードには台本行がありません")
		}
	case model.AudioJobTypeRemix:
		// voice_audio_id の存在確認
		if episode.VoiceAudioID == nil {
			return nil, apperror.ErrValidation.WithMessage("ボイス音声がありません。先に音声生成を実行してください")
		}
	}

	// BGM バリデーション（full / remix の場合）
	var bgmID *uuid.UUID
	var systemBgmID *uuid.UUID

	if jobType == model.AudioJobTypeFull || jobType == model.AudioJobTypeRemix {
		// bgmId と systemBgmId の同時指定チェック
		if req.BgmID != nil && req.SystemBgmID != nil {
			return nil, apperror.ErrValidation.WithMessage("bgmId と systemBgmId は同時に指定できません")
		}
		// どちらも指定されていない場合（full は必須、remix は BGM なしを許可）
		if req.BgmID == nil && req.SystemBgmID == nil {
			if jobType == model.AudioJobTypeFull {
				return nil, apperror.ErrValidation.WithMessage("bgmId または systemBgmId のいずれかを指定してください")
			}
		}

		if req.BgmID != nil {
			bid, err := uuid.Parse(*req.BgmID)
			if err != nil {
				return nil, apperror.ErrValidation.WithMessage("無効な bgmId です")
			}
			bgm, err := s.bgmRepo.FindByID(ctx, bid)
			if err != nil {
				return nil, err
			}
			if bgm.UserID != uid {
				return nil, apperror.ErrForbidden.WithMessage("この BGM へのアクセス権限がありません")
			}
			bgmID = &bid
		}

		if req.SystemBgmID != nil {
			sbid, err := uuid.Parse(*req.SystemBgmID)
			if err != nil {
				return nil, apperror.ErrValidation.WithMessage("無効な systemBgmId です")
			}
			systemBgm, err := s.systemBgmRepo.FindByID(ctx, sbid)
			if err != nil {
				return nil, err
			}
			if !systemBgm.IsActive {
				return nil, apperror.ErrNotFound.WithMessage("このシステム BGM は利用できません")
			}
			systemBgmID = &sbid
		}
	}

	// デフォルト値を設定
	bgmVolumeDB := defaultBgmVolumeDB
	if req.BgmVolumeDB != nil {
		bgmVolumeDB = *req.BgmVolumeDB
	}

	fadeOutMs := defaultFadeOutMs
	if req.FadeOutMs != nil {
		fadeOutMs = *req.FadeOutMs
	}

	paddingStartMs := defaultPaddingStartMs
	if req.PaddingStartMs != nil {
		paddingStartMs = *req.PaddingStartMs
	}

	paddingEndMs := defaultPaddingEndMs
	if req.PaddingEndMs != nil {
		paddingEndMs = *req.PaddingEndMs
	}

	// ジョブを作成
	job := &model.AudioJob{
		EpisodeID:      eid,
		UserID:         uid,
		Status:         model.AudioJobStatusPending,
		JobType:        jobType,
		Progress:       0,
		BgmID:          bgmID,
		SystemBgmID:    systemBgmID,
		BgmVolumeDB:    bgmVolumeDB,
		FadeOutMs:      fadeOutMs,
		PaddingStartMs: paddingStartMs,
		PaddingEndMs:   paddingEndMs,
	}

	if err := s.audioJobRepo.Create(ctx, job); err != nil {
		return nil, err
	}

	// Cloud Tasks が設定されている場合はエンキュー、そうでなければ goroutine で直接実行
	if s.tasksClient != nil {
		if err := s.tasksClient.EnqueueAudioJob(ctx, job.ID.String()); err != nil {
			log.Error("failed to enqueue job", "error", err, "job_id", job.ID)
			// エンキュー失敗時はジョブを失敗状態に更新（ベストエフォート）
			job.Status = model.AudioJobStatusFailed
			errMsg := "タスクのエンキューに失敗しました"
			errCode := "ENQUEUE_FAILED"
			job.ErrorMessage = &errMsg
			job.ErrorCode = &errCode
			_ = s.audioJobRepo.Update(ctx, job) //nolint:errcheck // best effort cleanup
			return nil, apperror.ErrInternal.WithMessage("音声生成タスクの登録に失敗しました").WithError(err)
		}
		log.Info("audio job created and enqueued", "job_id", job.ID, "episode_id", eid)
	} else {
		// ローカル開発モード: goroutine で直接実行
		log.Info("executing job directly as Cloud Tasks is not configured", "job_id", job.ID, "episode_id", eid)
		go func() {
			if err := s.ExecuteJob(context.Background(), job.ID.String()); err != nil {
				log.Error("failed to execute local job", "error", err, "job_id", job.ID)
			}
		}()
	}

	return s.toAudioJobResponse(ctx, job)
}

// GetJob は指定されたジョブの詳細を取得する
func (s *audioJobService) GetJob(ctx context.Context, userID, jobID string) (*response.AudioJobResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	jid, err := uuid.Parse(jobID)
	if err != nil {
		return nil, err
	}

	job, err := s.audioJobRepo.FindByID(ctx, jid)
	if err != nil {
		return nil, err
	}

	// オーナーチェック
	if job.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("このジョブへのアクセス権限がありません")
	}

	return s.toAudioJobResponse(ctx, job)
}

// ListMyJobs は指定されたユーザーのジョブ一覧を取得する
func (s *audioJobService) ListMyJobs(ctx context.Context, userID string, filter repository.AudioJobFilter) (*response.AudioJobListResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	jobs, err := s.audioJobRepo.FindByUserID(ctx, uid, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]response.AudioJobResponse, len(jobs))
	for i, job := range jobs {
		resp, err := s.toAudioJobResponse(ctx, &job)
		if err != nil {
			return nil, err
		}
		responses[i] = *resp
	}

	return &response.AudioJobListResponse{
		Data: responses,
	}, nil
}

// ExecuteJob は指定されたジョブを実行する（Cloud Tasks ワーカーから呼び出される）
func (s *audioJobService) ExecuteJob(ctx context.Context, jobID string) error {
	log := logger.FromContext(ctx)

	jid, err := uuid.Parse(jobID)
	if err != nil {
		return err
	}

	// ジョブを取得
	job, err := s.audioJobRepo.FindByID(ctx, jid)
	if err != nil {
		return err
	}

	// 既に完了、失敗、またはキャンセル済みの場合はスキップ
	if job.Status == model.AudioJobStatusCompleted ||
		job.Status == model.AudioJobStatusFailed ||
		job.Status == model.AudioJobStatusCanceled {
		log.Info("skipping job as it is already completed", "job_id", jobID, "status", job.Status)
		return nil
	}

	// キャンセル中の場合はキャンセル完了にする
	if job.Status == model.AudioJobStatusCanceling {
		log.Info("completing cancellation for job", "job_id", jobID)
		job.Status = model.AudioJobStatusCanceled
		now := time.Now().UTC()
		job.CompletedAt = &now
		if err := s.audioJobRepo.Update(ctx, job); err != nil {
			return err
		}
		s.notifyCanceled(job.ID.String(), job.UserID.String())
		return nil
	}

	// 処理開始
	now := time.Now().UTC()
	job.Status = model.AudioJobStatusProcessing
	job.StartedAt = &now
	if err := s.audioJobRepo.Update(ctx, job); err != nil {
		return err
	}

	// WebSocket で開始通知
	s.notifyProgress(job.ID.String(), job.UserID.String(), 0, "音声生成を開始しています...")

	// ジョブタイプに応じて処理を実行
	var execErr error
	switch job.JobType {
	case model.AudioJobTypeRemix:
		execErr = s.executeRemixInternal(ctx, job)
	case model.AudioJobTypeVoice, model.AudioJobTypeFull:
		execErr = s.executeJobInternal(ctx, job)
	default:
		execErr = s.executeJobInternal(ctx, job)
	}

	// 処理実行（エラー時はジョブを失敗状態に）
	if err := execErr; err != nil {
		// キャンセルによるエラーの場合は失敗扱いにしない
		if apperror.IsCode(err, apperror.CodeCanceled) {
			log.Info("job was canceled during execution", "job_id", jobID)
			return nil
		}
		log.Error("failed to execute job", "error", err, "job_id", jobID)
		s.failJob(ctx, job, err)
		return err
	}

	return nil
}

// executeJobInternal は音声生成処理を実行する
func (s *audioJobService) executeJobInternal(ctx context.Context, job *model.AudioJob) error {
	log := logger.FromContext(ctx)

	// エピソードと台本を取得
	episode, err := s.episodeRepo.FindByID(ctx, job.EpisodeID)
	if err != nil {
		return err
	}

	scriptLines, err := s.scriptLineRepo.FindByEpisodeIDWithVoice(ctx, job.EpisodeID)
	if err != nil {
		return err
	}

	if len(scriptLines) == 0 {
		return apperror.ErrValidation.WithMessage("台本行がありません")
	}

	// 進捗: 10%
	s.updateProgress(ctx, job, 10, "台本を読み込み中...")

	// TTS 用のデータを構築
	var turns []tts.SpeakerTurn
	speakerAliasMap := make(map[string]string)
	voiceConfigMap := make(map[string]string)
	speakerIndex := 1

	for _, line := range scriptLines {
		if line.Text == "" {
			continue
		}

		alias, exists := speakerAliasMap[line.Speaker.Name]
		if !exists {
			alias = fmt.Sprintf("speaker%d", speakerIndex)
			speakerAliasMap[line.Speaker.Name] = alias
			voiceConfigMap[alias] = line.Speaker.Voice.ProviderVoiceID
			speakerIndex++
		}

		turns = append(turns, tts.SpeakerTurn{
			Speaker: alias,
			Text:    line.Text,
			Emotion: line.Emotion,
		})
	}

	if len(turns) == 0 {
		return apperror.ErrValidation.WithMessage("音声生成に使用できる台本行がありません")
	}

	voiceConfigs := make([]tts.SpeakerVoiceConfig, 0, len(voiceConfigMap))
	for alias, voiceID := range voiceConfigMap {
		voiceConfigs = append(voiceConfigs, tts.SpeakerVoiceConfig{
			SpeakerAlias: alias,
			VoiceID:      voiceID,
		})
	}

	// 進捗: 20%
	s.updateProgress(ctx, job, 20, "音声を生成中...")

	// キャンセルチェック（TTS 開始前）
	if err := s.checkCanceled(ctx, job); err != nil {
		return err
	}

	// scriptLines の最初の行の Voice.Provider から TTS クライアントを取得
	provider := tts.Provider(scriptLines[0].Speaker.Voice.Provider)
	ttsClient, err := s.ttsRegistry.Get(provider)
	if err != nil {
		log.Error("TTS provider not available", "provider", provider)
		return apperror.ErrValidation.WithMessage(fmt.Sprintf("TTS プロバイダ %q が利用できません", provider)).WithError(err)
	}

	log.Info("generating audio", "total_turns", len(turns), "provider", provider)

	// TTS で音声を生成
	var result *tts.SynthesisResult
	switch {
	case len(voiceConfigs) == 1:
		// シングルスピーカー: 全ターンのテキストを連結して単一話者で合成
		var textBuilder strings.Builder
		for _, turn := range turns {
			text := turn.Text
			if turn.Emotion != nil && *turn.Emotion != "" {
				text = fmt.Sprintf("[%s] %s", *turn.Emotion, turn.Text)
			}
			textBuilder.WriteString(text + "\n")
		}
		result, err = ttsClient.Synthesize(ctx, textBuilder.String(), nil, voiceConfigs[0].VoiceID, scriptLines[0].Speaker.Voice.Gender)
	case provider == tts.ProviderGoogle:
		// Gemini: 話者別合成 + 無音分割再アセンブル
		result, err = s.synthesizeMultiSpeakerByReassembly(ctx, job, turns, voiceConfigs, ttsClient, scriptLines)
	default:
		// ElevenLabs 等: 従来方式
		result, err = ttsClient.SynthesizeMultiSpeaker(ctx, turns, voiceConfigs)
	}
	if err != nil {
		log.Error("TTS failed", "error", err)
		return apperror.ErrGenerationFailed.WithMessage("音声の生成に失敗しました").WithError(err)
	}

	// フォーマットに応じて MP3 に変換
	s.updateProgress(ctx, job, 45, "音声を変換中...")
	var voiceAudio []byte
	if result.Format == "pcm" {
		log.Info("audio synthesis succeeded, converting PCM to MP3", "pcm_size", len(result.Data))
		voiceAudio, err = s.ffmpegService.ConvertToMP3(ctx, result.Data, "pcm", result.SampleRate)
		if err != nil {
			log.Error("PCM to MP3 conversion failed", "error", err)
			return apperror.ErrInternal.WithMessage("音声フォーマットの変換に失敗しました").WithError(err)
		}
	} else {
		log.Info("audio synthesis succeeded", "format", result.Format, "size", len(result.Data))
		voiceAudio = result.Data
	}

	// 進捗: 50%
	s.updateProgress(ctx, job, 50, "音声生成完了")

	// ボイス音声を GCS に保存
	s.updateProgress(ctx, job, 52, "ボイス音声を保存中...")
	voiceAudioID := uuid.New()
	voiceAudioPath := storage.GenerateAudioPath(voiceAudioID.String())

	if _, err := s.storageClient.Upload(ctx, voiceAudio, voiceAudioPath, "audio/mpeg"); err != nil {
		log.Error("failed to upload voice audio", "error", err)
		return apperror.ErrInternal.WithMessage("ボイス音声のアップロードに失敗しました").WithError(err)
	}

	voiceDurationMs, err := audio.GetDurationMsE(voiceAudio)
	if err != nil {
		log.Warn("failed to get voice audio duration", "error", err)
	}

	voiceAudioRecord := &model.Audio{
		ID:         voiceAudioID,
		MimeType:   "audio/mpeg",
		Path:       voiceAudioPath,
		Filename:   voiceAudioID.String() + ".mp3",
		FileSize:   len(voiceAudio),
		DurationMs: voiceDurationMs,
	}

	if err := s.audioRepo.Create(ctx, voiceAudioRecord); err != nil {
		log.Error("failed to create voice audio record", "error", err)
		return apperror.ErrInternal.WithMessage("ボイス音声レコードの保存に失敗しました").WithError(err)
	}

	// エピソードにボイス音声を設定
	episode.VoiceAudioID = &voiceAudioID
	episode.VoiceAudio = nil

	// 最終的な音声データ
	var finalAudio []byte

	// キャンセルチェック（BGM ミキシング前）
	if err := s.checkCanceled(ctx, job); err != nil {
		return err
	}

	// type=full の場合は BGM ミキシング
	if job.JobType == model.AudioJobTypeFull && (job.BgmID != nil || job.SystemBgmID != nil) {
		s.updateProgress(ctx, job, 55, "BGM をミキシング中...")

		// BGM データを取得
		var bgmPath string
		if job.BgmID != nil {
			bgm, err := s.bgmRepo.FindByID(ctx, *job.BgmID)
			if err != nil {
				return err
			}
			bgmPath = bgm.Audio.Path
		} else if job.SystemBgmID != nil {
			systemBgm, err := s.systemBgmRepo.FindByID(ctx, *job.SystemBgmID)
			if err != nil {
				return err
			}
			bgmPath = systemBgm.Audio.Path
		}

		// GCS から BGM をダウンロード
		bgmData, err := s.downloadFromStorage(ctx, bgmPath)
		if err != nil {
			log.Error("failed to download BGM", "error", err, "path", bgmPath)
			return apperror.ErrInternal.WithMessage("BGM のダウンロードに失敗しました").WithError(err)
		}

		// 進捗: 70%
		s.updateProgress(ctx, job, 70, "BGM をミキシング中...")

		// FFmpeg でミキシング
		finalAudio, err = s.ffmpegService.MixAudioWithBGM(ctx, MixParams{
			VoiceData:       voiceAudio,
			BGMData:         bgmData,
			VoiceDurationMs: voiceDurationMs,
			BGMVolumeDB:     job.BgmVolumeDB,
			FadeOutMs:       job.FadeOutMs,
			PaddingStartMs:  job.PaddingStartMs,
			PaddingEndMs:    job.PaddingEndMs,
		})
		if err != nil {
			log.Error("FFmpeg mixing failed", "error", err)
			return apperror.ErrInternal.WithMessage("BGM のミキシングに失敗しました").WithError(err)
		}
	} else {
		// BGM なしの場合はそのまま
		finalAudio = voiceAudio
	}

	// 進捗: 85%
	s.updateProgress(ctx, job, 85, "音声をアップロード中...")

	// キャンセルチェック（アップロード前）
	if err := s.checkCanceled(ctx, job); err != nil {
		return err
	}

	// 新しい Audio ID を生成してアップロード
	audioID := uuid.New()
	audioPath := storage.GenerateAudioPath(audioID.String())

	if _, err := s.storageClient.Upload(ctx, finalAudio, audioPath, "audio/mpeg"); err != nil {
		log.Error("failed to upload audio", "error", err)
		return apperror.ErrInternal.WithMessage("音声のアップロードに失敗しました").WithError(err)
	}

	// 最終的な長さを取得（ミキシングした場合は変わる可能性がある）
	finalDurationMs, err := audio.GetDurationMsE(finalAudio)
	if err != nil {
		log.Warn("failed to get final audio duration, using estimate", "error", err)
		// ミキシングした場合は計算、そうでなければ voiceDurationMs を使用
		if voiceDurationMs > 0 {
			finalDurationMs = job.PaddingStartMs + voiceDurationMs + job.PaddingEndMs
		}
	}

	// Audio レコードを作成
	audioRecord := &model.Audio{
		ID:         audioID,
		MimeType:   "audio/mpeg",
		Path:       audioPath,
		Filename:   audioID.String() + ".mp3",
		FileSize:   len(finalAudio),
		DurationMs: finalDurationMs,
	}

	if err := s.audioRepo.Create(ctx, audioRecord); err != nil {
		log.Error("failed to create audio record", "error", err)
		return apperror.ErrInternal.WithMessage("音声レコードの保存に失敗しました").WithError(err)
	}

	// 進捗: 95%
	s.updateProgress(ctx, job, 95, "エピソードを更新中...")

	// エピソードを更新
	episode.FullAudioID = &audioID
	episode.FullAudio = nil

	// full の場合は BGM 情報をエピソードに記録
	if job.JobType == model.AudioJobTypeFull {
		episode.BgmID = job.BgmID
		episode.SystemBgmID = job.SystemBgmID
		episode.Bgm = nil
		episode.SystemBgm = nil
	}

	if err := s.episodeRepo.Update(ctx, episode); err != nil {
		return err
	}

	// キャンセルチェック（完了遷移前）
	if err := s.checkCanceled(ctx, job); err != nil {
		return err
	}

	// ジョブを完了状態に更新
	completedAt := time.Now().UTC()
	job.Status = model.AudioJobStatusCompleted
	job.Progress = 100
	job.CompletedAt = &completedAt
	job.ResultAudioID = &audioID

	if err := s.audioJobRepo.Update(ctx, job); err != nil {
		return err
	}

	// WebSocket で完了通知
	s.notifyCompleted(job.ID.String(), job.UserID.String(), audioRecord)

	log.Info("audio job completed successfully", "job_id", job.ID, "audio_id", audioID)

	return nil
}

// reassemblySegment は再アセンブル用のセグメント情報
type reassemblySegment struct {
	originalIndex int    // 元の台本での順序
	pcmData       []byte // 分割された PCM データ
}

// reassemblySpeakerGroup は話者ごとのグループ情報
type reassemblySpeakerGroup struct {
	alias           string // 話者エイリアス（speaker1 等）
	voiceID         string // Voice の ProviderVoiceID
	gender          model.Gender
	texts           []string // TTS に送るテキスト一覧（感情指示を含む）
	spokenTexts     []string // 実際に発話されるテキスト一覧（感情指示を除く、アライメント用）
	originalIndices []int    // 元の台本でのインデックス一覧
}

// synthesizeMultiSpeakerByReassembly は話者別にシングルスピーカー合成し、
// 無音分割で個別セグメントに分けた後、元の順序に再アセンブルする
func (s *audioJobService) synthesizeMultiSpeakerByReassembly(
	ctx context.Context,
	job *model.AudioJob,
	turns []tts.SpeakerTurn,
	voiceConfigs []tts.SpeakerVoiceConfig,
	ttsClient tts.Client,
	scriptLines []model.ScriptLine,
) (*tts.SynthesisResult, error) {
	log := logger.FromContext(ctx)

	// VoiceConfig のマップを構築
	voiceConfigMap := make(map[string]tts.SpeakerVoiceConfig)
	for _, vc := range voiceConfigs {
		voiceConfigMap[vc.SpeakerAlias] = vc
	}

	// 話者の Gender マップを構築
	genderMap := make(map[string]model.Gender)
	for _, line := range scriptLines {
		alias := ""
		for a, vc := range voiceConfigMap {
			if vc.VoiceID == line.Speaker.Voice.ProviderVoiceID {
				alias = a
				break
			}
		}
		if alias != "" {
			genderMap[alias] = line.Speaker.Voice.Gender
		}
	}

	// Step 1: 話者別にグループ化（元のインデックスを保持）
	speakerGroups := make(map[string]*reassemblySpeakerGroup)
	for i, turn := range turns {
		group, exists := speakerGroups[turn.Speaker]
		if !exists {
			vc := voiceConfigMap[turn.Speaker]
			group = &reassemblySpeakerGroup{
				alias:   turn.Speaker,
				voiceID: vc.VoiceID,
				gender:  genderMap[turn.Speaker],
			}
			speakerGroups[turn.Speaker] = group
		}

		// 発話テキスト（感情指示を除く）: アライメントに使用
		spokenText := turn.Text
		if !strings.HasSuffix(spokenText, "。") {
			spokenText += "。"
		}

		// TTS 用テキスト（感情指示を含む）
		ttsText := spokenText
		if turn.Emotion != nil && *turn.Emotion != "" {
			ttsText = fmt.Sprintf("[%s] %s", *turn.Emotion, spokenText)
		}

		group.texts = append(group.texts, ttsText)
		group.spokenTexts = append(group.spokenTexts, spokenText)
		group.originalIndices = append(group.originalIndices, i)
	}

	log.Info("reassembly: grouped turns by speaker",
		"speaker_count", len(speakerGroups),
		"total_turns", len(turns),
	)

	// Step 2: 話者ごとにシングルスピーカー合成（並列実行）
	type speakerResult struct {
		alias           string
		pcmData         []byte
		originalIndices []int
	}

	var mu sync.Mutex
	results := make(map[string]*speakerResult)

	eg, egCtx := errgroup.WithContext(ctx)

	const maxTTSRetries = 2

	for _, group := range speakerGroups {
		g := group
		eg.Go(func() error {
			// テキストを改行区切りで連結し、音声スタイルを先頭に付加
			fullText := tts.DefaultVoiceStyle + "\n\n" + strings.Join(g.texts, "\n")

			log.Debug("reassembly: synthesizing speaker",
				"alias", g.alias,
				"line_count", len(g.texts),
				"text_length", len(fullText),
			)

			var result *tts.SynthesisResult
			var lastErr error
			for attempt := 0; attempt <= maxTTSRetries; attempt++ {
				if attempt > 0 {
					log.Warn("reassembly: retrying TTS synthesis",
						"alias", g.alias,
						"attempt", attempt+1,
						"prev_error", lastErr,
					)
					// リトライ前に少し待つ
					select {
					case <-egCtx.Done():
						return fmt.Errorf("話者 %s の合成に失敗しました: %w", g.alias, egCtx.Err())
					case <-time.After(time.Duration(attempt*2) * time.Second):
					}
				}

				result, lastErr = ttsClient.Synthesize(egCtx, fullText, nil, g.voiceID, g.gender)
				if lastErr == nil {
					break
				}
			}
			if lastErr != nil {
				return fmt.Errorf("話者 %s の合成に失敗しました: %w", g.alias, lastErr)
			}

			if result.Format != "pcm" {
				return fmt.Errorf("話者 %s の出力フォーマットが pcm ではありません: %s", g.alias, result.Format)
			}

			mu.Lock()
			results[g.alias] = &speakerResult{
				alias:           g.alias,
				pcmData:         result.Data,
				originalIndices: g.originalIndices,
			}
			mu.Unlock()

			log.Debug("reassembly: speaker synthesis done",
				"alias", g.alias,
				"pcm_size", len(result.Data),
			)

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// Step 3: STT アライメント + silencedetect スナップのハイブリッド方式で行境界を特定し分割
	// デバッグ用: 分割前のスピーカー別オリジナル音源をファイルに保存
	debugDir := filepath.Join("tmp", "audio-debug", job.ID.String())
	if err := os.MkdirAll(debugDir, 0o755); err != nil {
		log.Warn("reassembly: failed to create debug directory", "error", err)
	} else {
		for alias, res := range results {
			wavData := audio.EncodeWAV(res.pcmData, 24000, 1, 2)
			debugPath := filepath.Join(debugDir, fmt.Sprintf("speaker_%s_original.wav", alias))
			if writeErr := os.WriteFile(debugPath, wavData, 0o644); writeErr != nil {
				log.Warn("reassembly: failed to write debug audio", "alias", alias, "error", writeErr)
			} else {
				log.Debug("reassembly: saved original speaker audio",
					"alias", alias,
					"path", debugPath,
					"pcm_bytes", len(res.pcmData),
				)
			}
		}
	}

	var allSegments []reassemblySegment

	for alias, res := range results {
		group := speakerGroups[alias]

		// STT で単語レベルのタイムスタンプを取得
		log.Debug("reassembly: recognizing speech for alignment", "alias", alias)
		sttWords, err := s.sttClient.RecognizeWithTimestamps(ctx, res.pcmData, 24000)
		if err != nil {
			return nil, fmt.Errorf("話者 %s の音声認識に失敗しました: %w", alias, err)
		}

		log.Debug("reassembly: STT result",
			"alias", alias,
			"word_count", len(sttWords),
		)
		for i, w := range sttWords {
			log.Debug("reassembly: STT word",
				"alias", alias,
				"index", i,
				"word", w.Word,
				"start_ms", w.StartTime.Milliseconds(),
				"end_ms", w.EndTime.Milliseconds(),
			)
		}

		// STT の WordTimestamp を audio パッケージの型に変換
		audioWords := make([]audio.WordTimestamp, len(sttWords))
		for i, w := range sttWords {
			audioWords[i] = audio.WordTimestamp{
				Word:      w.Word,
				StartTime: w.StartTime,
				EndTime:   w.EndTime,
			}
		}

		// DP アライメントで行境界を特定
		boundaries, err := audio.AlignTextToTimestamps(group.spokenTexts, audioWords)
		if err != nil {
			return nil, fmt.Errorf("話者 %s のテキストアライメントに失敗しました: %w", alias, err)
		}
		for i, b := range boundaries {
			text := ""
			if i < len(group.spokenTexts) {
				text = group.spokenTexts[i]
			}
			log.Debug("reassembly: DP alignment boundary",
				"alias", alias,
				"line", i,
				"text", text,
				"start_ms", b.StartTime.Milliseconds(),
				"end_ms", b.EndTime.Milliseconds(),
			)
		}

		// 先頭と末尾の境界を PCM データの実際の範囲に拡張する
		// STT の単語タイムスタンプは音声の先頭/末尾の余白をカバーしないため
		pcmDuration := time.Duration(float64(len(res.pcmData)) / float64(24000*1*2) * float64(time.Second))
		boundaries[0].StartTime = 0
		boundaries[len(boundaries)-1].EndTime = pcmDuration

		// STT アライメントで得た行境界をログ出力
		for i, b := range boundaries {
			log.Debug("reassembly: STT boundary",
				"alias", alias,
				"line", i,
				"start_ms", b.StartTime.Milliseconds(),
				"end_ms", b.EndTime.Milliseconds(),
			)
		}

		// 行が2行以上の場合、silencedetect で正確なカット位置にスナップする
		if len(group.spokenTexts) >= 2 {
			silences, silenceErr := audio.DetectSilenceIntervals(res.pcmData, audio.PCMSplitConfig{
				SampleRate:     24000,
				Channels:       1,
				BytesPerSample: 2,
				NoiseDB:        -30,
				MinSilenceSec:  0.2,
			})
			if silenceErr != nil {
				log.Warn("reassembly: silencedetect failed, using STT boundaries",
					"alias", alias,
					"error", silenceErr,
				)
			} else {
				// 検出された全無音区間をログ出力
				for j, s := range silences {
					log.Debug("reassembly: silence interval",
						"alias", alias,
						"index", j,
						"start_ms", int(s.StartSec*1000),
						"end_ms", int(s.EndSec*1000),
						"mid_ms", int((s.StartSec+s.EndSec)/2.0*1000),
					)
				}

				sttBoundaries := make([]audio.LineBoundary, len(boundaries))
				copy(sttBoundaries, boundaries)

				boundaries = audio.SnapBoundariesToSilence(boundaries, silences, 500*time.Millisecond)

				// スナップ前後の差分をログ出力
				for i := 0; i < len(boundaries)-1; i++ {
					before := sttBoundaries[i].EndTime.Milliseconds()
					after := boundaries[i].EndTime.Milliseconds()
					if before != after {
						log.Debug("reassembly: boundary snapped",
							"alias", alias,
							"cut", i,
							"before_ms", before,
							"after_ms", after,
							"delta_ms", after-before,
						)
					} else {
						log.Debug("reassembly: boundary not snapped",
							"alias", alias,
							"cut", i,
							"at_ms", before,
						)
					}
				}
			}
		}

		// タイムスタンプ境界で PCM を分割
		segments := audio.SplitPCMByTimestamps(res.pcmData, boundaries, 24000, 1, 2)

		for i, seg := range segments {
			allSegments = append(allSegments, reassemblySegment{
				originalIndex: res.originalIndices[i],
				pcmData:       seg,
			})
		}

		log.Debug("reassembly: alignment split done",
			"alias", alias,
			"segments", len(segments),
			"lines", len(group.texts),
		)
	}

	// Step 4: 元の順序で再アセンブル（セグメント間に 200ms 無音パディング挿入）
	sort.Slice(allSegments, func(i, j int) bool {
		return allSegments[i].originalIndex < allSegments[j].originalIndex
	})

	silencePadding := audio.GenerateSilencePCM(200, 24000, 1, 2)

	pcmParts := make([][]byte, 0, len(allSegments)*2)
	for i, seg := range allSegments {
		pcmParts = append(pcmParts, seg.pcmData)
		// 最後のセグメント以外にはパディングを挿入
		if i < len(allSegments)-1 {
			pcmParts = append(pcmParts, silencePadding)
		}
	}

	finalPCM := audio.ConcatPCM(pcmParts)

	log.Info("reassembly: completed",
		"total_segments", len(allSegments),
		"final_pcm_size", len(finalPCM),
	)

	return &tts.SynthesisResult{
		Data:       finalPCM,
		Format:     "pcm",
		SampleRate: 24000,
	}, nil
}

// downloadFromStorage は指定されたパスのファイルを GCS からダウンロードする
func (s *audioJobService) downloadFromStorage(ctx context.Context, path string) ([]byte, error) {
	// storage.Client に Download メソッドがある前提
	type downloader interface {
		Download(ctx context.Context, path string) ([]byte, error)
	}

	if d, ok := s.storageClient.(downloader); ok {
		return d.Download(ctx, path)
	}

	return nil, apperror.ErrInternal.WithMessage("ストレージクライアントが Download をサポートしていません")
}

// updateProgress はジョブの進捗を更新し WebSocket で通知する
//
// 進捗のみを更新し、ステータスなど他のフィールドは変更しない
func (s *audioJobService) updateProgress(ctx context.Context, job *model.AudioJob, progress int, message string) {
	job.Progress = progress
	_ = s.audioJobRepo.UpdateProgress(ctx, job.ID, progress) //nolint:errcheck // progress update is best effort
	s.notifyProgress(job.ID.String(), job.UserID.String(), progress, message)
}

// checkCanceled はジョブがキャンセルされていないかチェックする
//
// キャンセル中の場合は canceled に遷移させ、ErrCanceled を返す
func (s *audioJobService) checkCanceled(ctx context.Context, job *model.AudioJob) error {
	// 最新のステータスを取得
	latestJob, err := s.audioJobRepo.FindByID(ctx, job.ID)
	if err != nil {
		return err
	}

	if latestJob.Status == model.AudioJobStatusCanceling {
		// canceled に遷移
		latestJob.Status = model.AudioJobStatusCanceled
		now := time.Now().UTC()
		latestJob.CompletedAt = &now
		if err := s.audioJobRepo.Update(ctx, latestJob); err != nil {
			return err
		}
		s.notifyCanceled(latestJob.ID.String(), latestJob.UserID.String())
		return apperror.ErrCanceled.WithMessage("ジョブがキャンセルされました")
	}

	return nil
}

// failJob は指定されたジョブを失敗状態に更新する
func (s *audioJobService) failJob(ctx context.Context, job *model.AudioJob, err error) {
	log := logger.FromContext(ctx)
	completedAt := time.Now().UTC()
	job.Status = model.AudioJobStatusFailed
	job.CompletedAt = &completedAt

	var appErr *apperror.AppError
	if ok := errors.As(err, &appErr); ok {
		errCode := string(appErr.Code)
		job.ErrorCode = &errCode
		job.ErrorMessage = &appErr.Message
	} else {
		errCode := "INTERNAL_ERROR"
		errMsg := "内部エラーが発生しました"
		job.ErrorCode = &errCode
		job.ErrorMessage = &errMsg
	}

	_ = s.audioJobRepo.Update(ctx, job) //nolint:errcheck // fail update is best effort
	s.notifyFailed(job.ID.String(), job.UserID.String(), job.ErrorCode, job.ErrorMessage)

	// Slack アラート通知（ベストエフォート）
	if s.slackClient != nil {
		errCode := ""
		errMsg := ""
		if job.ErrorCode != nil {
			errCode = *job.ErrorCode
		}
		if job.ErrorMessage != nil {
			errMsg = *job.ErrorMessage
		}
		if alertErr := s.slackClient.SendAlert(ctx, slack.AlertNotification{
			JobID:        job.ID.String(),
			JobType:      "音声生成 (Audio)",
			ErrorCode:    errCode,
			ErrorMessage: errMsg,
			OccurredAt:   completedAt,
		}); alertErr != nil {
			log.Warn("failed to send slack alert for audio job", "error", alertErr, "job_id", job.ID)
		}
	}
}

// notifyProgress はジョブの進捗を WebSocket で通知する
func (s *audioJobService) notifyProgress(jobID, userID string, progress int, message string) {
	if s.wsHub == nil {
		return
	}
	s.wsHub.SendToUser(userID, websocket.Message{
		Type: "audio_progress",
		Payload: map[string]interface{}{
			"jobId":    jobID,
			"progress": progress,
			"message":  message,
		},
	})
}

// notifyCompleted はジョブの完了を WebSocket で通知する
func (s *audioJobService) notifyCompleted(jobID, userID string, audioModel *model.Audio) {
	log := logger.Default()
	if s.wsHub == nil {
		log.Warn("skipping completion notification as WebSocket Hub is not configured", "job_id", jobID, "user_id", userID)
		return
	}
	log.Info("notifying completion via WebSocket", "job_id", jobID, "user_id", userID, "audio_id", audioModel.ID.String())
	s.wsHub.SendToUser(userID, websocket.Message{
		Type: "audio_completed",
		Payload: map[string]interface{}{
			"jobId": jobID,
			"audio": map[string]interface{}{
				"id":         audioModel.ID.String(),
				"durationMs": audioModel.DurationMs,
			},
		},
	})
}

// CancelJob は指定されたジョブをキャンセルする
func (s *audioJobService) CancelJob(ctx context.Context, userID, jobID string) error {
	log := logger.FromContext(ctx)

	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	jid, err := uuid.Parse(jobID)
	if err != nil {
		return err
	}

	// ジョブを取得
	job, err := s.audioJobRepo.FindByID(ctx, jid)
	if err != nil {
		return err
	}

	// オーナーチェック
	if job.UserID != uid {
		return apperror.ErrForbidden.WithMessage("このジョブへのアクセス権限がありません")
	}

	// キャンセル可能なステータスか確認
	switch job.Status {
	case model.AudioJobStatusPending:
		// pending → canceled に遷移
		job.Status = model.AudioJobStatusCanceled
		now := time.Now().UTC()
		job.CompletedAt = &now
		if err := s.audioJobRepo.Update(ctx, job); err != nil {
			return err
		}
		s.notifyCanceled(job.ID.String(), job.UserID.String())
		log.Info("audio job canceled (was pending)", "job_id", jobID)
		return nil

	case model.AudioJobStatusProcessing:
		// processing → canceling に遷移
		job.Status = model.AudioJobStatusCanceling
		if err := s.audioJobRepo.Update(ctx, job); err != nil {
			return err
		}
		s.notifyCanceling(job.ID.String(), job.UserID.String())
		log.Info("audio job canceling (was processing)", "job_id", jobID)
		return nil

	case model.AudioJobStatusCanceling:
		// 既にキャンセル中
		return apperror.ErrValidation.WithMessage("このジョブは既にキャンセル中です")

	case model.AudioJobStatusCanceled:
		// 既にキャンセル済み
		return apperror.ErrValidation.WithMessage("このジョブは既にキャンセルされています")

	case model.AudioJobStatusCompleted, model.AudioJobStatusFailed:
		// 完了または失敗済みのジョブはキャンセル不可
		return apperror.ErrValidation.WithMessage("完了または失敗したジョブはキャンセルできません")

	default:
		return apperror.ErrInternal.WithMessage("不明なジョブステータスです")
	}
}

// notifyCanceling はジョブがキャンセル中になったことを WebSocket で通知する
func (s *audioJobService) notifyCanceling(jobID, userID string) {
	if s.wsHub == nil {
		return
	}
	s.wsHub.SendToUser(userID, websocket.Message{
		Type: "audio_canceling",
		Payload: map[string]interface{}{
			"jobId": jobID,
		},
	})
}

// notifyCanceled はジョブがキャンセルされたことを WebSocket で通知する
func (s *audioJobService) notifyCanceled(jobID, userID string) {
	if s.wsHub == nil {
		return
	}
	s.wsHub.SendToUser(userID, websocket.Message{
		Type: "audio_canceled",
		Payload: map[string]interface{}{
			"jobId": jobID,
		},
	})
}

// notifyFailed はジョブの失敗を WebSocket で通知する
func (s *audioJobService) notifyFailed(jobID, userID string, errorCode, errorMessage *string) {
	log := logger.Default()
	if s.wsHub == nil {
		log.Warn("skipping failure notification as WebSocket Hub is not configured", "job_id", jobID, "user_id", userID)
		return
	}
	code := ""
	msg := ""
	if errorCode != nil {
		code = *errorCode
	}
	if errorMessage != nil {
		msg = *errorMessage
	}
	log.Info("notifying failure via WebSocket", "job_id", jobID, "user_id", userID, "error_code", code, "error_message", msg)
	s.wsHub.SendToUser(userID, websocket.Message{
		Type: "audio_failed",
		Payload: map[string]interface{}{
			"jobId":        jobID,
			"errorCode":    code,
			"errorMessage": msg,
		},
	})
}

// executeRemixInternal は BGM リミックス処理を実行する
func (s *audioJobService) executeRemixInternal(ctx context.Context, job *model.AudioJob) error {
	log := logger.FromContext(ctx)

	// エピソードを取得
	episode, err := s.episodeRepo.FindByID(ctx, job.EpisodeID)
	if err != nil {
		return err
	}

	// voice_audio_id の存在確認
	if episode.VoiceAudioID == nil {
		return apperror.ErrValidation.WithMessage("ボイス音声がありません")
	}

	if episode.VoiceAudio == nil {
		return apperror.ErrInternal.WithMessage("ボイス音声の読み込みに失敗しました")
	}

	// 進捗: 10%
	s.updateProgress(ctx, job, 10, "ボイス音声をダウンロード中...")

	// ボイス音声を GCS からダウンロード
	voiceAudioData, err := s.downloadFromStorage(ctx, episode.VoiceAudio.Path)
	if err != nil {
		log.Error("failed to download voice audio", "error", err, "path", episode.VoiceAudio.Path)
		return apperror.ErrInternal.WithMessage("ボイス音声のダウンロードに失敗しました").WithError(err)
	}

	// ボイス音声の長さを取得
	voiceDurationMs, err := audio.GetDurationMsE(voiceAudioData)
	if err != nil {
		log.Error("failed to get voice audio duration", "error", err)
		return apperror.ErrInternal.WithMessage("音声長の取得に失敗しました").WithError(err)
	}

	var finalAudio []byte

	if job.BgmID == nil && job.SystemBgmID == nil {
		// BGM なし: ボイス音声をそのまま使用
		s.updateProgress(ctx, job, 50, "音声を処理中...")
		finalAudio = voiceAudioData
	} else {
		// BGM あり: ダウンロードしてミキシング
		// 進捗: 30%
		s.updateProgress(ctx, job, 30, "BGM をダウンロード中...")

		// キャンセルチェック
		if err := s.checkCanceled(ctx, job); err != nil {
			return err
		}

		// BGM データを取得（ジョブに紐づく BGM を使用）
		var bgmPath string
		if job.BgmID != nil {
			bgm, err := s.bgmRepo.FindByID(ctx, *job.BgmID)
			if err != nil {
				return err
			}
			bgmPath = bgm.Audio.Path
		} else if job.SystemBgmID != nil {
			systemBgm, err := s.systemBgmRepo.FindByID(ctx, *job.SystemBgmID)
			if err != nil {
				return err
			}
			bgmPath = systemBgm.Audio.Path
		}

		// GCS から BGM をダウンロード
		bgmData, err := s.downloadFromStorage(ctx, bgmPath)
		if err != nil {
			log.Error("failed to download BGM", "error", err, "path", bgmPath)
			return apperror.ErrInternal.WithMessage("BGM のダウンロードに失敗しました").WithError(err)
		}

		// 進捗: 50%
		s.updateProgress(ctx, job, 50, "BGM をミキシング中...")

		// キャンセルチェック
		if err := s.checkCanceled(ctx, job); err != nil {
			return err
		}

		// FFmpeg でミキシング
		finalAudio, err = s.ffmpegService.MixAudioWithBGM(ctx, MixParams{
			VoiceData:       voiceAudioData,
			BGMData:         bgmData,
			VoiceDurationMs: voiceDurationMs,
			BGMVolumeDB:     job.BgmVolumeDB,
			FadeOutMs:       job.FadeOutMs,
			PaddingStartMs:  job.PaddingStartMs,
			PaddingEndMs:    job.PaddingEndMs,
		})
		if err != nil {
			log.Error("FFmpeg mixing failed", "error", err)
			return apperror.ErrInternal.WithMessage("BGM のミキシングに失敗しました").WithError(err)
		}
	}

	// 進捗: 85%
	s.updateProgress(ctx, job, 85, "音声をアップロード中...")

	// キャンセルチェック
	if err := s.checkCanceled(ctx, job); err != nil {
		return err
	}

	// 新しい Audio ID を生成してアップロード
	audioID := uuid.New()
	audioPath := storage.GenerateAudioPath(audioID.String())

	if _, err := s.storageClient.Upload(ctx, finalAudio, audioPath, "audio/mpeg"); err != nil {
		log.Error("failed to upload audio", "error", err)
		return apperror.ErrInternal.WithMessage("音声のアップロードに失敗しました").WithError(err)
	}

	// 最終的な長さを取得
	finalDurationMs, err := audio.GetDurationMsE(finalAudio)
	if err != nil {
		log.Warn("failed to get final audio duration, using estimate", "error", err)
		finalDurationMs = job.PaddingStartMs + voiceDurationMs + job.PaddingEndMs
	}

	// Audio レコードを作成
	audioRecord := &model.Audio{
		ID:         audioID,
		MimeType:   "audio/mpeg",
		Path:       audioPath,
		Filename:   audioID.String() + ".mp3",
		FileSize:   len(finalAudio),
		DurationMs: finalDurationMs,
	}

	if err := s.audioRepo.Create(ctx, audioRecord); err != nil {
		log.Error("failed to create audio record", "error", err)
		return apperror.ErrInternal.WithMessage("音声レコードの保存に失敗しました").WithError(err)
	}

	// 進捗: 95%
	s.updateProgress(ctx, job, 95, "エピソードを更新中...")

	// エピソードを更新
	episode.FullAudioID = &audioID
	episode.FullAudio = nil
	// BGM 情報をエピソードに記録
	episode.BgmID = job.BgmID
	episode.SystemBgmID = job.SystemBgmID
	episode.Bgm = nil
	episode.SystemBgm = nil

	if err := s.episodeRepo.Update(ctx, episode); err != nil {
		return err
	}

	// キャンセルチェック（完了遷移前）
	if err := s.checkCanceled(ctx, job); err != nil {
		return err
	}

	// ジョブを完了状態に更新
	completedAt := time.Now().UTC()
	job.Status = model.AudioJobStatusCompleted
	job.Progress = 100
	job.CompletedAt = &completedAt
	job.ResultAudioID = &audioID

	if err := s.audioJobRepo.Update(ctx, job); err != nil {
		return err
	}

	// WebSocket で完了通知
	s.notifyCompleted(job.ID.String(), job.UserID.String(), audioRecord)

	log.Info("remix job completed successfully", "job_id", job.ID, "audio_id", audioID)

	return nil
}

// toAudioJobResponse は AudioJob をレスポンス DTO に変換する
func (s *audioJobService) toAudioJobResponse(ctx context.Context, job *model.AudioJob) (*response.AudioJobResponse, error) {
	resp := &response.AudioJobResponse{
		ID:             job.ID,
		EpisodeID:      job.EpisodeID,
		Status:         string(job.Status),
		JobType:        string(job.JobType),
		Progress:       job.Progress,
		BgmVolumeDB:    job.BgmVolumeDB,
		FadeOutMs:      job.FadeOutMs,
		PaddingStartMs: job.PaddingStartMs,
		PaddingEndMs:   job.PaddingEndMs,
		ErrorMessage:   job.ErrorMessage,
		ErrorCode:      job.ErrorCode,
		StartedAt:      job.StartedAt,
		CompletedAt:    job.CompletedAt,
		CreatedAt:      job.CreatedAt,
		UpdatedAt:      job.UpdatedAt,
	}

	// Episode 情報
	if job.Episode.ID != uuid.Nil {
		resp.Episode = &response.AudioJobEpisodeResponse{
			ID:    job.Episode.ID,
			Title: job.Episode.Title,
		}
		if job.Episode.Channel.ID != uuid.Nil {
			resp.Episode.Channel = &response.AudioJobChannelResponse{
				ID:   job.Episode.Channel.ID,
				Name: job.Episode.Channel.Name,
			}
		}
	}

	// ResultAudio の署名付き URL を生成
	if job.ResultAudio != nil {
		signedURL, err := s.storageClient.GenerateSignedURL(ctx, job.ResultAudio.Path, storage.SignedURLExpirationAudio)
		if err != nil {
			return nil, err
		}
		resp.ResultAudio = &response.AudioResponse{
			ID:         job.ResultAudio.ID,
			URL:        signedURL,
			MimeType:   job.ResultAudio.MimeType,
			FileSize:   job.ResultAudio.FileSize,
			DurationMs: job.ResultAudio.DurationMs,
		}
	}

	return resp, nil
}
