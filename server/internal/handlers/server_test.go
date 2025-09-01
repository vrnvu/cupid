package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vrnvu/cupid/internal/client"
	"github.com/vrnvu/cupid/internal/database"
)

// MockRepository is a mock implementation of the Repository interface
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

func (m *MockRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRepository) SearchReviewsByVector(ctx context.Context, queryEmbedding []float64, limit int, threshold float64) ([]client.Review, error) {
	args := m.Called(ctx, queryEmbedding, limit, threshold)
	return args.Get(0).([]client.Review), args.Error(1)
}

func (m *MockRepository) GetReviewsNeedingEmbeddings(ctx context.Context, limit int) ([]int, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]int), args.Error(1)
}

func TestServer_HealthHandler(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		method         string
		pingError      error
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name:           "healthy database",
			method:         "GET",
			pingError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]string{
				"status":  "healthy",
				"service": "cupid-api",
			},
		},
		{
			name:           "unhealthy database",
			method:         "GET",
			pingError:      errors.New("connection failed"),
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   nil,
		},
		{
			name:           "method not allowed",
			method:         "POST",
			pingError:      nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockRepo := new(MockRepository)

			// Only set up Ping mock expectation if the method is GET
			// For method not allowed, Ping should not be called
			if tt.method == "GET" {
				mockRepo.On("Ping", mock.Anything).Return(tt.pingError)
			}

			server := &Server{repository: mockRepo}
			req := httptest.NewRequest(tt.method, "/health", nil)
			w := httptest.NewRecorder()

			server.healthHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestServer_GetHotelHandler(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		hotelID        string
		hotel          *client.Property
		repoError      error
		expectedStatus int
	}{
		{
			name:    "successful request",
			hotelID: "123",
			hotel: &client.Property{
				HotelID:   123,
				HotelName: "Test Hotel",
				Rating:    4.5,
			},
			repoError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "hotel not found",
			hotelID:        "999",
			hotel:          nil,
			repoError:      database.ErrHotelNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid hotel ID",
			hotelID:        "invalid",
			hotel:          nil,
			repoError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repository error",
			hotelID:        "123",
			hotel:          nil,
			repoError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockRepo := new(MockRepository)
			if tt.hotelID != "invalid" {
				hotelID := 123
				if tt.hotelID == "999" {
					hotelID = 999
				}
				mockRepo.On("GetHotelByID", mock.Anything, hotelID).Return(tt.hotel, tt.repoError)
			}

			server := &Server{repository: mockRepo}
			req := httptest.NewRequest("GET", "/api/v1/hotels/"+tt.hotelID, nil)
			req.SetPathValue("hotelID", tt.hotelID)
			w := httptest.NewRecorder()

			server.getHotelHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response client.Property
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.hotel.HotelID, response.HotelID)
				assert.Equal(t, tt.hotel.HotelName, response.HotelName)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestServer_GetHotelReviewsHandler(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		hotelID        string
		reviews        []client.Review
		repoError      error
		expectedStatus int
		expectedCount  int
	}{
		{
			name:    "successful request with reviews",
			hotelID: "123",
			reviews: []client.Review{
				{ReviewerName: "John", Rating: 5, Title: "Great!"},
				{ReviewerName: "Jane", Rating: 4, Title: "Good"},
			},
			repoError:      nil,
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "successful request with no reviews",
			hotelID:        "123",
			reviews:        []client.Review{},
			repoError:      nil,
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "invalid hotel ID",
			hotelID:        "invalid",
			reviews:        nil,
			repoError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
		{
			name:           "repository error",
			hotelID:        "123",
			reviews:        nil,
			repoError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockRepo := new(MockRepository)
			if tt.hotelID != "invalid" {
				hotelID := 123
				mockRepo.On("GetHotelReviews", mock.Anything, hotelID).Return(tt.reviews, tt.repoError)
			}

			server := &Server{repository: mockRepo}
			req := httptest.NewRequest("GET", "/api/v1/hotels/"+tt.hotelID+"/reviews", nil)
			req.SetPathValue("hotelID", tt.hotelID)
			w := httptest.NewRecorder()

			server.getHotelReviewsHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, int(response["count"].(float64)))
				assert.Equal(t, 123, int(response["hotel_id"].(float64)))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
