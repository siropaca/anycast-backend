---
name: add-endpoint
description: 新しい API エンドポイントを追加する際のガイド。API 実装、エンドポイント追加、ハンドラー作成時に使用。
argument-hint: エンドポイントの仕様（例: "POST /channels/:channelId/episodes エピソード作成"）
---

# エンドポイント追加ガイド

新しい API エンドポイントを追加する際は、ドメインモデル駆動設計に従い、以下の順序で作業する。

---

## Phase 1: 設計ドキュメント更新

ドキュメントが先、実装が後。

### 1-1. ドメインモデル（必要な場合）

- **ファイル**: `docs/specs/domain-model.md`
- 新しいエンティティや属性の追加がある場合のみ更新

### 1-2. API ドキュメント

- **ファイル**: `docs/api/` 配下の該当ドキュメント（例: `episodes.md`, `channels.md`）
- 既存のドキュメントのフォーマットに合わせる:

```
## エンドポイント名

\```
METHOD /path/:param
\```

説明文。

**リクエスト:**（POST/PUT/PATCH の場合）
\```json
{
  "field": "value"
}
\```

**クエリパラメータ:**（GET でパラメータがある場合）

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| name | string | - | 説明 |

**レスポンス:**
\```json
{
  "data": { ... }
}
\```
```

### 1-3. API 一覧テーブル

- **ファイル**: `docs/api/README.md`
- API 一覧テーブルに行を追加（実装欄は空のまま）
- フォーマット: `| METHOD | \`/api/v1/path\` | 説明 | 権限 | | [詳細](./xxx.md#アンカー) |`

### 1-4. データベース設計（必要な場合）

- **ファイル**: `docs/specs/database.md`
- テーブルやカラムの追加がある場合のみ更新

---

## Phase 2: 実装

### 2-1. Request DTO

- **ファイル**: `internal/dto/request/` 配下
- `binding` タグでバリデーション

```go
// エピソード作成リクエスト
type CreateEpisodeRequest struct {
    Title       string  `json:"title" binding:"required,max=255"`
    Description string  `json:"description" binding:"max=2000"`
    ImageID     *string `json:"imageId" binding:"omitempty,uuid"`
}
```

- **クエリパラメータ**: `form` タグを使用

```go
type ListXxxRequest struct {
    PaginationRequest
    Status *string `form:"status" binding:"omitempty,oneof=published draft"`
}
```

- **nullable な更新フィールド**: `optional.Field[T]` を使用（`internal/pkg/optional`）

```go
type UpdateXxxRequest struct {
    ArtworkImageID optional.Field[string] `json:"artworkImageId"`
}
```

- **リレーション操作**: `connect` / `create` パターン

```go
type XxxInput struct {
    Connect []ConnectXxxInput `json:"connect"`
    Create  []CreateXxxInput  `json:"create"`
}
```

### 2-2. Response DTO（新規の場合）

- **ファイル**: `internal/dto/response/` 配下
- 常に値が存在するフィールド: `validate:"required"` タグ
- ポインタ型で `null` を返すフィールド（`omitempty` なし）: `extensions:"x-nullable"` タグ
- `omitempty` ありフィールド: `x-nullable` 不要

```go
type XxxResponse struct {
    ID        uuid.UUID  `json:"id" validate:"required"`
    Name      string     `json:"name" validate:"required"`
    Avatar    *string    `json:"avatar" extensions:"x-nullable"`
    Bio       *string    `json:"bio,omitempty" extensions:"x-nullable"`
    CreatedAt time.Time  `json:"createdAt" validate:"required"`
}
```

- 一覧レスポンス（ページネーション付き）:

```go
type XxxListWithPaginationResponse struct {
    Data       []XxxResponse      `json:"data" validate:"required"`
    Pagination PaginationResponse `json:"pagination" validate:"required"`
}
```

- 単体レスポンス（data ラッパー）:

```go
type XxxDataResponse struct {
    Data XxxResponse `json:"data" validate:"required"`
}
```

### 2-3. Repository（必要な場合）

- **ファイル**: `internal/repository/` 配下
- インターフェース定義 → 実装の順
- `internal/pkg/uuid` を使用（`github.com/google/uuid` は使わない）

### 2-4. Service

- **ファイル**: `internal/service/` 配下
- インターフェース定義 → 実装の順

```go
// XxxService は xxx 関連のビジネスロジックインターフェースを表す
type XxxService interface {
    GetXxx(ctx context.Context, userID, xxxID string) (*response.XxxDataResponse, error)
    CreateXxx(ctx context.Context, userID string, req request.CreateXxxRequest) (*response.XxxDataResponse, error)
}

type xxxService struct {
    xxxRepo   repository.XxxRepository
    // ...
}

// NewXxxService は xxxService を生成して XxxService として返す
func NewXxxService(xxxRepo repository.XxxRepository) XxxService {
    return &xxxService{xxxRepo: xxxRepo}
}
```

