package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/vrnvu/cupid/internal/ai"
	"github.com/vrnvu/cupid/internal/database"
	"github.com/vrnvu/cupid/internal/telemetry"
)

type ReviewData struct {
	ID      int
	Title   string
	Content string
}

func main() {
	openaiAPIKey, ok := os.LookupEnv("OPENAI_API_KEY")
	if !ok {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	hotelIDList := []int{1641879, 317597, 1202743}

	if os.Getenv("ENABLE_TELEMETRY") == "1" {
		otelShutdown, err := telemetry.ConfigureOpenTelemetry()
		if err != nil {
			log.Fatalf("failed to configure OpenTelemetry: %v", err)
		}
		defer otelShutdown()
	}

	dbConfig := database.Config{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefaultInt("DB_PORT", 5432),
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
	aiService := ai.NewService(openaiAPIKey)

	ctx := context.Background()

	log.Printf("Processing reviews for hotels: %v", hotelIDList)

	processed := 0
	for _, hotelID := range hotelIDList {
		log.Printf("Processing hotel %d...", hotelID)

		count, err := processHotelReviews(ctx, repository, aiService, hotelID)
		if err != nil {
			log.Printf("Failed to process hotel %d: %v", hotelID, err)
			continue
		}

		processed += count
		log.Printf("Processed %d reviews for hotel %d", count, hotelID)
	}

	log.Printf("Successfully processed %d reviews", processed)
}

func processHotelReviews(ctx context.Context, repo *database.HotelRepository, aiService ai.Service, hotelID int) (int, error) {
	reviews, err := getHotelReviewsNeedingEmbeddings(ctx, repo, hotelID)
	if err != nil {
		return 0, fmt.Errorf("failed to get reviews: %w", err)
	}

	if len(reviews) == 0 {
		return 0, nil
	}

	log.Printf("Found %d reviews needing embeddings for hotel %d", len(reviews), hotelID)

	processed := 0
	for _, review := range reviews {
		text := fmt.Sprintf("%s %s", review.Title, review.Content)
		embedding, err := aiService.GenerateEmbedding(ctx, text)
		if err != nil {
			log.Printf("Failed to generate embedding for review %d: %v", review.ID, err)
			if markErr := markReviewEmbeddingStatus(ctx, repo, review.ID, "failed"); markErr != nil {
				log.Printf("Failed to mark review %d as failed: %v", review.ID, markErr)
			}
			continue
		}

		if err := storeReviewEmbedding(ctx, repo, review.ID, embedding); err != nil {
			log.Printf("Failed to store embedding for review %d: %v", review.ID, err)
			continue
		}

		processed++
	}

	return processed, nil
}

func getHotelReviewsNeedingEmbeddings(ctx context.Context, repo *database.HotelRepository, hotelID int) ([]ReviewData, error) {
	query := `
		SELECT id, title, content 
		FROM reviews 
		WHERE hotel_id = $1 
		AND embedding_status IN ('pending', 'failed')
		AND content IS NOT NULL 
		AND LENGTH(TRIM(content)) > 0
		ORDER BY created_at ASC`

	rows, err := repo.GetDB().QueryContext(ctx, query, hotelID)
	if err != nil {
		return nil, fmt.Errorf("failed to query reviews: %w", err)
	}
	defer rows.Close()

	var reviews []ReviewData
	for rows.Next() {
		var review ReviewData
		if err := rows.Scan(&review.ID, &review.Title, &review.Content); err != nil {
			return nil, fmt.Errorf("failed to scan review: %w", err)
		}
		reviews = append(reviews, review)
	}

	return reviews, rows.Err()
}

func storeReviewEmbedding(ctx context.Context, repo *database.HotelRepository, reviewID int, embedding []float64) error {
	vectorStr := "[" + strings.Trim(strings.Join(strings.Fields(fmt.Sprint(embedding)), ","), "[]") + "]"

	query := `
		UPDATE reviews 
		SET embedding = $1::vector, 
		    embedding_status = 'completed', 
		    embedding_updated_at = NOW() 
		WHERE id = $2`

	_, err := repo.GetDB().ExecContext(ctx, query, vectorStr, reviewID)
	return err
}

func markReviewEmbeddingStatus(ctx context.Context, repo *database.HotelRepository, reviewID int, status string) error {
	query := `
		UPDATE reviews 
		SET embedding_status = $1, 
		    embedding_updated_at = NOW() 
		WHERE id = $2`

	_, err := repo.GetDB().ExecContext(ctx, query, status, reviewID)
	return err
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
