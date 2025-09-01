package database

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vrnvu/cupid/internal/client"
)

// randomID generates a random ID for testing
func randomID() int {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return int(n.Int64()) + 1000000
}

// createRandomProperty creates a minimal property with random values
func createRandomProperty() *client.Property {
	return &client.Property{
		HotelID:     randomID(),
		CupidID:     randomID(),
		HotelName:   "Random Hotel",
		Rating:      4.0,
		Stars:       4,
		ReviewCount: 100,
		Description: "Random hotel for testing",
		Address: client.Address{
			Address:    "Random Street",
			City:       "Random City",
			State:      "Random State",
			Country:    "RC",
			PostalCode: "12345",
		},
		Checkin: client.Checkin{
			CheckinStart: "14:00",
			Checkout:     "11:00",
		},
	}
}

// setupTestDB creates a database connection for testing
func setupTestDB(t *testing.T) *DB {
	config := Config{
		Host:     "localhost",
		Port:     5432,
		User:     "cupid",
		Password: "cupid123",
		DBName:   "cupid",
		SSLMode:  "disable",
	}

	db, err := NewConnection(config)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TestHotelRepository_StoreProperty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		property *client.Property
		wantErr  bool
	}{
		{
			name:     "store basic property",
			property: createRandomProperty(),
			wantErr:  false,
		},
		{
			name: "store property with custom values",
			property: func() *client.Property {
				p := createRandomProperty()
				p.HotelName = "Custom Hotel"
				p.Rating = 5.0
				p.Stars = 5
				return p
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			repo := NewHotelRepository(db)
			ctx := context.Background()

			err := repo.StoreProperty(ctx, tt.property)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify hotel was stored
			storedHotel, err := repo.GetHotelByID(ctx, tt.property.HotelID)
			require.NoError(t, err)
			assert.Equal(t, tt.property.HotelName, storedHotel.HotelName)
			assert.Equal(t, tt.property.Rating, storedHotel.Rating)
		})
	}
}

func TestHotelRepository_GetHotelByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(*testing.T, *HotelRepository) int
		wantName string
		wantErr  bool
	}{
		{
			name: "get existing hotel",
			setup: func(t *testing.T, repo *HotelRepository) int {
				property := createRandomProperty()
				property.HotelName = "Test Hotel"
				err := repo.StoreProperty(context.Background(), property)
				require.NoError(t, err)
				return property.HotelID
			},
			wantName: "Test Hotel",
			wantErr:  false,
		},
		{
			name: "get non-existing hotel",
			setup: func(_ *testing.T, _ *HotelRepository) int {
				return 999999
			},
			wantName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			repo := NewHotelRepository(db)
			ctx := context.Background()

			hotelID := tt.setup(t, repo)

			hotel, err := repo.GetHotelByID(ctx, hotelID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, hotel)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantName, hotel.HotelName)
		})
	}
}

func TestHotelRepository_GetHotels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupHotels func(*testing.T, *HotelRepository) []int
		limit       int
		offset      int
		expectedMin int
		expectedMax int
	}{
		{
			name: "get hotels with pagination",
			setupHotels: func(t *testing.T, repo *HotelRepository) []int {
				hotels := []int{}
				for i := 0; i < 5; i++ {
					property := createRandomProperty()
					property.HotelName = fmt.Sprintf("Hotel %d", i+1)
					err := repo.StoreProperty(context.Background(), property)
					require.NoError(t, err)
					hotels = append(hotels, property.HotelID)
				}
				return hotels
			},
			limit:       2,
			offset:      0,
			expectedMin: 2,
			expectedMax: 2,
		},
		{
			name: "get hotels with large limit",
			setupHotels: func(t *testing.T, repo *HotelRepository) []int {
				hotels := []int{}
				for i := 0; i < 3; i++ {
					property := createRandomProperty()
					property.HotelName = fmt.Sprintf("Large Hotel %d", i+1)
					err := repo.StoreProperty(context.Background(), property)
					require.NoError(t, err)
					hotels = append(hotels, property.HotelID)
				}
				return hotels
			},
			limit:       100,
			offset:      0,
			expectedMin: 3,
			expectedMax: 100,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			repo := NewHotelRepository(db)
			ctx := context.Background()

			tt.setupHotels(t, repo)

			hotels, err := repo.GetHotels(ctx, tt.limit, tt.offset)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(hotels), tt.expectedMin)
			if tt.expectedMax > 0 {
				assert.LessOrEqual(t, len(hotels), tt.expectedMax)
			}
		})
	}
}

