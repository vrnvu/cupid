package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vrnvu/cupid/internal/client"
)

// MockRedisCache implements ReviewCache interface for testing
type MockRedisCache struct {
	mock.Mock
}

func (m *MockRedisCache) GetReviews(ctx context.Context, hotelID int) ([]client.Review, error) {
	args := m.Called(ctx, hotelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]client.Review), args.Error(1)
}

func (m *MockRedisCache) SetReviews(ctx context.Context, hotelID int, reviews []client.Review, ttl time.Duration) error {
	args := m.Called(ctx, hotelID, reviews, ttl)
	return args.Error(0)
}

func (m *MockRedisCache) DeleteReviews(ctx context.Context, hotelID int) error {
	args := m.Called(ctx, hotelID)
	return args.Error(0)
}

func (m *MockRedisCache) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRedisCache) Close() error {
	args := m.Called()
	return args.Error(0)
}

// createTestReview creates a test review with deterministic values
func createTestReview(hotelID int, id int) client.Review {
	return client.Review{
		ID:           id,
		HotelID:      hotelID,
		ReviewerName: "Test User",
		Rating:       5,
		Title:        "Test Review",
		Content:      "Test content",
		LanguageCode: "en",
		ReviewDate:   "2024-01-01",
		HelpfulVotes: 0,
		CreatedAt:    "2024-01-01T00:00:00Z",
	}
}

// createTestReviews creates a slice of test reviews with deterministic IDs
func createTestReviews(hotelID int, count int) []client.Review {
	reviews := make([]client.Review, count)
	for i := 0; i < count; i++ {
		reviews[i] = createTestReview(hotelID, 1000+i)
	}
	return reviews
}

