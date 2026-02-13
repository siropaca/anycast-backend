package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/imagegen"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// 画像生成で返される MIME タイプと拡張子のマッピング
var imageGenMimeTypeToExt = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/webp": ".webp",
}

// ArtworkService はアートワーク生成サービスのインターフェースを表す
type ArtworkService interface {
	// GenerateChannelArtwork はチャンネルのアートワークを AI で生成する
	GenerateChannelArtwork(ctx context.Context, userID, channelID string, req request.GenerateChannelArtworkRequest) (*response.ImageUploadDataResponse, error)
	// GenerateEpisodeArtwork はエピソードのアートワークを AI で生成する
	GenerateEpisodeArtwork(ctx context.Context, userID, channelID, episodeID string, req request.GenerateEpisodeArtworkRequest) (*response.ImageUploadDataResponse, error)
}

type artworkService struct {
	channelRepo    repository.ChannelRepository
	episodeRepo    repository.EpisodeRepository
	imageRepo      repository.ImageRepository
	storageClient  storage.Client
	imagegenClient imagegen.Client
}

// NewArtworkService は artworkService を生成して ArtworkService として返す
func NewArtworkService(
	channelRepo repository.ChannelRepository,
	episodeRepo repository.EpisodeRepository,
	imageRepo repository.ImageRepository,
	storageClient storage.Client,
	imagegenClient imagegen.Client,
) ArtworkService {
	return &artworkService{
		channelRepo:    channelRepo,
		episodeRepo:    episodeRepo,
		imageRepo:      imageRepo,
		storageClient:  storageClient,
		imagegenClient: imagegenClient,
	}
}

// GenerateChannelArtwork はチャンネルのアートワークを AI で生成する
func (s *artworkService) GenerateChannelArtwork(ctx context.Context, userID, channelID string, req request.GenerateChannelArtworkRequest) (*response.ImageUploadDataResponse, error) {
	log := logger.FromContext(ctx)

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
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルのアートワーク生成権限がありません")
	}

	// プロンプト決定
	prompt := buildChannelArtworkPrompt(channel)
	if req.Prompt != nil && *req.Prompt != "" {
		prompt = *req.Prompt
	}

	log.Debug("generating channel artwork", "channel_id", channelID, "prompt", prompt)

	// 画像生成 → アップロード → DB 保存
	resp, err := s.generateAndSaveImage(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// setArtwork（デフォルト true）
	if req.SetArtwork == nil || *req.SetArtwork {
		channel.ArtworkID = &resp.Data.ID
		channel.Artwork = nil
		if err := s.channelRepo.Update(ctx, channel); err != nil {
			return nil, err
		}
		log.Debug("channel artwork updated", "channel_id", channelID, "image_id", resp.Data.ID)
	}

	return resp, nil
}

// GenerateEpisodeArtwork はエピソードのアートワークを AI で生成する
func (s *artworkService) GenerateEpisodeArtwork(ctx context.Context, userID, channelID, episodeID string, req request.GenerateEpisodeArtworkRequest) (*response.ImageUploadDataResponse, error) {
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
		return nil, apperror.ErrForbidden.WithMessage("このエピソードのアートワーク生成権限がありません")
	}

	// エピソードの存在確認
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("エピソードが見つかりません")
	}

	// プロンプト決定
	prompt := buildEpisodeArtworkPrompt(episode, channel)
	if req.Prompt != nil && *req.Prompt != "" {
		prompt = *req.Prompt
	}

	log.Debug("generating episode artwork", "episode_id", episodeID, "prompt", prompt)

	// 画像生成 → アップロード → DB 保存
	resp, err := s.generateAndSaveImage(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// setArtwork（デフォルト true）
	if req.SetArtwork == nil || *req.SetArtwork {
		episode.ArtworkID = &resp.Data.ID
		episode.Artwork = nil
		if err := s.episodeRepo.Update(ctx, episode); err != nil {
			return nil, err
		}
		log.Debug("episode artwork updated", "episode_id", episodeID, "image_id", resp.Data.ID)
	}

	return resp, nil
}

// generateAndSaveImage は画像を生成し、GCS にアップロードして DB に保存する
func (s *artworkService) generateAndSaveImage(ctx context.Context, prompt string) (*response.ImageUploadDataResponse, error) {
	log := logger.FromContext(ctx)

	// 画像生成
	result, err := s.imagegenClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// MIME タイプから拡張子を決定
	ext, ok := imageGenMimeTypeToExt[result.MimeType]
	if !ok {
		ext = ".png" // フォールバック
	}

	// 画像 ID の生成
	imageID := uuid.New()
	filename := fmt.Sprintf("artwork_%s%s", imageID.String()[:8], ext)

	// GCS へアップロード
	path := storage.GenerateImagePath(imageID.String(), ext)
	if _, err := s.storageClient.Upload(ctx, result.Data, path, result.MimeType); err != nil {
		return nil, err
	}

	// DB に保存
	image := &model.Image{
		ID:       imageID,
		MimeType: result.MimeType,
		Path:     path,
		Filename: filename,
		FileSize: len(result.Data),
	}

	if err := s.imageRepo.Create(ctx, image); err != nil {
		// DB 保存に失敗した場合は GCS のファイルを削除
		if deleteErr := s.storageClient.Delete(ctx, path); deleteErr != nil {
			log.Warn("failed to cleanup generated image", "error", deleteErr, "path", path)
		}
		return nil, err
	}

	// 署名付き URL を生成
	signedURL, err := s.storageClient.GenerateSignedURL(ctx, path, 1*time.Hour)
	if err != nil {
		return nil, err
	}

	return &response.ImageUploadDataResponse{
		Data: response.ImageUploadResponse{
			ID:       image.ID,
			MimeType: image.MimeType,
			URL:      signedURL,
			Filename: filepath.Base(image.Filename),
			FileSize: image.FileSize,
		},
	}, nil
}

// buildChannelArtworkPrompt はチャンネルのメタデータからアートワーク生成プロンプトを構築する
func buildChannelArtworkPrompt(channel *model.Channel) string {
	var b strings.Builder
	b.WriteString("ポッドキャストチャンネルのカバーアートを作成してください。\n")
	b.WriteString(fmt.Sprintf("チャンネル名: %s", channel.Name))

	if channel.Category.Name != "" {
		b.WriteString(fmt.Sprintf("、カテゴリ: %s", channel.Category.Name))
	}

	if channel.Description != "" {
		b.WriteString(fmt.Sprintf("、説明: %s", channel.Description))
	}

	b.WriteString("。\n")
	b.WriteString("正方形のアイコンとして適した、プロフェッショナルで目を引くデザインにしてください。\n")
	b.WriteString("テキストや文字は含めないでください。")
	return b.String()
}

// buildEpisodeArtworkPrompt はエピソードのメタデータからアートワーク生成プロンプトを構築する
func buildEpisodeArtworkPrompt(episode *model.Episode, channel *model.Channel) string {
	var b strings.Builder
	b.WriteString("ポッドキャストエピソードのカバーアートを作成してください。\n")
	b.WriteString(fmt.Sprintf("エピソード名: %s", episode.Title))

	if channel.Category.Name != "" {
		b.WriteString(fmt.Sprintf("、カテゴリ: %s", channel.Category.Name))
	}

	if episode.Description != "" {
		b.WriteString(fmt.Sprintf("、説明: %s", episode.Description))
	}

	b.WriteString("。\n")
	b.WriteString("正方形のアイコンとして適した、エピソードの内容を象徴するデザインにしてください。\n")
	b.WriteString("テキストや文字は含めないでください。")
	return b.String()
}
