package service

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// モックリポジトリ

type mockChannelRepository struct {
	mock.Mock
}

func (m *mockChannelRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Channel, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Channel), args.Error(1)
}

func (m *mockChannelRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filter repository.ChannelFilter) ([]model.Channel, int64, error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).([]model.Channel), args.Get(1).(int64), args.Error(2)
}

func (m *mockChannelRepository) FindPublishedByUserID(ctx context.Context, userID uuid.UUID) ([]model.Channel, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.Channel), args.Error(1)
}

func (m *mockChannelRepository) Search(ctx context.Context, filter repository.SearchChannelFilter) ([]model.Channel, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.Channel), args.Get(1).(int64), args.Error(2)
}

func (m *mockChannelRepository) Create(ctx context.Context, channel *model.Channel) error {
	args := m.Called(ctx, channel)
	return args.Error(0)
}

func (m *mockChannelRepository) Update(ctx context.Context, channel *model.Channel) error {
	args := m.Called(ctx, channel)
	return args.Error(0)
}

func (m *mockChannelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockChannelRepository) ReplaceChannelCharacters(ctx context.Context, channelID uuid.UUID, characterIDs []uuid.UUID) error {
	args := m.Called(ctx, channelID, characterIDs)
	return args.Error(0)
}

type mockEpisodeRepository struct {
	mock.Mock
}

func (m *mockEpisodeRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Episode, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Episode), args.Error(1)
}

func (m *mockEpisodeRepository) FindByChannelID(ctx context.Context, channelID uuid.UUID, filter repository.EpisodeFilter) ([]model.Episode, int64, error) {
	args := m.Called(ctx, channelID, filter)
	return args.Get(0).([]model.Episode), args.Get(1).(int64), args.Error(2)
}

func (m *mockEpisodeRepository) Create(ctx context.Context, episode *model.Episode) error {
	args := m.Called(ctx, episode)
	return args.Error(0)
}

func (m *mockEpisodeRepository) Update(ctx context.Context, episode *model.Episode) error {
	args := m.Called(ctx, episode)
	return args.Error(0)
}

func (m *mockEpisodeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockEpisodeRepository) CountPublishedByChannelIDs(ctx context.Context, channelIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	args := m.Called(ctx, channelIDs)
	return args.Get(0).(map[uuid.UUID]int), args.Error(1)
}

func (m *mockEpisodeRepository) IncrementPlayCount(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
