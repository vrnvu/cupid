package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/vrnvu/cupid/internal/database"
	"github.com/vrnvu/cupid/internal/telemetry"
)

type Server struct {
	repository database.Repository
}

func NewServer(repository database.Repository) http.Handler {
	server := &Server{repository: repository}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		handler := telemetry.NewHandler(http.HandlerFunc(server.healthHandler), "HealthCheck")
		handler.ServeHTTP(w, r)
	})
	mux.HandleFunc("GET /api/v1/hotels/{hotelID}", func(w http.ResponseWriter, r *http.Request) {
		handler := telemetry.NewHandler(http.HandlerFunc(server.getHotelHandler), "HotelHandler")
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
