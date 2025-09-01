package cache

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vrnvu/cupid/internal/client"
)

func TestRedisCache_GetReviews(t *testing.T) {
	t.Parallel()

	redisCache := NewRedisCache("localhost:6379")
	ctx := context.Background()

	if err := redisCache.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer redisCache.Close()

	hotelID := rand.Intn(1000000) + 1000000 //nolint:gosec // Test data only
	expectedReviews := []client.Review{
		{
			ID:           1,
			HotelID:      hotelID,
			ReviewerName: "John Doe",
			Rating:       5,
			Title:        "Great hotel!",
			Content:      "Amazing experience",
			LanguageCode: "en",
			ReviewDate:   "2024-01-15",
			HelpfulVotes: 10,
			CreatedAt:    "2024-01-15T10:00:00Z",
		},
		{
			ID:           2,
			HotelID:      hotelID,
			ReviewerName: "Jane Smith",
			Rating:       4,
			Title:        "Good stay",
			Content:      "Nice hotel with good service",
			LanguageCode: "en",
			ReviewDate:   "2024-01-14",
			HelpfulVotes: 5,
			CreatedAt:    "2024-01-14T15:30:00Z",
		},
	}

	_ = redisCache.DeleteReviews(ctx, hotelID)

	reviews, err := redisCache.GetReviews(ctx, hotelID)
	assert.NoError(t, err)
	assert.Nil(t, reviews)

	err = redisCache.SetReviews(ctx, hotelID, expectedReviews, 5*time.Second)
	require.NoError(t, err)

	reviews, err = redisCache.GetReviews(ctx, hotelID)
	assert.NoError(t, err)
	assert.Equal(t, expectedReviews, reviews)

	t.Cleanup(func() {
		_ = redisCache.DeleteReviews(ctx, hotelID)
	})
}

func TestRedisCache_SetReviews(t *testing.T) {
	t.Parallel()

	redisCache := NewRedisCache("localhost:6379")
	ctx := context.Background()

	if err := redisCache.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer redisCache.Close()

	hotelID := rand.Intn(1000000) + 2000000 //nolint:gosec // Test data only
	reviews := []client.Review{
		{
			ID:           1,
			HotelID:      hotelID,
			ReviewerName: "Test User",
			Rating:       5,
			Title:        "Test Review",
			Content:      "Test content",
			LanguageCode: "en",
			ReviewDate:   "2024-01-01",
			HelpfulVotes: 0,
			CreatedAt:    "2024-01-01T00:00:00Z",
		},
	}

	_ = redisCache.DeleteReviews(ctx, hotelID)

	err := redisCache.SetReviews(ctx, hotelID, reviews, 5*time.Second)
	assert.NoError(t, err)

	cachedReviews, err := redisCache.GetReviews(ctx, hotelID)
	assert.NoError(t, err)
	assert.Equal(t, reviews, cachedReviews)

	t.Cleanup(func() {
		_ = redisCache.DeleteReviews(ctx, hotelID)
	})
}

func TestRedisCache_DeleteReviews(t *testing.T) {
	t.Parallel()

	redisCache := NewRedisCache("localhost:6379")
	ctx := context.Background()

	if err := redisCache.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer redisCache.Close()

	hotelID := rand.Intn(1000000) + 3000000 //nolint:gosec // Test data only
	reviews := []client.Review{
		{
			ID:           1,
			HotelID:      hotelID,
			ReviewerName: "Delete Test",
			Rating:       3,
			Title:        "Will be deleted",
			Content:      "This will be deleted",
			LanguageCode: "en",
			ReviewDate:   "2024-01-01",
			HelpfulVotes: 0,
			CreatedAt:    "2024-01-01T00:00:00Z",
		},
	}

	_ = redisCache.DeleteReviews(ctx, hotelID)

	err := redisCache.SetReviews(ctx, hotelID, reviews, 5*time.Second)
	require.NoError(t, err)

	cachedReviews, err := redisCache.GetReviews(ctx, hotelID)
	assert.NoError(t, err)
	assert.Equal(t, reviews, cachedReviews)

	err = redisCache.DeleteReviews(ctx, hotelID)
	assert.NoError(t, err)

	cachedReviews, err = redisCache.GetReviews(ctx, hotelID)
	assert.NoError(t, err)
	assert.Nil(t, cachedReviews)

	t.Cleanup(func() {
		_ = redisCache.DeleteReviews(ctx, hotelID)
	})
}

func TestRedisCache_Ping(t *testing.T) {
	t.Parallel()

	redisCache := NewRedisCache("localhost:6379")
	ctx := context.Background()
	defer redisCache.Close()

	err := redisCache.Ping(ctx)
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}

	assert.NoError(t, err)
}

func TestRedisCache_Close(t *testing.T) {
	t.Parallel()

	redisCache := NewRedisCache("localhost:6379")
	ctx := context.Background()

	if err := redisCache.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping test")
	}

	err := redisCache.Close()
	assert.NoError(t, err)
}

func TestRedisCache_EmptyReviews(t *testing.T) {
	t.Parallel()

	redisCache := NewRedisCache("localhost:6379")
	ctx := context.Background()

	if err := redisCache.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer redisCache.Close()

	hotelID := rand.Intn(1000000) + 4000000 //nolint:gosec // Test data only
	emptyReviews := []client.Review{}

	_ = redisCache.DeleteReviews(ctx, hotelID)

	err := redisCache.SetReviews(ctx, hotelID, emptyReviews, 5*time.Second)
	assert.NoError(t, err)

	reviews, err := redisCache.GetReviews(ctx, hotelID)
	assert.NoError(t, err)
	assert.Equal(t, emptyReviews, reviews)

	t.Cleanup(func() {
		_ = redisCache.DeleteReviews(ctx, hotelID)
	})
}

func TestRedisCache_TTL(t *testing.T) {
	t.Parallel()

	redisCache := NewRedisCache("localhost:6379")
	ctx := context.Background()

	if err := redisCache.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer redisCache.Close()

	hotelID := rand.Intn(1000000) + 5000000 //nolint:gosec // Test data only
	reviews := []client.Review{
		{
			ID:           1,
			HotelID:      hotelID,
			ReviewerName: "TTL Test",
			Rating:       5,
			Title:        "TTL Test Review",
			Content:      "This will expire",
			LanguageCode: "en",
			ReviewDate:   "2024-01-01",
			HelpfulVotes: 0,
			CreatedAt:    "2024-01-01T00:00:00Z",
		},
	}

	_ = redisCache.DeleteReviews(ctx, hotelID)

	err := redisCache.SetReviews(ctx, hotelID, reviews, 100*time.Millisecond)
	require.NoError(t, err)

	cachedReviews, err := redisCache.GetReviews(ctx, hotelID)
	assert.NoError(t, err)
	assert.Equal(t, reviews, cachedReviews)

	time.Sleep(200 * time.Millisecond)

	cachedReviews, err = redisCache.GetReviews(ctx, hotelID)
	assert.NoError(t, err)
	assert.Nil(t, cachedReviews)

	t.Cleanup(func() {
		_ = redisCache.DeleteReviews(ctx, hotelID)
	})
}
