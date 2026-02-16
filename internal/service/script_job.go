package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/cloudtasks"
	"github.com/siropaca/anycast-backend/internal/infrastructure/llm"
	"github.com/siropaca/anycast-backend/internal/infrastructure/slack"
	"github.com/siropaca/anycast-backend/internal/infrastructure/websocket"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/script"
	"github.com/siropaca/anycast-backend/internal/pkg/tracer"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// ScriptJobService は非同期台本生成ジョブを管理するインターフェースを表す
type ScriptJobService interface {
	CreateJob(ctx context.Context, userID, channelID, episodeID string, req request.GenerateScriptAsyncRequest) (*response.ScriptJobResponse, error)
	GetJob(ctx context.Context, userID, jobID string) (*response.ScriptJobResponse, error)
	GetLatestJobByEpisode(ctx context.Context, userID, channelID, episodeID string) (*response.ScriptJobResponse, error)
	ListMyJobs(ctx context.Context, userID string, filter repository.ScriptJobFilter) (*response.ScriptJobListResponse, error)
	ExecuteJob(ctx context.Context, jobID string) error
	CancelJob(ctx context.Context, userID, jobID string) error
	GenerateScriptDirect(ctx context.Context, req request.GenerateScriptDirectRequest) (*response.GenerateScriptDirectResponse, error)
}

type scriptJobService struct {
	db             *gorm.DB
	scriptJobRepo  repository.ScriptJobRepository
	userRepo       repository.UserRepository
	channelRepo    repository.ChannelRepository
	episodeRepo    repository.EpisodeRepository
	scriptLineRepo repository.ScriptLineRepository
	llmRegistry    *llm.Registry
	tasksClient    cloudtasks.Client
	wsHub          *websocket.Hub
	traceMode      tracer.Mode
	slackClient    slack.Client
}

// NewScriptJobService は scriptJobService を生成して ScriptJobService として返す
func NewScriptJobService(
	db *gorm.DB,
	scriptJobRepo repository.ScriptJobRepository,
	userRepo repository.UserRepository,
	channelRepo repository.ChannelRepository,
	episodeRepo repository.EpisodeRepository,
	scriptLineRepo repository.ScriptLineRepository,
	llmRegistry *llm.Registry,
	tasksClient cloudtasks.Client,
	wsHub *websocket.Hub,
	traceMode tracer.Mode,
	slackClient slack.Client,
) ScriptJobService {
	return &scriptJobService{
		db:             db,
		scriptJobRepo:  scriptJobRepo,
		userRepo:       userRepo,
		channelRepo:    channelRepo,
		episodeRepo:    episodeRepo,
		scriptLineRepo: scriptLineRepo,
		llmRegistry:    llmRegistry,
		tasksClient:    tasksClient,
		wsHub:          wsHub,
		traceMode:      traceMode,
		slackClient:    slackClient,
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
		log.Info("executing script job directly as Cloud Tasks is not configured", "job_id", job.ID, "episode_id", eid)
		go func() {
			if err := s.ExecuteJob(context.Background(), job.ID.String()); err != nil {
				log.Error("failed to execute local script job", "error", err, "job_id", job.ID)
			}
		}()
	}

	return s.toScriptJobResponse(ctx, job)
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

// GetLatestJobByEpisode はエピソードの最新の完了済みジョブを取得する
func (s *scriptJobService) GetLatestJobByEpisode(ctx context.Context, userID, channelID, episodeID string) (*response.ScriptJobResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルへのアクセス権限がありません")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("このチャンネルにエピソードが見つかりません")
	}

	// 最新の完了済みジョブを取得
	job, err := s.scriptJobRepo.FindLatestCompletedByEpisodeID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if job == nil {
		return nil, nil //nolint:nilnil // 完了済みジョブがないのは正常な状態
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

	// 既に完了、失敗、またはキャンセル済みの場合はスキップ
	if job.Status == model.ScriptJobStatusCompleted ||
		job.Status == model.ScriptJobStatusFailed ||
		job.Status == model.ScriptJobStatusCanceled {
		log.Info("skipping script job as it is already completed", "job_id", jobID, "status", job.Status)
		return nil
	}

	// キャンセル中の場合はキャンセル完了にする
	if job.Status == model.ScriptJobStatusCanceling {
		log.Info("completing cancellation for script job", "job_id", jobID)
		job.Status = model.ScriptJobStatusCanceled
		now := time.Now().UTC()
		job.CompletedAt = &now
		if err := s.scriptJobRepo.Update(ctx, job); err != nil {
			return err
		}
		s.notifyCanceled(job.ID.String(), job.UserID.String())
		return nil
	}

	// 処理開始
	now := time.Now().UTC()
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
		// キャンセルによるエラーの場合は失敗扱いにしない
		if apperror.IsCode(err, apperror.CodeCanceled) {
			log.Info("script job was canceled during execution", "job_id", jobID)
			return nil
		}
		log.Error("failed to execute script job", "error", err, "job_id", jobID)
		s.failJob(ctx, job, err)
		return err
	}

	// WebSocket で完了通知
	s.notifyCompleted(job.ID.String(), job.UserID.String(), scriptLinesCount)

	log.Info("script job completed successfully", "job_id", job.ID, "lines_count", scriptLinesCount)

	return nil
}

