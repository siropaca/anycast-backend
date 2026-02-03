package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// FollowService はフォロー関連のビジネスロジックインターフェースを表す
type FollowService interface {
	ListFollows(ctx context.Context, userID string, limit, offset int) (*response.FollowListWithPaginationResponse, error)
	CreateFollow(ctx context.Context, userID, targetUsername string) (*response.FollowDataResponse, error)
	DeleteFollow(ctx context.Context, userID, targetUsername string) error
}

type followService struct {
	followRepo    repository.FollowRepository
	userRepo      repository.UserRepository
	storageClient storage.Client
}

// NewFollowService は followService を生成して FollowService として返す
func NewFollowService(
	followRepo repository.FollowRepository,
	userRepo repository.UserRepository,
	storageClient storage.Client,
) FollowService {
	return &followService{
		followRepo:    followRepo,
		userRepo:      userRepo,
		storageClient: storageClient,
	}
}

// ListFollows はフォロー中のユーザー一覧を取得する
func (s *followService) ListFollows(ctx context.Context, userID string, limit, offset int) (*response.FollowListWithPaginationResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	follows, total, err := s.followRepo.FindByUserID(ctx, uid, limit, offset)
	if err != nil {
		return nil, err
	}

	data := make([]response.FollowItemResponse, 0, len(follows))
	for _, follow := range follows {
		item := s.toFollowItemResponse(ctx, &follow)
		data = append(data, item)
	}

	return &response.FollowListWithPaginationResponse{
		Data: data,
		Pagination: response.PaginationResponse{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}, nil
}

// CreateFollow はフォローを登録する
func (s *followService) CreateFollow(ctx context.Context, userID, targetUsername string) (*response.FollowDataResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	targetUser, err := s.userRepo.FindByUsernameWithAvatar(ctx, targetUsername)
	if err != nil {
		return nil, err
	}

	if uid == targetUser.ID {
		return nil, apperror.ErrSelfFollowNotAllowed
	}

	exists, err := s.followRepo.ExistsByUserIDAndTargetUserID(ctx, uid, targetUser.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.ErrAlreadyFollowed
	}

	follow := &model.Follow{
		UserID:       uid,
		TargetUserID: targetUser.ID,
	}

	if err := s.followRepo.Create(ctx, follow); err != nil {
		return nil, err
	}

	return &response.FollowDataResponse{
		Data: response.FollowResponse{
			ID:           follow.ID,
			TargetUserID: follow.TargetUserID,
			CreatedAt:    follow.CreatedAt,
		},
	}, nil
}

// DeleteFollow はフォローを解除する
func (s *followService) DeleteFollow(ctx context.Context, userID, targetUsername string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	targetUser, err := s.userRepo.FindByUsernameWithAvatar(ctx, targetUsername)
	if err != nil {
		return err
	}

	return s.followRepo.DeleteByUserIDAndTargetUserID(ctx, uid, targetUser.ID)
}

// toFollowItemResponse は Follow を FollowItemResponse に変換する
func (s *followService) toFollowItemResponse(ctx context.Context, follow *model.Follow) response.FollowItemResponse {
	user := follow.TargetUser

	var avatar *response.AvatarResponse
	if user.Avatar != nil {
		if storage.IsExternalURL(user.Avatar.Path) {
			avatar = &response.AvatarResponse{
				ID:  user.Avatar.ID,
				URL: user.Avatar.Path,
			}
		} else {
			signedURL, err := s.storageClient.GenerateSignedURL(ctx, user.Avatar.Path, storage.SignedURLExpirationImage)
			if err == nil {
				avatar = &response.AvatarResponse{
					ID:  user.Avatar.ID,
					URL: signedURL,
				}
			}
		}
	}

	return response.FollowItemResponse{
		User: response.FollowUserResponse{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Avatar:      avatar,
		},
		FollowedAt: follow.CreatedAt,
	}
}
