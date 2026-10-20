package services

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// The dashboard calls the four /batch endpoints and POST /retrain/auto on every
// load. The batch results come from python-ai's training datasets, not from the
// tenant database, so they only change after a retrain - yet each call cost
// 0.3-1.9s because every load also started a full auto_retrain_all() in
// python-ai. AnalyticsCache keeps the batch bodies in Redis and lets at most one
// auto-retrain through per interval.
//
// Keys carry no tenant prefix: each tenant stack points at its own Redis logical
// DB (REDIS_DB), like the other services. Every Redis failure falls back to
// calling python-ai directly, so a Redis outage only costs speed.

const cacheKeyPrefix = "analytics:v1:"

var errCacheMiss = errors.New("cache miss")

// kvStore is the slice of Redis the cache needs; tests swap in a fake.
type kvStore interface {
	Get(ctx context.Context, key string) ([]byte, error) // errCacheMiss when absent
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
	SetNX(ctx context.Context, key string, ttl time.Duration) (bool, error)
}

type redisStore struct{ client *redis.Client }

func (s redisStore) Get(ctx context.Context, key string) ([]byte, error) {
	b, err := s.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, errCacheMiss
	}
	return b, err
}

func (s redisStore) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	return s.client.Set(ctx, key, val, ttl).Err()
}

func (s redisStore) SetNX(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return s.client.SetNX(ctx, key, time.Now().UTC().Format(time.RFC3339), ttl).Result()
}

type AnalyticsCache struct {
	store           kvStore // nil = caching disabled
	batchTTL        time.Duration
	retrainInterval time.Duration
}

// NewAnalyticsCacheFromEnv connects to Redis only when REDIS_HOST is set, so
// local runs and unit tests without Redis keep calling python-ai directly.
func NewAnalyticsCacheFromEnv() *AnalyticsCache {
	c := &AnalyticsCache{
		batchTTL:        envDuration("ANALYTICS_BATCH_CACHE_TTL", 15*time.Minute),
		retrainInterval: envDuration("ANALYTICS_RETRAIN_MIN_INTERVAL", 6*time.Hour),
	}

	host := os.Getenv("REDIS_HOST")
	if host == "" {
		log.Println("[analytics-cache] REDIS_HOST not set - batch cache and retrain throttle disabled")
		return c
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

	client := redis.NewClient(&redis.Options{
		Addr:         host + ":" + port,
		Password:     os.Getenv("REDIS_PASSWORD"),
		DB:           db,
		DialTimeout:  time.Second,
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 500 * time.Millisecond,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("[analytics-cache] Redis %s:%s db=%d unreachable (%v) - cache disabled", host, port, db, err)
		_ = client.Close()
		return c
	}
	log.Printf("[analytics-cache] Redis %s:%s db=%d - batch TTL %s, retrain at most every %s", host, port, db, c.batchTTL, c.retrainInterval)
	c.store = redisStore{client: client}
	return c
}

// GetOrLoad returns the cached body for name, or calls load and caches its
// result. Only valid JSON from a successful load is cached.
func (c *AnalyticsCache) GetOrLoad(name string, load func() ([]byte, error)) ([]byte, error) {
	if c == nil || c.store == nil || c.batchTTL <= 0 {
		return load()
	}
	key := cacheKeyPrefix + "batch:" + name

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if b, err := c.store.Get(ctx, key); err == nil {
		return b, nil
	} else if !errors.Is(err, errCacheMiss) {
		log.Printf("[analytics-cache] get %s: %v", key, err)
	}

	b, err := load()
	if err != nil || !json.Valid(b) {
		return b, err
	}
	setCtx, setCancel := context.WithTimeout(context.Background(), time.Second)
	defer setCancel()
	if err := c.store.Set(setCtx, key, b, c.batchTTL); err != nil {
		log.Printf("[analytics-cache] set %s: %v", key, err)
	}
	return b, nil
}

// AllowRetrain reports whether an auto-retrain may start now and, if so,
// claims the slot for retrainInterval. Without Redis every call is allowed,
// which is the old behaviour.
func (c *AnalyticsCache) AllowRetrain() bool {
	if c == nil || c.store == nil || c.retrainInterval <= 0 {
		return true
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ok, err := c.store.SetNX(ctx, cacheKeyPrefix+"retrain:auto:last", c.retrainInterval)
	if err != nil {
		log.Printf("[analytics-cache] retrain throttle: %v", err)
		return true
	}
	return ok
}

func envDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		log.Printf("[analytics-cache] invalid %s=%q, using %s", key, v, def)
		return def
	}
	return d
}