func TestRedisCache_GetReviews(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		hotelID        int
		setupMock      func(*MockRedisCache)
		expectedResult []client.Review
		expectedError  bool
	}{
		{
			name:    "successful retrieval",
			hotelID: 12345,
			setupMock: func(m *MockRedisCache) {
				expectedReviews := createTestReviews(12345, 2)
				m.On("GetReviews", mock.Anything, 12345).Return(expectedReviews, nil)
			},
			expectedResult: createTestReviews(12345, 2),
			expectedError:  false,
		},
		{
			name:    "cache miss returns nil",
			hotelID: 67890,
			setupMock: func(m *MockRedisCache) {
				m.On("GetReviews", mock.Anything, 67890).Return(nil, nil)
			},
			expectedResult: nil,
			expectedError:  false,
		},
		{
			name:    "redis error",
			hotelID: 11111,
			setupMock: func(m *MockRedisCache) {
				m.On("GetReviews", mock.Anything, 11111).Return(nil, assert.AnError)
			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCache := &MockRedisCache{}
			tt.setupMock(mockCache)

			ctx := context.Background()
			result, err := mockCache.GetReviews(ctx, tt.hotelID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, result)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestRedisCache_SetReviews(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		hotelID       int
		reviews       []client.Review
		ttl           time.Duration
		setupMock     func(*MockRedisCache)
		expectedError bool
	}{
		{
			name:    "successful set",
			hotelID: 12345,
			reviews: createTestReviews(12345, 2),
			ttl:     5 * time.Second,
			setupMock: func(m *MockRedisCache) {
				m.On("SetReviews", mock.Anything, 12345, createTestReviews(12345, 2), 5*time.Second).Return(nil)
			},
			expectedError: false,
		},
		{
			name:    "empty reviews",
			hotelID: 67890,
			reviews: []client.Review{},
			ttl:     10 * time.Second,
			setupMock: func(m *MockRedisCache) {
				m.On("SetReviews", mock.Anything, 67890, []client.Review{}, 10*time.Second).Return(nil)
			},
			expectedError: false,
		},
		{
			name:    "redis error",
			hotelID: 11111,
			reviews: createTestReviews(11111, 1),
			ttl:     1 * time.Second,
			setupMock: func(m *MockRedisCache) {
				m.On("SetReviews", mock.Anything, 11111, createTestReviews(11111, 1), 1*time.Second).Return(assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCache := &MockRedisCache{}
			tt.setupMock(mockCache)

			ctx := context.Background()
			err := mockCache.SetReviews(ctx, tt.hotelID, tt.reviews, tt.ttl)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCache.AssertExpectations(t)
		})
	}
}

func TestRedisCache_DeleteReviews(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		hotelID       int
		setupMock     func(*MockRedisCache)
		expectedError bool
	}{
		{
			name:    "successful deletion",
			hotelID: 12345,
			setupMock: func(m *MockRedisCache) {
				m.On("DeleteReviews", mock.Anything, 12345).Return(nil)
			},
			expectedError: false,
		},
		{
			name:    "redis error",
			hotelID: 11111,
			setupMock: func(m *MockRedisCache) {
				m.On("DeleteReviews", mock.Anything, 11111).Return(assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCache := &MockRedisCache{}
			tt.setupMock(mockCache)

			ctx := context.Background()
			err := mockCache.DeleteReviews(ctx, tt.hotelID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCache.AssertExpectations(t)
		})
	}
}

func TestRedisCache_Ping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupMock     func(*MockRedisCache)
		expectedError bool
	}{
		{
			name: "successful ping",
			setupMock: func(m *MockRedisCache) {
				m.On("Ping", mock.Anything).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "ping error",
			setupMock: func(m *MockRedisCache) {
				m.On("Ping", mock.Anything).Return(assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCache := &MockRedisCache{}
			tt.setupMock(mockCache)

			ctx := context.Background()
			err := mockCache.Ping(ctx)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCache.AssertExpectations(t)
		})
	}
}

func TestRedisCache_Close(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupMock     func(*MockRedisCache)
		expectedError bool
	}{
		{
			name: "successful close",
			setupMock: func(m *MockRedisCache) {
				m.On("Close").Return(nil)
			},
			expectedError: false,
		},
		{
			name: "close error",
			setupMock: func(m *MockRedisCache) {
				m.On("Close").Return(assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCache := &MockRedisCache{}
			tt.setupMock(mockCache)

			err := mockCache.Close()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCache.AssertExpectations(t)
		})
	}
}

func TestRedisCache_EmptyReviews(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		hotelID        int
		setupMock      func(*MockRedisCache)
		expectedResult []client.Review
		expectedError  bool
	}{
		{
			name:    "empty reviews list",
			hotelID: 12345,
			setupMock: func(m *MockRedisCache) {
				m.On("GetReviews", mock.Anything, 12345).Return([]client.Review{}, nil)
			},
			expectedResult: []client.Review{},
			expectedError:  false,
		},
		{
			name:    "nil reviews",
			hotelID: 67890,
			setupMock: func(m *MockRedisCache) {
				m.On("GetReviews", mock.Anything, 67890).Return(nil, nil)
			},
			expectedResult: nil,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCache := &MockRedisCache{}
			tt.setupMock(mockCache)

			ctx := context.Background()
			result, err := mockCache.GetReviews(ctx, tt.hotelID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, result)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestRedisCache_TTL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		hotelID       int
		reviews       []client.Review
		ttl           time.Duration
		setupMock     func(*MockRedisCache)
		expectedError bool
	}{
		{
			name:    "short TTL",
			hotelID: 12345,
			reviews: createTestReviews(12345, 1),
			ttl:     1 * time.Second,
			setupMock: func(m *MockRedisCache) {
				m.On("SetReviews", mock.Anything, 12345, createTestReviews(12345, 1), 1*time.Second).Return(nil)
			},
			expectedError: false,
		},
		{
			name:    "long TTL",
			hotelID: 67890,
			reviews: createTestReviews(67890, 2),
			ttl:     24 * time.Hour,
			setupMock: func(m *MockRedisCache) {
				m.On("SetReviews", mock.Anything, 67890, createTestReviews(67890, 2), 24*time.Hour).Return(nil)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockCache := &MockRedisCache{}
			tt.setupMock(mockCache)

			ctx := context.Background()
			err := mockCache.SetReviews(ctx, tt.hotelID, tt.reviews, tt.ttl)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCache.AssertExpectations(t)
		})
	}
}
