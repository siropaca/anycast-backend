package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/infrastructure/llm"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/script"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// LLM Client のモック
type mockLLMClient struct {
	mock.Mock
}

func (m *mockLLMClient) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	args := m.Called(ctx, systemPrompt, userPrompt)
	return args.String(0), args.Error(1)
}

func (m *mockLLMClient) ChatWithOptions(ctx context.Context, systemPrompt, userPrompt string, opts llm.ChatOptions) (string, error) {
	args := m.Called(ctx, systemPrompt, userPrompt, opts)
	return args.String(0), args.Error(1)
}

// ScriptJobRepository のモック
type mockScriptJobRepository struct {
	mock.Mock
}

func (m *mockScriptJobRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ScriptJob, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ScriptJob), args.Error(1)
}

func (m *mockScriptJobRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filter repository.ScriptJobFilter) ([]model.ScriptJob, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.ScriptJob), args.Error(1)
}

func (m *mockScriptJobRepository) FindByEpisodeID(ctx context.Context, episodeID uuid.UUID) ([]model.ScriptJob, error) {
	args := m.Called(ctx, episodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.ScriptJob), args.Error(1)
}

func (m *mockScriptJobRepository) FindPendingByEpisodeID(ctx context.Context, episodeID uuid.UUID) (*model.ScriptJob, error) {
	args := m.Called(ctx, episodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ScriptJob), args.Error(1)
}

func (m *mockScriptJobRepository) Create(ctx context.Context, job *model.ScriptJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *mockScriptJobRepository) Update(ctx context.Context, job *model.ScriptJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *mockScriptJobRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockScriptJobRepository) UpdateProgress(ctx context.Context, id uuid.UUID, progress int) error {
	args := m.Called(ctx, id, progress)
	return args.Error(0)
}

func TestScriptJobStatus(t *testing.T) {
	t.Run("ScriptJobStatus 定数が正しい", func(t *testing.T) {
		assert.Equal(t, model.ScriptJobStatus("pending"), model.ScriptJobStatusPending)
		assert.Equal(t, model.ScriptJobStatus("processing"), model.ScriptJobStatusProcessing)
		assert.Equal(t, model.ScriptJobStatus("canceling"), model.ScriptJobStatusCanceling)
		assert.Equal(t, model.ScriptJobStatus("completed"), model.ScriptJobStatusCompleted)
		assert.Equal(t, model.ScriptJobStatus("failed"), model.ScriptJobStatusFailed)
		assert.Equal(t, model.ScriptJobStatus("canceled"), model.ScriptJobStatusCanceled)
	})
}

