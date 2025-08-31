package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/vrnvu/cupid/internal/client"
)

// Error constants
var (
	ErrHotelNotFound      = errors.New("hotel not found")
	ErrDatabaseConnection = errors.New("database connection failed")
)

type Repository interface {
	StoreProperty(ctx context.Context, property *client.Property) error
	GetHotelByID(ctx context.Context, hotelID int) (*client.Property, error)
	Ping(ctx context.Context) error
}

type HotelRepository struct {
	db *DB
}

func NewHotelRepository(db *DB) *HotelRepository {
	return &HotelRepository{db: db}
}

func (r *HotelRepository) StoreProperty(ctx context.Context, property *client.Property) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	hotelID, err := r.storeHotel(ctx, tx, property)
	if err != nil {
		return fmt.Errorf("failed to store hotel: %w", err)
	}

	if err := r.storeAddress(ctx, tx, hotelID, &property.Address); err != nil {
		return fmt.Errorf("failed to store address: %w", err)
	}

	if err := r.storeCheckin(ctx, tx, hotelID, &property.Checkin); err != nil {
		return fmt.Errorf("failed to store checkin: %w", err)
	}

	if err := r.storePhotosBatch(ctx, tx, hotelID, property.Photos); err != nil {
		return fmt.Errorf("failed to store photos: %w", err)
	}

	if err := r.storeFacilitiesBatch(ctx, tx, hotelID, property.Facilities); err != nil {
		return fmt.Errorf("failed to store facilities: %w", err)
	}

	if err := r.storePoliciesBatch(ctx, tx, hotelID, property.Policies); err != nil {
		return fmt.Errorf("failed to store policies: %w", err)
	}

	if err := r.storeRoomsBatch(ctx, tx, hotelID, property.Rooms); err != nil {
		return fmt.Errorf("failed to store rooms: %w", err)
	}

	return tx.Commit()
}

func (r *HotelRepository) storeHotel(ctx context.Context, tx *sql.Tx, property *client.Property) (int, error) {
	query := `
		INSERT INTO hotels (
			hotel_id, cupid_id, main_image_th, hotel_type, hotel_type_id,
			chain, chain_id, latitude, longitude, hotel_name, phone, fax, email,
			stars, airport_code, rating, review_count, parking, group_room_min,
			child_allowed, pets_allowed, description, markdown_description, important_info
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
		ON CONFLICT (hotel_id) DO UPDATE SET
			updated_at = NOW(),
			main_image_th = EXCLUDED.main_image_th,
			hotel_name = EXCLUDED.hotel_name,
			phone = EXCLUDED.phone,
			fax = EXCLUDED.fax,
			email = EXCLUDED.email,
			rating = EXCLUDED.rating,
			review_count = EXCLUDED.review_count,
			description = EXCLUDED.description,
			markdown_description = EXCLUDED.markdown_description,
			important_info = EXCLUDED.important_info
		RETURNING hotel_id`

	var hotelID int
	err := tx.QueryRowContext(ctx, query,
		property.HotelID, property.CupidID, property.MainImageTh, property.HotelType, property.HotelTypeID,
		property.Chain, property.ChainID, property.Latitude, property.Longitude, property.HotelName,
		property.Phone, property.Fax, property.Email, property.Stars, property.AirportCode,
		property.Rating, property.ReviewCount, property.Parking, property.GroupRoomMin,
		property.ChildAllowed, property.PetsAllowed, property.Description, property.MarkdownDescription, property.ImportantInfo,
	).Scan(&hotelID)

	return hotelID, err
}

func (r *HotelRepository) storeAddress(ctx context.Context, tx *sql.Tx, hotelID int, address *client.Address) error {
	query := `
		INSERT INTO hotel_addresses (hotel_id, address, city, state, country, postal_code)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (hotel_id) DO UPDATE SET
			address = EXCLUDED.address,
			city = EXCLUDED.city,
			state = EXCLUDED.state,
			country = EXCLUDED.country,
			postal_code = EXCLUDED.postal_code`

	_, err := tx.ExecContext(ctx, query, hotelID, address.Address, address.City, address.State, address.Country, address.PostalCode)
	return err
}

func (r *HotelRepository) storeCheckin(ctx context.Context, tx *sql.Tx, hotelID int, checkin *client.Checkin) error {
	query := `
		INSERT INTO hotel_checkins (hotel_id, checkin_start, checkin_end, checkout, special_instructions)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (hotel_id) DO UPDATE SET
			checkin_start = EXCLUDED.checkin_start,
			checkin_end = EXCLUDED.checkin_end,
			checkout = EXCLUDED.checkout,
			special_instructions = EXCLUDED.special_instructions
		RETURNING id`

	var checkinID int
	err := tx.QueryRowContext(ctx, query, hotelID, checkin.CheckinStart, checkin.CheckinEnd, checkin.Checkout, checkin.SpecialInstructions).Scan(&checkinID)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM hotel_checkin_instructions WHERE hotel_checkin_id = $1", checkinID); err != nil {
		return err
	}

	if len(checkin.Instructions) > 0 {
		if err := r.storeCheckinInstructionsBatch(ctx, tx, checkinID, checkin.Instructions); err != nil {
			return err
		}
	}

	return nil
}

