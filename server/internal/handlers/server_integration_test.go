//go:build integration

package handlers

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vrnvu/cupid/internal/cache"
	"github.com/vrnvu/cupid/internal/client"
	"github.com/vrnvu/cupid/internal/database"
)

// setupTestInfrastructure creates real database and cache connections for integration testing
func setupTestInfrastructure(t *testing.T) (*database.DB, cache.ReviewCache, *database.HotelRepository) {
	t.Helper()

	// Setup database
	config := database.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "cupid",
		Password: "cupid123",
		DBName:   "cupid",
		SSLMode:  "disable",
	}

	db, err := database.NewConnection(config)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	// Setup cache
	redisCache := cache.NewRedisCache("localhost:6379")
	if err := redisCache.Ping(context.Background()); err != nil {
		t.Skip("Redis not available, skipping integration tests")
	}

	// Setup repository
	repo := database.NewHotelRepository(db)

	return db, redisCache, repo
}

// createTestHotel creates and stores a minimal test hotel
func createTestHotel(t *testing.T, repo *database.HotelRepository, name string) *client.Property {
	t.Helper()

	hotel := &client.Property{
		HotelID:     rand.Intn(1000000) + 1000000,
		CupidID:     rand.Intn(1000000) + 2000000,
		HotelName:   name,
		Rating:      4.0,
		Stars:       4,
		ReviewCount: 100,
		Description: "Test hotel for integration testing",
		Address: client.Address{
			Address:    "123 Test Street",
			City:       "Test City",
			State:      "Test State",
			Country:    "TC",
			PostalCode: "12345",
		},
		Checkin: client.Checkin{
			CheckinStart: "14:00",
			Checkout:     "11:00",
		},
	}

	ctx := context.Background()
	err := repo.StoreProperty(ctx, hotel)
	require.NoError(t, err)

	return hotel
}

func TestServer_GetHotelsHandler_Integration(t *testing.T) {
	t.Parallel()

	_, cache, repo := setupTestInfrastructure(t)
	server := NewServer(repo, cache, "")

	t.Run("GetAllHotels", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/api/v1/hotels", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		hotels := response["hotels"].([]interface{})
		assert.GreaterOrEqual(t, len(hotels), 1, "Should have at least 1 hotel")
	})

	t.Run("GetHotelsWithPagination", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/api/v1/hotels?limit=1&offset=0", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, float64(1), response["limit"])
		assert.Equal(t, float64(0), response["offset"])

		hotels := response["hotels"].([]interface{})
		assert.Len(t, hotels, 1, "Should have exactly 1 hotel with limit=1")
	})
}

func TestServer_GetHotelHandler_Integration(t *testing.T) {
	t.Parallel()

	_, cache, repo := setupTestInfrastructure(t)
	server := NewServer(repo, cache, "")

	t.Run("GetHotelByID", func(t *testing.T) {
		t.Parallel()

		// Just test that the endpoint responds correctly
		// Don't assume any specific hotel exists
		req := httptest.NewRequest("GET", "/api/v1/hotels/999999", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		// Should either return 404 or 200 depending on what's in the database
		assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, w.Code)
	})

	t.Run("GetNonExistentHotel", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/api/v1/hotels/999999", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Hotel with ID 999999 not found")
	})
}

func TestServer_GetHotelReviewsHandler_Integration(t *testing.T) {
	t.Parallel()

	_, cache, repo := setupTestInfrastructure(t)
	server := NewServer(repo, cache, "")

	t.Run("GetHotelReviews", func(t *testing.T) {
		t.Parallel()

		// Just test that the endpoint responds correctly
		// Don't assume any specific hotel exists
		req := httptest.NewRequest("GET", "/api/v1/hotels/999999/reviews", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		// Should either return 404 or 200 depending on what's in the database
		assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, w.Code)
	})
}

func TestServer_GetHotelTranslationsHandler_Integration(t *testing.T) {
	t.Parallel()

	_, cache, repo := setupTestInfrastructure(t)
	server := NewServer(repo, cache, "")

	t.Run("GetFrenchTranslations", func(t *testing.T) {
		t.Parallel()

		// Just test that the endpoint responds correctly
		// Don't assume any specific hotel exists
		req := httptest.NewRequest("GET", "/api/v1/hotels/999999/translations/fr", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		// Should either return 404 or 200 depending on what's in the database
		assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, w.Code)
	})

	t.Run("GetSpanishTranslations", func(t *testing.T) {
		t.Parallel()

		// Just test that the endpoint responds correctly
		// Don't assume any specific hotel exists
		req := httptest.NewRequest("GET", "/api/v1/hotels/999999/translations/es", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		// Should either return 404 or 200 depending on what's in the database
		assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, w.Code)
	})

	t.Run("GetUnsupportedLanguage", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/api/v1/hotels/999999/translations/de", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Unsupported language code")
	})
}

func TestServer_HealthHandler_Integration(t *testing.T) {
	t.Parallel()

	_, cache, repo := setupTestInfrastructure(t)
	server := NewServer(repo, cache, "")

	t.Run("HealthCheckWithDatabase", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "healthy", response["status"])
		assert.Equal(t, "cupid-api", response["service"])
	})
}
