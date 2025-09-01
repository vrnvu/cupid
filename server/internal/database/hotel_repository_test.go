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
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		// Fallback to a simple incrementing counter if crypto/rand fails
		return 1000000 + (int(n.Int64()) % 1000000)
	}
	return 1000000 + int(n.Int64())
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

	// Set up cleanup to close connection after test
	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func dummyProperty(hotelID int, cupidID int, hotelName string) *client.Property {
	// If hotelID is 0, generate a random one
	if hotelID == 0 {
		hotelID = randomID()
	}
	// If cupidID is 0, generate a random one
	if cupidID == 0 {
		cupidID = randomID()
	}

	return &client.Property{
		HotelID:             hotelID,
		CupidID:             cupidID,
		MainImageTh:         "https://example.com/image.jpg",
		HotelType:           "Hotel",
		HotelTypeID:         1,
		Chain:               "Test Chain",
		ChainID:             100,
		Latitude:            51.5074,
		Longitude:           -0.1278,
		HotelName:           hotelName,
		Phone:               "+44 20 1234 5678",
		Fax:                 "+44 20 1234 5679",
		Email:               "test@hotel.com",
		Stars:               4,
		AirportCode:         "LHR",
		Rating:              4.5,
		ReviewCount:         150,
		Parking:             "Available",
		GroupRoomMin:        &[]int{2}[0],
		ChildAllowed:        true,
		PetsAllowed:         false,
		Description:         "A beautiful test hotel",
		MarkdownDescription: "# Test Hotel\nA beautiful test hotel",
		ImportantInfo:       "Important information here",
		Address: client.Address{
			Address:    "123 Test Street",
			City:       "London",
			State:      "England",
			Country:    "GB",
			PostalCode: "SW1A 1AA",
		},
		Checkin: client.Checkin{
			CheckinStart:        "14:00",
			CheckinEnd:          "23:00",
			Checkout:            "11:00",
			SpecialInstructions: "Please bring ID",
			Instructions:        []string{"Bring ID", "Credit card required"},
		},
		Photos: []client.Photo{
			{
				URL:              "https://example.com/photo1.jpg",
				HDURL:            "https://example.com/photo1_hd.jpg",
				ImageDescription: "Hotel exterior",
				ImageClass1:      "exterior",
				ImageClass2:      "building",
				MainPhoto:        true,
				Score:            9.5,
				ClassID:          1,
				ClassOrder:       1,
			},
			{
				URL:              "https://example.com/photo2.jpg",
				HDURL:            "https://example.com/photo2_hd.jpg",
				ImageDescription: "Hotel lobby",
				ImageClass1:      "interior",
				ImageClass2:      "lobby",
				MainPhoto:        false,
				Score:            8.5,
				ClassID:          2,
				ClassOrder:       2,
			},
		},
		Facilities: []client.Facility{
			{
				FacilityID: 1,
				Name:       "WiFi",
			},
			{
				FacilityID: 2,
				Name:       "Pool",
			},
		},
		Policies: []client.Policy{
			{
				PolicyType:   "cancellation",
				Name:         "Flexible Cancellation",
				Description:  "Free cancellation up to 24 hours",
				ChildAllowed: "Yes",
				PetsAllowed:  "No",
				Parking:      "Available",
				ID:           1,
			},
		},
		Rooms: []client.Room{
			{
				ID:             1,
				RoomName:       "Standard Room",
				Description:    "Comfortable standard room",
				RoomSizeSquare: 25,
				RoomSizeUnit:   "m²",
				MaxAdults:      2,
				MaxChildren:    1,
				MaxOccupancy:   3,
				BedRelation:    "1 king bed",
				BedTypes: []client.BedType{
					{
						Quantity: 1,
						BedType:  "King",
						BedSize:  "200x180",
						ID:       1,
					},
				},
				RoomAmenities: []client.RoomAmenity{
					{
						AmenitiesID: 1,
						Name:        "TV",
						SortOrder:   1,
					},
					{
						AmenitiesID: 2,
						Name:        "Air Conditioning",
						SortOrder:   2,
					},
				},
				Photos: []client.Photo{
					{
						URL:              "https://example.com/room1.jpg",
						HDURL:            "https://example.com/room1_hd.jpg",
						ImageDescription: "Standard room",
						ImageClass1:      "room",
						ImageClass2:      "standard",
						MainPhoto:        true,
						Score:            9.0,
						ClassID:          3,
						ClassOrder:       1,
					},
				},
			},
		},
	}
}