func (r *HotelRepository) storeCheckinInstructionsBatch(ctx context.Context, tx *sql.Tx, checkinID int, instructions []string) error {
	if len(instructions) == 0 {
		return nil
	}

	query := "INSERT INTO hotel_checkin_instructions (hotel_checkin_id, instruction, sort_order) VALUES "
	values := make([]string, len(instructions))
	args := make([]interface{}, len(instructions)*3)

	for i, instruction := range instructions {
		values[i] = fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3)
		args[i*3] = checkinID
		args[i*3+1] = instruction
		args[i*3+2] = i
	}

	query += strings.Join(values, ", ")
	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *HotelRepository) storePhotosBatch(ctx context.Context, tx *sql.Tx, hotelID int, photos []client.Photo) error {
	if len(photos) == 0 {
		return nil
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM hotel_photos WHERE hotel_id = $1", hotelID); err != nil {
		return err
	}

	query := `
		INSERT INTO hotel_photos (hotel_id, url, hd_url, image_description, image_class1, image_class2, main_photo, score, class_id, class_order)
		VALUES `

	values := make([]string, len(photos))
	args := make([]interface{}, len(photos)*10)

	for i, photo := range photos {
		values[i] = fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			i*10+1, i*10+2, i*10+3, i*10+4, i*10+5, i*10+6, i*10+7, i*10+8, i*10+9, i*10+10)
		args[i*10] = hotelID
		args[i*10+1] = photo.URL
		args[i*10+2] = photo.HDURL
		args[i*10+3] = photo.ImageDescription
		args[i*10+4] = photo.ImageClass1
		args[i*10+5] = photo.ImageClass2
		args[i*10+6] = photo.MainPhoto
		args[i*10+7] = photo.Score
		args[i*10+8] = photo.ClassID
		args[i*10+9] = photo.ClassOrder
	}

	query += strings.Join(values, ", ")
	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *HotelRepository) storeFacilitiesBatch(ctx context.Context, tx *sql.Tx, hotelID int, facilities []client.Facility) error {
	if len(facilities) == 0 {
		return nil
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM hotel_facilities WHERE hotel_id = $1", hotelID); err != nil {
		return err
	}

	query := `INSERT INTO hotel_facilities (hotel_id, facility_id, name) VALUES `

	values := make([]string, len(facilities))
	args := make([]interface{}, len(facilities)*3)

	for i, facility := range facilities {
		values[i] = fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3)
		args[i*3] = hotelID
		args[i*3+1] = facility.FacilityID
		args[i*3+2] = facility.Name
	}

	query += strings.Join(values, ", ")
	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *HotelRepository) storePoliciesBatch(ctx context.Context, tx *sql.Tx, hotelID int, policies []client.Policy) error {
	if len(policies) == 0 {
		return nil
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM hotel_policies WHERE hotel_id = $1", hotelID); err != nil {
		return err
	}

	query := `
		INSERT INTO hotel_policies (hotel_id, policy_type, name, description, child_allowed, pets_allowed, parking, cupid_policy_id)
		VALUES `

	values := make([]string, len(policies))
	args := make([]interface{}, len(policies)*8)

	for i, policy := range policies {
		values[i] = fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			i*8+1, i*8+2, i*8+3, i*8+4, i*8+5, i*8+6, i*8+7, i*8+8)
		args[i*8] = hotelID
		args[i*8+1] = policy.PolicyType
		args[i*8+2] = policy.Name
		args[i*8+3] = policy.Description
		args[i*8+4] = policy.ChildAllowed
		args[i*8+5] = policy.PetsAllowed
		args[i*8+6] = policy.Parking
		args[i*8+7] = policy.ID
	}

	query += strings.Join(values, ", ")
	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *HotelRepository) storeRoomsBatch(ctx context.Context, tx *sql.Tx, hotelID int, rooms []client.Room) error {
	if len(rooms) == 0 {
		return nil
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM hotel_rooms WHERE hotel_id = $1", hotelID); err != nil {
		return err
	}

	roomQuery := `
		INSERT INTO hotel_rooms (hotel_id, cupid_room_id, room_name, description, room_size_square, room_size_unit, max_adults, max_children, max_occupancy, bed_relation)
		VALUES `

	roomValues := make([]string, len(rooms))
	roomArgs := make([]interface{}, len(rooms)*10)

	for i, room := range rooms {
		roomValues[i] = fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			i*10+1, i*10+2, i*10+3, i*10+4, i*10+5, i*10+6, i*10+7, i*10+8, i*10+9, i*10+10)
		roomArgs[i*10] = hotelID
		roomArgs[i*10+1] = room.ID
		roomArgs[i*10+2] = room.RoomName
		roomArgs[i*10+3] = room.Description
		roomArgs[i*10+4] = room.RoomSizeSquare
		roomArgs[i*10+5] = room.RoomSizeUnit
		roomArgs[i*10+6] = room.MaxAdults
		roomArgs[i*10+7] = room.MaxChildren
		roomArgs[i*10+8] = room.MaxOccupancy
		roomArgs[i*10+9] = room.BedRelation
	}

	roomQuery += strings.Join(roomValues, ", ") + " RETURNING id, cupid_room_id"

	rows, err := tx.QueryContext(ctx, roomQuery, roomArgs...)
	if err != nil {
		return err
	}
	defer rows.Close()

	roomIDMap := make(map[int]int)
	for rows.Next() {
		var roomID int
		var cupidRoomID int
		if err := rows.Scan(&roomID, &cupidRoomID); err != nil {
			return err
		}
		roomIDMap[cupidRoomID] = roomID
	}

	if err := rows.Err(); err != nil {
		return err
	}

	for _, room := range rooms {
		roomID := roomIDMap[room.ID]

		if err := r.storeBedTypesBatch(ctx, tx, roomID, room.BedTypes); err != nil {
			return err
		}

		if err := r.storeRoomAmenitiesBatch(ctx, tx, roomID, room.RoomAmenities); err != nil {
			return err
		}

		if err := r.storeRoomPhotosBatch(ctx, tx, roomID, room.Photos); err != nil {
			return err
		}
	}

	return nil
}

