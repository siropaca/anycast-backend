package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/cloudtasks"
	"github.com/siropaca/anycast-backend/internal/infrastructure/llm"
	"github.com/siropaca/anycast-backend/internal/infrastructure/websocket"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/script"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// ScriptJobService は非同期台本生成ジョブを管理するインターフェースを表す
type ScriptJobService interface {
	CreateJob(ctx context.Context, userID, channelID, episodeID string, req request.GenerateScriptAsyncRequest) (*response.ScriptJobResponse, error)
	GetJob(ctx context.Context, userID, jobID string) (*response.ScriptJobResponse, error)
	ListMyJobs(ctx context.Context, userID string, filter repository.ScriptJobFilter) (*response.ScriptJobListResponse, error)
	ExecuteJob(ctx context.Context, jobID string) error
}

type scriptJobService struct {
	db             *gorm.DB
	scriptJobRepo  repository.ScriptJobRepository
	userRepo       repository.UserRepository
	channelRepo    repository.ChannelRepository
	episodeRepo    repository.EpisodeRepository
	scriptLineRepo repository.ScriptLineRepository
	llmClient      llm.Client
	tasksClient    cloudtasks.Client
	wsHub          *websocket.Hub
}

// NewScriptJobService は scriptJobService を生成して ScriptJobService として返す
func NewScriptJobService(
	db *gorm.DB,
	scriptJobRepo repository.ScriptJobRepository,
	userRepo repository.UserRepository,
	channelRepo repository.ChannelRepository,
	episodeRepo repository.EpisodeRepository,
	scriptLineRepo repository.ScriptLineRepository,
	llmClient llm.Client,
	tasksClient cloudtasks.Client,
	wsHub *websocket.Hub,
) ScriptJobService {
	return &scriptJobService{
		db:             db,
		scriptJobRepo:  scriptJobRepo,
		userRepo:       userRepo,
		channelRepo:    channelRepo,
		episodeRepo:    episodeRepo,
		scriptLineRepo: scriptLineRepo,
		llmClient:      llmClient,
		tasksClient:    tasksClient,
		wsHub:          wsHub,
	}
}

// CreateJob は非同期台本生成ジョブを作成して返す
func (s *scriptJobService) CreateJob(ctx context.Context, userID, channelID, episodeID string, req request.GenerateScriptAsyncRequest) (*response.ScriptJobResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("このエピソードの台本生成権限がありません")
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
	pendingJob, err := s.scriptJobRepo.FindPendingByEpisodeID(ctx, eid)
	if err != nil {
		return nil, err
	}
	if pendingJob != nil {
		return nil, apperror.ErrValidation.WithMessage("このエピソードは既に台本生成中です")
	}

	// チャンネルにキャラクターが紐づいていることを確認
	if len(channel.ChannelCharacters) == 0 {
		return nil, apperror.ErrValidation.WithMessage("このチャンネルにはキャラクターが設定されていません")
	}

	// デフォルト値を設定
	durationMinutes := defaultDurationMinutes
	if req.DurationMinutes != nil {
		durationMinutes = *req.DurationMinutes
	}

	// ジョブを作成
	job := &model.ScriptJob{
		EpisodeID:       eid,
		UserID:          uid,
		Status:          model.ScriptJobStatusPending,
		Progress:        0,
		Prompt:          req.Prompt,
		DurationMinutes: durationMinutes,
		WithEmotion:     req.WithEmotion,
	}

	if err := s.scriptJobRepo.Create(ctx, job); err != nil {
		return nil, err
	}

	// Cloud Tasks が設定されている場合はエンキュー、そうでなければ goroutine で直接実行
	if s.tasksClient != nil {
		if err := s.tasksClient.EnqueueScriptJob(ctx, job.ID.String()); err != nil {
			log.Error("failed to enqueue script job", "error", err, "job_id", job.ID)
			// エンキュー失敗時はジョブを失敗状態に更新（ベストエフォート）
			job.Status = model.ScriptJobStatusFailed
			errMsg := "タスクのエンキューに失敗しました"
			errCode := "ENQUEUE_FAILED"
			job.ErrorMessage = &errMsg
			job.ErrorCode = &errCode
			_ = s.scriptJobRepo.Update(ctx, job) //nolint:errcheck // best effort cleanup
			return nil, apperror.ErrInternal.WithMessage("台本生成タスクの登録に失敗しました").WithError(err)
		}
		log.Info("script job created and enqueued", "job_id", job.ID, "episode_id", eid)
	} else {
		// ローカル開発モード: goroutine で直接実行
		log.Info("cloud tasks not configured, executing script job directly", "job_id", job.ID, "episode_id", eid)
		go func() {
			if err := s.ExecuteJob(context.Background(), job.ID.String()); err != nil {
				log.Error("local script job execution failed", "error", err, "job_id", job.ID)
			}
		}()
	}

	return &response.ScriptJobResponse{
		ID:       job.ID,
		Status:   string(job.Status),
		Progress: job.Progress,
	}, nil
}

