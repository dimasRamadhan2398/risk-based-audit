package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Repository provides common Redis operations with JSON serialization
type Repository struct {
	client *Client
}

// NewRepository creates a new Redis repository
func NewRepository(client *Client) *Repository {
	return &Repository{client: client}
}

// Set stores a value with optional expiration
func (r *Repository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := r.client.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

// Get retrieves a value and unmarshals it into the target
func (r *Repository) Get(ctx context.Context, key string, target interface{}) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrKeyNotFound
		}
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
	}

	return nil
}

// GetRaw retrieves raw bytes for a key
func (r *Repository) GetRaw(ctx context.Context, key string) ([]byte, error) {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrKeyNotFound
		}
		return nil, fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return data, nil
}

// SetRaw stores raw bytes with optional expiration
func (r *Repository) SetRaw(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	if err := r.client.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set raw key %s: %w", key, err)
	}
	return nil
}

// Delete removes a key
func (r *Repository) Delete(ctx context.Context, keys ...string) error {
	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}
	return nil
}

// Exists checks if keys exist
func (r *Repository) Exists(ctx context.Context, keys ...string) (int64, error) {
	count, err := r.client.Exists(ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to check existence: %w", err)
	}
	return count, nil
}

// Expire sets expiration on a key
func (r *Repository) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if err := r.client.Expire(ctx, key, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set expiration on key %s: %w", key, err)
	}
	return nil
}

// TTL returns the time to live for a key
func (r *Repository) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for key %s: %w", key, err)
	}
	return ttl, nil
}

// Keys returns all keys matching a pattern
func (r *Repository) Keys(ctx context.Context, pattern string) ([]string, error) {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %w", err)
	}
	return keys, nil
}

// Increment atomically increments a key
func (r *Repository) Increment(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}
	return val, nil
}

// Decrement atomically decrements a key
func (r *Repository) Decrement(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Decr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement key %s: %w", key, err)
	}
	return val, nil
}

// IncrementBy atomically increments a key by a delta
func (r *Repository) IncrementBy(ctx context.Context, key string, delta int64) (int64, error) {
	val, err := r.client.IncrBy(ctx, key, delta).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s by %d: %w", key, delta, err)
	}
	return val, nil
}

// SetNX sets a key only if it doesn't exist (for distributed locks)
func (r *Repository) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	ok, err := r.client.SetNX(ctx, key, data, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to setNX key %s: %w", key, err)
	}
	return ok, nil
}

// SetNXRaw sets raw bytes only if key doesn't exist
func (r *Repository) SetNXRaw(ctx context.Context, key string, value []byte, expiration time.Duration) (bool, error) {
	ok, err := r.client.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to setNX raw key %s: %w", key, err)
	}
	return ok, nil
}

// HashSet sets hash fields
func (r *Repository) HashSet(ctx context.Context, key string, values map[string]interface{}) error {
	if err := r.client.HSet(ctx, key, values).Err(); err != nil {
		return fmt.Errorf("failed to set hash %s: %w", key, err)
	}
	return nil
}

// HashGet gets a hash field
func (r *Repository) HashGet(ctx context.Context, key, field string) (string, error) {
	val, err := r.client.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrKeyNotFound
		}
		return "", fmt.Errorf("failed to get hash field %s:%s: %w", key, field, err)
	}
	return val, nil
}

// HashGetAll gets all hash fields
func (r *Repository) HashGetAll(ctx context.Context, key string) (map[string]string, error) {
	val, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get all hash fields for %s: %w", key, err)
	}
	return val, nil
}

// HashDelete deletes hash fields
func (r *Repository) HashDelete(ctx context.Context, key string, fields ...string) error {
	if err := r.client.HDel(ctx, key, fields...).Err(); err != nil {
		return fmt.Errorf("failed to delete hash fields from %s: %w", key, err)
	}
	return nil
}

// HashExists checks if a hash field exists
func (r *Repository) HashExists(ctx context.Context, key, field string) (bool, error) {
	exists, err := r.client.HExists(ctx, key, field).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check hash field existence %s:%s: %w", key, field, err)
	}
	return exists, nil
}

// ListPush pushes values to a list
func (r *Repository) ListPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	length, err := r.client.RPush(ctx, key, values...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to push to list %s: %w", key, err)
	}
	return length, nil
}

// ListRange returns list elements in a range
func (r *Repository) ListRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	elements, err := r.client.LRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get list range %s: %w", key, err)
	}
	return elements, nil
}

// ListLength returns the length of a list
func (r *Repository) ListLength(ctx context.Context, key string) (int64, error) {
	length, err := r.client.LLen(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get list length %s: %w", key, err)
	}
	return length, nil
}

// SortedSetAdd adds members to a sorted set
func (r *Repository) SortedSetAdd(ctx context.Context, key string, members ...redis.Z) (int64, error) {
	count, err := r.client.ZAdd(ctx, key, members...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to add to sorted set %s: %w", key, err)
	}
	return count, nil
}

// SortedSetRangeByScore returns members by score range
func (r *Repository) SortedSetRangeByScore(ctx context.Context, key string, min, max string, offset, count int64) ([]string, error) {
	members, err := r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: offset,
		Count:  count,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to range sorted set %s: %w", key, err)
	}
	return members, nil
}

// SortedSetRemove removes members from a sorted set
func (r *Repository) SortedSetRemove(ctx context.Context, key string, members ...interface{}) (int64, error) {
	removed, err := r.client.ZRem(ctx, key, members...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to remove from sorted set %s: %w", key, err)
	}
	return removed, nil
}

// Ping checks if the Redis connection is alive
func (r *Repository) Ping(ctx context.Context) error {
	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping redis: %w", err)
	}
	return nil
}

// Client returns the underlying Redis client for advanced operations
func (r *Repository) Client() *Client {
	return r.client
}

// ErrKeyNotFound is returned when a key doesn't exist
var ErrKeyNotFound = fmt.Errorf("key not found")

// Helper functions for common patterns

// CacheWithFunc is a helper for cache-aside pattern
func (r *Repository) CacheWithFunc(ctx context.Context, key string, expiration time.Duration, fetchFn func() (interface{}, error), target interface{}) error {
	// Try to get from cache
	err := r.Get(ctx, key, target)
	if err == nil {
		return nil // Cache hit
	}
	if err != ErrKeyNotFound {
		// Some other error, log and continue to fetch
	}

	// Cache miss, fetch data
	data, err := fetchFn()
	if err != nil {
		return err
	}

	// Store in cache
	if err := r.Set(ctx, key, data, expiration); err != nil {
		// Log error but don't fail
		return nil
	}

	// Unmarshal into target
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(dataBytes, target)
}
