package service

import (
	"context"
	"io"

	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/slack"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// FeedbackService はフィードバック関連のビジネスロジックインターフェースを表す
type FeedbackService interface {
	CreateFeedback(ctx context.Context, userID string, input CreateFeedbackInput) (*response.FeedbackDataResponse, error)
}

// CreateFeedbackInput はフィードバック作成の入力を表す
type CreateFeedbackInput struct {
	Content    string
	Screenshot *UploadImageInput
	PageURL    *string
	UserAgent  *string
}

type feedbackService struct {
	feedbackRepo  repository.FeedbackRepository
	imageRepo     repository.ImageRepository
	userRepo      repository.UserRepository
	storageClient storage.Client
	slackClient   slack.Client
}

// NewFeedbackService は feedbackService を生成して FeedbackService として返す
func NewFeedbackService(
	feedbackRepo repository.FeedbackRepository,
	imageRepo repository.ImageRepository,
	userRepo repository.UserRepository,
	storageClient storage.Client,
	slackClient slack.Client,
) FeedbackService {
	return &feedbackService{
		feedbackRepo:  feedbackRepo,
		imageRepo:     imageRepo,
		userRepo:      userRepo,
		storageClient: storageClient,
		slackClient:   slackClient,
	}
}

// CreateFeedback はフィードバックを作成する
func (s *feedbackService) CreateFeedback(ctx context.Context, userID string, input CreateFeedbackInput) (*response.FeedbackDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	// ユーザーの存在確認（外部キー制約があるため必須）
	user, err := s.userRepo.FindByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	feedback := &model.Feedback{
		UserID:    uid,
		Content:   input.Content,
		PageURL:   input.PageURL,
		UserAgent: input.UserAgent,
	}

	// スクリーンショットがある場合はアップロード
	var screenshotURL *string
	if input.Screenshot != nil {
		image, err := s.uploadScreenshot(ctx, input.Screenshot)
		if err != nil {
			return nil, err
		}
		feedback.ScreenshotID = &image.ID

		// 署名付き URL を生成
		url, err := s.storageClient.GenerateSignedURL(ctx, image.Path, storage.SignedURLExpirationImage)
		if err != nil {
			logger.FromContext(ctx).Warn("failed to generate signed URL for screenshot", "error", err)
		} else {
			screenshotURL = &url
		}
	}

	// フィードバックを保存
	if err := s.feedbackRepo.Create(ctx, feedback); err != nil {
		return nil, err
	}

	// Slack 通知（非同期で実行し、エラーは無視）
	if s.slackClient.IsFeedbackEnabled() {
		go func() {
			notification := slack.FeedbackNotification{
				UserEmail:     user.Email,
				UserName:      user.DisplayName,
				Content:       input.Content,
				ScreenshotURL: screenshotURL,
				PageURL:       input.PageURL,
				UserAgent:     input.UserAgent,
				CreatedAt:     feedback.CreatedAt,
			}
			if err := s.slackClient.SendFeedback(context.Background(), notification); err != nil {
				logger.Default().Warn("failed to send slack notification", "error", err)
			}
		}()
	}

	// レスポンスを構築
	resp := response.FeedbackResponse{
		ID:        feedback.ID,
		Content:   feedback.Content,
		PageURL:   feedback.PageURL,
		UserAgent: feedback.UserAgent,
		CreatedAt: feedback.CreatedAt,
	}

	if screenshotURL != nil && feedback.ScreenshotID != nil {
		resp.Screenshot = &response.ArtworkResponse{
			ID:  *feedback.ScreenshotID,
			URL: *screenshotURL,
		}
	}

	return &response.FeedbackDataResponse{
		Data: resp,
	}, nil
}

// screenshotMimeTypes はスクリーンショットで許可される MIME タイプ
var screenshotMimeTypes = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/webp": ".webp",
}

// uploadScreenshot はスクリーンショットをアップロードする
func (s *feedbackService) uploadScreenshot(ctx context.Context, input *UploadImageInput) (*model.Image, error) {
	log := logger.FromContext(ctx)

	// 拡張子を取得
	ext, ok := screenshotMimeTypes[input.ContentType]
	if !ok {
		ext = ".png" // フォールバック
	}

	// ファイルを読み込み
	data, err := io.ReadAll(input.File)
	if err != nil {
		log.Error("failed to read screenshot data", "error", err)
		return nil, err
	}

	// 画像 ID の生成
	imageID := uuid.New()

	// GCS にアップロード
	path := storage.GenerateImagePath(imageID.String(), ext)
	if _, err := s.storageClient.Upload(ctx, data, path, input.ContentType); err != nil {
		return nil, err
	}

	// DB に保存
	image := &model.Image{
		ID:       imageID,
		MimeType: input.ContentType,
		Path:     path,
		Filename: input.Filename,
		FileSize: input.FileSize,
	}
	if err := s.imageRepo.Create(ctx, image); err != nil {
		// DB 保存に失敗した場合は GCS のファイルを削除
		if deleteErr := s.storageClient.Delete(ctx, path); deleteErr != nil {
			log.Warn("failed to cleanup uploaded screenshot", "error", deleteErr, "path", path)
		}
		return nil, err
	}

	return image, nil
}
