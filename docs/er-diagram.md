# Cupid Hotel Management System - Entity Relationship Diagram

## Interactive ER Diagram

```mermaid
erDiagram
    hotels ||--o{ hotel_addresses : "has one"
    hotels ||--o{ hotel_checkins : "has one"
    hotels ||--o{ hotel_photos : "has many"
    hotels ||--o{ hotel_facilities : "has many"
    hotels ||--o{ hotel_policies : "has many"
    hotels ||--o{ hotel_rooms : "has many"
    hotels ||--o{ reviews : "has many"
    hotels ||--o{ translations : "has many"
    
    hotel_checkins ||--o{ hotel_checkin_instructions : "has many"
    hotel_rooms ||--o{ room_bed_types : "has many"
    hotel_rooms ||--o{ room_amenities : "has many"
    hotel_rooms ||--o{ room_photos : "has many"
    
    reviews ||--o{ translations : "has many"
    hotel_rooms ||--o{ translations : "has many"
    hotel_facilities ||--o{ translations : "has many"
    
    hotels {
        serial id PK
        integer hotel_id UK
        integer cupid_id
        text main_image_th
        varchar hotel_type
        integer hotel_type_id
        varchar chain
        integer chain_id
        decimal latitude
        decimal longitude
        varchar hotel_name
        varchar phone
        varchar fax
        varchar email
        integer stars
        varchar airport_code
        decimal rating
        integer review_count
        varchar parking
        integer group_room_min
        boolean child_allowed
        boolean pets_allowed
        text description
        text markdown_description
        text important_info
        timestamp created_at
        timestamp updated_at
    }
    
    hotel_addresses {
        serial id PK
        integer hotel_id FK
        text address
        varchar city
        varchar state
        varchar country
        varchar postal_code
        timestamp created_at
    }
    
    hotel_checkins {
        serial id PK
        integer hotel_id FK
        varchar checkin_start
        varchar checkin_end
        varchar checkout
        text special_instructions
        timestamp created_at
    }
    
    hotel_checkin_instructions {
        serial id PK
        integer hotel_checkin_id FK
        text instruction
        integer sort_order
    }
    
    hotel_photos {
        serial id PK
        integer hotel_id FK
        text url
        text hd_url
        text image_description
        varchar image_class1
        varchar image_class2
        boolean main_photo
        decimal score
        integer class_id
        integer class_order
        timestamp created_at
    }
    
    hotel_facilities {
        serial id PK
        integer hotel_id FK
        integer facility_id
        varchar name
        timestamp created_at
    }
    
    hotel_policies {
        serial id PK
        integer hotel_id FK
        varchar policy_type
        varchar name
        text description
        varchar child_allowed
        varchar pets_allowed
        varchar parking
        integer cupid_policy_id
        timestamp created_at
    }
    
    hotel_rooms {
        serial id PK
        integer hotel_id FK
        integer cupid_room_id
        varchar room_name
        text description
        integer room_size_square
        varchar room_size_unit
        integer max_adults
        integer max_children
        integer max_occupancy
        varchar bed_relation
        timestamp created_at
    }
    
    room_bed_types {
        serial id PK
        integer room_id FK
        integer quantity
        varchar bed_type
        varchar bed_size
        integer cupid_bed_id
        timestamp created_at
    }
    
    room_amenities {
        serial id PK
        integer room_id FK
        integer amenities_id
        varchar name
        integer sort_order
        timestamp created_at
    }
    
    room_photos {
        serial id PK
        integer room_id FK
        text url
        text hd_url
        text image_description
        varchar image_class1
        varchar image_class2
        boolean main_photo
        decimal score
        integer class_id
        integer class_order
        timestamp created_at
    }
    
    translations {
        serial id PK
        varchar entity_type
        integer entity_id
        varchar language_code
        varchar field_name
        text translated_text
        timestamp created_at
        timestamp updated_at
    }
    
    reviews {
        serial id PK
        integer hotel_id FK
        varchar reviewer_name
        integer rating
        varchar title
        text content
        varchar language_code
        date review_date
        integer helpful_votes
        timestamp created_at
        vector embedding
        varchar embedding_status
        timestamp embedding_updated_at
    }
```

## Key Relationships

### One-to-One Relationships
- **hotels ↔ hotel_addresses**: Each hotel has exactly one address
- **hotels ↔ hotel_checkins**: Each hotel has exactly one check-in policy

### One-to-Many Relationships
- **hotels → hotel_photos**: One hotel can have multiple photos
- **hotels → hotel_facilities**: One hotel can have multiple facilities
- **hotels → hotel_policies**: One hotel can have multiple policies
- **hotels → hotel_rooms**: One hotel can have multiple rooms
- **hotels → reviews**: One hotel can have multiple reviews
- **hotels → translations**: One hotel can have multiple translations

### Many-to-Many Relationships (via junction tables)
- **hotels ↔ translations**: Hotels can have translations in multiple languages
- **hotel_rooms ↔ translations**: Rooms can have translations in multiple languages
- **reviews ↔ translations**: Reviews can have translations in multiple languages

## Database Schema Features

### Multi-Language Support
- **translations** table supports English (en), French (fr), Spanish (es)
- Generic entity system for hotels, rooms, facilities, and reviews
- Field-level translation granularity

### AI-Powered Search
- **reviews** table includes vector embeddings (1536 dimensions)
- HNSW index for fast semantic similarity search
- Embedding status tracking for pipeline management

### Performance Optimizations
- Proper indexing on foreign keys and frequently queried fields
- Geographic indexing for location-based queries
- Vector search optimization for AI features

### Data Integrity
- Foreign key constraints with cascading deletes
- Check constraints for data validation
- Unique constraints to prevent duplicates
- Automatic timestamp management via triggers
