//go:build integration

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

func TestRedisCache_Integration(t *testing.T) {
	t.Parallel()

	redisCache := NewRedisCache("localhost:6379")
	ctx := context.Background()

	if err := redisCache.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}
	defer redisCache.Close()

	// Generate random data - don't assume clean state
	hotelID := rand.Intn(1000000) + 1000000   //nolint:gosec // Test data only
	reviewID1 := rand.Intn(1000000) + 1000000 //nolint:gosec // Test data only
	reviewID2 := rand.Intn(1000000) + 2000000 //nolint:gosec // Test data only

	expectedReviews := []client.Review{
		{
			ID:           reviewID1,
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
			ID:           reviewID2,
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

	// Don't assume clean state - clean up after ourselves
	t.Cleanup(func() {
		_ = redisCache.DeleteReviews(ctx, hotelID)
	})

	// Test cache miss - don't assume it's empty
	reviews, err := redisCache.GetReviews(ctx, hotelID)
	assert.NoError(t, err)
	// Note: we don't assert it's nil because the database might be dirty

	// Test set and get
	err = redisCache.SetReviews(ctx, hotelID, expectedReviews, 5*time.Second)
	require.NoError(t, err)

	reviews, err = redisCache.GetReviews(ctx, hotelID)
	assert.NoError(t, err)
	assert.Equal(t, expectedReviews, reviews)

	// Test deletion
	err = redisCache.DeleteReviews(ctx, hotelID)
	assert.NoError(t, err)

	reviews, err = redisCache.GetReviews(ctx, hotelID)
	assert.NoError(t, err)
	assert.Nil(t, reviews)
}

func TestRedisCache_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	redisCache := NewRedisCache("localhost:6379")
	ctx := context.Background()

	if err := redisCache.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}
	defer redisCache.Close()

	// Generate random data - don't assume clean state
	hotelID := rand.Intn(1000000) + 2000000 //nolint:gosec // Test data only
	reviewID := rand.Intn(1000000) + 3000000

	reviews := []client.Review{
		{
			ID:           reviewID,
			HotelID:      hotelID,
			ReviewerName: "Concurrent User",
			Rating:       5,
			Title:        "Concurrent Test",
			Content:      "Testing concurrent access",
			LanguageCode: "en",
			ReviewDate:   "2024-01-01",
			HelpfulVotes: 0,
			CreatedAt:    "2024-01-01T00:00:00Z",
		},
	}

	// Don't assume clean state - clean up after ourselves
	t.Cleanup(func() {
		_ = redisCache.DeleteReviews(ctx, hotelID)
	})

	// Test concurrent reads and writes
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Set reviews
			err := redisCache.SetReviews(ctx, hotelID, reviews, 10*time.Second)
			assert.NoError(t, err)

			// Get reviews
			cachedReviews, err := redisCache.GetReviews(ctx, hotelID)
			assert.NoError(t, err)
			assert.Equal(t, reviews, cachedReviews)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
