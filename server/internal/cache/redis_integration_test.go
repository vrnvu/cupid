//go:build integration
// +build integration

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

// TestRedisCache_Integration runs only when Redis is available
// Run with: go test -tags=integration ./internal/cache/...
func TestRedisCache_Integration(t *testing.T) {
	t.Parallel()

	redisCache := NewRedisCache("localhost:6379")
	ctx := context.Background()

	// Skip if Redis is not available
	if err := redisCache.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}
	defer redisCache.Close()

	// Use random hotel ID to avoid conflicts
	hotelID := rand.Intn(1000000) + 6000000
	reviews := []client.Review{
		{
			ID:           1,
			HotelID:      hotelID,
			ReviewerName: "Integration Test User",
			Rating:       5,
			Title:        "Integration Test Review",
			Content:      "This is an integration test",
			LanguageCode: "en",
			ReviewDate:   "2024-01-01",
			HelpfulVotes: 0,
			CreatedAt:    "2024-01-01T00:00:00Z",
		},
	}

	// Clean up any existing data
	_ = redisCache.DeleteReviews(ctx, hotelID)

	// Test full cycle: Set -> Get -> Delete -> Get
	err := redisCache.SetReviews(ctx, hotelID, reviews, 10*time.Second)
	require.NoError(t, err)

	// Verify they were set
	cachedReviews, err := redisCache.GetReviews(ctx, hotelID)
	assert.NoError(t, err)
	assert.Equal(t, reviews, cachedReviews)

	// Delete them
	err = redisCache.DeleteReviews(ctx, hotelID)
	assert.NoError(t, err)

	// Verify they're gone
	cachedReviews, err = redisCache.GetReviews(ctx, hotelID)
	assert.NoError(t, err)
	assert.Nil(t, cachedReviews)

	// Clean up after test
	t.Cleanup(func() {
		_ = redisCache.DeleteReviews(ctx, hotelID)
	})
}

// TestRedisCache_ConcurrentAccess tests concurrent access to Redis
func TestRedisCache_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	redisCache := NewRedisCache("localhost:6379")
	ctx := context.Background()

	if err := redisCache.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}
	defer redisCache.Close()

	// Test concurrent access with different hotel IDs
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer func() { done <- true }()

			hotelID := rand.Intn(1000000) + 7000000 + goroutineID
			reviews := []client.Review{
				{
					ID:           goroutineID,
					HotelID:      hotelID,
					ReviewerName: "Concurrent Test User",
					Rating:       4,
					Title:        "Concurrent Test Review",
					Content:      "This is a concurrent test",
					LanguageCode: "en",
					ReviewDate:   "2024-01-01",
					HelpfulVotes: 0,
					CreatedAt:    "2024-01-01T00:00:00Z",
				},
			}

			// Clean up any existing data
			_ = redisCache.DeleteReviews(ctx, hotelID)

			// Set reviews
			err := redisCache.SetReviews(ctx, hotelID, reviews, 5*time.Second)
			assert.NoError(t, err)

			// Get reviews
			cachedReviews, err := redisCache.GetReviews(ctx, hotelID)
			assert.NoError(t, err)
			assert.Equal(t, reviews, cachedReviews)

			// Clean up
			_ = redisCache.DeleteReviews(ctx, hotelID)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