func (r *HotelRepository) storeBedTypesBatch(ctx context.Context, tx *sql.Tx, roomID int, bedTypes []client.BedType) error {
	if len(bedTypes) == 0 {
		return nil
	}

	query := `INSERT INTO room_bed_types (room_id, quantity, bed_type, bed_size, cupid_bed_id) VALUES `

	values := make([]string, len(bedTypes))
	args := make([]interface{}, len(bedTypes)*5)

	for i, bedType := range bedTypes {
		values[i] = fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", i*5+1, i*5+2, i*5+3, i*5+4, i*5+5)
		args[i*5] = roomID
		args[i*5+1] = bedType.Quantity
		args[i*5+2] = bedType.BedType
		args[i*5+3] = bedType.BedSize
		args[i*5+4] = bedType.ID
	}

	query += strings.Join(values, ", ")
	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *HotelRepository) storeRoomAmenitiesBatch(ctx context.Context, tx *sql.Tx, roomID int, amenities []client.RoomAmenity) error {
	if len(amenities) == 0 {
		return nil
	}

	query := `INSERT INTO room_amenities (room_id, amenities_id, name, sort_order) VALUES `

	values := make([]string, len(amenities))
	args := make([]interface{}, len(amenities)*4)

	for i, amenity := range amenities {
		values[i] = fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4)
		args[i*4] = roomID
		args[i*4+1] = amenity.AmenitiesID
		args[i*4+2] = amenity.Name
		args[i*4+3] = amenity.Sort
	}

	query += strings.Join(values, ", ")
	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *HotelRepository) storeRoomPhotosBatch(ctx context.Context, tx *sql.Tx, roomID int, photos []client.Photo) error {
	if len(photos) == 0 {
		return nil
	}

	query := `
		INSERT INTO room_photos (room_id, url, hd_url, image_description, image_class1, image_class2, main_photo, score, class_id, class_order)
		VALUES `

	values := make([]string, len(photos))
	args := make([]interface{}, len(photos)*10)

	for i, photo := range photos {
		values[i] = fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			i*10+1, i*10+2, i*10+3, i*10+4, i*10+5, i*10+6, i*10+7, i*10+8, i*10+9, i*10+10)
		args[i*10] = roomID
		args[i*10+1] = photo.URL
		args[i*10+2] = photo.HDURL
		args[i*10+3] = photo.ImageDescription
		args[i*10+4] = photo.ImageClass1
		args[i*10+5] = photo.ImageClass2
		args[i*10+6] = photo.MainPhoto
		args[i*10+7] = photo.Score
		args[i*10+8] = photo.ClassID
		args[i*10+9] = photo.ClassOrder
	}

	query += strings.Join(values, ", ")
	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *HotelRepository) GetHotelByID(ctx context.Context, hotelID int) (*client.Property, error) {
	query := `SELECT hotel_id, cupid_id, hotel_name, rating, review_count FROM hotels WHERE hotel_id = $1`

	var property client.Property
	err := r.db.QueryRowContext(ctx, query, hotelID).Scan(
		&property.HotelID, &property.CupidID, &property.HotelName, &property.Rating, &property.ReviewCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrHotelNotFound
		}
		return nil, err
	}

	return &property, nil
}

func (r *HotelRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}