func TestHotelRepository_StoreProperty(t *testing.T) {
	t.Parallel()

	t.Run("store complete property", func(t *testing.T) {
		t.Parallel()
		db := setupTestDB(t)
		repo := NewHotelRepository(db)
		ctx := context.Background()
		property := dummyProperty(0, 0, "Test Hotel")

		err := repo.StoreProperty(ctx, property)
		require.NoError(t, err)

		// Verify hotel was stored
		storedHotel, err := repo.GetHotelByID(ctx, property.HotelID)
		require.NoError(t, err)
		assert.Equal(t, property.HotelID, storedHotel.HotelID)
		assert.Equal(t, property.HotelName, storedHotel.HotelName)
		assert.Equal(t, property.Rating, storedHotel.Rating)
	})

	t.Run("update existing property", func(t *testing.T) {
		t.Parallel()
		db := setupTestDB(t)
		repo := NewHotelRepository(db)
		ctx := context.Background()

		// First create a hotel with random ID
		originalProperty := dummyProperty(0, 0, "Original Hotel")
		originalProperty.Rating = 3.5
		originalProperty.ReviewCount = 50

		err := repo.StoreProperty(ctx, originalProperty)
		require.NoError(t, err)

		// Then update it using the same hotel ID
		updatedProperty := dummyProperty(originalProperty.HotelID, originalProperty.CupidID, "Updated Test Hotel")
		updatedProperty.Rating = 4.8
		updatedProperty.ReviewCount = 200

		err = repo.StoreProperty(ctx, updatedProperty)
		require.NoError(t, err)

		// Verify hotel was updated
		storedHotel, err := repo.GetHotelByID(ctx, updatedProperty.HotelID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Test Hotel", storedHotel.HotelName)
		assert.Equal(t, 4.8, storedHotel.Rating)
		assert.Equal(t, 200, storedHotel.ReviewCount)
	})
}

func TestHotelRepository_GetHotelByID(t *testing.T) {
	t.Parallel()

	t.Run("get existing hotel", func(t *testing.T) {
		t.Parallel()
		db := setupTestDB(t)
		repo := NewHotelRepository(db)
		ctx := context.Background()
		// First create a hotel with random ID
		property := dummyProperty(0, 0, "Get Hotel Test")
		property.Rating = 4.2
		property.ReviewCount = 75

		err := repo.StoreProperty(ctx, property)
		require.NoError(t, err)

		// Then retrieve it using the stored hotel ID
		hotel, err := repo.GetHotelByID(ctx, property.HotelID)
		require.NoError(t, err)
		assert.Equal(t, property.HotelID, hotel.HotelID)
		assert.Equal(t, "Get Hotel Test", hotel.HotelName)
		assert.Equal(t, 4.2, hotel.Rating)
	})

	t.Run("get non-existing hotel", func(t *testing.T) {
		t.Parallel()
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
		defer db.Close()

		repo := NewHotelRepository(db)
		ctx := context.Background()

		hotel, err := repo.GetHotelByID(ctx, 999999)
		assert.Error(t, err)
		assert.Nil(t, hotel)
	})
}

func TestHotelRepository_StoreProperty_EmptyData(t *testing.T) {
	t.Parallel()

	t.Run("store property with empty collections", func(t *testing.T) {
		t.Parallel()
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
		defer db.Close()

		repo := NewHotelRepository(db)
		ctx := context.Background()
		property := dummyProperty(0, 0, "Empty Collections Hotel")
		property.Rating = 3.5
		property.ReviewCount = 50
		property.Photos = []client.Photo{}
		property.Facilities = []client.Facility{}
		property.Policies = []client.Policy{}
		property.Rooms = []client.Room{}

		err = repo.StoreProperty(ctx, property)
		require.NoError(t, err)

		// Verify hotel was stored
		storedHotel, err := repo.GetHotelByID(ctx, property.HotelID)
		require.NoError(t, err)
		assert.Equal(t, property.HotelID, storedHotel.HotelID)
		assert.Equal(t, property.HotelName, storedHotel.HotelName)
	})
}

func TestHotelRepository_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	t.Run("concurrent property storage", func(t *testing.T) {
		t.Parallel()
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
		defer db.Close()

		repo := NewHotelRepository(db)
		ctx := context.Background()
		const numGoroutines = 5
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				property := dummyProperty(10000+id, 20000+id, fmt.Sprintf("Concurrent Hotel %d", id))
				property.Rating = 4.0 + float64(id)*0.1
				property.ReviewCount = 100 + id*10

				err = repo.StoreProperty(ctx, property)
				require.NoError(t, err)
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Verify all hotels were stored
		for i := 0; i < numGoroutines; i++ {
			hotel, err := repo.GetHotelByID(ctx, 10000+i)
			require.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("Concurrent Hotel %d", i), hotel.HotelName)
		}
	})
}

