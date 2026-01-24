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
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/infrastructure/tts"
	"github.com/siropaca/anycast-backend/internal/infrastructure/websocket"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/audio"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

const (
	// TTS API のテキスト入力制限（バイト）
	// Google Cloud TTS の制限は 4000 バイトだが、安全マージンを取る
	maxTTSInputBytes = 3500

	// TTS API で各ターンに追加されるポーズ文字列のバイト数
	// " [medium pause]" = 15 バイト
	turnPauseOverheadBytes = 15
)

// AudioJobService は非同期音声生成ジョブを管理するインターフェースを表す
type AudioJobService interface {
	CreateJob(ctx context.Context, userID, channelID, episodeID string, req request.GenerateAudioAsyncRequest) (*response.AudioJobResponse, error)
	GetJob(ctx context.Context, userID, jobID string) (*response.AudioJobResponse, error)
	ListMyJobs(ctx context.Context, userID string, filter repository.AudioJobFilter) (*response.AudioJobListResponse, error)
	ExecuteJob(ctx context.Context, jobID string) error
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

	// 台本行の存在確認
	scriptLines, err := s.scriptLineRepo.FindByEpisodeIDWithVoice(ctx, eid)
	if err != nil {
		return nil, err
	}

	if len(scriptLines) == 0 {
		return nil, apperror.ErrValidation.WithMessage("このエピソードには台本行がありません")
	}

	// デフォルト値を設定
	bgmVolumeDB := -15.0
	if req.BgmVolumeDB != nil {
		bgmVolumeDB = *req.BgmVolumeDB
	}

	fadeOutMs := 3000
	if req.FadeOutMs != nil {
		fadeOutMs = *req.FadeOutMs
	}

	paddingStartMs := 500
	if req.PaddingStartMs != nil {
		paddingStartMs = *req.PaddingStartMs
	}

	paddingEndMs := 1000
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
		Progress:       0,
		VoiceStyle:     voiceStyle,
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
			log.Error("ジョブのエンキューに失敗しました", "error", err, "job_id", job.ID)
			// エンキュー失敗時はジョブを失敗状態に更新（ベストエフォート）
			job.Status = model.AudioJobStatusFailed
			errMsg := "タスクのエンキューに失敗しました"
			errCode := "ENQUEUE_FAILED"
			job.ErrorMessage = &errMsg
			job.ErrorCode = &errCode
			_ = s.audioJobRepo.Update(ctx, job) //nolint:errcheck // best effort cleanup
			return nil, apperror.ErrInternal.WithMessage("音声生成タスクの登録に失敗しました").WithError(err)
		}
		log.Info("音声ジョブを作成しエンキューしました", "job_id", job.ID, "episode_id", eid)
	} else {
		// ローカル開発モード: goroutine で直接実行
		log.Info("Cloud Tasks 未設定のためジョブを直接実行します", "job_id", job.ID, "episode_id", eid)
		go func() {
			if err := s.ExecuteJob(context.Background(), job.ID.String()); err != nil {
				log.Error("ローカルジョブの実行に失敗しました", "error", err, "job_id", job.ID)
			}
		}()
	}

	return &response.AudioJobResponse{
		ID:       job.ID,
		Status:   string(job.Status),
		Progress: job.Progress,
	}, nil
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

	// 既に完了または失敗している場合はスキップ
	if job.Status == model.AudioJobStatusCompleted || job.Status == model.AudioJobStatusFailed {
		log.Info("ジョブは既に完了しているためスキップします", "job_id", jobID, "status", job.Status)
		return nil
	}

	// 処理開始
	now := time.Now()
	job.Status = model.AudioJobStatusProcessing
	job.StartedAt = &now
	if err := s.audioJobRepo.Update(ctx, job); err != nil {
		return err
	}

	// WebSocket で開始通知
	s.notifyProgress(job.ID.String(), job.UserID.String(), 0, "音声生成を開始しています...")

	// 処理実行（エラー時はジョブを失敗状態に）
	if err := s.executeJobInternal(ctx, job); err != nil {
		log.Error("ジョブの実行に失敗しました", "error", err, "job_id", jobID)
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

	// TTS で音声を生成（チャンク分割対応）
	var voiceStyle *string
	if job.VoiceStyle != "" {
		voiceStyle = &job.VoiceStyle
	}

	// ターンをチャンクに分割
	chunks := splitTurnsIntoChunks(turns, maxTTSInputBytes)
	log.Info("ターンをチャンクに分割しました", "total_turns", len(turns), "chunks", len(chunks))

	// チャンク分割の詳細をログ出力
	if len(chunks) > 1 {
		turnIndex := 0
		for i, chunk := range chunks {
			chunkBytes := 0
			for _, turn := range chunk {
				chunkBytes += len(turn.Text)
			}
			firstText := truncateText(chunk[0].Text, 30)
			lastText := truncateText(chunk[len(chunk)-1].Text, 30)
			log.Info("チャンク詳細",
				"chunk", i+1,
				"turns", fmt.Sprintf("%d-%d", turnIndex+1, turnIndex+len(chunk)),
				"turn_count", len(chunk),
				"bytes", chunkBytes,
				"first", firstText,
				"last", lastText,
			)
			turnIndex += len(chunk)
		}
	}

	var voiceAudio []byte
	if len(chunks) == 1 {
		// 単一チャンクの場合はそのまま TTS を呼び出し
		voiceAudio, err = s.ttsClient.SynthesizeMultiSpeaker(ctx, chunks[0], voiceConfigs, voiceStyle)
		if err != nil {
			log.Error("TTS が失敗しました", "error", err)
			return apperror.ErrGenerationFailed.WithMessage("音声の生成に失敗しました").WithError(err)
		}
	} else {
		// 複数チャンクの場合は順次 TTS を呼び出して結合
		audioChunks := make([][]byte, 0, len(chunks))
		for i, chunk := range chunks {
			progress := 20 + (i * 25 / len(chunks))
			s.updateProgress(ctx, job, progress, fmt.Sprintf("音声を生成中... (%d/%d)", i+1, len(chunks)))

			log.Info("チャンクを合成中", "chunk", i+1, "total", len(chunks), "turns", len(chunk))
			chunkAudio, err := s.ttsClient.SynthesizeMultiSpeaker(ctx, chunk, voiceConfigs, voiceStyle)
			if err != nil {
				log.Error("チャンクの TTS が失敗しました", "error", err, "chunk", i+1)
				return apperror.ErrGenerationFailed.WithMessage(
					fmt.Sprintf("音声の生成に失敗しました (チャンク %d/%d)", i+1, len(chunks)),
				).WithError(err)
			}
			audioChunks = append(audioChunks, chunkAudio)
		}

		// FFmpeg で音声を結合
		s.updateProgress(ctx, job, 45, "音声を結合中...")
		voiceAudio, err = s.ffmpegService.ConcatAudio(ctx, audioChunks)
		if err != nil {
			log.Error("音声結合が失敗しました", "error", err)
			return apperror.ErrInternal.WithMessage("音声の結合に失敗しました").WithError(err)
		}
	}

	// 進捗: 50%
	s.updateProgress(ctx, job, 50, "音声生成完了")

	// 最終的な音声データ
	var finalAudio []byte
	voiceDurationMs := 0

	// BGM がある場合はミキシング
	if episode.BgmID != nil || episode.SystemBgmID != nil {
		s.updateProgress(ctx, job, 55, "BGM をミキシング中...")

		// BGM データを取得
		var bgmPath string
		if episode.BgmID != nil {
			bgm, err := s.bgmRepo.FindByID(ctx, *episode.BgmID)
			if err != nil {
				return err
			}
			bgmPath = bgm.Audio.Path
		} else if episode.SystemBgmID != nil {
			systemBgm, err := s.systemBgmRepo.FindByID(ctx, *episode.SystemBgmID)
			if err != nil {
				return err
			}
			bgmPath = systemBgm.Audio.Path
		}

		// GCS から BGM をダウンロード
		bgmData, err := s.downloadFromStorage(ctx, bgmPath)
		if err != nil {
			log.Error("BGM のダウンロードに失敗しました", "error", err, "path", bgmPath)
			return apperror.ErrInternal.WithMessage("BGM のダウンロードに失敗しました").WithError(err)
		}

		// 音声の長さを取得
		voiceDurationMs, err = audio.GetDurationMsE(voiceAudio)
		if err != nil {
			log.Error("音声長の取得に失敗しました", "error", err)
			return apperror.ErrInternal.WithMessage("音声長の取得に失敗しました").WithError(err)
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
			log.Error("FFmpeg ミキシングが失敗しました", "error", err)
			return apperror.ErrInternal.WithMessage("BGM のミキシングに失敗しました").WithError(err)
		}
	} else {
		// BGM なしの場合はそのまま
		finalAudio = voiceAudio
	}

	// 進捗: 85%
	s.updateProgress(ctx, job, 85, "音声をアップロード中...")

	// 新しい Audio ID を生成してアップロード
	audioID := uuid.New()
	audioPath := storage.GenerateAudioPath(audioID.String())

	if _, err := s.storageClient.Upload(ctx, finalAudio, audioPath, "audio/mpeg"); err != nil {
		log.Error("音声のアップロードに失敗しました", "error", err)
		return apperror.ErrInternal.WithMessage("音声のアップロードに失敗しました").WithError(err)
	}

	// 最終的な長さを取得（ミキシングした場合は変わる可能性がある）
	finalDurationMs, err := audio.GetDurationMsE(finalAudio)
	if err != nil {
		log.Warn("最終音声長の取得に失敗、推定値を使用", "error", err)
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
		log.Error("音声レコードの作成に失敗しました", "error", err)
		return apperror.ErrInternal.WithMessage("音声レコードの保存に失敗しました").WithError(err)
	}

	// 進捗: 95%
	s.updateProgress(ctx, job, 95, "エピソードを更新中...")

	// エピソードを更新
	episode.FullAudioID = &audioID
	episode.FullAudio = nil
	episode.AudioOutdated = false
	if job.VoiceStyle != "" {
		episode.VoiceStyle = job.VoiceStyle
	}

	if err := s.episodeRepo.Update(ctx, episode); err != nil {
		return err
	}

	// ジョブを完了状態に更新
	completedAt := time.Now()
	job.Status = model.AudioJobStatusCompleted
	job.Progress = 100
	job.CompletedAt = &completedAt
	job.ResultAudioID = &audioID

	if err := s.audioJobRepo.Update(ctx, job); err != nil {
		return err
	}

	// WebSocket で完了通知
	s.notifyCompleted(job.ID.String(), job.UserID.String(), audioRecord)

	log.Info("音声ジョブが正常に完了しました", "job_id", job.ID, "audio_id", audioID)

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
func (s *audioJobService) updateProgress(ctx context.Context, job *model.AudioJob, progress int, message string) {
	job.Progress = progress
	_ = s.audioJobRepo.Update(ctx, job) //nolint:errcheck // progress update is best effort
	s.notifyProgress(job.ID.String(), job.UserID.String(), progress, message)
}

// failJob は指定されたジョブを失敗状態に更新する
func (s *audioJobService) failJob(ctx context.Context, job *model.AudioJob, err error) {
	completedAt := time.Now()
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
}

// notifyProgress はジョブの進捗を WebSocket で通知する
func (s *audioJobService) notifyProgress(jobID, userID string, progress int, message string) {
	if s.wsHub == nil {
		return
	}
	s.wsHub.SendToUser(userID, websocket.Message{
		Type: "progress",
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
		log.Warn("WebSocket Hub が未設定のため完了通知をスキップしました", "job_id", jobID, "user_id", userID)
		return
	}
	log.Info("WebSocket で完了を通知します", "job_id", jobID, "user_id", userID, "audio_id", audioModel.ID.String())
	s.wsHub.SendToUser(userID, websocket.Message{
		Type: "completed",
		Payload: map[string]interface{}{
			"jobId": jobID,
			"audio": map[string]interface{}{
				"id":         audioModel.ID.String(),
				"durationMs": audioModel.DurationMs,
			},
		},
	})
}

// notifyFailed はジョブの失敗を WebSocket で通知する
func (s *audioJobService) notifyFailed(jobID, userID string, errorCode, errorMessage *string) {
	log := logger.Default()
	if s.wsHub == nil {
		log.Warn("WebSocket Hub が未設定のため失敗通知をスキップしました", "job_id", jobID, "user_id", userID)
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
	log.Info("WebSocket で失敗を通知します", "job_id", jobID, "user_id", userID, "error_code", code, "error_message", msg)
	s.wsHub.SendToUser(userID, websocket.Message{
		Type: "failed",
		Payload: map[string]interface{}{
			"jobId":        jobID,
			"errorCode":    code,
			"errorMessage": msg,
		},
	})
}

// truncateText はテキストを指定した長さで切り詰める（日本語対応）
// maxRunes で最大文字数（runeベース）を指定し、超過時は "..." を付加する。
func truncateText(text string, maxRunes int) string {
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	return string(runes[:maxRunes]) + "..."
}

// splitTurnsIntoChunks はターンをバイトサイズ制限に基づいてチャンクに分割する
// maxBytes で1チャンクの最大バイト数を指定する。
func splitTurnsIntoChunks(turns []tts.SpeakerTurn, maxBytes int) [][]tts.SpeakerTurn {
	if len(turns) == 0 {
		return nil
	}

	var chunks [][]tts.SpeakerTurn
	var currentChunk []tts.SpeakerTurn
	currentBytes := 0

	for _, turn := range turns {
		// ターンのバイト数を計算
		// - emotion がある場合は [emotion] も含む
		// - TTS API で追加されるポーズ " [medium pause]" を含む
		turnBytes := len(turn.Text) + turnPauseOverheadBytes
		if turn.Emotion != nil && *turn.Emotion != "" {
			turnBytes += len(*turn.Emotion) + 3 // "[]" と空白
		}

		// 現在のチャンクに追加すると制限を超える場合は新しいチャンクを開始
		if len(currentChunk) > 0 && currentBytes+turnBytes > maxBytes {
			chunks = append(chunks, currentChunk)
			currentChunk = nil
			currentBytes = 0
		}

		currentChunk = append(currentChunk, turn)
		currentBytes += turnBytes
	}

	// 最後のチャンクを追加
	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

// toAudioJobResponse は AudioJob をレスポンス DTO に変換する
func (s *audioJobService) toAudioJobResponse(ctx context.Context, job *model.AudioJob) (*response.AudioJobResponse, error) {
	resp := &response.AudioJobResponse{
		ID:             job.ID,
		EpisodeID:      job.EpisodeID,
		Status:         string(job.Status),
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
