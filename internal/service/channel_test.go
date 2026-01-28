package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// storageClient のモック
type mockStorageClient struct {
	mock.Mock
}

func (m *mockStorageClient) Upload(ctx context.Context, data []byte, path, contentType string) (string, error) {
	args := m.Called(ctx, data, path, contentType)
	return args.String(0), args.Error(1)
}

func (m *mockStorageClient) GenerateSignedURL(ctx context.Context, path string, expiration time.Duration) (string, error) {
	args := m.Called(ctx, path, expiration)
	return args.String(0), args.Error(1)
}

func (m *mockStorageClient) Delete(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

// episodeRepository のモック
type mockEpisodeRepositoryForChannel struct {
	mock.Mock
}

func (m *mockEpisodeRepositoryForChannel) FindByID(ctx context.Context, id uuid.UUID) (*model.Episode, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Episode), args.Error(1)
}

func (m *mockEpisodeRepositoryForChannel) FindByChannelID(ctx context.Context, channelID uuid.UUID, filter repository.EpisodeFilter) ([]model.Episode, int64, error) {
	args := m.Called(ctx, channelID, filter)
	return args.Get(0).([]model.Episode), args.Get(1).(int64), args.Error(2)
}

func (m *mockEpisodeRepositoryForChannel) Create(ctx context.Context, episode *model.Episode) error {
	args := m.Called(ctx, episode)
	return args.Error(0)
}

func (m *mockEpisodeRepositoryForChannel) Update(ctx context.Context, episode *model.Episode) error {
	args := m.Called(ctx, episode)
	return args.Error(0)
}