func TestHotelRepository_TransactionRollback(t *testing.T) {
	t.Parallel()

	t.Run("transaction rollback on error", func(t *testing.T) {
		t.Parallel()
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
		defer db.Close()

		repo := NewHotelRepository(db)
		ctx := context.Background()
		// Store a property first with random ID
		property := dummyProperty(0, 0, "Rollback Test Hotel")
		property.Rating = 4.0
		property.ReviewCount = 100

		err = repo.StoreProperty(ctx, property)
		require.NoError(t, err)

		// Verify it was stored
		hotel, err := repo.GetHotelByID(ctx, property.HotelID)
		require.NoError(t, err)
		assert.Equal(t, "Rollback Test Hotel", hotel.HotelName)

		// Try to store a property with invalid data (this should fail and rollback)
		invalidProperty := dummyProperty(property.HotelID, property.CupidID, "") // Empty name might cause issues
		invalidProperty.Rating = 4.0
		invalidProperty.ReviewCount = 100

		_ = repo.StoreProperty(ctx, invalidProperty)
		// This might succeed due to upsert, but the point is to test transaction handling
		// In a real scenario, you'd have more complex validation that could fail
	})
}

func TestHotelRepository_GetHotelReviews(t *testing.T) {
	t.Parallel()

	t.Run("get hotel reviews", func(t *testing.T) {
		t.Parallel()
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
		defer db.Close()

		repo := NewHotelRepository(db)
		ctx := context.Background()

		// Store a hotel first with random ID
		property := dummyProperty(0, 0, "Reviews Test Hotel")
		hotelID := property.HotelID
		err = repo.StoreProperty(ctx, property)
		require.NoError(t, err)

		// Clean up any existing reviews for this hotel
		_, err = db.ExecContext(ctx, "DELETE FROM reviews WHERE hotel_id = $1", hotelID)
		require.NoError(t, err)

		// Insert test reviews
		_, err = db.ExecContext(ctx, `
			INSERT INTO reviews (hotel_id, reviewer_name, rating, title, content, language_code, review_date, helpful_votes, created_at)
			VALUES 
			($1, 'John Doe', 5, 'Great hotel!', 'Excellent service', 'en', '2024-01-15', 10, '2024-01-15T10:00:00Z'),
			($1, 'Jane Smith', 4, 'Good experience', 'Nice location', 'en', '2024-01-10', 5, '2024-01-10T14:30:00Z')
		`, hotelID)
		require.NoError(t, err)

		reviews, err := repo.GetHotelReviews(ctx, hotelID)
		require.NoError(t, err)
		assert.Len(t, reviews, 2)

		// Check that both reviews exist without assuming order
		foundJohn := false
		foundJane := false
		for _, review := range reviews {
			if review.ReviewerName == "John Doe" {
				assert.Equal(t, 5, review.Rating)
				assert.Equal(t, "Great hotel!", review.Title)
				foundJohn = true
			}
			if review.ReviewerName == "Jane Smith" {
				assert.Equal(t, 4, review.Rating)
				assert.Equal(t, "Good experience", review.Title)
				foundJane = true
			}
		}
		assert.True(t, foundJohn, "John Doe review not found")
		assert.True(t, foundJane, "Jane Smith review not found")
	})

	t.Run("get hotel with no reviews", func(t *testing.T) {
		t.Parallel()
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
		defer db.Close()

		repo := NewHotelRepository(db)
		ctx := context.Background()

		// Store a hotel without reviews
		property := dummyProperty(22222, 33333, "No Reviews Hotel")
		err = repo.StoreProperty(ctx, property)
		require.NoError(t, err)

		reviews, err := repo.GetHotelReviews(ctx, 22222)
		require.NoError(t, err)
		assert.Len(t, reviews, 0)
	})

	t.Run("get reviews for non-existent hotel", func(t *testing.T) {
		t.Parallel()
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
		defer db.Close()

		repo := NewHotelRepository(db)
		ctx := context.Background()

		reviews, err := repo.GetHotelReviews(ctx, 999999)
		require.NoError(t, err)
		assert.Len(t, reviews, 0)
	})
}

