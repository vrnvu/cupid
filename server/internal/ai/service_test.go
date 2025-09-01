package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	service := NewService("test-api-key")
	embeddingService, ok := service.(*EmbeddingService)
	if !ok {
		t.Fatal("Expected service to be *EmbeddingService")
	}

	if embeddingService.apiKey != "test-api-key" {
		t.Errorf("Expected apiKey to be 'test-api-key', got %s", embeddingService.apiKey)
	}

	if embeddingService.baseURL != "https://api.openai.com/v1" {
		t.Errorf("Expected baseURL to be 'https://api.openai.com/v1', got %s", embeddingService.baseURL)
	}

	if embeddingService.model != "text-embedding-3-small" {
		t.Errorf("Expected model to be 'text-embedding-3-small', got %s", embeddingService.model)
	}

	if embeddingService.client == nil {
		t.Error("Expected client to be initialized")
	}
}

func TestGetModelInfo(t *testing.T) {
	t.Parallel()

	service := NewService("test-api-key")
	model, dimensions := service.GetModelInfo()

	if model != "text-embedding-3-small" {
		t.Errorf("Expected model to be 'text-embedding-3-small', got %s", model)
	}

	if dimensions != 1536 {
		t.Errorf("Expected dimensions to be 1536, got %d", dimensions)
	}
}

