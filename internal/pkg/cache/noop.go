package cache

import (
	"context"
	"time"
)

// noopClient は Redis 未設定時に使用する no-op キャッシュクライアント。
// すべての操作を何もせずに返し、常にキャッシュミスとなる。
type noopClient struct{}

func (c *noopClient) Get(_ context.Context, _ string, _ any) (bool, error) {
	return false, nil
}

func (c *noopClient) Set(_ context.Context, _ string, _ any, _ time.Duration) error {
	return nil
}

func (c *noopClient) Delete(_ context.Context, _ ...string) error {
	return nil
}

func (c *noopClient) DeleteByPrefix(_ context.Context, _ string) error {
	return nil
}

func (c *noopClient) Close() error {
	return nil
}