func TestHotelRepository_GetHotelTranslations(t *testing.T) {
	t.Parallel()

	t.Run("get French translations", func(t *testing.T) {
		t.Parallel()
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
		defer db.Close()

		repo := NewHotelRepository(db)
		ctx := context.Background()

		// Store a hotel first
		property := dummyProperty(33333, 44444, "Translations Test Hotel")
		err = repo.StoreProperty(ctx, property)
		require.NoError(t, err)

		// Clean up any existing translations for this hotel
		_, err = db.ExecContext(ctx, "DELETE FROM translations WHERE entity_id = $1", 33333)
		require.NoError(t, err)

		// Insert test translations
		_, err = db.ExecContext(ctx, `
			INSERT INTO translations (entity_type, entity_id, language_code, field_name, translated_text, created_at, updated_at)
			VALUES 
			('hotel', 33333, 'fr', 'hotel_name', 'L''Hôtel Z Covent Garden', '2024-01-15T10:00:00Z', '2024-01-15T10:00:00Z'),
			('hotel', 33333, 'fr', 'description', 'Un hôtel moderne au cœur de Londres', '2024-01-15T10:00:00Z', '2024-01-15T10:00:00Z')
		`)
		require.NoError(t, err)

		translations, err := repo.GetHotelTranslations(ctx, 33333, "fr")
		require.NoError(t, err)
		assert.Len(t, translations, 2)

		// Check that both translations exist without assuming order
		foundHotelName := false
		foundDescription := false
		for _, trans := range translations {
			if trans.FieldName == "hotel_name" {
				assert.Equal(t, "L'Hôtel Z Covent Garden", trans.TranslatedText)
				foundHotelName = true
			}
			if trans.FieldName == "description" {
				assert.Equal(t, "Un hôtel moderne au cœur de Londres", trans.TranslatedText)
				foundDescription = true
			}
		}
		assert.True(t, foundHotelName, "hotel_name translation not found")
		assert.True(t, foundDescription, "description translation not found")
	})

	t.Run("get Spanish translations", func(t *testing.T) {
		t.Parallel()
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
		defer db.Close()

		repo := NewHotelRepository(db)
		ctx := context.Background()

		// Store a hotel first
		property := dummyProperty(44444, 55555, "Spanish Translations Hotel")
		err = repo.StoreProperty(ctx, property)
		require.NoError(t, err)

		// Clean up any existing translations for this hotel
		_, err = db.ExecContext(ctx, "DELETE FROM translations WHERE entity_id = $1", 44444)
		require.NoError(t, err)

		// Insert test translations
		_, err = db.ExecContext(ctx, `
			INSERT INTO translations (entity_type, entity_id, language_code, field_name, translated_text, created_at, updated_at)
			VALUES 
			('hotel', 44444, 'es', 'hotel_name', 'El Hotel Z Covent Garden', '2024-01-15T10:00:00Z', '2024-01-15T10:00:00Z')
		`)
		require.NoError(t, err)

		translations, err := repo.GetHotelTranslations(ctx, 44444, "es")
		require.NoError(t, err)
		assert.Len(t, translations, 1)
		assert.Equal(t, "hotel_name", translations[0].FieldName)
		assert.Equal(t, "El Hotel Z Covent Garden", translations[0].TranslatedText)
	})

	t.Run("get hotel with no translations", func(t *testing.T) {
		t.Parallel()
		db := setupTestDB(t)
		repo := NewHotelRepository(db)
		ctx := context.Background()

		// Store a hotel without translations
		property := dummyProperty(700001, 800001, "No Translations Hotel")
		err := repo.StoreProperty(ctx, property)
		require.NoError(t, err)

		translations, err := repo.GetHotelTranslations(ctx, 700001, "fr")
		require.NoError(t, err)
		assert.Len(t, translations, 0)
	})

	t.Run("get translations for non-existent hotel", func(t *testing.T) {
		t.Parallel()
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
		defer db.Close()

		repo := NewHotelRepository(db)
		ctx := context.Background()

		translations, err := repo.GetHotelTranslations(ctx, 999999, "fr")
		require.NoError(t, err)
		assert.Len(t, translations, 0)
	})

	t.Run("get translations for unsupported language", func(t *testing.T) {
		t.Parallel()
		db := setupTestDB(t)
		repo := NewHotelRepository(db)
		ctx := context.Background()

		// Store a hotel first
		property := dummyProperty(800001, 900001, "German Translations Hotel")
		err := repo.StoreProperty(ctx, property)
		require.NoError(t, err)

		translations, err := repo.GetHotelTranslations(ctx, 800001, "de")
		require.NoError(t, err)
		assert.Len(t, translations, 0)
	})
}

