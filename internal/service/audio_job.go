package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/cloudtasks"
	"github.com/siropaca/anycast-backend/internal/infrastructure/slack"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
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
	ttsClient      tts.Client
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
	ttsClient tts.Client,
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
		ttsClient:      ttsClient,
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

	voiceStyle := ""
	if req.VoiceStyle != nil {
		voiceStyle = *req.VoiceStyle
	}

	// ジョブを作成
	job := &model.AudioJob{
		EpisodeID:      eid,
		UserID:         uid,
		Status:         model.AudioJobStatusPending,
		JobType:        jobType,
		Progress:       0,
		VoiceStyle:     voiceStyle,
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

	// TTS で音声を生成
	var voiceStyle *string
	if job.VoiceStyle != "" {
		voiceStyle = &job.VoiceStyle
	}

	log.Info("generating audio", "total_turns", len(turns))

	// TTS で音声を生成
	result, err := s.ttsClient.SynthesizeMultiSpeaker(ctx, turns, voiceConfigs, voiceStyle)
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
	if job.VoiceStyle != "" {
		episode.VoiceStyle = job.VoiceStyle
	}

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
		VoiceStyle:     job.VoiceStyle,
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
