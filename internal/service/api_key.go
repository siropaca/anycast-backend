package service

import (
	"context"
	"time"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/apikey"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// APIKeyService は API キー管理のサービスインターフェース
type APIKeyService interface {
	Create(ctx context.Context, userID string, req request.CreateAPIKeyRequest) (*response.APIKeyCreatedResponse, error)
	List(ctx context.Context, userID string) ([]response.APIKeyResponse, error)
	Delete(ctx context.Context, userID, apiKeyID string) error
	Authenticate(ctx context.Context, plainKey string) (string, error)
}

type apiKeyService struct {
	apiKeyRepo repository.APIKeyRepository
}

// NewAPIKeyService は APIKeyService の実装を返す
func NewAPIKeyService(apiKeyRepo repository.APIKeyRepository) APIKeyService {
	return &apiKeyService{
		apiKeyRepo: apiKeyRepo,
	}
}

// Create は新しい API キーを作成し、平文キーを含むレスポンスを返す
func (s *apiKeyService) Create(ctx context.Context, userID string, req request.CreateAPIKeyRequest) (*response.APIKeyCreatedResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	// 同一ユーザー内で同じ名前の API キーが存在するかチェック
	exists, err := s.apiKeyRepo.ExistsByUserIDAndName(ctx, uid, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.ErrDuplicateName.WithMessage("同名の API キーが既に存在します")
	}

	result, err := apikey.Generate()
	if err != nil {
		logger.FromContext(ctx).Error("failed to generate api key", "error", err)
		return nil, apperror.ErrInternal.WithMessage("API キーの生成に失敗しました").WithError(err)
	}

	ak := &model.APIKey{
		UserID:  uid,
		Name:    req.Name,
		KeyHash: result.Hash,
		Prefix:  result.Prefix,
	}

	if err := s.apiKeyRepo.Create(ctx, ak); err != nil {
		return nil, err
	}

	return &response.APIKeyCreatedResponse{
		ID:        ak.ID,
		Name:      ak.Name,
		Key:       result.PlainText,
		Prefix:    ak.Prefix,
		CreatedAt: ak.CreatedAt,
	}, nil
}

// List はユーザーの全 API キーを取得する
func (s *apiKeyService) List(ctx context.Context, userID string) ([]response.APIKeyResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	apiKeys, err := s.apiKeyRepo.FindByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}

	responses := make([]response.APIKeyResponse, len(apiKeys))
	for i, ak := range apiKeys {
		responses[i] = response.APIKeyResponse{
			ID:         ak.ID,
			Name:       ak.Name,
			Prefix:     ak.Prefix,
			LastUsedAt: ak.LastUsedAt,
			CreatedAt:  ak.CreatedAt,
		}
	}

	return responses, nil
}

// Delete はユーザーが所有する API キーを削除する
func (s *apiKeyService) Delete(ctx context.Context, userID, apiKeyID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	akID, err := uuid.Parse(apiKeyID)
	if err != nil {
		return err
	}

	// 所有者チェック
	if _, err := s.apiKeyRepo.FindByUserIDAndID(ctx, uid, akID); err != nil {
		return err
	}

	return s.apiKeyRepo.Delete(ctx, akID)
}

// Authenticate は API キーの平文を検証し、対応するユーザー ID を返す
// 認証成功時、非同期で lastUsedAt を更新する
func (s *apiKeyService) Authenticate(ctx context.Context, plainKey string) (string, error) {
	hash := apikey.HashKey(plainKey)

	ak, err := s.apiKeyRepo.FindByKeyHash(ctx, hash)
	if err != nil {
		return "", apperror.ErrUnauthorized.WithMessage("無効な API キーです")
	}

	// 非同期で lastUsedAt を更新
	go func() {
		bgCtx := context.Background()
		now := time.Now().UTC()
		if err := s.apiKeyRepo.UpdateLastUsedAt(bgCtx, ak.ID, now); err != nil {
			logger.FromContext(bgCtx).Error("failed to update api key last_used_at", "error", err, "id", ak.ID)
		}
	}()

	return ak.UserID.String(), nil
}
