package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Service defines the interface for AI operations
type Service interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float64, error)
	GenerateEmbeddings(ctx context.Context, texts []string) ([][]float64, error)
	GetModelInfo() (string, int)
}

// EmbeddingService handles AI operations like generating embeddings
type EmbeddingService struct {
	apiKey  string
	client  *http.Client
	baseURL string
	model   string
}

// EmbeddingRequest represents the request to OpenAI embedding API
type EmbeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

// EmbeddingResponse represents the response from OpenAI embedding API
type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// NewService creates a new AI service instance
func NewService(apiKey string) Service {
	return &EmbeddingService{
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: "https://api.openai.com/v1",
		model:   "text-embedding-3-small", // 1536 dimensions
	}
}

// GenerateEmbedding generates an embedding for a single text
func (s *EmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	embeddings, err := s.GenerateEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return embeddings[0], nil
}

// GenerateEmbeddings generates embeddings for multiple texts
func (s *EmbeddingService) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	// Filter out empty texts
	var validTexts []string
	for _, text := range texts {
		if trimmed := strings.TrimSpace(text); len(trimmed) > 0 {
			validTexts = append(validTexts, trimmed)
		}
	}

	if len(validTexts) == 0 {
		return nil, fmt.Errorf("no valid texts provided")
	}

	reqBody := EmbeddingRequest{
		Input: validTexts,
		Model: s.model,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var embeddingResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(embeddingResp.Data) != len(validTexts) {
		return nil, fmt.Errorf("mismatch between input texts and returned embeddings: %d texts, %d embeddings",
			len(validTexts), len(embeddingResp.Data))
	}

	// Extract embeddings in the same order as input texts
	embeddings := make([][]float64, len(validTexts))
	for _, data := range embeddingResp.Data {
		if data.Index >= len(embeddings) {
			return nil, fmt.Errorf("embedding index %d out of range", data.Index)
		}
		embeddings[data.Index] = data.Embedding
	}

	return embeddings, nil
}

// GetModelInfo returns information about the current model
func (s *EmbeddingService) GetModelInfo() (string, int) {
	return s.model, 1536 // text-embedding-3-small has 1536 dimensions
}
