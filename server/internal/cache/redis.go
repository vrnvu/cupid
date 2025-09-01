package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vrnvu/cupid/internal/client"
)

// ReviewCache defines the interface for caching hotel reviews
type ReviewCache interface {
	GetReviews(ctx context.Context, hotelID int) ([]client.Review, error)
	SetReviews(ctx context.Context, hotelID int, reviews []client.Review, ttl time.Duration) error
	DeleteReviews(ctx context.Context, hotelID int) error
	Ping(ctx context.Context) error
	Close() error
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr string) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		PoolSize:     10,
		MinIdleConns: 2,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	return &RedisCache{client: rdb}
}

func (r *RedisCache) GetReviews(ctx context.Context, hotelID int) ([]client.Review, error) {
	key := fmt.Sprintf("reviews:hotel:%d", hotelID)

	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	var reviews []client.Review
	if err := json.Unmarshal([]byte(val), &reviews); err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}

	return reviews, nil
}

func (r *RedisCache) SetReviews(ctx context.Context, hotelID int, reviews []client.Review, ttl time.Duration) error {
	key := fmt.Sprintf("reviews:hotel:%d", hotelID)

	data, err := json.Marshal(reviews)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisCache) DeleteReviews(ctx context.Context, hotelID int) error {
	key := fmt.Sprintf("reviews:hotel:%d", hotelID)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}
