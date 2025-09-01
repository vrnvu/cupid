# Table Schema Quick Reference

## Core Tables

| Table | Primary Key | Foreign Keys | Key Fields | Purpose |
|-------|-------------|--------------|------------|---------|
| `hotels` | `id` (SERIAL) | - | `hotel_id`, `cupid_id`, `hotel_name` | Main hotel information |
| `hotel_addresses` | `id` (SERIAL) | `hotel_id` → `hotels.hotel_id` | `address`, `city`, `country` | Hotel location details |
| `hotel_checkins` | `id` (SERIAL) | `hotel_id` → `hotels.hotel_id` | `checkin_start`, `checkout` | Check-in/out policies |
| `hotel_photos` | `id` (SERIAL) | `hotel_id` → `hotels.hotel_id` | `url`, `main_photo`, `score` | Hotel image gallery |
| `hotel_facilities` | `id` (SERIAL) | `hotel_id` → `hotels.hotel_id` | `facility_id`, `name` | Hotel amenities |
| `hotel_policies` | `id` (SERIAL) | `hotel_id` → `hotels.hotel_id` | `policy_type`, `name` | Hotel rules |
| `hotel_rooms` | `id` (SERIAL) | `hotel_id` → `hotels.hotel_id` | `cupid_room_id`, `room_name` | Room configurations |
| `room_bed_types` | `id` (SERIAL) | `room_id` → `hotel_rooms.id` | `bed_type`, `quantity` | Bed configurations |
| `room_amenities` | `id` (SERIAL) | `room_id` → `hotel_rooms.id` | `amenities_id`, `name` | Room amenities |
| `room_photos` | `id` (SERIAL) | `room_id` → `hotel_rooms.id` | `url`, `main_photo` | Room images |
| `translations` | `id` (SERIAL) | - | `entity_type`, `entity_id`, `language_code` | Multi-language content |
| `reviews` | `id` (SERIAL) | `hotel_id` → `hotels.hotel_id` | `rating`, `content`, `embedding` | Customer feedback |

## Key Relationships

### Hotel Hierarchy
```
hotels (1) ←→ (1) hotel_addresses
hotels (1) ←→ (1) hotel_checkins
hotels (1) ←→ (N) hotel_photos
hotels (1) ←→ (N) hotel_facilities
hotels (1) ←→ (N) hotel_policies
hotels (1) ←→ (N) hotel_rooms
hotels (1) ←→ (N) reviews
```

### Room Hierarchy
```
hotel_rooms (1) ←→ (N) room_bed_types
hotel_rooms (1) ←→ (N) room_amenities
hotel_rooms (1) ←→ (N) room_photos
```

### Translation System
```
translations ←→ hotels (via entity_type='hotel', entity_id=hotel_id)
translations ←→ hotel_rooms (via entity_type='room', entity_id=room_id)
translations ←→ reviews (via entity_type='review', entity_id=review_id)
translations ←→ hotel_facilities (via entity_type='facility', entity_id=facility_id)
```

## Common Query Patterns

### Get Hotel with All Related Data
```sql
SELECT h.*, ha.*, hc.*, 
       array_agg(DISTINCT hp.url) as photo_urls,
       array_agg(DISTINCT hf.name) as facilities
FROM hotels h
LEFT JOIN hotel_addresses ha ON h.hotel_id = ha.hotel_id
LEFT JOIN hotel_checkins hc ON h.hotel_id = hc.hotel_id
LEFT JOIN hotel_photos hp ON h.hotel_id = hp.hotel_id
LEFT JOIN hotel_facilities hf ON h.hotel_id = hf.hotel_id
WHERE h.hotel_id = $1
GROUP BY h.id, ha.id, hc.id;
```

### Get Hotel Translations
```sql
SELECT field_name, translated_text, language_code
FROM translations
WHERE entity_type = 'hotel' 
  AND entity_id = $1 
  AND language_code = $2;
```

### Vector Search for Similar Reviews
```sql
SELECT content, rating, 
       1 - (embedding <=> $1) as similarity
FROM reviews
WHERE embedding_status = 'completed'
ORDER BY embedding <=> $1
LIMIT 10;
```

## Indexes

| Index | Table | Columns | Purpose |
|-------|-------|---------|---------|
| `idx_hotels_hotel_id` | `hotels` | `hotel_id` | Primary lookup |
| `idx_hotels_chain_id` | `hotels` | `chain_id` | Chain-based queries |
| `idx_hotels_location` | `hotels` | `latitude, longitude` | Geographic queries |
| `idx_hotel_photos_hotel_id` | `hotel_photos` | `hotel_id` | Photo lookups |
| `idx_hotel_rooms_hotel_id` | `hotel_rooms` | `hotel_id` | Room lookups |
| `idx_translations_entity` | `translations` | `entity_type, entity_id, language_code` | Translation lookups |
| `idx_reviews_hotel_id` | `reviews` | `hotel_id` | Review lookups |
| `idx_reviews_rating` | `reviews` | `rating` | Rating-based queries |
| `idx_reviews_embedding_hnsw` | `reviews` | `embedding` | Vector similarity search |
| `idx_reviews_embedding_status` | `reviews` | `embedding_status` | Pipeline filtering |

## Data Types

### Common Field Types
- **IDs**: `SERIAL` (auto-increment), `INTEGER` (external references)
- **Text**: `VARCHAR(n)` (limited), `TEXT` (unlimited)
- **Numbers**: `INTEGER`, `DECIMAL(p,s)` (precision, scale)
- **Booleans**: `BOOLEAN` (true/false)
- **Dates**: `DATE`, `TIMESTAMP WITH TIME ZONE`
- **Vectors**: `vector(1536)` (AI embeddings)

### Constraints
- **Primary Keys**: All tables have `id` as SERIAL primary key
- **Foreign Keys**: Proper referential integrity with CASCADE deletes
- **Unique Constraints**: Prevent duplicates where appropriate
- **Check Constraints**: Rating validation (1-5 scale)
- **NOT NULL**: Required fields properly constrained

## Triggers

| Trigger | Table | Function | Purpose |
|---------|-------|----------|---------|
| `update_hotels_updated_at` | `hotels` | `update_updated_at_column()` | Auto-update timestamps |
| `update_translations_updated_at` | `translations` | `update_updated_at_column()` | Auto-update timestamps |
| `update_reviews_embedding_status` | `reviews` | `update_embedding_status()` | Track embedding updates |

## Extensions

- **uuid-ossp**: UUID generation support
- **vector**: Vector similarity search (pgvector)

## Performance Notes

- Connection pool: 25 max connections
- Vector search: HNSW index for fast similarity
- Geographic queries: Composite index on lat/lng
- Batch operations: Support for bulk data sync
- Parallel testing: All tests use `t.Parallel()`