// executeJobInternal は台本生成処理を多段階ワークフローで実行する
//
// Phase 1: ブリーフ正規化 → Phase 2: 素材+アウトライン → Phase 3: 台本ドラフト → Phase 4: リライト → Phase 5: QA+パッチ
func (s *scriptJobService) executeJobInternal(ctx context.Context, job *model.ScriptJob) (int, error) {
	log := logger.FromContext(ctx)

	// 進捗: 5% - データ取得
	s.updateProgress(ctx, job, 5, "データを準備中...")

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

	// ユーザー情報を取得
	user, err := s.userRepo.FindByID(ctx, job.UserID)
	if err != nil {
		return 0, err
	}

	// エピソード番号を算出（同チャンネル内で何話目か）
	countBefore, err := s.episodeRepo.CountByChannelIDBeforeCreatedAt(ctx, episode.ChannelID, episode.CreatedAt)
	if err != nil {
		return 0, fmt.Errorf("エピソード番号の算出に失敗: %w", err)
	}
	episodeNumber := int(countBefore) + 1

	// 許可された話者名のリストを作成
	allowedSpeakers := make([]string, len(channel.ChannelCharacters))
	speakerMap := make(map[string]*model.Character, len(channel.ChannelCharacters))
	for i, cc := range channel.ChannelCharacters {
		allowedSpeakers[i] = cc.Character.Name
		speakerMap[cc.Character.Name] = &channel.ChannelCharacters[i].Character
	}

	// ===== Phase 1: ブリーフ正規化 =====
	s.updateProgress(ctx, job, 10, "ブリーフを正規化中...")

	briefInput := script.BriefInput{
		EpisodeTitle:       episode.Title,
		EpisodeDescription: episode.Description,
		DurationMinutes:    job.DurationMinutes,
		EpisodeNumber:      episodeNumber,
		ChannelName:        channel.Name,
		ChannelDescription: channel.Description,
		ChannelCategory:    channel.Category.Name,
		ChannelStyleGuide:  channel.UserPrompt,
		MasterGuide:        user.UserPrompt,
		Theme:              job.Prompt,
		WithEmotion:        job.WithEmotion,
	}

	for _, cc := range channel.ChannelCharacters {
		briefInput.Characters = append(briefInput.Characters, script.BriefInputCharacter{
			Name:    cc.Character.Name,
			Gender:  string(cc.Character.Voice.Gender),
			Persona: cc.Character.Persona,
		})
	}

	brief := script.NormalizeBrief(briefInput)

	briefJSON, err := brief.ToJSON()
	if err != nil {
		return 0, fmt.Errorf("ブリーフの JSON 変換に失敗: %w", err)
	}

	// トレーサーを生成
	t := tracer.New(s.traceMode, episode.Title)

	log.Info("brief normalized", "talk_mode", brief.Constraints.TalkMode, "characters", len(brief.Characters))
	t.Trace("phase1", "brief", briefJSON)
	t.Flush("phase1")

	// ===== Phase 2: 素材+アウトライン生成 =====
	s.updateProgress(ctx, job, 15, "素材とアウトラインを生成中...")

	// キャンセルチェック（Phase 2 開始前）
	if err := s.checkCanceled(ctx, job); err != nil {
		return 0, err
	}

	phase2Output, err := s.executePhase2(ctx, briefJSON, t)
	if err != nil {
		return 0, err
	}

	s.updateProgress(ctx, job, 35, "素材とアウトライン生成完了...")

	// ===== Phase 3: 台本ドラフト生成 =====
	s.updateProgress(ctx, job, 40, "台本ドラフトを生成中...")

	// キャンセルチェック（Phase 3 開始前）
	if err := s.checkCanceled(ctx, job); err != nil {
		return 0, err
	}

	generatedText, err := s.executePhase3(ctx, brief, phase2Output, t)
	if err != nil {
		return 0, err
	}

	// 進捗: 65% - Phase 3 完了 + パース
	s.updateProgress(ctx, job, 65, "台本をパース中...")

	// 生成されたテキストをパース
	parseResult := script.Parse(generatedText, allowedSpeakers)

	// パースエラーがある場合（全行失敗の場合のみエラー）
	if len(parseResult.Lines) == 0 && parseResult.HasErrors() {
		return 0, apperror.ErrGenerationFailed.WithMessage("生成された台本のパースに失敗しました")
	}

	phase3CharCount := countTotalChars(parseResult.Lines)
	log.Info("script parsed", "lines", len(parseResult.Lines), "errors", len(parseResult.Errors),
		"total_chars", phase3CharCount, "target_chars", job.DurationMinutes*script.CharsPerMinute)

	// ===== Phase 4: リライト =====
	s.updateProgress(ctx, job, 68, "台本をリライト中...")

	// キャンセルチェック（Phase 4 開始前）
	if err := s.checkCanceled(ctx, job); err != nil {
		return 0, err
	}

	rewrittenText, err := s.executePhase4(ctx, generatedText, brief, t)
	if err != nil {
		log.Warn("Phase 4 rewrite failed, using original draft", "error", err)
		rewrittenText = generatedText
	} else {
		// リライト後のテキストを再パース
		rewriteResult := script.Parse(rewrittenText, allowedSpeakers)
		if len(rewriteResult.Lines) > 0 {
			parseResult = rewriteResult
			phase4CharCount := countTotalChars(parseResult.Lines)
			log.Info("Phase 4 rewrite applied", "lines", len(parseResult.Lines),
				"total_chars", phase4CharCount, "target_chars", job.DurationMinutes*script.CharsPerMinute)
		} else {
			log.Warn("Phase 4 rewrite parse failed, using original draft")
			rewrittenText = generatedText
		}
	}

	s.updateProgress(ctx, job, 78, "リライト完了...")

	// ===== Phase 5: QA 検証+パッチ修正 =====
	s.updateProgress(ctx, job, 80, "品質チェック中...")

	parsedLines := s.executePhase5(ctx, job, parseResult.Lines, brief, allowedSpeakers, rewrittenText, t)

	// 進捗: 90% - DB 保存
	s.updateProgress(ctx, job, 90, "台本を保存中...")

	// キャンセルチェック（保存前）
	if err := s.checkCanceled(ctx, job); err != nil {
		return 0, err
	}

	// ScriptLine モデルに変換
	scriptLines := make([]model.ScriptLine, len(parsedLines))
	for i, line := range parsedLines {
		speaker := speakerMap[line.SpeakerName]
		scriptLines[i] = model.ScriptLine{
			EpisodeID: job.EpisodeID,
			LineOrder: i,
			SpeakerID: speaker.ID,
			Text:      line.Text,
			Emotion:   line.Emotion,
		}
	}

	// トランザクションで既存行削除・新規作成を実行
	err = s.db.Transaction(func(tx *gorm.DB) error {
		txScriptLineRepo := repository.NewScriptLineRepository(tx)

		if err := txScriptLineRepo.DeleteByEpisodeID(ctx, job.EpisodeID); err != nil {
			return err
		}

		if _, err := txScriptLineRepo.CreateBatch(ctx, scriptLines); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	// 進捗: 95% - 完了処理
	s.updateProgress(ctx, job, 95, "完了処理中...")

	// キャンセルチェック（完了遷移前）
	if err := s.checkCanceled(ctx, job); err != nil {
		return 0, err
	}

	// ジョブを完了状態に更新
	completedAt := time.Now().UTC()
	job.Status = model.ScriptJobStatusCompleted
	job.Progress = 100
	job.CompletedAt = &completedAt

	if err := s.scriptJobRepo.Update(ctx, job); err != nil {
		return 0, err
	}

	return len(scriptLines), nil
}

// executePhase2 は Phase 2（素材+アウトライン生成）を実行する
//
// 最大2回リトライし、全失敗時はエラーを返す
func (s *scriptJobService) executePhase2(ctx context.Context, briefJSON string, t tracer.Tracer) (*script.Phase2Output, error) {
	log := logger.FromContext(ctx)

	client, err := s.llmRegistry.Get(phase2Config.Provider)
	if err != nil {
		return nil, fmt.Errorf("phase 2 LLM client: %w", err)
	}

	temp := phase2Config.Temperature
	opts := llm.ChatOptions{Temperature: &temp, EnableWebSearch: true}

	t.Trace("phase2", "model_info", s.llmRegistry.GetModelInfo(phase2Config.Provider))
	t.Trace("phase2", "system_prompt", phase2SystemPrompt)
	t.Trace("phase2", "user_prompt", briefJSON)

	var lastErr error
	for attempt := 1; attempt <= 2; attempt++ {
		log.Debug("executing Phase 2", "attempt", attempt, "provider", phase2Config.Provider)

		result, err := client.ChatWithOptions(ctx, phase2SystemPrompt, briefJSON, opts)
		if err != nil {
			log.Warn("Phase 2 LLM call failed", "attempt", attempt, "error", err)
			lastErr = err
			continue
		}

		t.Trace("phase2", "response", result)

		output, err := script.ParsePhase2Output(result)
		if err != nil {
			log.Warn("Phase 2 output parse failed", "attempt", attempt, "error", err, "response_length", len(result))
			lastErr = err
			continue
		}

		parsedJSON, _ := json.Marshal(output) //nolint:errcheck // trace data
		t.Trace("phase2", "parsed_output", string(parsedJSON))
		t.Flush("phase2")

		log.Info("Phase 2 completed", "attempt", attempt)
		return output, nil
	}

	t.Flush("phase2")
	log.Error("Phase 2 failed after retries", "error", lastErr)
	return nil, apperror.ErrGenerationFailed.WithMessage("素材とアウトラインの生成に失敗しました").WithError(lastErr)
}

// executePhase3 は Phase 3（台本ドラフト生成）を実行する
func (s *scriptJobService) executePhase3(ctx context.Context, brief script.Brief, phase2 *script.Phase2Output, t tracer.Tracer) (string, error) {
	log := logger.FromContext(ctx)

	client, err := s.llmRegistry.Get(phase3Config.Provider)
	if err != nil {
		return "", fmt.Errorf("phase 3 LLM client: %w", err)
	}

	sysPrompt := getPhase3SystemPrompt(brief.Constraints.TalkMode, brief.Constraints.WithEmotion, brief.Episode.DurationMinutes, brief.Episode.EpisodeNumber)
	userPrompt := buildPhase3UserPrompt(brief, phase2)

	t.Trace("phase3", "model_info", s.llmRegistry.GetModelInfo(phase3Config.Provider))
	t.Trace("phase3", "system_prompt", sysPrompt)
	t.Trace("phase3", "user_prompt", userPrompt)

	temp := phase3Config.Temperature
	opts := llm.ChatOptions{Temperature: &temp}

	result, err := client.ChatWithOptions(ctx, sysPrompt, userPrompt, opts)
	if err != nil {
		t.Flush("phase3")
		return "", err
	}

	t.Trace("phase3", "response", result)
	t.Flush("phase3")

	log.Info("Phase 3 completed")

	return result, nil
}

// executePhase4 は Phase 4（リライト）を実行する
//
// 台本ドラフトの会話の流れ・自然さ・面白さを改善する
func (s *scriptJobService) executePhase4(ctx context.Context, draftText string, brief script.Brief, t tracer.Tracer) (string, error) {
	log := logger.FromContext(ctx)

	client, err := s.llmRegistry.Get(phase4Config.Provider)
	if err != nil {
		return "", fmt.Errorf("phase 4 LLM client: %w", err)
	}

	briefJSON, err := brief.ToJSON()
	if err != nil {
		return "", fmt.Errorf("ブリーフの JSON 変換に失敗: %w", err)
	}

	userPrompt := "## ブリーフ\n" + briefJSON + "\n\n## ドラフト台本\n" + draftText

	t.Trace("phase4", "model_info", s.llmRegistry.GetModelInfo(phase4Config.Provider))
	t.Trace("phase4", "system_prompt", phase4SystemPrompt)
	t.Trace("phase4", "user_prompt", userPrompt)

	temp := phase4Config.Temperature
	opts := llm.ChatOptions{Temperature: &temp}

	result, err := client.ChatWithOptions(ctx, phase4SystemPrompt, userPrompt, opts)
	if err != nil {
		t.Flush("phase4")
		return "", err
	}

	t.Trace("phase4", "response", result)
	t.Flush("phase4")

	log.Info("Phase 4 completed")

	return result, nil
}

// executePhase5 は Phase 5（QA 検証+パッチ修正）を実行する
//
// コード定量チェック → 不合格時のみ LLM パッチ修正（最大1回）
func (s *scriptJobService) executePhase5(ctx context.Context, job *model.ScriptJob, lines []script.ParsedLine, brief script.Brief, allowedSpeakers []string, originalText string, t tracer.Tracer) []script.ParsedLine {
	log := logger.FromContext(ctx)

	config := script.ValidatorConfig{
		TalkMode:        brief.Constraints.TalkMode,
		DurationMinutes: brief.Episode.DurationMinutes,
	}

	// model_info を先に記録（パッチ修正が実行される場合のため）
	if s.llmRegistry != nil {
		t.Trace("phase5", "model_info", s.llmRegistry.GetModelInfo(phase5Config.Provider))
	}

	// 1回目のチェック
	result := script.Validate(lines, config)
	if result.Passed {
		log.Info("Phase 5: QA check passed")
		t.Trace("phase5", "qa_result", "passed")
		t.Flush("phase5")
		return lines
	}

	log.Info("Phase 5: QA check failed", "issues", len(result.Issues))

	issuesJSON, _ := json.Marshal(result.Issues) //nolint:errcheck // trace data
	t.Trace("phase5", "qa_result", string(issuesJSON))

	// キャンセルチェック（パッチ修正前）— job が nil の場合はスキップ
	if job != nil {
		if err := s.checkCanceled(ctx, job); err != nil {
			log.Info("Phase 5: canceled before patch")
			t.Flush("phase5")
			return lines
		}

		// 進捗: 85% - パッチ修正
		s.updateProgress(ctx, job, 85, "品質パッチを適用中...")
	}

	// LLM パッチ修正
	client, err := s.llmRegistry.Get(phase5Config.Provider)
	if err != nil {
		log.Warn("Phase 5: LLM client not available", "error", err)
		t.Flush("phase5")
		return lines
	}

	patchPrompt := buildPhase5UserPrompt(originalText, result.Issues)
	temp := phase5Config.Temperature
	opts := llm.ChatOptions{Temperature: &temp}

	t.Trace("phase5", "system_prompt", phase5SystemPrompt)
	t.Trace("phase5", "user_prompt", patchPrompt)

	patchedText, err := client.ChatWithOptions(ctx, phase5SystemPrompt, patchPrompt, opts)
	if err != nil {
		log.Warn("Phase 5: patch LLM call failed", "error", err)
		t.Flush("phase5")
		return lines
	}

	t.Trace("phase5", "response", patchedText)

	// パッチ結果をパース
	patchedResult := script.Parse(patchedText, allowedSpeakers)
	if len(patchedResult.Lines) == 0 {
		log.Warn("Phase 5: patched script parse failed")
		t.Flush("phase5")
		return lines
	}

	// 2回目のチェック
	result2 := script.Validate(patchedResult.Lines, config)
	if !result2.Passed {
		log.Warn("Phase 5: patched script still has issues", "issues", len(result2.Issues))
	} else {
		log.Info("Phase 5: patched script passed QA")
	}

	t.Flush("phase5")

	// パッチ後の結果をそのまま採用（合格/不合格に関わらず）
	return patchedResult.Lines
}

// buildPhase3UserPrompt は Phase 3 用のユーザープロンプトを構築する
func buildPhase3UserPrompt(brief script.Brief, phase2 *script.Phase2Output) string {
	var sb strings.Builder

	// ブリーフ情報
	briefJSON, err := brief.ToJSON()
	if err != nil {
		return ""
	}
	sb.WriteString("## ブリーフ\n")
	sb.WriteString(briefJSON)
	sb.WriteString("\n\n")

	// Phase 2 の出力を埋め込む
	phase2JSON, err := json.Marshal(phase2)
	if err != nil {
		return ""
	}
	sb.WriteString("## 素材とアウトライン\n")
	sb.WriteString(string(phase2JSON))

	return sb.String()
}

// buildPhase5UserPrompt は Phase 5 用のユーザープロンプトを構築する
func buildPhase5UserPrompt(scriptText string, issues []script.ValidationIssue) string {
	var sb strings.Builder

	sb.WriteString("## 台本\n")
	sb.WriteString(scriptText)
	sb.WriteString("\n\n")

	sb.WriteString("## 問題箇所\n")
	for _, issue := range issues {
		if issue.Line > 0 {
			sb.WriteString(fmt.Sprintf("- [行%d] %s: %s\n", issue.Line, issue.Check, issue.Message))
		} else {
			sb.WriteString(fmt.Sprintf("- [全体] %s: %s\n", issue.Check, issue.Message))
		}
	}

	return sb.String()
}

// updateProgress はジョブの進捗を更新し WebSocket で通知する
//
// 進捗のみを更新し、ステータスなど他のフィールドは変更しない
func (s *scriptJobService) updateProgress(ctx context.Context, job *model.ScriptJob, progress int, message string) {
	job.Progress = progress
	_ = s.scriptJobRepo.UpdateProgress(ctx, job.ID, progress) //nolint:errcheck // progress update is best effort
	s.notifyProgress(job.ID.String(), job.UserID.String(), progress, message)
}

// checkCanceled はジョブがキャンセルされていないかチェックする
//
// キャンセル中の場合は canceled に遷移させ、ErrCanceled を返す
func (s *scriptJobService) checkCanceled(ctx context.Context, job *model.ScriptJob) error {
	// 最新のステータスを取得
	latestJob, err := s.scriptJobRepo.FindByID(ctx, job.ID)
	if err != nil {
		return err
	}

	if latestJob.Status == model.ScriptJobStatusCanceling {
		// canceled に遷移
		latestJob.Status = model.ScriptJobStatusCanceled
		now := time.Now().UTC()
		latestJob.CompletedAt = &now
		if err := s.scriptJobRepo.Update(ctx, latestJob); err != nil {
			return err
		}
		s.notifyCanceled(latestJob.ID.String(), latestJob.UserID.String())
		return apperror.ErrCanceled.WithMessage("ジョブがキャンセルされました")
	}

	return nil
}

// failJob は指定されたジョブを失敗状態に更新する
func (s *scriptJobService) failJob(ctx context.Context, job *model.ScriptJob, err error) {
	log := logger.FromContext(ctx)
	completedAt := time.Now().UTC()
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
			JobType:      "台本生成 (Script)",
			ErrorCode:    errCode,
			ErrorMessage: errMsg,
			OccurredAt:   completedAt,
		}); alertErr != nil {
			log.Warn("failed to send slack alert for script job", "error", alertErr, "job_id", job.ID)
		}
	}
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

