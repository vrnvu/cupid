# Cupid Database Documentation

This directory contains comprehensive documentation for the Cupid hotel management system database schema.

## Documentation Files

### [Database Schema Documentation](database-schema.md)
Complete database schema documentation including:
- Detailed table descriptions
- Relationship explanations
- Performance optimizations
- Security considerations
- Migration strategy

### [Entity Relationship Diagram](er-diagram.md)
Interactive ER diagram showing:
- All tables and their relationships
- Field definitions and data types
- Key relationship patterns
- Database features overview

### [Table Schema Quick Reference](table-schema-reference.md)
Developer-friendly quick reference including:
- Table summaries and purposes
- Common query patterns
- Index information
- Performance notes

## Database Overview

The Cupid system uses PostgreSQL with a normalized schema design supporting:

- **Hotel Management**: Properties, rooms, facilities, policies
- **Multi-Language Support**: English, French, Spanish translations
- **AI-Powered Search**: Vector embeddings for semantic review search
- **Media Management**: Photo galleries for hotels and rooms
- **Review System**: Customer feedback with ratings and content

## Key Features

### Hotel Data Model
- Comprehensive hotel information storage
- Geographic coordinates for location queries
- Chain and type categorization
- Rating and review metrics

### Localization System
- Generic translation entity system
- Field-level translation granularity
- Support for multiple languages
- Flexible content management

### AI Integration
- Vector embeddings for reviews
- HNSW indexing for fast similarity search
- Embedding pipeline status tracking
- Semantic search capabilities

### Media Management
- Hotel and room photo galleries
- Image classification and scoring
- HD and standard resolution support
- Main photo selection

## Quick Start

1. **View ER Diagram**: Open `er-diagram.md` to see the complete database structure
2. **Reference Schema**: Use `table-schema-reference.md` for quick lookups
3. **Deep Dive**: Read `database-schema.md` for comprehensive understanding

## Database Extensions

- **pgvector**: Vector similarity search
- **uuid-ossp**: UUID generation support

## Performance Features

- Optimized indexing strategy
- Connection pooling (25 max connections)
- Geographic query optimization
- Vector search acceleration

## Development Notes

- All tests use `t.Parallel()` for concurrent execution
- Database cleanup between tests with `t.Cleanup()`
- Comprehensive foreign key constraints
- Automatic timestamp management via triggers
