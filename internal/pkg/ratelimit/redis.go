package ratelimit

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

type RedisRateLimiter struct {
	redisdb   *redis.Client
	secWindow int64
	url       string
}

func NewRedisRateLimiter(url string, secWindow int64) (*RedisRateLimiter, error) {
	if secWindow <= 0 {
		return nil, fmt.Errorf("secWindow must be > 0")
	}
	if url == "" {
		return nil, fmt.Errorf("no redis url")
	}
	redisdb := redis.NewClient(&redis.Options{
		Addr:               url,
		MaxRetries:         3,
		MinIdleConns:       2,
		IdleTimeout:        5 * time.Minute,
		IdleCheckFrequency: time.Minute,
		PoolSize:           30,
	})
	return &RedisRateLimiter{redisdb: redisdb, secWindow: secWindow, url: url}, nil
}

func (r *RedisRateLimiter) Validate(key string, limit int64, quota int64) (bool, int64, int64, error) {
	now := time.Now().Unix()
	at := now / r.secWindow
	key = fmt.Sprintf("%s:%d", key, at)
	val, err := r.redisdb.Get(key).Result()
	if err != nil && err != redis.Nil {
		return false, 0, 0, fmt.Errorf("can't get key: %v", err)
	}
	valInt, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		valInt = 0
	}
	valInt = valInt + quota
	if valInt >= limit {
		return false, 0, ((at + 1) * r.secWindow) - now, nil
	}

	_, err = r.redisdb.TxPipelined(func(pipe redis.Pipeliner) error {
		_ = pipe.IncrBy(key, quota)
		pipe.Expire(key, time.Second*time.Duration(r.secWindow))
		return nil
	})
	if err != nil {
		return false, 0, 0, fmt.Errorf("can't update rate limiter: %v", err)
	}
	return true, limit - valInt, 0, nil
}

func (r *RedisRateLimiter) Info(pr string) string {
	return pr + fmt.Sprintf("RedisRateLimiter(%s, %d)", r.url, r.secWindow)
}