func TestHotelRepository_StoreReviews(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		reviews []client.Review
		wantErr bool
	}{
		{
			name: "store single review",
			reviews: []client.Review{
				{
					ReviewerName: "Test User",
					Rating:       5,
					Title:        "Great stay",
					Content:      "Excellent service",
					LanguageCode: "en",
					ReviewDate:   "2024-01-15",
					HelpfulVotes: 10,
				},
			},
			wantErr: false,
		},
		{
			name: "store multiple reviews",
			reviews: []client.Review{
				{
					ReviewerName: "User 1",
					Rating:       4,
					Title:        "Good stay",
					Content:      "Nice hotel",
					LanguageCode: "en",
					ReviewDate:   "2024-01-10",
					HelpfulVotes: 5,
				},
				{
					ReviewerName: "User 2",
					Rating:       5,
					Title:        "Amazing stay",
					Content:      "Perfect experience",
					LanguageCode: "en",
					ReviewDate:   "2024-01-12",
					HelpfulVotes: 15,
				},
			},
			wantErr: false,
		},
		{
			name:    "store empty reviews",
			reviews: []client.Review{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			repo := NewHotelRepository(db)
			ctx := context.Background()

			// Create a test hotel first
			property := createRandomProperty()
			err := repo.StoreProperty(ctx, property)
			require.NoError(t, err)

			err = repo.StoreReviews(ctx, property.HotelID, tt.reviews)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify reviews were stored
			storedReviews, err := repo.GetHotelReviews(ctx, property.HotelID)
			require.NoError(t, err)
			assert.Len(t, storedReviews, len(tt.reviews))
		})
	}
}

func TestHotelRepository_StoreTranslations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		translations []client.Translation
		wantErr      bool
	}{
		{
			name: "store French translations",
			translations: []client.Translation{
				{
					LanguageCode:   "fr",
					FieldName:      "hotel_name",
					TranslatedText: "Hôtel de Test",
				},
				{
					LanguageCode:   "fr",
					FieldName:      "description",
					TranslatedText: "Un hôtel magnifique",
				},
			},
			wantErr: false,
		},
		{
			name: "store Spanish translations",
			translations: []client.Translation{
				{
					LanguageCode:   "es",
					FieldName:      "hotel_name",
					TranslatedText: "Hotel de Prueba",
				},
			},
			wantErr: false,
		},
		{
			name:         "store empty translations",
			translations: []client.Translation{},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			repo := NewHotelRepository(db)
			ctx := context.Background()

			// Create a test hotel first
			property := createRandomProperty()
			err := repo.StoreProperty(ctx, property)
			require.NoError(t, err)

			err = repo.StoreTranslations(ctx, property.HotelID, tt.translations)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify translations were stored for each language
			languages := make(map[string]bool)
			for _, trans := range tt.translations {
				languages[trans.LanguageCode] = true
			}

			for lang := range languages {
				storedTranslations, err := repo.GetHotelTranslations(ctx, property.HotelID, lang)
				require.NoError(t, err)
				assert.Greater(t, len(storedTranslations), 0, "Should have translations for language %s", lang)
			}
		})
	}
}

func TestHotelRepository_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := NewHotelRepository(db)
	ctx := context.Background()

	const numGoroutines = 5
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			property := createRandomProperty()
			property.HotelName = fmt.Sprintf("Concurrent Hotel %d", id)
			property.Rating = 4.0 + float64(id)*0.1

			err := repo.StoreProperty(ctx, property)
			require.NoError(t, err)

			// Verify it was stored
			storedHotel, err := repo.GetHotelByID(ctx, property.HotelID)
			require.NoError(t, err)
			assert.Equal(t, property.HotelName, storedHotel.HotelName)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
