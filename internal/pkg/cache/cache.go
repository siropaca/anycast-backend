package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

// Client はキャッシュ操作のインターフェース
type Client interface {
	Get(ctx context.Context, key string, dest any) (bool, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	DeleteByPrefix(ctx context.Context, prefix string) error
	Close() error
}

type redisClient struct {
	rdb *redis.Client
}

// New は Redis に接続してキャッシュクライアントを返す。
// redisURL が空の場合は no-op クライアントを返す。
func New(ctx context.Context, redisURL string) (Client, error) {
	if redisURL == "" {
		return &noopClient{}, nil
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(opts)

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	logger.Default().Info("Redis connected", "addr", opts.Addr)
	return &redisClient{rdb: rdb}, nil
}

// Get はキャッシュからキーに対応する値を取得して dest にデコードする。
// キャッシュヒット時は true を返す。
// ミスまたはエラー時は false を返し、エラーはログに記録してフォールバックを促す。
func (c *redisClient) Get(ctx context.Context, key string, dest any) (bool, error) {
	val, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		logger.FromContext(ctx).Warn("cache get failed, falling back to DB", "key", key, "error", err)
		return false, nil
	}

	if err := json.Unmarshal(val, dest); err != nil {
		logger.FromContext(ctx).Warn("cache unmarshal failed", "key", key, "error", err)
		return false, nil
	}

	return true, nil
}

// Set はキーに対応する値を JSON シリアライズしてキャッシュに保存する
func (c *redisClient) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		logger.FromContext(ctx).Warn("cache marshal failed", "key", key, "error", err)
		return nil
	}

	if err := c.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		logger.FromContext(ctx).Warn("cache set failed", "key", key, "error", err)
	}

	return nil
}

// Delete は指定されたキーをキャッシュから削除する
func (c *redisClient) Delete(ctx context.Context, keys ...string) error {
	if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
		logger.FromContext(ctx).Warn("cache delete failed", "keys", keys, "error", err)
	}

	return nil
}

// DeleteByPrefix は指定されたプレフィックスに一致するキーをすべて削除する
func (c *redisClient) DeleteByPrefix(ctx context.Context, prefix string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := c.rdb.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			logger.FromContext(ctx).Warn("cache scan failed", "prefix", prefix, "error", err)
			return nil
		}
		if len(keys) > 0 {
			if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
				logger.FromContext(ctx).Warn("cache delete by prefix failed", "prefix", prefix, "error", err)
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

// Close は Redis 接続を閉じる
func (c *redisClient) Close() error {
	return c.rdb.Close()
}