func TestGenerateEmbedding_Success(t *testing.T) {
	t.Parallel()

	mockResponse := EmbeddingResponse{
		Object: "list",
		Data: []struct {
			Object    string    `json:"object"`
			Index     int       `json:"index"`
			Embedding []float64 `json:"embedding"`
		}{
			{
				Object:    "embedding",
				Index:     0,
				Embedding: []float64{0.1, 0.2, 0.3, 0.4, 0.5},
			},
		},
		Model: "text-embedding-3-small",
		Usage: struct {
			PromptTokens int `json:"prompt_tokens"`
			TotalTokens  int `json:"total_tokens"`
		}{
			PromptTokens: 5,
			TotalTokens:  5,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/embeddings" {
			t.Errorf("Expected /embeddings path, got %s", r.URL.Path)
		}

		authHeader := r.Header.Get("Authorization")
		expectedAuth := "Bearer test-api-key"
		if authHeader != expectedAuth {
			t.Errorf("Expected Authorization header %s, got %s", expectedAuth, authHeader)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		var req EmbeddingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if len(req.Input) != 1 {
			t.Errorf("Expected 1 input text, got %d", len(req.Input))
		}
		if req.Input[0] != "test text" {
			t.Errorf("Expected input text 'test text', got %s", req.Input[0])
		}
		if req.Model != "text-embedding-3-small" {
			t.Errorf("Expected model 'text-embedding-3-small', got %s", req.Model)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(mockResponse); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	service := NewService("test-api-key")
	embeddingService := service.(*EmbeddingService)
	embeddingService.baseURL = server.URL

	ctx := context.Background()
	embedding, err := service.GenerateEmbedding(ctx, "test text")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(embedding) != 5 {
		t.Errorf("Expected embedding length 5, got %d", len(embedding))
	}

	expectedEmbedding := []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	for i, val := range embedding {
		if val != expectedEmbedding[i] {
			t.Errorf("Expected embedding[%d] to be %f, got %f", i, expectedEmbedding[i], val)
		}
	}
}

func TestGenerateEmbedding_EmptyText(t *testing.T) {
	t.Parallel()

	service := NewService("test-api-key")
	ctx := context.Background()

	_, err := service.GenerateEmbedding(ctx, "")
	if err == nil {
		t.Error("Expected error for empty text, got nil")
	}

	if !strings.Contains(err.Error(), "no valid texts provided") {
		t.Errorf("Expected error message about no valid texts, got %v", err)
	}
}

func TestGenerateEmbedding_WhitespaceOnly(t *testing.T) {
	t.Parallel()

	service := NewService("test-api-key")
	ctx := context.Background()

	_, err := service.GenerateEmbedding(ctx, "   \n\t   ")
	if err == nil {
		t.Error("Expected error for whitespace-only text, got nil")
	}

	if !strings.Contains(err.Error(), "no valid texts provided") {
		t.Errorf("Expected error message about no valid texts, got %v", err)
	}
}

func TestGenerateEmbeddings_Success(t *testing.T) {
	t.Parallel()

	mockResponse := EmbeddingResponse{
		Object: "list",
		Data: []struct {
			Object    string    `json:"object"`
			Index     int       `json:"index"`
			Embedding []float64 `json:"embedding"`
		}{
			{
				Object:    "embedding",
				Index:     0,
				Embedding: []float64{0.1, 0.2, 0.3},
			},
			{
				Object:    "embedding",
				Index:     1,
				Embedding: []float64{0.4, 0.5, 0.6},
			},
		},
		Model: "text-embedding-3-small",
		Usage: struct {
			PromptTokens int `json:"prompt_tokens"`
			TotalTokens  int `json:"total_tokens"`
		}{
			PromptTokens: 10,
			TotalTokens:  10,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req EmbeddingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if len(req.Input) != 2 {
			t.Errorf("Expected 2 input texts, got %d", len(req.Input))
		}
		if req.Input[0] != "text one" {
			t.Errorf("Expected first input 'text one', got %s", req.Input[0])
		}
		if req.Input[1] != "text two" {
			t.Errorf("Expected second input 'text two', got %s", req.Input[1])
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(mockResponse); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	service := NewService("test-api-key")
	embeddingService := service.(*EmbeddingService)
	embeddingService.baseURL = server.URL

	ctx := context.Background()
	texts := []string{"text one", "text two"}
	embeddings, err := service.GenerateEmbeddings(ctx, texts)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(embeddings) != 2 {
		t.Errorf("Expected 2 embeddings, got %d", len(embeddings))
	}

	if len(embeddings[0]) != 3 {
		t.Errorf("Expected first embedding length 3, got %d", len(embeddings[0]))
	}
	expectedFirst := []float64{0.1, 0.2, 0.3}
	for i, val := range embeddings[0] {
		if val != expectedFirst[i] {
			t.Errorf("Expected first embedding[%d] to be %f, got %f", i, expectedFirst[i], val)
		}
	}

	if len(embeddings[1]) != 3 {
		t.Errorf("Expected second embedding length 3, got %d", len(embeddings[1]))
	}
	expectedSecond := []float64{0.4, 0.5, 0.6}
	for i, val := range embeddings[1] {
		if val != expectedSecond[i] {
			t.Errorf("Expected second embedding[%d] to be %f, got %f", i, expectedSecond[i], val)
		}
	}
}

func TestGenerateEmbeddings_EmptyInput(t *testing.T) {
	t.Parallel()

	service := NewService("test-api-key")
	ctx := context.Background()

	_, err := service.GenerateEmbeddings(ctx, []string{})
	if err == nil {
		t.Error("Expected error for empty input, got nil")
	}

	if !strings.Contains(err.Error(), "no texts provided") {
		t.Errorf("Expected error message about no texts provided, got %v", err)
	}
}

func TestGenerateEmbeddings_FiltersEmptyTexts(t *testing.T) {
	t.Parallel()

	mockResponse := EmbeddingResponse{
		Object: "list",
		Data: []struct {
			Object    string    `json:"object"`
			Index     int       `json:"index"`
			Embedding []float64 `json:"embedding"`
		}{
			{
				Object:    "embedding",
				Index:     0,
				Embedding: []float64{0.1, 0.2, 0.3},
			},
		},
		Model: "text-embedding-3-small",
		Usage: struct {
			PromptTokens int `json:"prompt_tokens"`
			TotalTokens  int `json:"total_tokens"`
		}{
			PromptTokens: 5,
			TotalTokens:  5,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req EmbeddingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if len(req.Input) != 1 {
			t.Errorf("Expected 1 input text after filtering, got %d", len(req.Input))
		}
		if req.Input[0] != "valid text" {
			t.Errorf("Expected input 'valid text', got %s", req.Input[0])
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(mockResponse); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	service := NewService("test-api-key")
	embeddingService := service.(*EmbeddingService)
	embeddingService.baseURL = server.URL

	ctx := context.Background()
	texts := []string{"", "valid text", "   ", "\n\t"}
	embeddings, err := service.GenerateEmbeddings(ctx, texts)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(embeddings) != 1 {
		t.Errorf("Expected 1 embedding after filtering, got %d", len(embeddings))
	}
}

func TestGenerateEmbeddings_APIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("Internal Server Error")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	service := NewService("test-api-key")
	embeddingService := service.(*EmbeddingService)
	embeddingService.baseURL = server.URL

	ctx := context.Background()
	_, err := service.GenerateEmbeddings(ctx, []string{"test text"})

	if err == nil {
		t.Error("Expected error for API failure, got nil")
	}

	if !strings.Contains(err.Error(), "API request failed with status 500") {
		t.Errorf("Expected error message about API failure, got %v", err)
	}
}

func TestGenerateEmbeddings_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte("invalid json")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	service := NewService("test-api-key")
	embeddingService := service.(*EmbeddingService)
	embeddingService.baseURL = server.URL

	ctx := context.Background()
	_, err := service.GenerateEmbeddings(ctx, []string{"test text"})

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}

	if !strings.Contains(err.Error(), "failed to decode response") {
		t.Errorf("Expected error message about decoding response, got %v", err)
	}
}

func TestGenerateEmbeddings_MismatchedResponse(t *testing.T) {
	t.Parallel()

	mockResponse := EmbeddingResponse{
		Object: "list",
		Data: []struct {
			Object    string    `json:"object"`
			Index     int       `json:"index"`
			Embedding []float64 `json:"embedding"`
		}{
			{
				Object:    "embedding",
				Index:     0,
				Embedding: []float64{0.1, 0.2, 0.3},
			},
		},
		Model: "text-embedding-3-small",
		Usage: struct {
			PromptTokens int `json:"prompt_tokens"`
			TotalTokens  int `json:"total_tokens"`
		}{
			PromptTokens: 5,
			TotalTokens:  5,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(mockResponse); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	service := NewService("test-api-key")
	embeddingService := service.(*EmbeddingService)
	embeddingService.baseURL = server.URL

	ctx := context.Background()
	_, err := service.GenerateEmbeddings(ctx, []string{"text one", "text two"})

	if err == nil {
		t.Error("Expected error for mismatched response, got nil")
	}

	if !strings.Contains(err.Error(), "mismatch between input texts and returned embeddings") {
		t.Errorf("Expected error message about mismatch, got %v", err)
	}
}