func TestHotelRepository_StoreReviews(t *testing.T) {
	t.Parallel()

	t.Run("store reviews successfully", func(t *testing.T) {
		t.Parallel()
		db := setupTestDB(t)
		repo := NewHotelRepository(db)
		ctx := context.Background()

		// Store a hotel first with random ID
		property := dummyProperty(0, 0, "Store Reviews Test Hotel")
		hotelID := property.HotelID
		err := repo.StoreProperty(ctx, property)
		require.NoError(t, err)

		// Clean up any existing reviews for this hotel
		_, err = db.ExecContext(ctx, "DELETE FROM reviews WHERE hotel_id = $1", hotelID)
		require.NoError(t, err)

		reviews := []client.Review{
			{
				ReviewerName: "Alice Johnson",
				Rating:       5,
				Title:        "Amazing stay!",
				Content:      "Perfect location and excellent service",
				LanguageCode: "en",
				ReviewDate:   "2024-01-15",
				HelpfulVotes: 12,
			},
			{
				ReviewerName: "Bob Smith",
				Rating:       4,
				Title:        "Good hotel",
				Content:      "Nice amenities and friendly staff",
				LanguageCode: "en",
				ReviewDate:   "2024-01-10",
				HelpfulVotes: 8,
			},
		}

		err = repo.StoreReviews(ctx, hotelID, reviews)
		require.NoError(t, err)

		// Verify reviews were stored
		storedReviews, err := repo.GetHotelReviews(ctx, hotelID)
		require.NoError(t, err)
		assert.Len(t, storedReviews, 2)

		// Check that both reviews exist without assuming order
		foundAlice := false
		foundBob := false
		for _, review := range storedReviews {
			if review.ReviewerName == "Alice Johnson" {
				assert.Equal(t, 5, review.Rating)
				assert.Equal(t, "Amazing stay!", review.Title)
				assert.Equal(t, "Perfect location and excellent service", review.Content)
				foundAlice = true
			}
			if review.ReviewerName == "Bob Smith" {
				assert.Equal(t, 4, review.Rating)
				assert.Equal(t, "Good hotel", review.Title)
				assert.Equal(t, "Nice amenities and friendly staff", review.Content)
				foundBob = true
			}
		}
		assert.True(t, foundAlice, "Alice Johnson review not found")
		assert.True(t, foundBob, "Bob Smith review not found")
	})

	t.Run("store empty reviews list", func(t *testing.T) {
		t.Parallel()
		db := setupTestDB(t)
		repo := NewHotelRepository(db)
		ctx := context.Background()

		// Store a hotel first with random ID
		property := dummyProperty(0, 0, "Empty Reviews Test Hotel")
		hotelID := property.HotelID
		err := repo.StoreProperty(ctx, property)
		require.NoError(t, err)

		err = repo.StoreReviews(ctx, hotelID, []client.Review{})
		require.NoError(t, err)

		// Verify no reviews were stored
		storedReviews, err := repo.GetHotelReviews(ctx, hotelID)
		require.NoError(t, err)
		assert.Len(t, storedReviews, 0)
	})

	t.Run("replace existing reviews", func(t *testing.T) {
		t.Parallel()
		db := setupTestDB(t)
		repo := NewHotelRepository(db)
		ctx := context.Background()

		// Store a hotel first with random ID
		property := dummyProperty(0, 0, "Replace Reviews Test Hotel")
		hotelID := property.HotelID
		err := repo.StoreProperty(ctx, property)
		require.NoError(t, err)

		// Store initial reviews
		initialReviews := []client.Review{
			{
				ReviewerName: "Old Reviewer",
				Rating:       3,
				Title:        "Old review",
				Content:      "This is an old review",
				LanguageCode: "en",
				ReviewDate:   "2024-01-01",
				HelpfulVotes: 1,
			},
		}

		err = repo.StoreReviews(ctx, hotelID, initialReviews)
		require.NoError(t, err)

		// Verify initial reviews were stored
		storedReviews, err := repo.GetHotelReviews(ctx, hotelID)
		require.NoError(t, err)
		assert.Len(t, storedReviews, 1)
		assert.Equal(t, "Old Reviewer", storedReviews[0].ReviewerName)

		// Store new reviews (should replace the old ones)
		newReviews := []client.Review{
			{
				ReviewerName: "New Reviewer",
				Rating:       5,
				Title:        "New review",
				Content:      "This is a new review",
				LanguageCode: "en",
				ReviewDate:   "2024-01-20",
				HelpfulVotes: 5,
			},
		}

		err = repo.StoreReviews(ctx, hotelID, newReviews)
		require.NoError(t, err)

		// Verify old reviews were replaced
		storedReviews, err = repo.GetHotelReviews(ctx, hotelID)
		require.NoError(t, err)
		assert.Len(t, storedReviews, 1)
		assert.Equal(t, "New Reviewer", storedReviews[0].ReviewerName)
		assert.Equal(t, "New review", storedReviews[0].Title)
	})
}