// GetJob は指定されたジョブの詳細を取得する
func (s *scriptJobService) GetJob(ctx context.Context, userID, jobID string) (*response.ScriptJobResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	jid, err := uuid.Parse(jobID)
	if err != nil {
		return nil, err
	}

	job, err := s.scriptJobRepo.FindByID(ctx, jid)
	if err != nil {
		return nil, err
	}

	// オーナーチェック
	if job.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("このジョブへのアクセス権限がありません")
	}

	return s.toScriptJobResponse(ctx, job)
}

// ListMyJobs は指定されたユーザーのジョブ一覧を取得する
func (s *scriptJobService) ListMyJobs(ctx context.Context, userID string, filter repository.ScriptJobFilter) (*response.ScriptJobListResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	jobs, err := s.scriptJobRepo.FindByUserID(ctx, uid, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]response.ScriptJobResponse, len(jobs))
	for i, job := range jobs {
		resp, err := s.toScriptJobResponse(ctx, &job)
		if err != nil {
			return nil, err
		}
		responses[i] = *resp
	}

	return &response.ScriptJobListResponse{
		Data: responses,
	}, nil
}

// ExecuteJob は指定されたジョブを実行する（Cloud Tasks ワーカーから呼び出される）
func (s *scriptJobService) ExecuteJob(ctx context.Context, jobID string) error {
	log := logger.FromContext(ctx)

	jid, err := uuid.Parse(jobID)
	if err != nil {
		return err
	}

	// ジョブを取得
	job, err := s.scriptJobRepo.FindByID(ctx, jid)
	if err != nil {
		return err
	}

	// 既に完了または失敗している場合はスキップ
	if job.Status == model.ScriptJobStatusCompleted || job.Status == model.ScriptJobStatusFailed {
		log.Info("script job already finished, skipping", "job_id", jobID, "status", job.Status)
		return nil
	}

	// 処理開始
	now := time.Now()
	job.Status = model.ScriptJobStatusProcessing
	job.StartedAt = &now
	if err := s.scriptJobRepo.Update(ctx, job); err != nil {
		return err
	}

	// WebSocket で開始通知
	s.notifyProgress(job.ID.String(), job.UserID.String(), 0, "台本生成を開始しています...")

	// 処理実行（エラー時はジョブを失敗状態に）
	scriptLinesCount, err := s.executeJobInternal(ctx, job)
	if err != nil {
		log.Error("script job execution failed", "error", err, "job_id", jobID)
		s.failJob(ctx, job, err)
		return err
	}

	// WebSocket で完了通知
	s.notifyCompleted(job.ID.String(), job.UserID.String(), scriptLinesCount)

	log.Info("script job completed successfully", "job_id", job.ID, "lines_count", scriptLinesCount)

	return nil
}

