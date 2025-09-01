package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	service := NewService("test-api-key")
	assert.NotNil(t, service, "Service should be created successfully")
}

func TestGetModelInfo(t *testing.T) {
	t.Parallel()

	service := NewService("test-api-key")
	model, dimensions := service.GetModelInfo()

	assert.Equal(t, "text-embedding-3-small", model)
	assert.Equal(t, 1536, dimensions)
}

func TestGenerateEmbedding_Validation(t *testing.T) {
	t.Parallel()

	service := NewService("test-api-key")
	ctx := context.Background()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty text",
			input:   "",
			wantErr: true,
			errMsg:  "no valid texts provided",
		},
		{
			name:    "whitespace only",
			input:   "   \n\t   ",
			wantErr: true,
			errMsg:  "no valid texts provided",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := service.GenerateEmbedding(ctx, tt.input)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestGenerateEmbeddings_Validation(t *testing.T) {
	t.Parallel()

	service := NewService("test-api-key")
	ctx := context.Background()

	tests := []struct {
		name    string
		inputs  []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty input",
			inputs:  []string{},
			wantErr: true,
			errMsg:  "no texts provided",
		},
		{
			name:    "all empty strings",
			inputs:  []string{"", "   ", "\n\t"},
			wantErr: true,
			errMsg:  "no valid texts provided",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := service.GenerateEmbeddings(ctx, tt.inputs)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}