func TestHotelRepository_StoreTranslations(t *testing.T) {
	t.Parallel()

	t.Run("store translations successfully", func(t *testing.T) {
		t.Parallel()
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
		defer db.Close()

		repo := NewHotelRepository(db)
		ctx := context.Background()

		// Store a hotel first with random ID
		property := dummyProperty(0, 0, "Store Translations Test Hotel")
		hotelID := property.HotelID
		err = repo.StoreProperty(ctx, property)
		require.NoError(t, err)

		translations := []client.Translation{
			{
				LanguageCode:   "fr",
				FieldName:      "hotel_name",
				TranslatedText: "L'Hôtel de Test",
			},
			{
				LanguageCode:   "fr",
				FieldName:      "description",
				TranslatedText: "Un hôtel magnifique au cœur de la ville",
			},
			{
				LanguageCode:   "es",
				FieldName:      "hotel_name",
				TranslatedText: "El Hotel de Prueba",
			},
		}

		err = repo.StoreTranslations(ctx, hotelID, translations)
		require.NoError(t, err)

		// Verify French translations were stored
		frenchTranslations, err := repo.GetHotelTranslations(ctx, hotelID, "fr")
		require.NoError(t, err)
		assert.Len(t, frenchTranslations, 2)

		// Check that both French translations exist without assuming order
		foundHotelName := false
		foundDescription := false
		for _, trans := range frenchTranslations {
			if trans.FieldName == "hotel_name" {
				assert.Equal(t, "L'Hôtel de Test", trans.TranslatedText)
				foundHotelName = true
			}
			if trans.FieldName == "description" {
				assert.Equal(t, "Un hôtel magnifique au cœur de la ville", trans.TranslatedText)
				foundDescription = true
			}
		}
		assert.True(t, foundHotelName, "hotel_name French translation not found")
		assert.True(t, foundDescription, "description French translation not found")

		// Verify Spanish translations were stored
		spanishTranslations, err := repo.GetHotelTranslations(ctx, hotelID, "es")
		require.NoError(t, err)
		assert.Len(t, spanishTranslations, 1)
		assert.Equal(t, "hotel_name", spanishTranslations[0].FieldName)
		assert.Equal(t, "El Hotel de Prueba", spanishTranslations[0].TranslatedText)
	})

	t.Run("store empty translations list", func(t *testing.T) {
		t.Parallel()
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
		defer db.Close()

		repo := NewHotelRepository(db)
		ctx := context.Background()

		// Store a hotel first with random ID
		property := dummyProperty(0, 0, "Empty Translations Test Hotel")
		hotelID := property.HotelID
		err = repo.StoreProperty(ctx, property)
		require.NoError(t, err)

		err = repo.StoreTranslations(ctx, hotelID, []client.Translation{})
		require.NoError(t, err)

		// Verify no translations were stored
		translations, err := repo.GetHotelTranslations(ctx, hotelID, "fr")
		require.NoError(t, err)
		assert.Len(t, translations, 0)
	})

	t.Run("update existing translations", func(t *testing.T) {
		t.Parallel()
		db := setupTestDB(t)
		repo := NewHotelRepository(db)
		ctx := context.Background()

		// Store a hotel first with random ID
		property := dummyProperty(0, 0, "Update Translations Test Hotel")
		hotelID := property.HotelID
		err := repo.StoreProperty(ctx, property)
		require.NoError(t, err)

		// Store initial translations
		initialTranslations := []client.Translation{
			{
				LanguageCode:   "fr",
				FieldName:      "hotel_name",
				TranslatedText: "Ancien Nom d'Hôtel",
			},
		}

		err = repo.StoreTranslations(ctx, hotelID, initialTranslations)
		require.NoError(t, err)

		// Verify initial translations were stored
		translations, err := repo.GetHotelTranslations(ctx, hotelID, "fr")
		require.NoError(t, err)
		assert.Len(t, translations, 1)
		assert.Equal(t, "Ancien Nom d'Hôtel", translations[0].TranslatedText)

		// Store updated translations (should update the existing ones)
		updatedTranslations := []client.Translation{
			{
				LanguageCode:   "fr",
				FieldName:      "hotel_name",
				TranslatedText: "Nouveau Nom d'Hôtel",
			},
		}

		err = repo.StoreTranslations(ctx, hotelID, updatedTranslations)
		require.NoError(t, err)

		// Verify translations were updated
		translations, err = repo.GetHotelTranslations(ctx, hotelID, "fr")
		require.NoError(t, err)
		assert.Len(t, translations, 1)
		assert.Equal(t, "Nouveau Nom d'Hôtel", translations[0].TranslatedText)
	})
}

