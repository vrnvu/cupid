package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vrnvu/cupid/internal/client"
	"github.com/vrnvu/cupid/internal/database"
)

type mockHotelRepository struct {
	hotels  map[int]*client.Property
	pingErr error
}

func (m *mockHotelRepository) StoreProperty(_ context.Context, _ *client.Property) error {
	return nil
}

func (m *mockHotelRepository) GetHotelByID(_ context.Context, hotelID int) (*client.Property, error) {
	if hotel, exists := m.hotels[hotelID]; exists {
		return hotel, nil
	}
	return nil, database.ErrHotelNotFound
}

func (m *mockHotelRepository) Ping(_ context.Context) error {
	return m.pingErr
}

func TestHealthHandler(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		method     string
		pingErr    error
		wantStatus int
		wantBody   map[string]string
	}{
		{
			name:       "healthy service",
			method:     http.MethodGet,
			pingErr:    nil,
			wantStatus: http.StatusOK,
			wantBody: map[string]string{
				"status":  "healthy",
				"service": "cupid-api",
			},
		},
		{
			name:       "database connection failed",
			method:     http.MethodGet,
			pingErr:    database.ErrDatabaseConnection,
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   nil,
		},
		{
			name:       "method not allowed",
			method:     http.MethodPost,
			pingErr:    nil,
			wantStatus: http.StatusMethodNotAllowed,
			wantBody:   nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mockHotelRepository{pingErr: tc.pingErr}
			server := NewServer(mockRepo)

			req, err := http.NewRequestWithContext(context.Background(), tc.method, "/health", nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			rec := httptest.NewRecorder()
			server.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("unexpected status code: got %d, want %d", rec.Code, tc.wantStatus)
			}

			if tc.wantBody != nil {
				var response map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				for key, expectedValue := range tc.wantBody {
					if response[key] != expectedValue {
						t.Errorf("unexpected value for %s: got %s, want %s", key, response[key], expectedValue)
					}
				}
			}
		})
	}
}

func TestHotelHandler(t *testing.T) {
	t.Parallel()

	mockHotels := map[int]*client.Property{
		1641879: {
			HotelID:     1641879,
			CupidID:     12345,
			HotelName:   "Test Hotel",
			Rating:      4.5,
			ReviewCount: 150,
		},
	}

	testCases := []struct {
		name       string
		method     string
		hotelID    string
		wantStatus int
		wantHotel  *client.Property
	}{
		{
			name:       "get existing hotel",
			method:     http.MethodGet,
			hotelID:    "1641879",
			wantStatus: http.StatusOK,
			wantHotel:  mockHotels[1641879],
		},
		{
			name:       "hotel not found",
			method:     http.MethodGet,
			hotelID:    "999999",
			wantStatus: http.StatusNotFound,
			wantHotel:  nil,
		},
		{
			name:       "invalid hotel ID format",
			method:     http.MethodGet,
			hotelID:    "invalid",
			wantStatus: http.StatusBadRequest,
			wantHotel:  nil,
		},
		{
			name:       "method not allowed",
			method:     http.MethodPost,
			hotelID:    "1641879",
			wantStatus: http.StatusMethodNotAllowed,
			wantHotel:  nil,
		},
		{
			name:       "missing hotel ID",
			method:     http.MethodGet,
			hotelID:    "",
			wantStatus: http.StatusNotFound,
			wantHotel:  nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mockHotelRepository{hotels: mockHotels}
			server := NewServer(mockRepo)

			url := "/api/v1/hotels/"
			if tc.hotelID != "" {
				url += tc.hotelID
			}

			req, err := http.NewRequestWithContext(context.Background(), tc.method, url, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			rec := httptest.NewRecorder()
			server.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("unexpected status code: got %d, want %d", rec.Code, tc.wantStatus)
			}

			if tc.wantHotel != nil {
				var response client.Property
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if response.HotelID != tc.wantHotel.HotelID {
					t.Errorf("unexpected hotel ID: got %d, want %d", response.HotelID, tc.wantHotel.HotelID)
				}
				if response.HotelName != tc.wantHotel.HotelName {
					t.Errorf("unexpected hotel name: got %s, want %s", response.HotelName, tc.wantHotel.HotelName)
				}
				if response.Rating != tc.wantHotel.Rating {
					t.Errorf("unexpected rating: got %f, want %f", response.Rating, tc.wantHotel.Rating)
				}
				if response.ReviewCount != tc.wantHotel.ReviewCount {
					t.Errorf("unexpected review count: got %d, want %d", response.ReviewCount, tc.wantHotel.ReviewCount)
				}
			}
		})
	}
}
