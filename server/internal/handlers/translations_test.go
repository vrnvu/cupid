package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vrnvu/cupid/internal/client"
)

func TestServer_GetHotelTranslationsHandler(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		hotelID        string
		language       string
		translations   []client.Translation
		repoError      error
		expectedStatus int
		expectedCount  int
	}{
		{
			name:     "successful request with French translations",
			hotelID:  "123",
			language: "fr",
			translations: []client.Translation{
				{FieldName: "hotel_name", TranslatedText: "L'Hôtel"},
				{FieldName: "description", TranslatedText: "Un bel hôtel"},
			},
			repoError:      nil,
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "successful request with no translations",
			hotelID:        "123",
			language:       "es",
			translations:   []client.Translation{},
			repoError:      nil,
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "invalid hotel ID",
			hotelID:        "invalid",
			language:       "fr",
			translations:   nil,
			repoError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
		{
			name:           "unsupported language",
			hotelID:        "123",
			language:       "de",
			translations:   nil,
			repoError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
		{
			name:           "repository error",
			hotelID:        "123",
			language:       "fr",
			translations:   nil,
			repoError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockRepo := new(MockRepository)
			if tt.hotelID != "invalid" && tt.language != "de" {
				hotelID := 123
				mockRepo.On("GetHotelTranslations", mock.Anything, hotelID, tt.language).Return(tt.translations, tt.repoError)
			}

			server := &Server{repository: mockRepo}
			req := httptest.NewRequest("GET", "/api/v1/hotels/"+tt.hotelID+"/translations/"+tt.language, nil)
			req.SetPathValue("hotelID", tt.hotelID)
			req.SetPathValue("language", tt.language)
			w := httptest.NewRecorder()

			server.getHotelTranslationsHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, int(response["count"].(float64)))
				assert.Equal(t, 123, int(response["hotel_id"].(float64)))
				assert.Equal(t, tt.language, response["language"])
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