func (m *mockEpisodeRepositoryForChannel) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockEpisodeRepositoryForChannel) IncrementPlayCount(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestToChannelResponse(t *testing.T) {
	now := time.Now()
	categoryID := uuid.New()
	channelID := uuid.New()
	userID := uuid.New()
	artworkID := uuid.New()
	voiceID := uuid.New()
	characterID := uuid.New()

	baseChannel := &model.Channel{
		ID:          channelID,
		UserID:      userID,
		Name:        "Test Channel",
		Description: "Test Description",
		UserPrompt:  "Test User Prompt",
		CategoryID:  categoryID,
		PublishedAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
		Category: model.Category{
			ID:   categoryID,
			Slug: "technology",
			Name: "テクノロジー",
		},
		ChannelCharacters: []model.ChannelCharacter{
			{
				ChannelID:   channelID,
				CharacterID: characterID,
				Character: model.Character{
					ID:      characterID,
					Name:    "太郎",
					Persona: "明るい性格",
					VoiceID: voiceID,
					Voice: model.Voice{
						ID:     voiceID,
						Name:   "ja-JP-Wavenet-C",
						Gender: model.GenderMale,
					},
				},
			},
		},
	}

	t.Run("isOwner が true の場合、userPrompt が含まれる", func(t *testing.T) {
		mockStorage := new(mockStorageClient)
		mockEpisodeRepo := new(mockEpisodeRepositoryForChannel)
		mockEpisodeRepo.On("FindByChannelID", mock.Anything, channelID, mock.Anything).Return([]model.Episode{}, int64(0), nil)
		svc := &channelService{storageClient: mockStorage, episodeRepo: mockEpisodeRepo}
		ctx := context.Background()

		resp, err := svc.toChannelResponse(ctx, baseChannel, true)

		assert.NoError(t, err)
		assert.Equal(t, channelID, resp.ID)
		assert.Equal(t, "Test Channel", resp.Name)
		assert.Equal(t, "Test Description", resp.Description)
		assert.Equal(t, "Test User Prompt", resp.UserPrompt)
		assert.Equal(t, categoryID, resp.Category.ID)
		assert.Equal(t, "technology", resp.Category.Slug)
		assert.Equal(t, "テクノロジー", resp.Category.Name)
		assert.NotNil(t, resp.PublishedAt)
		assert.Len(t, resp.Characters, 1)
		assert.Equal(t, characterID, resp.Characters[0].ID)
		assert.Equal(t, "太郎", resp.Characters[0].Name)
		assert.Equal(t, "明るい性格", resp.Characters[0].Persona)
		assert.Equal(t, voiceID, resp.Characters[0].Voice.ID)
		assert.Equal(t, "ja-JP-Wavenet-C", resp.Characters[0].Voice.Name)
		assert.Equal(t, "male", resp.Characters[0].Voice.Gender)
		assert.NotNil(t, resp.Episodes)
		assert.Len(t, resp.Episodes, 0)
	})

	t.Run("isOwner が false の場合、userPrompt が空文字になる", func(t *testing.T) {
		mockStorage := new(mockStorageClient)
		mockEpisodeRepo := new(mockEpisodeRepositoryForChannel)
		mockEpisodeRepo.On("FindByChannelID", mock.Anything, channelID, mock.Anything).Return([]model.Episode{}, int64(0), nil)
		svc := &channelService{storageClient: mockStorage, episodeRepo: mockEpisodeRepo}
		ctx := context.Background()

		resp, err := svc.toChannelResponse(ctx, baseChannel, false)

		assert.NoError(t, err)
		assert.Equal(t, channelID, resp.ID)
		assert.Equal(t, "Test Channel", resp.Name)
		assert.Equal(t, "Test Description", resp.Description)
		assert.Equal(t, "", resp.UserPrompt)
	})

	t.Run("Artwork が nil の場合、レスポンスの Artwork も nil", func(t *testing.T) {
		mockStorage := new(mockStorageClient)
		mockEpisodeRepo := new(mockEpisodeRepositoryForChannel)
		mockEpisodeRepo.On("FindByChannelID", mock.Anything, channelID, mock.Anything).Return([]model.Episode{}, int64(0), nil)
		svc := &channelService{storageClient: mockStorage, episodeRepo: mockEpisodeRepo}
		ctx := context.Background()

		channel := *baseChannel
		channel.Artwork = nil

		resp, err := svc.toChannelResponse(ctx, &channel, true)

		assert.NoError(t, err)
		assert.Nil(t, resp.Artwork)
	})

	t.Run("Artwork がある場合、署名 URL が生成される", func(t *testing.T) {
		mockStorage := new(mockStorageClient)
		mockStorage.On("GenerateSignedURL", mock.Anything, "images/artwork.png", storage.SignedURLExpirationImage).Return("https://signed-url.example.com/artwork.png", nil)
		mockEpisodeRepo := new(mockEpisodeRepositoryForChannel)
		mockEpisodeRepo.On("FindByChannelID", mock.Anything, channelID, mock.Anything).Return([]model.Episode{}, int64(0), nil)
		svc := &channelService{storageClient: mockStorage, episodeRepo: mockEpisodeRepo}
		ctx := context.Background()

		channel := *baseChannel
		channel.ArtworkID = &artworkID
		channel.Artwork = &model.Image{
			ID:   artworkID,
			Path: "images/artwork.png",
		}

		resp, err := svc.toChannelResponse(ctx, &channel, true)

		assert.NoError(t, err)
		assert.NotNil(t, resp.Artwork)
		assert.Equal(t, artworkID, resp.Artwork.ID)
		assert.Equal(t, "https://signed-url.example.com/artwork.png", resp.Artwork.URL)
		mockStorage.AssertExpectations(t)
	})

	t.Run("PublishedAt が nil の場合、レスポンスの PublishedAt も nil", func(t *testing.T) {
		mockStorage := new(mockStorageClient)
		mockEpisodeRepo := new(mockEpisodeRepositoryForChannel)
		mockEpisodeRepo.On("FindByChannelID", mock.Anything, channelID, mock.Anything).Return([]model.Episode{}, int64(0), nil)
		svc := &channelService{storageClient: mockStorage, episodeRepo: mockEpisodeRepo}
		ctx := context.Background()

		channel := *baseChannel
		channel.PublishedAt = nil

		resp, err := svc.toChannelResponse(ctx, &channel, true)

		assert.NoError(t, err)
		assert.Nil(t, resp.PublishedAt)
	})
}

func TestToChannelResponses(t *testing.T) {
	now := time.Now()
	categoryID := uuid.New()

	channels := []model.Channel{
		{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			Name:        "Channel 1",
			Description: "Description 1",
			UserPrompt:  "Prompt 1",
			CategoryID:  categoryID,
			CreatedAt:   now,
			UpdatedAt:   now,
			Category: model.Category{
				ID:   categoryID,
				Slug: "tech",
				Name: "Tech",
			},
		},
		{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			Name:        "Channel 2",
			Description: "Description 2",
			UserPrompt:  "Prompt 2",
			CategoryID:  categoryID,
			CreatedAt:   now,
			UpdatedAt:   now,
			Category: model.Category{
				ID:   categoryID,
				Slug: "tech",
				Name: "Tech",
			},
		},
	}

	t.Run("複数チャンネルを正しく変換する", func(t *testing.T) {
		mockStorage := new(mockStorageClient)
		mockEpisodeRepo := new(mockEpisodeRepositoryForChannel)
		mockEpisodeRepo.On("FindByChannelID", mock.Anything, mock.Anything, mock.Anything).Return([]model.Episode{}, int64(0), nil)
		svc := &channelService{storageClient: mockStorage, episodeRepo: mockEpisodeRepo}
		ctx := context.Background()

		result, err := svc.toChannelResponses(ctx, channels)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "Channel 1", result[0].Name)
		assert.Equal(t, "Channel 2", result[1].Name)
	})

	t.Run("オーナーとして扱われるため userPrompt が含まれる", func(t *testing.T) {
		mockStorage := new(mockStorageClient)
		mockEpisodeRepo := new(mockEpisodeRepositoryForChannel)
		mockEpisodeRepo.On("FindByChannelID", mock.Anything, mock.Anything, mock.Anything).Return([]model.Episode{}, int64(0), nil)
		svc := &channelService{storageClient: mockStorage, episodeRepo: mockEpisodeRepo}
		ctx := context.Background()

		result, err := svc.toChannelResponses(ctx, channels)

		assert.NoError(t, err)
		assert.Equal(t, "Prompt 1", result[0].UserPrompt)
		assert.Equal(t, "Prompt 2", result[1].UserPrompt)
	})

	t.Run("空のスライスの場合、空のスライスを返す", func(t *testing.T) {
		mockStorage := new(mockStorageClient)
		svc := &channelService{storageClient: mockStorage}
		ctx := context.Background()

		result, err := svc.toChannelResponses(ctx, []model.Channel{})

		assert.NoError(t, err)
		assert.Len(t, result, 0)
		assert.NotNil(t, result)
	})
}