// CancelJob は指定されたジョブをキャンセルする
func (s *scriptJobService) CancelJob(ctx context.Context, userID, jobID string) error {
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
	job, err := s.scriptJobRepo.FindByID(ctx, jid)
	if err != nil {
		return err
	}

	// オーナーチェック
	if job.UserID != uid {
		return apperror.ErrForbidden.WithMessage("このジョブへのアクセス権限がありません")
	}

	// キャンセル可能なステータスか確認
	switch job.Status {
	case model.ScriptJobStatusPending:
		// pending → canceled に遷移
		job.Status = model.ScriptJobStatusCanceled
		now := time.Now().UTC()
		job.CompletedAt = &now
		if err := s.scriptJobRepo.Update(ctx, job); err != nil {
			return err
		}
		s.notifyCanceled(job.ID.String(), job.UserID.String())
		log.Info("script job canceled (was pending)", "job_id", jobID)
		return nil

	case model.ScriptJobStatusProcessing:
		// processing → canceling に遷移
		job.Status = model.ScriptJobStatusCanceling
		if err := s.scriptJobRepo.Update(ctx, job); err != nil {
			return err
		}
		s.notifyCanceling(job.ID.String(), job.UserID.String())
		log.Info("script job canceling (was processing)", "job_id", jobID)
		return nil

	case model.ScriptJobStatusCanceling:
		// 既にキャンセル中
		return apperror.ErrValidation.WithMessage("このジョブは既にキャンセル中です")

	case model.ScriptJobStatusCanceled:
		// 既にキャンセル済み
		return apperror.ErrValidation.WithMessage("このジョブは既にキャンセルされています")

	case model.ScriptJobStatusCompleted, model.ScriptJobStatusFailed:
		// 完了または失敗済みのジョブはキャンセル不可
		return apperror.ErrValidation.WithMessage("完了または失敗したジョブはキャンセルできません")

	default:
		return apperror.ErrInternal.WithMessage("不明なジョブステータスです")
	}
}

