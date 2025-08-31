-- Initial schema for Cupid hotel properties
-- This migration creates the core tables for storing hotel data

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Hotels table
CREATE TABLE hotels (
    id SERIAL PRIMARY KEY,
    hotel_id INTEGER UNIQUE NOT NULL,
    cupid_id INTEGER NOT NULL,
    main_image_th TEXT,
    hotel_type VARCHAR(100),
    hotel_type_id INTEGER,
    chain VARCHAR(255),
    chain_id INTEGER,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    hotel_name VARCHAR(500) NOT NULL,
    phone VARCHAR(50),
    fax VARCHAR(50),
    email VARCHAR(255),
    stars INTEGER,
    airport_code VARCHAR(10),
    rating DECIMAL(3, 2),
    review_count INTEGER DEFAULT 0,
    parking VARCHAR(50),
    group_room_min INTEGER,
    child_allowed BOOLEAN DEFAULT false,
    pets_allowed BOOLEAN DEFAULT false,
    description TEXT,
    markdown_description TEXT,
    important_info TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Hotel addresses table
CREATE TABLE hotel_addresses (
    id SERIAL PRIMARY KEY,
    hotel_id INTEGER REFERENCES hotels(hotel_id) ON DELETE CASCADE UNIQUE,
    address TEXT,
    city VARCHAR(255),
    state VARCHAR(255),
    country VARCHAR(10),
    postal_code VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Hotel check-in information
CREATE TABLE hotel_checkins (
    id SERIAL PRIMARY KEY,
    hotel_id INTEGER REFERENCES hotels(hotel_id) ON DELETE CASCADE UNIQUE,
    checkin_start VARCHAR(10),
    checkin_end VARCHAR(10),
    checkout VARCHAR(10),
    special_instructions TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Hotel check-in instructions
CREATE TABLE hotel_checkin_instructions (
    id SERIAL PRIMARY KEY,
    hotel_checkin_id INTEGER REFERENCES hotel_checkins(id) ON DELETE CASCADE,
    instruction TEXT NOT NULL,
    sort_order INTEGER DEFAULT 0
);

-- Hotel photos table
CREATE TABLE hotel_photos (
    id SERIAL PRIMARY KEY,
    hotel_id INTEGER REFERENCES hotels(hotel_id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    hd_url TEXT,
    image_description TEXT,
    image_class1 VARCHAR(100),
    image_class2 VARCHAR(100),
    main_photo BOOLEAN DEFAULT false,
    score DECIMAL(5, 2),
    class_id INTEGER,
    class_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(hotel_id, url)
);

-- Hotel facilities table
CREATE TABLE hotel_facilities (
    id SERIAL PRIMARY KEY,
    hotel_id INTEGER REFERENCES hotels(hotel_id) ON DELETE CASCADE,
    facility_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(hotel_id, facility_id)
);

-- Hotel policies table
CREATE TABLE hotel_policies (
    id SERIAL PRIMARY KEY,
    hotel_id INTEGER REFERENCES hotels(hotel_id) ON DELETE CASCADE,
    policy_type VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    child_allowed VARCHAR(50),
    pets_allowed VARCHAR(50),
    parking VARCHAR(50),
    cupid_policy_id INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(hotel_id, cupid_policy_id)
);

-- Hotel rooms table
CREATE TABLE hotel_rooms (
    id SERIAL PRIMARY KEY,
    hotel_id INTEGER REFERENCES hotels(hotel_id) ON DELETE CASCADE,
    cupid_room_id INTEGER NOT NULL,
    room_name VARCHAR(255) NOT NULL,
    description TEXT,
    room_size_square INTEGER,
    room_size_unit VARCHAR(10),
    max_adults INTEGER DEFAULT 1,
    max_children INTEGER DEFAULT 0,
    max_occupancy INTEGER DEFAULT 1,
    bed_relation VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(hotel_id, cupid_room_id)
);

-- Room bed types table
CREATE TABLE room_bed_types (
    id SERIAL PRIMARY KEY,
    room_id INTEGER REFERENCES hotel_rooms(id) ON DELETE CASCADE,
    quantity INTEGER DEFAULT 1,
    bed_type VARCHAR(100) NOT NULL,
    bed_size VARCHAR(100),
    cupid_bed_id INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(room_id, cupid_bed_id)
);

-- Room amenities table
CREATE TABLE room_amenities (
    id SERIAL PRIMARY KEY,
    room_id INTEGER REFERENCES hotel_rooms(id) ON DELETE CASCADE,
    amenities_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(room_id, amenities_id)
);

-- Room photos table
CREATE TABLE room_photos (
    id SERIAL PRIMARY KEY,
    room_id INTEGER REFERENCES hotel_rooms(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    hd_url TEXT,
    image_description TEXT,
    image_class1 VARCHAR(100),
    image_class2 VARCHAR(100),
    main_photo BOOLEAN DEFAULT false,
    score DECIMAL(5, 2),
    class_id INTEGER,
    class_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(room_id, url)
);

-- Translations table for multi-language support
CREATE TABLE translations (
    id SERIAL PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL, -- 'hotel', 'room', 'facility', etc.
    entity_id INTEGER NOT NULL,
    language_code VARCHAR(10) NOT NULL, -- 'en', 'fr', 'es'
    field_name VARCHAR(100) NOT NULL, -- 'name', 'description', etc.
    translated_text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(entity_type, entity_id, language_code, field_name)
);

-- Reviews table
CREATE TABLE reviews (
    id SERIAL PRIMARY KEY,
    hotel_id INTEGER REFERENCES hotels(hotel_id) ON DELETE CASCADE,
    reviewer_name VARCHAR(255),
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    title VARCHAR(500),
    content TEXT,
    language_code VARCHAR(10) DEFAULT 'en',
    review_date DATE,
    helpful_votes INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX idx_hotels_hotel_id ON hotels(hotel_id);
CREATE INDEX idx_hotels_chain_id ON hotels(chain_id);
CREATE INDEX idx_hotels_location ON hotels(latitude, longitude);
CREATE INDEX idx_hotel_photos_hotel_id ON hotel_photos(hotel_id);
CREATE INDEX idx_hotel_rooms_hotel_id ON hotel_rooms(hotel_id);
CREATE INDEX idx_translations_entity ON translations(entity_type, entity_id, language_code);
CREATE INDEX idx_reviews_hotel_id ON reviews(hotel_id);
CREATE INDEX idx_reviews_rating ON reviews(rating);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add updated_at triggers
CREATE TRIGGER update_hotels_updated_at BEFORE UPDATE ON hotels
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_translations_updated_at BEFORE UPDATE ON translations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
