package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisCache implements distributed Redis cache
type RedisCache struct {
	client     *redis.Client
	defaultTTL time.Duration
	logger     *zap.Logger
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(addresses []string, password string, database int, defaultTTL time.Duration, poolSize int, maxRetries int, logger *zap.Logger) (*RedisCache, error) {
	// For simplicity, using single-node Redis client
	// In production, use Redis Cluster for high availability
	client := redis.NewClient(&redis.Options{
		Addr:       addresses[0], // Use first address
		Password:   password,
		DB:         database,
		PoolSize:   poolSize,
		MaxRetries: maxRetries,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	logger.Info("redis cache initialized",
		zap.Strings("addresses", addresses),
		zap.Int("database", database),
		zap.Duration("default_ttl", defaultTTL),
	)

	return &RedisCache{
		client:     client,
		defaultTTL: defaultTTL,
		logger:     logger,
	}, nil
}

// Get retrieves a value from cache
func (rc *RedisCache) Get(ctx context.Context, key string) (string, error) {
	value, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrCacheMiss
		}
		rc.logger.Error("redis get error", zap.Error(err), zap.String("key", key))
		return "", err
	}

	return value, nil
}

// Set stores a value in cache with TTL
func (rc *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if ttl == 0 {
		ttl = rc.defaultTTL
	}

	if err := rc.client.Set(ctx, key, value, ttl).Err(); err != nil {
		rc.logger.Error("redis set error", zap.Error(err), zap.String("key", key))
		return err
	}

	return nil
}

// Delete removes a value from cache
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	if err := rc.client.Del(ctx, key).Err(); err != nil {
		rc.logger.Error("redis delete error", zap.Error(err), zap.String("key", key))
		return err
	}

	return nil
}

// DeletePattern removes all keys matching a pattern
func (rc *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	// Scan for matching keys
	var cursor uint64
	var keys []string

	for {
		var batch []string
		var err error

		batch, cursor, err = rc.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			rc.logger.Error("redis scan error", zap.Error(err), zap.String("pattern", pattern))
			return err
		}

		keys = append(keys, batch...)

		if cursor == 0 {
			break
		}
	}

	// Delete matching keys
	if len(keys) > 0 {
		if err := rc.client.Del(ctx, keys...).Err(); err != nil {
			rc.logger.Error("redis delete pattern error", zap.Error(err), zap.String("pattern", pattern))
			return err
		}

		rc.logger.Info("deleted keys by pattern",
			zap.String("pattern", pattern),
			zap.Int("count", len(keys)),
		)
	}

	return nil
}

// Exists checks if a key exists in cache
func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := rc.client.Exists(ctx, key).Result()
	if err != nil {
		rc.logger.Error("redis exists error", zap.Error(err), zap.String("key", key))
		return false, err
	}

	return count > 0, nil
}

// Close closes the Redis connection
func (rc *RedisCache) Close() error {
	rc.logger.Info("redis cache closing")
	return rc.client.Close()
}

// Ping tests Redis connectivity
func (rc *RedisCache) Ping(ctx context.Context) error {
	return rc.client.Ping(ctx).Err()
}

// Stats returns Redis cache statistics
func (rc *RedisCache) Stats(ctx context.Context) map[string]interface{} {
	info, err := rc.client.Info(ctx, "stats").Result()
	if err != nil {
		rc.logger.Error("redis stats error", zap.Error(err))
		return nil
	}

	return map[string]interface{}{
		"info": info,
	}
}
