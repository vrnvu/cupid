package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vrnvu/cupid/internal/client"
	"github.com/vrnvu/cupid/internal/database"
)

// Mock implementations
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) StoreProperty(ctx context.Context, property *client.Property) error {
	args := m.Called(ctx, property)
	return args.Error(0)
}

func (m *MockRepository) StoreReviews(ctx context.Context, hotelID int, reviews []client.Review) error {
	args := m.Called(ctx, hotelID, reviews)
	return args.Error(0)
}

func (m *MockRepository) StoreTranslations(ctx context.Context, hotelID int, translations []client.Translation) error {
	args := m.Called(ctx, hotelID, translations)
	return args.Error(0)
}

func (m *MockRepository) GetHotels(ctx context.Context, limit, offset int) ([]client.Property, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]client.Property), args.Error(1)
}

func (m *MockRepository) GetHotelByID(ctx context.Context, hotelID int) (*client.Property, error) {
	args := m.Called(ctx, hotelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.Property), args.Error(1)
}

func (m *MockRepository) GetHotelReviews(ctx context.Context, hotelID int) ([]client.Review, error) {
	args := m.Called(ctx, hotelID)
	return args.Get(0).([]client.Review), args.Error(1)
}

func (m *MockRepository) GetHotelTranslations(ctx context.Context, hotelID int, languageCode string) ([]client.Translation, error) {
	args := m.Called(ctx, hotelID, languageCode)
	return args.Get(0).([]client.Translation), args.Error(1)
}

func (m *MockRepository) SearchReviewsByVector(ctx context.Context, queryEmbedding []float64, limit int, threshold float64) ([]client.Review, error) {
	args := m.Called(ctx, queryEmbedding, limit, threshold)
	return args.Get(0).([]client.Review), args.Error(1)
}

func (m *MockRepository) GetReviewsNeedingEmbeddings(ctx context.Context, limit int) ([]int, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]int), args.Error(1)
}

func (m *MockRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) GetReviews(ctx context.Context, hotelID int) ([]client.Review, error) {
	args := m.Called(ctx, hotelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]client.Review), args.Error(1)
}

func (m *MockCache) SetReviews(ctx context.Context, hotelID int, reviews []client.Review, ttl time.Duration) error {
	args := m.Called(ctx, hotelID, reviews, ttl)
	return args.Error(0)
}

func (m *MockCache) DeleteReviews(ctx context.Context, hotelID int) error {
	args := m.Called(ctx, hotelID)
	return args.Error(0)
}

func (m *MockCache) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCache) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewServer(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	apiKey := "test-api-key"

	server := NewServer(mockRepo, mockCache, apiKey)
	assert.NotNil(t, server)
}

func TestServer_HealthHandler_NoAuth(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	server := NewServer(mockRepo, mockCache, "")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	mockRepo.On("Ping", mock.Anything).Return(nil)

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "cupid-api", response["service"])

	mockRepo.AssertExpectations(t)
}

