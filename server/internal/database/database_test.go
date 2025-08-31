package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConnection(t *testing.T) {
	t.Parallel()

	t.Run("successful connection", func(t *testing.T) {
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

		assert.NotNil(t, db)
		assert.NotNil(t, db.DB)
	})

	t.Run("connection with invalid credentials", func(t *testing.T) {
		config := Config{
			Host:     "localhost",
			Port:     5432,
			User:     "invalid_user",
			Password: "invalid_password",
			DBName:   "cupid",
			SSLMode:  "disable",
		}

		db, err := NewConnection(config)
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "failed to ping database")
	})

	t.Run("connection with invalid host", func(t *testing.T) {
		config := Config{
			Host:     "invalid-host",
			Port:     5432,
			User:     "cupid",
			Password: "cupid123",
			DBName:   "cupid",
			SSLMode:  "disable",
		}

		db, err := NewConnection(config)
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "failed to ping database")
	})

	t.Run("connection with invalid port", func(t *testing.T) {
		config := Config{
			Host:     "localhost",
			Port:     9999, // Invalid port
			User:     "cupid",
			Password: "cupid123",
			DBName:   "cupid",
			SSLMode:  "disable",
		}

		db, err := NewConnection(config)
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "failed to ping database")
	})
}

func TestDB_Ping(t *testing.T) {
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

	t.Run("successful ping", func(t *testing.T) {
		ctx := context.Background()
		err := db.Ping(ctx)
		assert.NoError(t, err)
	})

	t.Run("ping with timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := db.Ping(ctx)
		assert.NoError(t, err)
	})

	t.Run("ping with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := db.Ping(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestDB_Close(t *testing.T) {
	t.Parallel()

	config := Config{
		Host:     "localhost",
		Port:     5432,
		User:     "cupid",
		Password: "cupid123",
		DBName:   "cupid",
		SSLMode:  "disable",
	}

	t.Run("close connection", func(t *testing.T) {
		db, err := NewConnection(config)
		require.NoError(t, err)

		// Verify connection is working
		err = db.Ping(context.Background())
		assert.NoError(t, err)

		// Close connection
		err = db.Close()
		assert.NoError(t, err)

		// Verify connection is closed
		err = db.Ping(context.Background())
		assert.Error(t, err)
	})

	t.Run("close already closed connection", func(t *testing.T) {
		db, err := NewConnection(config)
		require.NoError(t, err)

		// Close once
		err = db.Close()
		assert.NoError(t, err)

		// Close again
		err = db.Close()
		assert.NoError(t, err) // Should not error on second close
	})
}

func TestConnectionPoolSettings(t *testing.T) {
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

	t.Run("verify connection pool settings", func(t *testing.T) {
		// These settings are set in NewConnection
		stats := db.DB.Stats()
		assert.Equal(t, 25, stats.MaxOpenConnections)
		// Note: MaxIdleConnections is not available in sql.DBStats
		// We can only verify MaxOpenConnections and current stats
	})

	t.Run("connection pool stats", func(t *testing.T) {
		stats := db.DB.Stats()
		assert.GreaterOrEqual(t, stats.OpenConnections, 0)
		assert.GreaterOrEqual(t, stats.InUse, 0)
		assert.GreaterOrEqual(t, stats.Idle, 0)
	})
}

func TestConcurrentConnections(t *testing.T) {
	t.Parallel()

	config := Config{
		Host:     "localhost",
		Port:     5432,
		User:     "cupid",
		Password: "cupid123",
		DBName:   "cupid",
		SSLMode:  "disable",
	}

	t.Run("multiple concurrent connections", func(t *testing.T) {
		const numConnections = 5
		connections := make([]*DB, numConnections)
		errors := make(chan error, numConnections)

		// Create multiple connections concurrently
		for i := 0; i < numConnections; i++ {
			go func(id int) {
				db, err := NewConnection(config)
				if err != nil {
					errors <- err
					return
				}
				connections[id] = db
				errors <- nil
			}(i)
		}

		// Wait for all connections to be established
		for i := 0; i < numConnections; i++ {
			err := <-errors
			assert.NoError(t, err)
		}

		// Test ping on all connections
		for i, db := range connections {
			if db != nil {
				err := db.Ping(context.Background())
				assert.NoError(t, err, "Connection %d failed ping", i)
				db.Close()
			}
		}
	})
}

func TestConnectionStringFormat(t *testing.T) {
	t.Parallel()

	t.Run("DSN format verification", func(t *testing.T) {
		config := Config{
			Host:     "test-host",
			Port:     5433,
			User:     "test-user",
			Password: "test-pass",
			DBName:   "test-db",
			SSLMode:  "require",
		}

		// This would normally fail to connect, but we're testing the DSN format
		// The error should be about connection, not DSN format
		db, err := NewConnection(config)
		if err != nil {
			// Should fail with connection error, not DSN format error
			assert.Contains(t, err.Error(), "failed to ping database")
		} else {
			db.Close()
		}
	})
}
