//go:build integration

package ai

import (
	"context"
	"os"
	"testing"
)

func TestAIService_Integration(t *testing.T) {
	t.Parallel()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	service := NewService(apiKey)
	ctx := context.Background()

	t.Run("GenerateEmbedding_RealAPI", func(t *testing.T) {
		t.Parallel()

		text := "This is a test review about a great hotel with excellent service and amazing breakfast."
		embedding, err := service.GenerateEmbedding(ctx, text)

		if err != nil {
			t.Fatalf("Failed to generate embedding: %v", err)
		}

		// Verify embedding properties
		if len(embedding) != 1536 {
			t.Errorf("Expected embedding length 1536, got %d", len(embedding))
		}

		// Check that embedding values are reasonable (not all zeros)
		hasNonZero := false
		for _, val := range embedding {
			if val != 0.0 {
				hasNonZero = true
				break
			}
		}
		if !hasNonZero {
			t.Error("Expected embedding to have non-zero values")
		}

		// Check that values are in reasonable range for embeddings
		for i, val := range embedding {
			if val < -10 || val > 10 {
				t.Errorf("Embedding[%d] value %f seems out of range", i, val)
			}
		}
	})

	t.Run("GenerateEmbeddings_RealAPI", func(t *testing.T) {
		t.Parallel()

		texts := []string{
			"Amazing hotel with great service and clean rooms.",
			"Terrible experience, dirty rooms and rude staff.",
			"Average hotel, nothing special but not bad either.",
		}

		embeddings, err := service.GenerateEmbeddings(ctx, texts)

		if err != nil {
			t.Fatalf("Failed to generate embeddings: %v", err)
		}

		if len(embeddings) != len(texts) {
			t.Errorf("Expected %d embeddings, got %d", len(texts), len(embeddings))
		}

		// Verify each embedding
		for i, embedding := range embeddings {
			if len(embedding) != 1536 {
				t.Errorf("Embedding %d: expected length 1536, got %d", i, len(embedding))
			}

			// Check for reasonable values
			hasNonZero := false
			for _, val := range embedding {
				if val != 0.0 {
					hasNonZero = true
					break
				}
			}
			if !hasNonZero {
				t.Errorf("Embedding %d: expected non-zero values", i)
			}
		}

		// Verify that different texts produce different embeddings
		// (they should be different, but we'll just check they're not identical)
		if len(embeddings) >= 2 {
			identical := true
			for i := 0; i < 1536; i++ {
				if embeddings[0][i] != embeddings[1][i] {
					identical = false
					break
				}
			}
			if identical {
				t.Error("Expected different texts to produce different embeddings")
			}
		}
	})

	t.Run("GenerateEmbedding_EmptyText", func(t *testing.T) {
		t.Parallel()

		_, err := service.GenerateEmbedding(ctx, "")
		if err == nil {
			t.Error("Expected error for empty text")
		}
	})

	t.Run("GenerateEmbedding_WhitespaceOnly", func(t *testing.T) {
		t.Parallel()

		_, err := service.GenerateEmbedding(ctx, "   \n\t   ")
		if err == nil {
			t.Error("Expected error for whitespace-only text")
		}
	})

	t.Run("GenerateEmbeddings_FiltersEmptyTexts", func(t *testing.T) {
		t.Parallel()

		texts := []string{"", "Valid text here", "   ", "\n\t", "Another valid text"}
		embeddings, err := service.GenerateEmbeddings(ctx, texts)

		if err != nil {
			t.Fatalf("Failed to generate embeddings: %v", err)
		}

		// Should only have embeddings for the 2 valid texts
		if len(embeddings) != 2 {
			t.Errorf("Expected 2 embeddings after filtering, got %d", len(embeddings))
		}
	})

	t.Run("ModelInfo", func(t *testing.T) {
		t.Parallel()

		model, dimensions := service.GetModelInfo()
		if model != "text-embedding-3-small" {
			t.Errorf("Expected model 'text-embedding-3-small', got %s", model)
		}
		if dimensions != 1536 {
			t.Errorf("Expected dimensions 1536, got %d", dimensions)
		}
	})
}