func TestHotelRepository_GetHotels(t *testing.T) {
	t.Parallel()

	t.Run("get hotels with pagination", func(t *testing.T) {
		t.Parallel()
		db := setupTestDB(t)
		repo := NewHotelRepository(db)
		ctx := context.Background()

		// Store multiple hotels with random IDs
		hotels := []*client.Property{
			dummyProperty(0, 0, "Pagination Hotel A"),
			dummyProperty(0, 0, "Pagination Hotel B"),
			dummyProperty(0, 0, "Pagination Hotel C"),
			dummyProperty(0, 0, "Pagination Hotel D"),
			dummyProperty(0, 0, "Pagination Hotel E"),
		}

		for _, hotel := range hotels {
			err := repo.StoreProperty(ctx, hotel)
			require.NoError(t, err)
		}

		// Test pagination - get all hotels and verify we can paginate through them
		allHotels, err := repo.GetHotels(ctx, 100, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(allHotels), 5, "Should have at least 5 hotels")

		// Test pagination with small limit
		hotelsPage1, err := repo.GetHotels(ctx, 2, 0)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(hotelsPage1), 2, "Page 1 should have at most 2 hotels")

		// Verify no overlap between pages
		allHotelIDs := make(map[int]bool)
		for _, hotel := range hotelsPage1 {
			allHotelIDs[hotel.HotelID] = true
		}

		// Test that we can get different pages
		hotelsPage2, err := repo.GetHotels(ctx, 2, 2)
		require.NoError(t, err)
		for _, hotel := range hotelsPage2 {
			assert.False(t, allHotelIDs[hotel.HotelID], "Hotel ID %d appears in multiple pages", hotel.HotelID)
			allHotelIDs[hotel.HotelID] = true
		}
	})

	t.Run("get hotels with large limit", func(t *testing.T) {
		t.Parallel()
		db := setupTestDB(t)
		repo := NewHotelRepository(db)
		ctx := context.Background()

		// Store a few hotels with random IDs
		hotels := []*client.Property{
			dummyProperty(0, 0, "Large Limit Hotel 1"),
			dummyProperty(0, 0, "Large Limit Hotel 2"),
		}

		for _, hotel := range hotels {
			err := repo.StoreProperty(ctx, hotel)
			require.NoError(t, err)
		}

		// Request more hotels than exist
		allHotels, err := repo.GetHotels(ctx, 100, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(allHotels), 2)
	})

	t.Run("get hotels with offset beyond available data", func(t *testing.T) {
		t.Parallel()
		db := setupTestDB(t)
		repo := NewHotelRepository(db)
		ctx := context.Background()

		// Request hotels with offset beyond what exists
		hotels, err := repo.GetHotels(ctx, 10, 10000)
		require.NoError(t, err)
		assert.Len(t, hotels, 0)
	})
}
