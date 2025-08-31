package client

import (
	"encoding/json"
)

// Property represents a hotel property from the Cupid API
type Property struct {
	HotelID             int        `json:"hotel_id"`
	CupidID             int        `json:"cupid_id"`
	MainImageTh         string     `json:"main_image_th"`
	HotelType           string     `json:"hotel_type"`
	HotelTypeID         int        `json:"hotel_type_id"`
	Chain               string     `json:"chain"`
	ChainID             int        `json:"chain_id"`
	Latitude            float64    `json:"latitude"`
	Longitude           float64    `json:"longitude"`
	HotelName           string     `json:"hotel_name"`
	Phone               string     `json:"phone"`
	Fax                 string     `json:"fax"`
	Email               string     `json:"email"`
	Address             Address    `json:"address"`
	Stars               int        `json:"stars"`
	AirportCode         string     `json:"airport_code"`
	Rating              float64    `json:"rating"`
	ReviewCount         int        `json:"review_count"`
	Checkin             Checkin    `json:"checkin"`
	Parking             string     `json:"parking"`
	GroupRoomMin        *int       `json:"group_room_min"`
	ChildAllowed        bool       `json:"child_allowed"`
	PetsAllowed         bool       `json:"pets_allowed"`
	Photos              []Photo    `json:"photos"`
	Description         string     `json:"description"`
	MarkdownDescription string     `json:"markdown_description"`
	ImportantInfo       string     `json:"important_info"`
	Facilities          []Facility `json:"facilities"`
	Policies            []Policy   `json:"policies"`
	Rooms               []Room     `json:"rooms"`
}

// Address represents a hotel address
type Address struct {
	Address    string `json:"address"`
	City       string `json:"city"`
	State      string `json:"state"`
	Country    string `json:"country"`
	PostalCode string `json:"postal_code"`
}

// Checkin represents check-in/check-out information
type Checkin struct {
	CheckinStart        string   `json:"checkin_start"`
	CheckinEnd          string   `json:"checkin_end"`
	Checkout            string   `json:"checkout"`
	Instructions        []string `json:"instructions"`
	SpecialInstructions string   `json:"special_instructions"`
}

// Photo represents a hotel photo
type Photo struct {
	URL              string  `json:"url"`
	HDURL            string  `json:"hd_url"`
	ImageDescription string  `json:"image_description"`
	ImageClass1      string  `json:"image_class1"`
	ImageClass2      string  `json:"image_class2"`
	MainPhoto        bool    `json:"main_photo"`
	Score            float64 `json:"score"`
	ClassID          int     `json:"class_id"`
	ClassOrder       int     `json:"class_order"`
}

// Facility represents a hotel facility
type Facility struct {
	FacilityID int    `json:"facility_id"`
	Name       string `json:"name"`
}

// Policy represents a hotel policy
type Policy struct {
	PolicyType   string `json:"policy_type"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	ChildAllowed string `json:"child_allowed"`
	PetsAllowed  string `json:"pets_allowed"`
	Parking      string `json:"parking"`
	ID           int    `json:"id"`
}

// Room represents a hotel room
type Room struct {
	ID             int           `json:"id"`
	RoomName       string        `json:"room_name"`
	Description    string        `json:"description"`
	RoomSizeSquare int           `json:"room_size_square"`
	RoomSizeUnit   string        `json:"room_size_unit"`
	HotelID        string        `json:"hotel_id"`
	MaxAdults      int           `json:"max_adults"`
	MaxChildren    int           `json:"max_children"`
	MaxOccupancy   int           `json:"max_occupancy"`
	BedRelation    string        `json:"bed_relation"`
	BedTypes       []BedType     `json:"bed_types"`
	RoomAmenities  []RoomAmenity `json:"room_amenities"`
	Photos         []Photo       `json:"photos"`
	Views          []interface{} `json:"views"`
}

// BedType represents a bed type in a room
type BedType struct {
	Quantity int    `json:"quantity"`
	BedType  string `json:"bed_type"`
	BedSize  string `json:"bed_size"`
	ID       int    `json:"id"`
}

// RoomAmenity represents an amenity in a room
type RoomAmenity struct {
	AmenitiesID int    `json:"amenities_id"`
	Name        string `json:"name"`
	Sort        int    `json:"sort"`
}

// ParseProperty parses JSON data into a Property struct
func ParseProperty(data []byte) (*Property, error) {
	var p Property
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// JSON returns the Property as JSON bytes
func (p *Property) JSON() ([]byte, error) {
	return json.Marshal(p)
}