// notifyCanceling はジョブがキャンセル中になったことを WebSocket で通知する
func (s *scriptJobService) notifyCanceling(jobID, userID string) {
	if s.wsHub == nil {
		return
	}
	s.wsHub.SendToUser(userID, websocket.Message{
		Type: "script_canceling",
		Payload: map[string]interface{}{
			"jobId": jobID,
		},
	})
}

// notifyCanceled はジョブがキャンセルされたことを WebSocket で通知する
func (s *scriptJobService) notifyCanceled(jobID, userID string) {
	if s.wsHub == nil {
		return
	}
	s.wsHub.SendToUser(userID, websocket.Message{
		Type: "script_canceled",
		Payload: map[string]interface{}{
			"jobId": jobID,
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

// GenerateScriptDirect は DB を使わずにリクエストパラメータのみで台本を生成する（開発用）
func (s *scriptJobService) GenerateScriptDirect(ctx context.Context, req request.GenerateScriptDirectRequest) (*response.GenerateScriptDirectResponse, error) {
	log := logger.FromContext(ctx)

	// リクエスト → BriefInput に変換
	briefInput := script.BriefInput{
		EpisodeTitle:       req.EpisodeTitle,
		EpisodeDescription: req.EpisodeDescription,
		DurationMinutes:    req.DurationMinutes,
		EpisodeNumber:      req.EpisodeNumber,
		ChannelName:        req.ChannelName,
		ChannelDescription: req.ChannelDescription,
		ChannelCategory:    req.ChannelCategory,
		ChannelStyleGuide:  req.ChannelStyleGuide,
		MasterGuide:        req.MasterGuide,
		Theme:              req.Theme,
		WithEmotion:        req.WithEmotion,
	}

	for _, c := range req.Characters {
		briefInput.Characters = append(briefInput.Characters, script.BriefInputCharacter{
			Name:    c.Name,
			Gender:  c.Gender,
			Persona: c.Persona,
		})
	}

	// 許可された話者名のリストを作成
	allowedSpeakers := make([]string, len(req.Characters))
	for i, c := range req.Characters {
		allowedSpeakers[i] = c.Name
	}

	// ===== Phase 1: ブリーフ正規化 =====
	brief := script.NormalizeBrief(briefInput)

	briefJSON, err := brief.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("ブリーフの JSON 変換に失敗: %w", err)
	}

	t := tracer.New(s.traceMode, req.EpisodeTitle)

	log.Info("brief normalized", "talk_mode", brief.Constraints.TalkMode, "characters", len(brief.Characters))
	t.Trace("phase1", "brief", briefJSON)
	t.Flush("phase1")

	// ===== Phase 2: 素材+アウトライン生成 =====
	phase2Output, err := s.executePhase2(ctx, briefJSON, t)
	if err != nil {
		return nil, err
	}

	// ===== Phase 3: 台本ドラフト生成 =====
	generatedText, err := s.executePhase3(ctx, brief, phase2Output, t)
	if err != nil {
		return nil, err
	}

	parseResult := script.Parse(generatedText, allowedSpeakers)

	if len(parseResult.Lines) == 0 && parseResult.HasErrors() {
		return nil, apperror.ErrGenerationFailed.WithMessage("生成された台本のパースに失敗しました")
	}

	log.Info("script parsed", "lines", len(parseResult.Lines), "errors", len(parseResult.Errors),
		"total_chars", countTotalChars(parseResult.Lines), "target_chars", req.DurationMinutes*script.CharsPerMinute)

	// ===== Phase 4: リライト =====
	rewrittenText, err := s.executePhase4(ctx, generatedText, brief, t)
	if err != nil {
		log.Warn("Phase 4 rewrite failed, using original draft", "error", err)
		rewrittenText = generatedText
	} else {
		rewriteResult := script.Parse(rewrittenText, allowedSpeakers)
		if len(rewriteResult.Lines) > 0 {
			parseResult = rewriteResult
		} else {
			log.Warn("Phase 4 rewrite parse failed, using original draft")
			rewrittenText = generatedText
		}
	}

	// ===== Phase 5: QA 検証+パッチ修正 =====
	parsedLines := s.executePhase5(ctx, nil, parseResult.Lines, brief, allowedSpeakers, rewrittenText, t)

	// ParsedLine → FormatLine に変換してテキスト化
	formatLines := make([]script.FormatLine, len(parsedLines))
	for i, line := range parsedLines {
		formatLines[i] = script.FormatLine(line)
	}

	return &response.GenerateScriptDirectResponse{
		Script: script.Format(formatLines),
	}, nil
}

// countTotalChars は台本全行の合計文字数を返す
func countTotalChars(lines []script.ParsedLine) int {
	var total int
	for _, line := range lines {
		total += utf8.RuneCountInString(line.Text)
	}
	return total
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
