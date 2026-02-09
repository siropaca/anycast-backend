package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// UserRepository のモック
type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepository) FindByIDWithAvatar(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *mockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *mockUserRepository) FindByUsernameWithAvatar(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepository) Search(ctx context.Context, filter repository.SearchUserFilter) ([]model.User, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.User), args.Get(1).(int64), args.Error(2)
}

func (m *mockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupAdminRouter(userRepo *mockUserRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Admin(userRepo))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})
	return r
}

func TestAdmin(t *testing.T) {
	t.Run("コンテキストにユーザー ID がない場合は 403 を返す", func(t *testing.T) {
		mockRepo := new(mockUserRepository)
		router := setupAdminRouter(mockRepo)

		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "FORBIDDEN")
	})

	t.Run("無効なユーザー ID 形式の場合は 403 を返す", func(t *testing.T) {
		mockRepo := new(mockUserRepository)

		gin.SetMode(gin.TestMode)
		r := gin.New()
		// ユーザー ID をコンテキストに設定するミドルウェア
		r.Use(func(c *gin.Context) {
			c.Set(string(UserIDKey), "invalid-uuid")
			c.Next()
		})
		r.Use(Admin(mockRepo))
		r.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "FORBIDDEN")
	})

	t.Run("ユーザーが見つからない場合は 403 を返す", func(t *testing.T) {
		userID := uuid.New()
		mockRepo := new(mockUserRepository)
		mockRepo.On("FindByID", mock.Anything, userID).Return(nil, apperror.ErrNotFound)

		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set(string(UserIDKey), userID.String())
			c.Next()
		})
		r.Use(Admin(mockRepo))
		r.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "FORBIDDEN")
		mockRepo.AssertExpectations(t)
	})

	t.Run("一般ユーザーの場合は 403 を返す", func(t *testing.T) {
		userID := uuid.New()
		mockRepo := new(mockUserRepository)
		mockRepo.On("FindByID", mock.Anything, userID).Return(&model.User{
			ID:   userID,
			Role: model.RoleUser,
		}, nil)

		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set(string(UserIDKey), userID.String())
			c.Next()
		})
		r.Use(Admin(mockRepo))
		r.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "FORBIDDEN")
		mockRepo.AssertExpectations(t)
	})

	t.Run("管理者ユーザーの場合は次のハンドラーが呼ばれる", func(t *testing.T) {
		userID := uuid.New()
		mockRepo := new(mockUserRepository)
		mockRepo.On("FindByID", mock.Anything, userID).Return(&model.User{
			ID:   userID,
			Role: model.RoleAdmin,
		}, nil)

		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set(string(UserIDKey), userID.String())
			c.Next()
		})
		r.Use(Admin(mockRepo))
		r.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "ok")
		mockRepo.AssertExpectations(t)
	})
}
