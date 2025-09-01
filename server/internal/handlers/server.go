package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/vrnvu/cupid/internal/cache"
	"github.com/vrnvu/cupid/internal/client"
	"github.com/vrnvu/cupid/internal/database"
	"github.com/vrnvu/cupid/internal/telemetry"
)

type Server struct {
	repository database.Repository
	cache      cache.ReviewCache
}

func NewServer(repository database.Repository, cache cache.ReviewCache) http.Handler {
	server := &Server{repository: repository, cache: cache}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		handler := telemetry.NewHandler(http.HandlerFunc(server.healthHandler), "HealthCheck")
		handler.ServeHTTP(w, r)
	})
	mux.HandleFunc("GET /api/v1/hotels", func(w http.ResponseWriter, r *http.Request) {
		handler := telemetry.NewHandler(http.HandlerFunc(server.getHotelsHandler), "HotelsHandler")
		handler.ServeHTTP(w, r)
	})

	mux.HandleFunc("GET /api/v1/hotels/{hotelID}", func(w http.ResponseWriter, r *http.Request) {
		handler := telemetry.NewHandler(http.HandlerFunc(server.getHotelHandler), "HotelHandler")
		handler.ServeHTTP(w, r)
	})
	mux.HandleFunc("GET /api/v1/hotels/{hotelID}/reviews", func(w http.ResponseWriter, r *http.Request) {
		handler := telemetry.NewHandler(http.HandlerFunc(server.getHotelReviewsHandler), "HotelReviewsHandler")
		handler.ServeHTTP(w, r)
	})
	mux.HandleFunc("GET /api/v1/hotels/{hotelID}/translations/{language}", func(w http.ResponseWriter, r *http.Request) {
		handler := telemetry.NewHandler(http.HandlerFunc(server.getHotelTranslationsHandler), "HotelTranslationsHandler")
		handler.ServeHTTP(w, r)
	})
	mux.HandleFunc("GET /api/v1/reviews/search", func(w http.ResponseWriter, r *http.Request) {
		handler := telemetry.NewHandler(http.HandlerFunc(server.searchReviewsHandler), "SearchReviewsHandler")
		handler.ServeHTTP(w, r)
	})

	return mux
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	if err := s.repository.Ping(ctx); err != nil {
		http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "cupid-api",
	}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) getHotelsHandler(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	ctx := r.Context()
	hotels, err := s.repository.GetHotels(ctx, limit, offset)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"hotels": hotels,
		"count":  len(hotels),
		"limit":  limit,
		"offset": offset,
	}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) getHotelHandler(w http.ResponseWriter, r *http.Request) {
	hotelIDStr := r.PathValue("hotelID")
	if hotelIDStr == "" {
		http.Error(w, "Invalid hotel ID", http.StatusBadRequest)
		return
	}

	hotelID, err := strconv.Atoi(hotelIDStr)
	if err != nil {
		http.Error(w, "Invalid hotel ID format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	hotel, err := s.repository.GetHotelByID(ctx, hotelID)
	if err != nil {
		if errors.Is(err, database.ErrHotelNotFound) {
			http.Error(w, fmt.Sprintf("Hotel with ID %d not found", hotelID), http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(hotel); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) getHotelReviewsHandler(w http.ResponseWriter, r *http.Request) {
	hotelIDStr := r.PathValue("hotelID")
	if hotelIDStr == "" {
		http.Error(w, "Invalid hotel ID", http.StatusBadRequest)
		return
	}

	hotelID, err := strconv.Atoi(hotelIDStr)
	if err != nil {
		http.Error(w, "Invalid hotel ID format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	var reviews []client.Review
	var fromCache bool

	if s.cache != nil {
		cachedReviews, err := s.cache.GetReviews(ctx, hotelID)
		if err == nil && cachedReviews != nil {
			reviews = cachedReviews
			fromCache = true
		}
	}

	if !fromCache {
		reviews, err = s.repository.GetHotelReviews(ctx, hotelID)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if s.cache != nil {
			cacheErr := s.cache.SetReviews(ctx, hotelID, reviews, 5*time.Minute)
			if cacheErr != nil {
				// Log cache error but don't fail the request
				fmt.Printf("Warning: Failed to cache reviews for hotel %d: %v\n", hotelID, cacheErr)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"hotel_id":   hotelID,
		"reviews":    reviews,
		"count":      len(reviews),
		"from_cache": fromCache,
		"cached_at":  time.Now().Format(time.RFC3339),
	}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) getHotelTranslationsHandler(w http.ResponseWriter, r *http.Request) {
	hotelIDStr := r.PathValue("hotelID")
	if hotelIDStr == "" {
		http.Error(w, "Invalid hotel ID", http.StatusBadRequest)
		return
	}

	hotelID, err := strconv.Atoi(hotelIDStr)
	if err != nil {
		http.Error(w, "Invalid hotel ID format", http.StatusBadRequest)
		return
	}

	languageCode := r.PathValue("language")
	if languageCode == "" {
		http.Error(w, "Invalid language code", http.StatusBadRequest)
		return
	}

	// Validate language code
	validLanguages := map[string]bool{"fr": true, "es": true, "en": true}
	if !validLanguages[languageCode] {
		http.Error(w, "Unsupported language code. Supported: fr, es, en", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	translations, err := s.repository.GetHotelTranslations(ctx, hotelID, languageCode)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"hotel_id":     hotelID,
		"language":     languageCode,
		"translations": translations,
		"count":        len(translations),
	}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) searchReviewsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 100 {
				limit = 100 // max limit
			}
		}
	}

	// Parse threshold parameter
	thresholdStr := r.URL.Query().Get("threshold")
	threshold := 0.7 // default similarity threshold
	if thresholdStr != "" {
		if parsedThreshold, err := strconv.ParseFloat(thresholdStr, 64); err == nil && parsedThreshold > 0 {
			threshold = parsedThreshold
		}
	}

	// TODO: Generate embedding for the query text and perform vector search
	// For now, return a placeholder response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"query":     query,
		"limit":     limit,
		"threshold": threshold,
		"message":   "Vector search endpoint ready - embedding generation not yet implemented",
		"reviews":   []client.Review{},
		"count":     0,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
