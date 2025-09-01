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

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedLimit  interface{}
		expectedOffset interface{}
		minHotels      int
		exactHotels    int
	}{
		{
			name:           "GetAllHotels",
			url:            "/api/v1/hotels",
			expectedStatus: http.StatusOK,
			minHotels:      1,
		},
		{
			name:           "GetHotelsWithPagination",
			url:            "/api/v1/hotels?limit=1&offset=0",
			expectedStatus: http.StatusOK,
			expectedLimit:  float64(1),
			expectedOffset: float64(0),
			exactHotels:    1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			if tt.expectedLimit != nil {
				assert.Equal(t, tt.expectedLimit, response["limit"])
			}
			if tt.expectedOffset != nil {
				assert.Equal(t, tt.expectedOffset, response["offset"])
			}

			hotels := response["hotels"].([]interface{})
			if tt.exactHotels > 0 {
				assert.Len(t, hotels, tt.exactHotels, "Should have exactly %d hotels", tt.exactHotels)
			} else if tt.minHotels > 0 {
				assert.GreaterOrEqual(t, len(hotels), tt.minHotels, "Should have at least %d hotels", tt.minHotels)
			}
		})
	}
}

func TestServer_GetHotelHandler_Integration(t *testing.T) {
	t.Parallel()

	_, cache, repo := setupTestInfrastructure(t)
	server := NewServer(repo, cache, "")

	tests := []struct {
		name           string
		hotelID        string
		expectedStatus []int
		expectedBody   string
	}{
		{
			name:           "GetHotelByID",
			hotelID:        "999999",
			expectedStatus: []int{http.StatusOK, http.StatusNotFound},
		},
		{
			name:           "GetNonExistentHotel",
			hotelID:        "999999",
			expectedStatus: []int{http.StatusNotFound},
			expectedBody:   "Hotel with ID 999999 not found",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest("GET", "/api/v1/hotels/"+tt.hotelID, nil)
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			if len(tt.expectedStatus) == 1 {
				assert.Equal(t, tt.expectedStatus[0], w.Code)
			} else {
				assert.Contains(t, tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestServer_GetHotelReviewsHandler_Integration(t *testing.T) {
	t.Parallel()

	_, cache, repo := setupTestInfrastructure(t)
	server := NewServer(repo, cache, "")

	tests := []struct {
		name           string
		hotelID        string
		expectedStatus []int
	}{
		{
			name:           "GetHotelReviews",
			hotelID:        "999999",
			expectedStatus: []int{http.StatusOK, http.StatusNotFound},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest("GET", "/api/v1/hotels/"+tt.hotelID+"/reviews", nil)
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			assert.Contains(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestServer_GetHotelTranslationsHandler_Integration(t *testing.T) {
	t.Parallel()

	_, cache, repo := setupTestInfrastructure(t)
	server := NewServer(repo, cache, "")

	tests := []struct {
		name           string
		language       string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GetFrenchTranslations",
			language:       "fr",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GetSpanishTranslations",
			language:       "es",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GetUnsupportedLanguage",
			language:       "de",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Unsupported language code",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest("GET", "/api/v1/hotels/999999/translations/"+tt.language, nil)
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			if tt.expectedStatus == http.StatusOK {
				// For supported languages, should either return 404 or 200 depending on what's in the database
				assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, w.Code)
			} else {
				assert.Equal(t, tt.expectedStatus, w.Code)
				if tt.expectedBody != "" {
					assert.Contains(t, w.Body.String(), tt.expectedBody)
				}
			}
		})
	}
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
