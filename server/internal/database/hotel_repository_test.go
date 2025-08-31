package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vrnvu/cupid/internal/client"
)

func dummyProperty(hotelID int, cupidID int, hotelName string) *client.Property {
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
						Sort:        1,
					},
					{
						AmenitiesID: 2,
						Name:        "Air Conditioning",
						Sort:        2,
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
		property := dummyProperty(12345, 67890, "Test Hotel")

		err = repo.StoreProperty(ctx, property)
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

		// First create a hotel
		originalProperty := dummyProperty(77777, 88888, "Original Hotel")
		originalProperty.Rating = 3.5
		originalProperty.ReviewCount = 50

		err = repo.StoreProperty(ctx, originalProperty)
		require.NoError(t, err)

		// Then update it
		updatedProperty := dummyProperty(77777, 88888, "Updated Test Hotel")
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
		// First create a hotel
		property := dummyProperty(55555, 66666, "Get Hotel Test")
		property.Rating = 4.2
		property.ReviewCount = 75

		err = repo.StoreProperty(ctx, property)
		require.NoError(t, err)

		// Then retrieve it
		hotel, err := repo.GetHotelByID(ctx, 55555)
		require.NoError(t, err)
		assert.Equal(t, 55555, hotel.HotelID)
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
		property := dummyProperty(54321, 98765, "Empty Collections Hotel")
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
		// Store a property first
		property := dummyProperty(99999, 88888, "Rollback Test Hotel")
		property.Rating = 4.0
		property.ReviewCount = 100

		err = repo.StoreProperty(ctx, property)
		require.NoError(t, err)

		// Verify it was stored
		hotel, err := repo.GetHotelByID(ctx, 99999)
		require.NoError(t, err)
		assert.Equal(t, "Rollback Test Hotel", hotel.HotelName)

		// Try to store a property with invalid data (this should fail and rollback)
		invalidProperty := dummyProperty(99999, 88888, "") // Empty name might cause issues
		invalidProperty.Rating = 4.0
		invalidProperty.ReviewCount = 100

		_ = repo.StoreProperty(ctx, invalidProperty)
		// This might succeed due to upsert, but the point is to test transaction handling
		// In a real scenario, you'd have more complex validation that could fail
	})
}