- エラーハンドリング: `apperror` パッケージを使用

### 2-5. Handler

- **ファイル**: `internal/handler/` 配下
- **Swagger アノテーション必須**（godoc コメントの形式）

```go
// XxxHandler godoc
// @Summary 要約
// @Description 説明
// @Tags タグ名
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param request body request.CreateXxxRequest true "リクエスト"
// @Success 200 {object} response.XxxDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /path/{param} [method]
```

- **認証パターン**:
  - 認証必須: `userID, ok := middleware.GetUserID(c); if !ok { Error(c, apperror.ErrUnauthorized); return }`
  - 任意認証: `userID, _ := middleware.GetUserID(c)`

- **パスパラメータの空チェック**:

```go
channelID := c.Param("channelId")
if channelID == "" {
    Error(c, apperror.ErrValidation.WithMessage("channelId は必須です"))
    return
}
```

- **リクエストボディのバインド**:

```go
var req request.CreateXxxRequest
if err := c.ShouldBindJSON(&req); err != nil {
    Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
    return
}
```

- **クエリパラメータのバインド**: `c.ShouldBindQuery(&req)`

- **エラーレスポンス**: `Error(c, err)` ヘルパーを使用
- **成功レスポンス**: `c.JSON(http.StatusOK, result)` または `c.JSON(http.StatusCreated, result)`

### 2-6. Router

- **ファイル**: `internal/router/router.go`
- 認証必須: `authenticated` グループに追加
- 任意認証: `optionalAuth` グループに追加
- 既存の同カテゴリのエンドポイント近くに配置

### 2-7. DI Container（必要な場合）

- **ファイル**: `internal/di/container.go`
- Repository → Service → Handler の順に初期化
- `Container` 構造体にハンドラーフィールドを追加

---

## Phase 3: テスト・仕上げ

### 3-1. ハンドラーテスト

- **ファイル**: `internal/handler/*_test.go`
- `testify/mock` でサービスをモック

```go
// XxxService のモック
type mockXxxService struct {
    mock.Mock
}

func (m *mockXxxService) GetXxx(ctx context.Context, ...) (*response.XxxDataResponse, error) {
    args := m.Called(ctx, ...)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*response.XxxDataResponse), args.Error(1)
}
```

- **ルーターセットアップ**:

```go
// テスト用のルーターをセットアップする
func setupXxxRouter(h *XxxHandler) *gin.Engine {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.GET("/path/:param", h.Method)
    return r
}

// 認証済みルーターをセットアップする
func setupAuthenticatedXxxRouter(h *XxxHandler, userID string) *gin.Engine {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.Use(func(c *gin.Context) {
        c.Set(string(middleware.UserIDKey), userID)
        c.Next()
    })
    r.POST("/path/:param", h.Method)
    return r
}
```

- **テスト関数**: `t.Run` でサブテストに分割、テスト名は日本語

```go
func TestXxxHandler_Method(t *testing.T) {
    t.Run("正常にリソースを取得できる", func(t *testing.T) {
        mockSvc := new(mockXxxService)
        // モック設定
        mockSvc.On("Method", mock.Anything, ...).Return(result, nil)

        handler := NewXxxHandler(mockSvc)
        router := setupAuthenticatedXxxRouter(handler, userID)

        w := httptest.NewRecorder()
        req := httptest.NewRequest("GET", "/path/"+id, http.NoBody)
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusOK, w.Code)
        mockSvc.AssertExpectations(t)
    })
}
```

### 3-2. Swagger 再生成

```bash
make swagger
```

### 3-3. `.http` ファイル更新

- **ファイル**: `http/` 配下の該当ファイル
- フォーマット:

```http
@baseUrl = http://localhost:8081/api/v1

# トークン生成: make token
@token = YOUR_TOKEN_HERE

### エンドポイント名
METHOD {{baseUrl}}/path/YOUR_ID_HERE
Content-Type: application/json
Authorization: Bearer {{token}}

{
  "field": "value"
}
```

### 3-4. API 一覧テーブル更新

- `docs/api/README.md` の実装欄を ✅ に更新

---

## コーディング規約チェックリスト

- [ ] Go コメントに `@param` / `@returns` などの JSDoc タグを使っていない
- [ ] エラーメッセージは日本語
- [ ] ログメッセージは英語
- [ ] 新しいフィールドを `fieldNameMap`（`internal/handler/validation.go`）に追加した
- [ ] `internal/pkg/uuid` を使用している（`github.com/google/uuid` ではない）
- [ ] エクスポートされるシンボルのコメントはシンボル名から始めている
- [ ] `make swagger` で Swagger ドキュメントを再生成した
- [ ] `http/` ファイルを更新した
- [ ] `docs/api/README.md` の実装欄を ✅ に更新した
