package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/vrnvu/cupid/internal/client"
	"github.com/vrnvu/cupid/internal/database"
	"github.com/vrnvu/cupid/internal/telemetry"
)

var allHotelIDs = []int{
	1641879, 317597, 1202743, 1037179, 1154868, 1270324, 1305326, 1617655, 1975211, 2017823, 1503950, 1033299, 378772, 1563003, 1085875, 828917, 830417, 838887, 1702062, 1144294, 1738870, 898052, 906450, 906467, 2241195, 1244595, 1277032, 956026, 957111, 152896, 896868, 982911, 986491, 986622, 988544, 989315, 989544, 990223, 990341, 990370, 990490, 990609, 990629, 1259611, 991819, 992027, 992851, 993851, 994085, 994333, 994495, 994903, 995227, 995787, 996977, 1186578, 999444, 1000017, 1000051, 1198750, 1001100, 1001296, 1001402, 1002200, 1003142, 1004288, 1006404, 1006602, 1006810, 1006887, 1007101, 1007269, 1007466, 1011203, 1011644, 1011945, 1012047, 1012140, 1012944, 1023527, 1013529, 1013584, 1014383, 1015094, 1016591, 1016611, 1017019, 1017039, 1017044, 1018030, 1018130, 1018251, 1018402, 1018946, 1019473, 1020332, 1020335, 1020386, 1021856, 1022380,
}

type EndpointType string

const (
	ContentEndpoint      EndpointType = "content"
	ReviewsEndpoint      EndpointType = "reviews"
	TranslationsEndpoint EndpointType = "translations"
)

func main() {
	var endpointType string
	flag.StringVar(&endpointType, "e", "content", "Endpoint type: content, reviews, or translations")
	flag.Parse()

	et := EndpointType(endpointType)
	if et != ContentEndpoint && et != ReviewsEndpoint && et != TranslationsEndpoint {
		log.Fatalf("Invalid endpoint type: %s. Must be one of: content, reviews, translations", endpointType)
	}

	cupidSandboxAPI, ok := os.LookupEnv("CUPID_SANDBOX_API")
	if !ok {
		panic("CUPID_SANDBOX_API env var was not set")
	}

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

	singleHotelID := os.Getenv("HOTEL_ID")
	if singleHotelID != "" {
		hotelID, err := strconv.Atoi(singleHotelID)
		if err != nil {
			log.Fatalf("invalid hotel ID: %s", singleHotelID)
		}
		log.Printf("Starting sync for hotel %d", hotelID)
		if err := syncHotel(hotelID, baseURL, cupidSandboxAPI, et); err != nil {
			log.Printf("Failed to sync hotel %d: %v", hotelID, err)
		}
		log.Printf("Completed sync for hotel %d", hotelID)
	} else {
		log.Printf("Starting batch sync of %d hotels", len(allHotelIDs))
		successCount := 0

		for i, hotelID := range allHotelIDs {
			log.Printf("Processing hotel %d (%d/%d)", hotelID, i+1, len(allHotelIDs))
			if err := syncHotel(hotelID, baseURL, cupidSandboxAPI, et); err == nil {
				successCount++
			} else {
				log.Printf("Failed to sync hotel %d: %v", hotelID, err)
			}
			time.Sleep(100 * time.Millisecond)
		}

		log.Printf("Batch sync completed: %d successful, %d failed", successCount, len(allHotelIDs)-successCount)
	}
}

func syncHotel(hotelID int, baseURL, apiKey string, endpointType EndpointType) error {
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
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	repository := database.NewHotelRepository(db)

	c, err := client.New(baseURL,
		client.WithTimeout(10*time.Second),
		client.WithUserAgent("cupid-data-sync/1.0"),
		client.WithConnectionClose(),
	)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	headers := make(http.Header)
	headers.Add("accept", "application/json")
	headers.Add("x-api-key", apiKey)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	switch endpointType {
	case ContentEndpoint:
		return syncHotelContent(ctx, c, headers, hotelID, repository)
	case ReviewsEndpoint:
		return syncHotelReviews(ctx, c, headers, hotelID, repository)
	case TranslationsEndpoint:
		return syncHotelTranslations(ctx, c, headers, hotelID, repository)
	default:
		return fmt.Errorf("unknown endpoint type: %s", endpointType)
	}
}

func syncHotelContent(ctx context.Context, c *client.Client, headers http.Header, hotelID int, repository *database.HotelRepository) error {
	path := fmt.Sprintf("/v3.0/property/%d", hotelID)

	body, resp, err := c.Do(ctx, http.MethodGet, path, nil, headers)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	property, err := client.ParseProperty(body)
	if err != nil {
		return fmt.Errorf("failed to parse property: %w", err)
	}

	if err := repository.StoreProperty(ctx, property); err != nil {
		return fmt.Errorf("failed to store property: %w", err)
	}

	return nil
}

func syncHotelReviews(ctx context.Context, c *client.Client, headers http.Header, hotelID int, repository *database.HotelRepository) error {
	reviewCount := 100
	path := fmt.Sprintf("/v3.0/property/reviews/%d/%d", hotelID, reviewCount)

	body, resp, err := c.Do(ctx, http.MethodGet, path, nil, headers)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if len(body) > 0 {
		reviews, err := client.ParseReviews(body)
		if err != nil {
			return fmt.Errorf("failed to parse reviews: %w", err)
		}

		if len(reviews) > 0 {
			if err := repository.StoreReviews(ctx, hotelID, reviews); err != nil {
				return fmt.Errorf("failed to store reviews: %w", err)
			}
		}
	}

	return nil
}

func syncHotelTranslations(ctx context.Context, c *client.Client, headers http.Header, hotelID int, repository *database.HotelRepository) error {
	languages := []string{"fr", "es", "en"}
	var allTranslations []client.Translation

	for _, lang := range languages {
		path := fmt.Sprintf("/v3.0/property/%d/lang/%s", hotelID, lang)

		body, resp, err := c.Do(ctx, http.MethodGet, path, nil, headers)
		if err != nil {
			log.Printf("Failed to get %s translations for hotel %d: %v", lang, hotelID, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK && len(body) > 0 {
			translations, err := client.ParseTranslations(body)
			if err != nil {
				log.Printf("Failed to parse %s translations for hotel %d: %v", lang, hotelID, err)
				continue
			}

			for _, translation := range translations {
				translation.LanguageCode = lang
				allTranslations = append(allTranslations, translation)
			}
		}
	}

	if len(allTranslations) > 0 {
		if err := repository.StoreTranslations(ctx, hotelID, allTranslations); err != nil {
			return fmt.Errorf("failed to store translations: %w", err)
		}
	}

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
