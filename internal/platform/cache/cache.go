package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	// ErrCacheMiss is returned when a key is not found in the cache
	ErrCacheMiss = errors.New("cache miss")
	// ErrCacheDisabled is returned when the cache is not configured
	ErrCacheDisabled = errors.New("cache disabled")
)

// Config holds the Redis cache configuration
type Config struct {
	Addr     string
	Password string
	DB       int
	Prefix   string // Key prefix for namespacing
}

// Cache defines the interface for cache operations
type Cache interface {
	// Get retrieves a value from the cache
	Get(ctx context.Context, key string) (string, error)
	// Set stores a value in the cache with TTL
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error
	// Exists checks if a key exists in the cache
	Exists(ctx context.Context, key string) (bool, error)
	// Clear removes all keys with the configured prefix
	Clear(ctx context.Context) error
	// IsEnabled returns whether the cache is enabled
	IsEnabled() bool
}

// RedisCache implements the Cache interface using Redis
type RedisCache struct {
	client *redis.Client
	prefix string
}

// NewRedisCache creates a new Redis cache client
func NewRedisCache(cfg Config) (Cache, error) {
	if cfg.Addr == "" {
		// Return disabled cache if no address configured
		return &disabledCache{}, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		prefix: cfg.Prefix,
	}, nil
}

func (c *RedisCache) key(k string) string {
	if c.prefix == "" {
		return k
	}
	return c.prefix + ":" + k
}

func (c *RedisCache) IsEnabled() bool {
	return true
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, c.key(key)).Result()
	if err == redis.Nil {
		return "", ErrCacheMiss
	}
	if err != nil {
		return "", fmt.Errorf("cache get error: %w", err)
	}
	return val, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if err := c.client.Set(ctx, c.key(key), value, ttl).Err(); err != nil {
		return fmt.Errorf("cache set error: %w", err)
	}
	return nil
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, c.key(key)).Err(); err != nil {
		return fmt.Errorf("cache delete error: %w", err)
	}
	return nil
}

func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, c.key(key)).Result()
	if err != nil {
		return false, fmt.Errorf("cache exists error: %w", err)
	}
	return result > 0, nil
}

func (c *RedisCache) Clear(ctx context.Context) error {
	if c.prefix == "" {
		return errors.New("cannot clear cache without prefix")
	}

	// Find all keys with prefix
	iter := c.client.Scan(ctx, 0, c.prefix+":*", 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("cache clear error: %w", err)
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("cache clear error: %w", err)
	}
	return nil
}

func (c *RedisCache) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// disabledCache is a no-op implementation for when cache is disabled
type disabledCache struct{}

func (d *disabledCache) IsEnabled() bool                              { return false }
func (d *disabledCache) Get(ctx context.Context, key string) (string, error) {
	return "", ErrCacheDisabled
}
func (d *disabledCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return ErrCacheDisabled
}
func (d *disabledCache) Delete(ctx context.Context, key string) error {
	return ErrCacheDisabled
}
func (d *disabledCache) Exists(ctx context.Context, key string) (bool, error) {
	return false, ErrCacheDisabled
}
func (d *disabledCache) Clear(ctx context.Context) error {
	return ErrCacheDisabled
}

// Cache keys for RBAC
const (
	CacheKeyUserRoles      = "rbac:user:roles:%s"      // user_id
	CacheKeyRolePerms      = "rbac:role:perms:%s:%s"   // tenant_id:role_id
	CacheKeyResourcePerms  = "rbac:resource:perms:%s"  // resource_name
	CacheKeyUserPerms      = "rbac:user:perms:%s:%s"   // tenant_id:user_id
	CacheDefaultTTL        = 15 * time.Minute
	CacheShortTTL          = 5 * time.Minute
	CacheLongTTL           = 1 * time.Hour
)