// executeJobInternal は台本生成処理を実行する
func (s *scriptJobService) executeJobInternal(ctx context.Context, job *model.ScriptJob) (int, error) {
	log := logger.FromContext(ctx)

	// 進捗: 10%
	s.updateProgress(ctx, job, 10, "データを準備中...")

	// チャンネルを取得（キャラクター情報含む）
	channel, err := s.channelRepo.FindByID(ctx, job.Episode.ChannelID)
	if err != nil {
		return 0, err
	}

	// エピソードを取得
	episode, err := s.episodeRepo.FindByID(ctx, job.EpisodeID)
	if err != nil {
		return 0, err
	}

	// ユーザー情報を取得（userPrompt 用）
	user, err := s.userRepo.FindByID(ctx, job.UserID)
	if err != nil {
		return 0, err
	}

	// 進捗: 20%
	s.updateProgress(ctx, job, 20, "プロンプトを構築中...")

	// withEmotion に応じてシステムプロンプトを選択
	sysPrompt := systemPromptWithoutEmotion
	if job.WithEmotion {
		sysPrompt = systemPromptWithEmotion
	}

	// LLM 用ユーザープロンプトを構築
	userPrompt := s.buildUserPrompt(user, channel, episode, job.Prompt, job.DurationMinutes)

	// 進捗: 30%
	s.updateProgress(ctx, job, 30, "AI で台本を生成中...")

	// LLM で台本生成
	generatedText, err := s.llmClient.GenerateScript(ctx, sysPrompt, userPrompt)
	if err != nil {
		return 0, err
	}

	// 進捗: 70%
	s.updateProgress(ctx, job, 70, "台本をパース中...")

	// 許可された話者名のリストを作成
	allowedSpeakers := make([]string, len(channel.ChannelCharacters))
	speakerMap := make(map[string]*model.Character, len(channel.ChannelCharacters))
	for i, cc := range channel.ChannelCharacters {
		allowedSpeakers[i] = cc.Character.Name
		speakerMap[cc.Character.Name] = &channel.ChannelCharacters[i].Character
	}

	// 生成されたテキストをパース
	parseResult := script.Parse(generatedText, allowedSpeakers)

	// パースエラーがある場合（全行失敗の場合のみエラー）
	if len(parseResult.Lines) == 0 && parseResult.HasErrors() {
		return 0, apperror.ErrGenerationFailed.WithMessage("生成された台本のパースに失敗しました")
	}

	log.Info("script parsed", "lines", len(parseResult.Lines), "errors", len(parseResult.Errors))

	// 進捗: 85%
	s.updateProgress(ctx, job, 85, "台本を保存中...")

	// ScriptLine モデルに変換
	scriptLines := make([]model.ScriptLine, len(parseResult.Lines))
	for i, line := range parseResult.Lines {
		speaker := speakerMap[line.SpeakerName]
		scriptLines[i] = model.ScriptLine{
			EpisodeID: job.EpisodeID,
			LineOrder: i,
			SpeakerID: speaker.ID,
			Text:      line.Text,
			Emotion:   line.Emotion,
		}
	}

	// トランザクションで既存行削除・新規作成・エピソード更新を実行
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// トランザクション内で使うリポジトリを作成
		txScriptLineRepo := repository.NewScriptLineRepository(tx)
		txEpisodeRepo := repository.NewEpisodeRepository(tx)

		// 既存の台本行を削除
		if err := txScriptLineRepo.DeleteByEpisodeID(ctx, job.EpisodeID); err != nil {
			return err
		}

		// 新しい台本行を一括作成
		if _, err := txScriptLineRepo.CreateBatch(ctx, scriptLines); err != nil {
			return err
		}

		// episode.userPrompt を更新
		episode.UserPrompt = job.Prompt
		if err := txEpisodeRepo.Update(ctx, episode); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	// 進捗: 95%
	s.updateProgress(ctx, job, 95, "完了処理中...")

	// ジョブを完了状態に更新
	completedAt := time.Now()
	job.Status = model.ScriptJobStatusCompleted
	job.Progress = 100
	job.CompletedAt = &completedAt

	if err := s.scriptJobRepo.Update(ctx, job); err != nil {
		return 0, err
	}

	return len(scriptLines), nil
}

// buildUserPrompt は LLM 用のユーザープロンプトを構築する
// プロンプトは User → Channel → Episode の順で結合（追記）される
func (s *scriptJobService) buildUserPrompt(user *model.User, channel *model.Channel, episode *model.Episode, prompt string, durationMinutes int) string {
	var sb strings.Builder

	// ユーザー設定（ユーザーレベルのプロンプト）
	if user.UserPrompt != "" {
		sb.WriteString("## ユーザー設定\n")
		sb.WriteString(user.UserPrompt)
		sb.WriteString("\n\n")
	}

	// チャンネル情報
	sb.WriteString("## チャンネル情報\n")
	sb.WriteString(fmt.Sprintf("チャンネル名: %s\n", channel.Name))
	if channel.Description != "" {
		sb.WriteString(fmt.Sprintf("説明: %s\n", channel.Description))
	}
	sb.WriteString(fmt.Sprintf("カテゴリー: %s\n", channel.Category.Name))
	sb.WriteString("\n")

	// チャンネル設定
	if channel.UserPrompt != "" {
		sb.WriteString("## チャンネル設定\n")
		sb.WriteString(channel.UserPrompt)
		sb.WriteString("\n\n")
	}

	// 登場人物
	sb.WriteString("## 登場人物\n")
	for _, cc := range channel.ChannelCharacters {
		if cc.Character.Persona != "" {
			sb.WriteString(fmt.Sprintf("- %s（%s）: %s\n", cc.Character.Name, cc.Character.Voice.Gender, cc.Character.Persona))
		} else {
			sb.WriteString(fmt.Sprintf("- %s（%s）\n", cc.Character.Name, cc.Character.Voice.Gender))
		}
	}
	sb.WriteString("\n")

	// エピソード情報
	sb.WriteString("## エピソード情報\n")
	sb.WriteString(fmt.Sprintf("タイトル: %s\n", episode.Title))
	if episode.Description != "" {
		sb.WriteString(fmt.Sprintf("説明: %s\n", episode.Description))
	}
	sb.WriteString("\n")

	// エピソードの長さ
	sb.WriteString(fmt.Sprintf("## エピソードの長さ\n%d分\n\n", durationMinutes))

	// 今回のテーマ
	sb.WriteString("## 今回のテーマ\n")
	sb.WriteString(prompt)

	return sb.String()
}