func TestScriptJobService_CancelJob(t *testing.T) {
	userID := uuid.New()
	jobID := uuid.New()
	episodeID := uuid.New()

	t.Run("pending ジョブをキャンセルすると canceled になる", func(t *testing.T) {
		mockRepo := new(mockScriptJobRepository)
		job := &model.ScriptJob{
			ID:        jobID,
			EpisodeID: episodeID,
			UserID:    userID,
			Status:    model.ScriptJobStatusPending,
		}
		mockRepo.On("FindByID", mock.Anything, jobID).Return(job, nil)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(j *model.ScriptJob) bool {
			return j.Status == model.ScriptJobStatusCanceled
		})).Return(nil)

		svc := &scriptJobService{scriptJobRepo: mockRepo}
		err := svc.CancelJob(context.Background(), userID.String(), jobID.String())

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("processing ジョブをキャンセルすると canceling になる", func(t *testing.T) {
		mockRepo := new(mockScriptJobRepository)
		job := &model.ScriptJob{
			ID:        jobID,
			EpisodeID: episodeID,
			UserID:    userID,
			Status:    model.ScriptJobStatusProcessing,
		}
		mockRepo.On("FindByID", mock.Anything, jobID).Return(job, nil)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(j *model.ScriptJob) bool {
			return j.Status == model.ScriptJobStatusCanceling
		})).Return(nil)

		svc := &scriptJobService{scriptJobRepo: mockRepo}
		err := svc.CancelJob(context.Background(), userID.String(), jobID.String())

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("completed ジョブはキャンセルできない", func(t *testing.T) {
		mockRepo := new(mockScriptJobRepository)
		job := &model.ScriptJob{
			ID:        jobID,
			EpisodeID: episodeID,
			UserID:    userID,
			Status:    model.ScriptJobStatusCompleted,
		}
		mockRepo.On("FindByID", mock.Anything, jobID).Return(job, nil)

		svc := &scriptJobService{scriptJobRepo: mockRepo}
		err := svc.CancelJob(context.Background(), userID.String(), jobID.String())

		assert.Error(t, err)
		var appErr *apperror.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, apperror.CodeValidation, appErr.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("他のユーザーのジョブはキャンセルできない", func(t *testing.T) {
		mockRepo := new(mockScriptJobRepository)
		otherUserID := uuid.New()
		job := &model.ScriptJob{
			ID:        jobID,
			EpisodeID: episodeID,
			UserID:    otherUserID,
			Status:    model.ScriptJobStatusPending,
		}
		mockRepo.On("FindByID", mock.Anything, jobID).Return(job, nil)

		svc := &scriptJobService{scriptJobRepo: mockRepo}
		err := svc.CancelJob(context.Background(), userID.String(), jobID.String())

		assert.Error(t, err)
		var appErr *apperror.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, apperror.CodeForbidden, appErr.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestScriptJobService_updateProgress(t *testing.T) {
	jobID := uuid.New()
	userID := uuid.New()
	episodeID := uuid.New()

	t.Run("UpdateProgress は進捗のみを更新しステータスを上書きしない", func(t *testing.T) {
		mockRepo := new(mockScriptJobRepository)
		job := &model.ScriptJob{
			ID:        jobID,
			EpisodeID: episodeID,
			UserID:    userID,
			Status:    model.ScriptJobStatusProcessing,
			Progress:  30,
		}

		// UpdateProgress が呼ばれることを確認（Update ではなく）
		mockRepo.On("UpdateProgress", mock.Anything, jobID, 50).Return(nil)

		svc := &scriptJobService{scriptJobRepo: mockRepo}
		svc.updateProgress(context.Background(), job, 50, "処理中...")

		// job.Progress がメモリ上で更新されていることを確認
		assert.Equal(t, 50, job.Progress)
		// Status は変更されていないことを確認
		assert.Equal(t, model.ScriptJobStatusProcessing, job.Status)
		mockRepo.AssertExpectations(t)
	})
}

func TestScriptJobService_checkCanceled(t *testing.T) {
	jobID := uuid.New()
	userID := uuid.New()
	episodeID := uuid.New()

	t.Run("canceling 状態のジョブは ErrCanceled を返す", func(t *testing.T) {
		mockRepo := new(mockScriptJobRepository)
		job := &model.ScriptJob{
			ID:        jobID,
			EpisodeID: episodeID,
			UserID:    userID,
			Status:    model.ScriptJobStatusCanceling,
		}
		mockRepo.On("FindByID", mock.Anything, jobID).Return(job, nil)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(j *model.ScriptJob) bool {
			return j.Status == model.ScriptJobStatusCanceled
		})).Return(nil)

		svc := &scriptJobService{scriptJobRepo: mockRepo}
		err := svc.checkCanceled(context.Background(), job)

		assert.Error(t, err)
		var appErr *apperror.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, apperror.CodeCanceled, appErr.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("processing 状態のジョブはエラーを返さない", func(t *testing.T) {
		mockRepo := new(mockScriptJobRepository)
		job := &model.ScriptJob{
			ID:        jobID,
			EpisodeID: episodeID,
			UserID:    userID,
			Status:    model.ScriptJobStatusProcessing,
		}
		mockRepo.On("FindByID", mock.Anything, jobID).Return(job, nil)

		svc := &scriptJobService{scriptJobRepo: mockRepo}
		err := svc.checkCanceled(context.Background(), job)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

// validPhase2JSON は Phase 2 テスト用の有効な JSON レスポンス
var validPhase2JSON = `{
	"grounding": {
		"definitions": [
			{"term": "AI", "definition": "人工知能"},
			{"term": "ML", "definition": "機械学習"},
			{"term": "DL", "definition": "深層学習"}
		],
		"examples": [
			{"id": "ex1", "situation": "企業の AI 導入", "detail": "導入率が前年比30%増加"},
			{"id": "ex2", "situation": "自動運転", "detail": "レベル4の実証実験が10都市で開始"},
			{"id": "ex3", "situation": "医療 AI", "detail": "がん検出精度が95%に到達"}
		],
		"pitfalls": [
			{"id": "pit1", "misconception": "AI は万能", "reality": "特定タスクに特化"},
			{"id": "pit2", "misconception": "AI は仕事を奪う", "reality": "新しい仕事も生まれる"},
			{"id": "pit3", "misconception": "AI は人間と同じ", "reality": "統計的パターン認識"}
		],
		"questions": [
			{"id": "q1", "question": "AI はどこまで進化するの？"},
			{"id": "q2", "question": "私たちの生活はどう変わる？"},
			{"id": "q3", "question": "AI 時代に必要なスキルは？"}
		],
		"action_steps": [
			{"id": "act1", "step": "ChatGPT を使ってみる"},
			{"id": "act2", "step": "AI 関連のニュースを読む"},
			{"id": "act3", "step": "プログラミングを学ぶ"}
		]
	},
	"outline": {
		"opening": {"hook": "AI って本当にすごいの？"},
		"blocks": [
			{
				"block_number": 1,
				"topic": "AI の現状",
				"example_ids": ["ex1"],
				"pitfall_ids": ["pit1"],
				"action_step_ids": ["act1"],
				"question_ids": ["q1"]
			},
			{
				"block_number": 2,
				"topic": "AI の影響",
				"example_ids": ["ex2"],
				"pitfall_ids": ["pit2"],
				"action_step_ids": ["act2"],
				"question_ids": ["q2"]
			},
			{
				"block_number": 3,
				"topic": "AI の未来",
				"example_ids": ["ex3"],
				"pitfall_ids": ["pit3"],
				"action_step_ids": ["act3"],
				"question_ids": ["q3"]
			}
		],
		"closing": {
			"summary": "AI は着実に進化している",
			"takeaway": "まずは身近なところから触れてみよう"
		}
	}
}`

func TestScriptJobService_executePhase2(t *testing.T) {
	t.Run("正常系: 1回目で成功", func(t *testing.T) {
		mockLLM := new(mockLLMClient)
		mockLLM.On("ChatWithOptions", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(validPhase2JSON, nil).Once()

		registry := llm.NewRegistry()
		registry.Register(llm.ProviderOpenAI, mockLLM)
		svc := &scriptJobService{llmRegistry: registry}
		output, err := svc.executePhase2(context.Background(), `{"theme":"test"}`)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Len(t, output.Outline.Blocks, 3)
		mockLLM.AssertExpectations(t)
	})

	t.Run("リトライ後成功: 1回目失敗、2回目成功", func(t *testing.T) {
		mockLLM := new(mockLLMClient)
		mockLLM.On("ChatWithOptions", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return("", fmt.Errorf("API error")).Once()
		mockLLM.On("ChatWithOptions", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(validPhase2JSON, nil).Once()

		registry := llm.NewRegistry()
		registry.Register(llm.ProviderOpenAI, mockLLM)
		svc := &scriptJobService{llmRegistry: registry}
		output, err := svc.executePhase2(context.Background(), `{"theme":"test"}`)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		mockLLM.AssertExpectations(t)
	})

	t.Run("全失敗でエラーを返す", func(t *testing.T) {
		mockLLM := new(mockLLMClient)
		mockLLM.On("ChatWithOptions", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return("invalid json", nil).Times(2)

		registry := llm.NewRegistry()
		registry.Register(llm.ProviderOpenAI, mockLLM)
		svc := &scriptJobService{llmRegistry: registry}
		output, err := svc.executePhase2(context.Background(), `{"theme":"test"}`)

		assert.Error(t, err)
		assert.Nil(t, output)
		var appErr *apperror.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, apperror.CodeGenerationFailed, appErr.Code)
		mockLLM.AssertExpectations(t)
	})
}

func TestScriptJobService_executePhase4(t *testing.T) {
	jobID := uuid.New()
	userID := uuid.New()
	episodeID := uuid.New()

	t.Run("QA 合格でパッチなし", func(t *testing.T) {
		mockRepo := new(mockScriptJobRepository)
		svc := &scriptJobService{scriptJobRepo: mockRepo}

		// 合格するデータ（十分な行数、ゆらぎあり、句点なし）
		lines := make([]script.ParsedLine, 0, 44)
		speakers := []string{"太郎", "花子"}
		texts := []string{
			"これは最初のセリフですね",
			"はい、そうですよ、今日もよろしくお願いしますね、頑張っていきましょう",
			"今日のテーマについてですが",
			"えーと、具体的にはどういうことなんでしょうか、もうちょっと教えてください",
			"簡単に言うとこういうこと",
			"なるほどね、それはすごく面白い話ですね、初めて知りましたよ",
			"そうなんです、意外でしょう",
			"ちょっと驚きましたね、初めて聞きましたし本当にびっくりしています",
			"もう少し教えてくれますか",
			"もちろんです、具体例を出してみますね、分かりやすいでしょう",
			"すごいですね、これは",
		}
		for i := 0; i < 44; i++ {
			lines = append(lines, script.ParsedLine{
				SpeakerName: speakers[i%2],
				Text:        texts[i%len(texts)],
			})
		}

		brief := script.Brief{
			Episode:     script.BriefEpisode{DurationMinutes: 10},
			Constraints: script.BriefConstraints{TalkMode: script.TalkModeDialogue},
		}

		job := &model.ScriptJob{
			ID:        jobID,
			EpisodeID: episodeID,
			UserID:    userID,
			Status:    model.ScriptJobStatusProcessing,
		}

		result := svc.executePhase4(context.Background(), job, lines, brief, speakers, "original text")
		assert.Equal(t, len(lines), len(result))
	})

	t.Run("パッチ修正で合格", func(t *testing.T) {
		mockRepo := new(mockScriptJobRepository)
		mockLLM := new(mockLLMClient)

		// checkCanceled 用
		mockRepo.On("FindByID", mock.Anything, jobID).Return(&model.ScriptJob{
			ID:     jobID,
			Status: model.ScriptJobStatusProcessing,
		}, nil)
		// updateProgress 用
		mockRepo.On("UpdateProgress", mock.Anything, jobID, 85).Return(nil)

		// 修正されたテキストを返す（句点なし版）
		patchedText := ""
		for i := 0; i < 44; i++ {
			speaker := "太郎"
			if i%2 == 1 {
				speaker = "花子"
			}
			texts := []string{
				"これは修正後のセリフです",
				"はい、そうですよ、修正してもらいました、よかったですね",
				"今日のテーマについてです",
				"えーと、具体的にはどういうことなんでしょうか、教えてほしいですね",
				"簡単に言うとこういうこと",
				"なるほどね、それはすごく面白い話ですね、感動しました",
				"そうなんです、意外でしょう",
				"ちょっと驚きましたね、それは初めて聞きましたし本当にすごいです",
				"もう少し教えてくれますか",
				"もちろん、具体例を出すとこんな感じですね、どうでしょうか",
				"これはすごいですね",
			}
			patchedText += fmt.Sprintf("%s: %s\n", speaker, texts[i%len(texts)])
		}

		mockLLM.On("ChatWithOptions", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(patchedText, nil)

		registry := llm.NewRegistry()
		registry.Register(llm.ProviderOpenAI, mockLLM)
		svc := &scriptJobService{
			scriptJobRepo: mockRepo,
			llmRegistry:   registry,
		}

		// 不合格データ（句点あり）
		lines := []script.ParsedLine{
			{SpeakerName: "太郎", Text: "これはセリフです。"},
		}

		brief := script.Brief{
			Episode:     script.BriefEpisode{DurationMinutes: 10},
			Constraints: script.BriefConstraints{TalkMode: script.TalkModeDialogue},
		}

		job := &model.ScriptJob{
			ID:        jobID,
			EpisodeID: episodeID,
			UserID:    userID,
			Status:    model.ScriptJobStatusProcessing,
		}

		result := svc.executePhase4(context.Background(), job, lines, brief, []string{"太郎", "花子"}, "太郎: これはセリフです。")
		// パッチ結果が返ることを確認
		assert.Greater(t, len(result), len(lines))
		mockLLM.AssertExpectations(t)
	})

	t.Run("パッチ後も不合格でも結果を採用", func(t *testing.T) {
		mockRepo := new(mockScriptJobRepository)
		mockLLM := new(mockLLMClient)

		mockRepo.On("FindByID", mock.Anything, jobID).Return(&model.ScriptJob{
			ID:     jobID,
			Status: model.ScriptJobStatusProcessing,
		}, nil)
		mockRepo.On("UpdateProgress", mock.Anything, jobID, 85).Return(nil)

		// パッチ修正でも不合格な結果を返す（まだ句点あり）
		mockLLM.On("ChatWithOptions", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return("太郎: まだ句点がある。", nil)

		registry := llm.NewRegistry()
		registry.Register(llm.ProviderOpenAI, mockLLM)
		svc := &scriptJobService{
			scriptJobRepo: mockRepo,
			llmRegistry:   registry,
		}

		lines := []script.ParsedLine{
			{SpeakerName: "太郎", Text: "元のセリフです。"},
		}

		brief := script.Brief{
			Episode:     script.BriefEpisode{DurationMinutes: 1},
			Constraints: script.BriefConstraints{TalkMode: script.TalkModeMonologue},
		}

		job := &model.ScriptJob{
			ID:        jobID,
			EpisodeID: episodeID,
			UserID:    userID,
			Status:    model.ScriptJobStatusProcessing,
		}

		result := svc.executePhase4(context.Background(), job, lines, brief, []string{"太郎"}, "太郎: 元のセリフです。")
		// パッチ結果が返ることを確認（不合格でも採用）
		assert.NotEmpty(t, result)
		mockLLM.AssertExpectations(t)
	})
}

func TestBuildPhase3UserPrompt(t *testing.T) {
	t.Run("ブリーフと Phase 2 出力からプロンプトを構築できる", func(t *testing.T) {
		brief := script.Brief{
			Episode: script.BriefEpisode{Title: "テスト", DurationMinutes: 10},
			Channel: script.BriefChannel{Name: "テスト", Category: "テスト"},
			Theme:   "AIの未来",
			Constraints: script.BriefConstraints{
				TalkMode: script.TalkModeDialogue,
			},
		}

		phase2 := &script.Phase2Output{
			Grounding: script.Grounding{
				Examples: []script.Example{{ID: "ex1", Situation: "テスト", Detail: "詳細"}},
			},
			Outline: script.Outline{
				Opening: script.Opening{Hook: "掴み"},
				Blocks: []script.OutlineBlock{
					{BlockNumber: 1, Topic: "トピック1"},
					{BlockNumber: 2, Topic: "トピック2"},
					{BlockNumber: 3, Topic: "トピック3"},
				},
				Closing: script.Closing{Summary: "まとめ"},
			},
		}

		result := buildPhase3UserPrompt(brief, phase2)

		assert.Contains(t, result, "ブリーフ")
		assert.Contains(t, result, "素材とアウトライン")
		assert.Contains(t, result, "AIの未来")
	})
}

func TestBuildPhase4UserPrompt(t *testing.T) {
	t.Run("台本と問題箇所からプロンプトを構築できる", func(t *testing.T) {
		issues := []script.ValidationIssue{
			{Check: "trailing_period", Line: 3, Message: "セリフの末尾に句点があります"},
			{Check: "minimum_lines", Line: 0, Message: "行数不足"},
		}

		result := buildPhase4UserPrompt("太郎: テスト", issues)

		assert.Contains(t, result, "台本")
		assert.Contains(t, result, "太郎: テスト")
		assert.Contains(t, result, "問題箇所")
		assert.Contains(t, result, "[行3]")
		assert.Contains(t, result, "[全体]")
	})
}