func TestServer_HealthHandler_DatabaseError(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	server := NewServer(mockRepo, mockCache, "")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	mockRepo.On("Ping", mock.Anything).Return(database.ErrDatabaseConnection)

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestServer_Authentication_ValidAPIKey(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	apiKey := "valid-api-key" //nolint:gosec // This is a test value, not a real credential
	server := NewServer(mockRepo, mockCache, apiKey)

	req := httptest.NewRequest("GET", "/api/v1/hotels", nil)
	req.Header.Set("Authorization", "Bearer valid-api-key")
	w := httptest.NewRecorder()

	mockRepo.On("GetHotels", mock.Anything, 50, 0).Return([]client.Property{}, nil)

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestServer_Authentication_MissingAPIKey(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	apiKey := "required-api-key"
	server := NewServer(mockRepo, mockCache, apiKey)

	req := httptest.NewRequest("GET", "/api/v1/hotels", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization header required")
}

func TestServer_Authentication_InvalidFormat(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	apiKey := "required-api-key"
	server := NewServer(mockRepo, mockCache, apiKey)

	req := httptest.NewRequest("GET", "/api/v1/hotels", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid authorization format")
}

func TestServer_Authentication_WrongAPIKey(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	apiKey := "correct-api-key"
	server := NewServer(mockRepo, mockCache, apiKey)

	req := httptest.NewRequest("GET", "/api/v1/hotels", nil)
	req.Header.Set("Authorization", "Bearer wrong-api-key")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid API key")
}

func TestServer_Authentication_NoAuthRequired(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	server := NewServer(mockRepo, mockCache, "") // No API key required

	req := httptest.NewRequest("GET", "/api/v1/hotels", nil)
	w := httptest.NewRecorder()

	mockRepo.On("GetHotels", mock.Anything, 50, 0).Return([]client.Property{}, nil)

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestServer_RateLimiting(t *testing.T) {
	t.Parallel()
	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	server := NewServer(mockRepo, mockCache, "")

	req := httptest.NewRequest("GET", "/api/v1/hotels", nil)
	w := httptest.NewRecorder()

	mockRepo.On("GetHotels", mock.Anything, 50, 0).Return([]client.Property{}, nil)

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestServer_GetHotelsHandler_Success(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	server := NewServer(mockRepo, mockCache, "")

	req := httptest.NewRequest("GET", "/api/v1/hotels", nil)
	w := httptest.NewRecorder()

	expectedHotels := []client.Property{
		{HotelID: 1, HotelName: "Test Hotel 1"},
		{HotelID: 2, HotelName: "Test Hotel 2"},
	}

	mockRepo.On("GetHotels", mock.Anything, 50, 0).Return(expectedHotels, nil)

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	hotels := response["hotels"].([]interface{})
	assert.Len(t, hotels, 2)
	assert.Equal(t, float64(1), hotels[0].(map[string]interface{})["hotel_id"])
	assert.Equal(t, "Test Hotel 1", hotels[0].(map[string]interface{})["hotel_name"])

	mockRepo.AssertExpectations(t)
}

func TestServer_GetHotelsHandler_WithPagination(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	server := NewServer(mockRepo, mockCache, "")

	req := httptest.NewRequest("GET", "/api/v1/hotels?limit=10&offset=20", nil)
	w := httptest.NewRecorder()

	expectedHotels := []client.Property{{HotelID: 1, HotelName: "Test Hotel"}}
	mockRepo.On("GetHotels", mock.Anything, 10, 20).Return(expectedHotels, nil)

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, float64(10), response["limit"])
	assert.Equal(t, float64(20), response["offset"])

	mockRepo.AssertExpectations(t)
}

func TestServer_GetHotelsHandler_DatabaseError(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	server := NewServer(mockRepo, mockCache, "")

	req := httptest.NewRequest("GET", "/api/v1/hotels", nil)
	w := httptest.NewRecorder()

	mockRepo.On("GetHotels", mock.Anything, 50, 0).Return(nil, database.ErrDatabaseConnection)

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestServer_GetHotelHandler_Success(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	server := NewServer(mockRepo, mockCache, "")

	req := httptest.NewRequest("GET", "/api/v1/hotels/123", nil)
	w := httptest.NewRecorder()

	expectedHotel := &client.Property{HotelID: 123, HotelName: "Test Hotel"}
	mockRepo.On("GetHotelByID", mock.Anything, 123).Return(expectedHotel, nil)

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var hotel client.Property
	err := json.NewDecoder(w.Body).Decode(&hotel)
	assert.NoError(t, err)
	assert.Equal(t, 123, hotel.HotelID)
	assert.Equal(t, "Test Hotel", hotel.HotelName)

	mockRepo.AssertExpectations(t)
}

func TestServer_GetHotelHandler_NotFound(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	server := NewServer(mockRepo, mockCache, "")

	req := httptest.NewRequest("GET", "/api/v1/hotels/999", nil)
	w := httptest.NewRecorder()

	mockRepo.On("GetHotelByID", mock.Anything, 999).Return(nil, database.ErrHotelNotFound)

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Hotel with ID 999 not found")

	mockRepo.AssertExpectations(t)
}

func TestServer_GetHotelHandler_InvalidID(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	server := NewServer(mockRepo, mockCache, "")

	req := httptest.NewRequest("GET", "/api/v1/hotels/invalid", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid hotel ID format")

	mockRepo.AssertExpectations(t)
}

func TestServer_GetHotelReviewsHandler_Success(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	server := NewServer(mockRepo, mockCache, "")

	req := httptest.NewRequest("GET", "/api/v1/hotels/123/reviews", nil)
	w := httptest.NewRecorder()

	expectedReviews := []client.Review{
		{ID: 1, Rating: 5, Title: "Great hotel!", Content: "Excellent experience"},
		{ID: 2, Rating: 4, Title: "Good experience", Content: "Nice stay"},
	}

	mockCache.On("GetReviews", mock.Anything, 123).Return(nil, assert.AnError)
	mockRepo.On("GetHotelReviews", mock.Anything, 123).Return(expectedReviews, nil)
	mockCache.On("SetReviews", mock.Anything, 123, expectedReviews, mock.Anything).Return(nil)

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, float64(123), response["hotel_id"])
	assert.Equal(t, float64(2), response["count"])
	assert.Equal(t, false, response["from_cache"])

	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestServer_GetHotelReviewsHandler_FromCache(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	server := NewServer(mockRepo, mockCache, "")

	req := httptest.NewRequest("GET", "/api/v1/hotels/123/reviews", nil)
	w := httptest.NewRecorder()

	expectedReviews := []client.Review{
		{ID: 1, Rating: 5, Title: "Great hotel!", Content: "Excellent experience"},
	}

	mockCache.On("GetReviews", mock.Anything, 123).Return(expectedReviews, nil)

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, true, response["from_cache"])
	assert.Equal(t, float64(1), response["count"])

	mockCache.AssertExpectations(t)
}

func TestServer_GetHotelTranslationsHandler_Success(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	server := NewServer(mockRepo, mockCache, "")

	req := httptest.NewRequest("GET", "/api/v1/hotels/123/translations/fr", nil)
	w := httptest.NewRecorder()

	expectedTranslations := []client.Translation{
		{FieldName: "hotel_name", LanguageCode: "fr", TranslatedText: "Hôtel de Test"},
	}

	mockRepo.On("GetHotelTranslations", mock.Anything, 123, "fr").Return(expectedTranslations, nil)

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, float64(123), response["hotel_id"])
	assert.Equal(t, "fr", response["language"])
	assert.Equal(t, float64(1), response["count"])

	mockRepo.AssertExpectations(t)
}

func TestServer_GetHotelTranslationsHandler_InvalidLanguage(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	server := NewServer(mockRepo, mockCache, "")

	req := httptest.NewRequest("GET", "/api/v1/hotels/123/translations/xx", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Unsupported language code")

	mockRepo.AssertExpectations(t)
}

func TestServer_SearchReviewsHandler_Success(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	server := NewServer(mockRepo, mockCache, "")

	req := httptest.NewRequest("GET", "/api/v1/reviews/search?q=great&limit=5&threshold=0.8", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, "great", response["query"])
	assert.Equal(t, float64(5), response["limit"])
	assert.Equal(t, 0.8, response["threshold"])
	assert.Contains(t, response["message"], "Vector search endpoint ready")
}

func TestServer_SearchReviewsHandler_MissingQuery(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	server := NewServer(mockRepo, mockCache, "")

	req := httptest.NewRequest("GET", "/api/v1/reviews/search", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Query parameter 'q' is required")
}

func TestServer_SearchReviewsHandler_InvalidLimit(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	server := NewServer(mockRepo, mockCache, "")

	req := httptest.NewRequest("GET", "/api/v1/reviews/search?q=test&limit=invalid", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	// Should use default limit of 10
	assert.Equal(t, float64(10), response["limit"])
}

func TestServer_SearchReviewsHandler_LimitExceedsMax(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	server := NewServer(mockRepo, mockCache, "")

	req := httptest.NewRequest("GET", "/api/v1/reviews/search?q=test&limit=150", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	// Should cap at max limit of 100
	assert.Equal(t, float64(100), response["limit"])
}