// updateProgress はジョブの進捗を更新し WebSocket で通知する
func (s *scriptJobService) updateProgress(ctx context.Context, job *model.ScriptJob, progress int, message string) {
	job.Progress = progress
	_ = s.scriptJobRepo.Update(ctx, job) //nolint:errcheck // progress update is best effort
	s.notifyProgress(job.ID.String(), job.UserID.String(), progress, message)
}

// failJob は指定されたジョブを失敗状態に更新する
func (s *scriptJobService) failJob(ctx context.Context, job *model.ScriptJob, err error) {
	completedAt := time.Now()
	job.Status = model.ScriptJobStatusFailed
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

	_ = s.scriptJobRepo.Update(ctx, job) //nolint:errcheck // fail update is best effort
	s.notifyFailed(job.ID.String(), job.UserID.String(), job.ErrorCode, job.ErrorMessage)
}

// notifyProgress はジョブの進捗を WebSocket で通知する
func (s *scriptJobService) notifyProgress(jobID, userID string, progress int, message string) {
	if s.wsHub == nil {
		return
	}
	s.wsHub.SendToUser(userID, websocket.Message{
		Type: "script_progress",
		Payload: map[string]interface{}{
			"jobId":    jobID,
			"progress": progress,
			"message":  message,
		},
	})
}

// notifyCompleted はジョブの完了を WebSocket で通知する
func (s *scriptJobService) notifyCompleted(jobID, userID string, scriptLinesCount int) {
	if s.wsHub == nil {
		return
	}
	s.wsHub.SendToUser(userID, websocket.Message{
		Type: "script_completed",
		Payload: map[string]interface{}{
			"jobId":            jobID,
			"scriptLinesCount": scriptLinesCount,
		},
	})
}

// notifyFailed はジョブの失敗を WebSocket で通知する
func (s *scriptJobService) notifyFailed(jobID, userID string, errorCode, errorMessage *string) {
	if s.wsHub == nil {
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
	s.wsHub.SendToUser(userID, websocket.Message{
		Type: "script_failed",
		Payload: map[string]interface{}{
			"jobId":        jobID,
			"errorCode":    code,
			"errorMessage": msg,
		},
	})
}

// toScriptJobResponse は ScriptJob をレスポンス DTO に変換する
func (s *scriptJobService) toScriptJobResponse(ctx context.Context, job *model.ScriptJob) (*response.ScriptJobResponse, error) {
	resp := &response.ScriptJobResponse{
		ID:              job.ID,
		EpisodeID:       job.EpisodeID,
		Status:          string(job.Status),
		Progress:        job.Progress,
		Prompt:          job.Prompt,
		DurationMinutes: job.DurationMinutes,
		WithEmotion:     job.WithEmotion,
		ErrorMessage:    job.ErrorMessage,
		ErrorCode:       job.ErrorCode,
		StartedAt:       job.StartedAt,
		CompletedAt:     job.CompletedAt,
		CreatedAt:       job.CreatedAt,
		UpdatedAt:       job.UpdatedAt,
	}

	// Episode 情報
	if job.Episode.ID != uuid.Nil {
		resp.Episode = &response.ScriptJobEpisodeResponse{
			ID:    job.Episode.ID,
			Title: job.Episode.Title,
		}
		if job.Episode.Channel.ID != uuid.Nil {
			resp.Episode.Channel = &response.ScriptJobChannelResponse{
				ID:   job.Episode.Channel.ID,
				Name: job.Episode.Channel.Name,
			}
		}
	}

	// 完了時は台本行数を取得
	if job.Status == model.ScriptJobStatusCompleted {
		scriptLines, err := s.scriptLineRepo.FindByEpisodeID(ctx, job.EpisodeID)
		if err == nil {
			count := len(scriptLines)
			resp.ScriptLinesCount = &count
		}
	}

	return resp, nil
}
