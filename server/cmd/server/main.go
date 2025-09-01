package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vrnvu/cupid/internal/cache"
	"github.com/vrnvu/cupid/internal/database"
	"github.com/vrnvu/cupid/internal/handlers"
	"github.com/vrnvu/cupid/internal/telemetry"
)

func main() {
	if os.Getenv("ENABLE_TELEMETRY") == "1" {
		otelShutdown, err := telemetry.ConfigureOpenTelemetry()
		if err != nil {
			log.Fatalf("failed to configure OpenTelemetry: %v", err)
		}
		defer otelShutdown()
	}

	dbConfig := database.Config{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     5432,
		User:     getEnvOrDefault("DB_USER", "cupid"),
		Password: getEnvOrDefault("DB_PASSWORD", "cupid123"),
		DBName:   getEnvOrDefault("DB_NAME", "cupid"),
		SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
	}

	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	repository := database.NewHotelRepository(db)

	redisAddr := getEnvOrDefault("REDIS_HOST", "localhost") + ":" + getEnvOrDefault("REDIS_PORT", "6379")
	redisCache := cache.NewRedisCache(redisAddr)
	defer redisCache.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisCache.Ping(ctx); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
		log.Println("Continuing without cache...")
		redisCache = nil
	} else {
		log.Println("Redis cache connected successfully")
	}

	server := handlers.NewServer(repository, redisCache)

	port := getEnvOrDefault("PORT", "8080")
	addr := ":" + port

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      server,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting server on %s", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
