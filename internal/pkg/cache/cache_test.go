package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_EmptyURL_ReturnsNoopClient(t *testing.T) {
	client, err := New(context.Background(), "")
	require.NoError(t, err)
	assert.IsType(t, &noopClient{}, client)
}

func TestNew_InvalidURL_ReturnsError(t *testing.T) {
	_, err := New(context.Background(), "not-a-valid-url")
	assert.Error(t, err)
}

func TestNoopClient_GetAlwaysMisses(t *testing.T) {
	client := &noopClient{}
	var dest string
	hit, err := client.Get(context.Background(), "key", &dest)
	assert.NoError(t, err)
	assert.False(t, hit)
}

func TestNoopClient_SetDoesNothing(t *testing.T) {
	client := &noopClient{}
	err := client.Set(context.Background(), "key", "value", time.Minute)
	assert.NoError(t, err)
}

func TestNoopClient_DeleteDoesNothing(t *testing.T) {
	client := &noopClient{}
	err := client.Delete(context.Background(), "key1", "key2")
	assert.NoError(t, err)
}

func TestNoopClient_CloseDoesNothing(t *testing.T) {
	client := &noopClient{}
	err := client.Close()
	assert.NoError(t, err)
}
