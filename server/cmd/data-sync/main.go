package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/vrnvu/cupid/internal/client"
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
	log.Printf("status=%d body=%s", resp.StatusCode, string(body))
}
