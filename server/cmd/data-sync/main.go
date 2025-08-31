package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/vrnvu/cupid/internal/client"
	"github.com/vrnvu/cupid/internal/database"
	"github.com/vrnvu/cupid/internal/telemetry"
)

func main() {
	cupidSandboxAPI, ok := os.LookupEnv("CUPID_SANDBOX_API")
	if !ok {
		panic("CUPID_SANDBOX_API env var was not set")
	}

	// Enable telemetry only if explicitly opted in
	if os.Getenv("ENABLE_TELEMETRY") == "1" {
		otelShutdown, err := telemetry.ConfigureOpenTelemetry()
		if err != nil {
			log.Fatalf("failed to configure OpenTelemetry: %v", err)
		}
		defer otelShutdown()
	}

	baseURL := os.Getenv("CUPID_BASE_URL")
	if baseURL == "" {
		baseURL = "https://content-api.cupid.travel"
	}

	hotelID := os.Getenv("HOTEL_ID")
	if hotelID == "" {
		hotelID = "1641879"
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

	c, err := client.New(baseURL,
		client.WithTimeout(5*time.Second),
		client.WithUserAgent("cupid-data-sync/1.0"),
		client.WithConnectionClose(),
	)
	if err != nil {
		panic(err)
	}

	headers := make(http.Header)
	headers.Add("accept", "application/json")
	headers.Add("x-api-key", cupidSandboxAPI)

	path := fmt.Sprintf("/v3.0/property/%s", hotelID)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	body, resp, err := c.Do(ctx, http.MethodGet, path, nil, headers)
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("status=%d", resp.StatusCode)

	property, err := client.ParseProperty(body)
	if err != nil {
		log.Fatalf("failed to parse property: %v", err)
	}

	log.Printf("parsed property: %s", property.HotelName)

	if err := repository.StoreProperty(ctx, property); err != nil {
		log.Fatalf("failed to store property: %v", err)
	}

	log.Printf("successfully stored property %d: %s", property.HotelID, property.HotelName)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