func TestToCharacterResponsesFromChannelCharacters(t *testing.T) {
	voiceID1 := uuid.New()
	voiceID2 := uuid.New()
	charID1 := uuid.New()
	charID2 := uuid.New()
	channelID := uuid.New()

	channelCharacters := []model.ChannelCharacter{
		{
			ChannelID:   channelID,
			CharacterID: charID1,
			Character: model.Character{
				ID:      charID1,
				Name:    "太郎",
				Persona: "明るい性格",
				VoiceID: voiceID1,
				Voice: model.Voice{
					ID:     voiceID1,
					Name:   "ja-JP-Wavenet-C",
					Gender: model.GenderMale,
				},
			},
		},
		{
			ChannelID:   channelID,
			CharacterID: charID2,
			Character: model.Character{
				ID:      charID2,
				Name:    "花子",
				Persona: "落ち着いた性格",
				VoiceID: voiceID2,
				Voice: model.Voice{
					ID:     voiceID2,
					Name:   "ja-JP-Wavenet-B",
					Gender: model.GenderFemale,
				},
			},
		},
	}

	t.Run("複数キャラクターを正しく変換する", func(t *testing.T) {
		svc := &channelService{}
		result := svc.toCharacterResponsesFromChannelCharacters(channelCharacters)

		assert.Len(t, result, 2)
		assert.Equal(t, charID1, result[0].ID)
		assert.Equal(t, "太郎", result[0].Name)
		assert.Equal(t, "明るい性格", result[0].Persona)
		assert.Equal(t, voiceID1, result[0].Voice.ID)
		assert.Equal(t, "ja-JP-Wavenet-C", result[0].Voice.Name)
		assert.Equal(t, "male", result[0].Voice.Gender)

		assert.Equal(t, charID2, result[1].ID)
		assert.Equal(t, "花子", result[1].Name)
		assert.Equal(t, "落ち着いた性格", result[1].Persona)
		assert.Equal(t, voiceID2, result[1].Voice.ID)
		assert.Equal(t, "ja-JP-Wavenet-B", result[1].Voice.Name)
		assert.Equal(t, "female", result[1].Voice.Gender)
	})

	t.Run("空のスライスの場合、空のスライスを返す", func(t *testing.T) {
		svc := &channelService{}
		result := svc.toCharacterResponsesFromChannelCharacters([]model.ChannelCharacter{})

		assert.Len(t, result, 0)
		assert.NotNil(t, result)
		assert.Equal(t, []response.CharacterResponse{}, result)
	})
}
