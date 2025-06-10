package cache

import (
	"context"
	"errors" // Added for custom error
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrNotFound is returned when a key is not found in the cache.
var ErrNotFound = errors.New("cache: key not found")

// Cache defines the interface for a caching layer.
type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error // Added Delete for completeness
}

// redisCache implements the Cache interface using Redis.
type redisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis client and returns a Cache implementation.
func NewRedisCache(addr string, password string, db int) (Cache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,     // e.g., "localhost:6379"
		Password: password, // no password set if empty
		DB:       db,       // use default DB
	})

	// Ping the Redis server to ensure connectivity.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	return &redisCache{client: rdb}, nil
}

// Set stores a key-value pair in Redis with an expiration time.
func (rc *redisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return rc.client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value from Redis by key.
// It returns ErrNotFound if the key does not exist.
func (rc *redisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrNotFound // Standardize to our own ErrNotFound
		}
		return "", err
	}
	return val, nil
}

// Delete removes a key from Redis.
func (rc *redisCache) Delete(ctx context.Context, key string) error {
	return rc.client.Del(ctx, key).Err()
}

// Close closes the Redis client connection.
// It's good practice to have a way to close the client, especially for graceful shutdowns.
func (rc *redisCache) Close() error {
	if rc.client != nil {
		return rc.client.Close()
	}
	return nil
}
