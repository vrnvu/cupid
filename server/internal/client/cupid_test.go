package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockCupidAPI struct {
	statusCode int
	body       string
	headers    map[string]string
}

func (m *mockCupidAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for k, v := range m.headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(m.statusCode)
	if m.body != "" {
		w.Write([]byte(m.body))
	}
}

func TestClient_Do_StatusMapping(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		statusCode int
		body       string
		expectErr  error
	}{
		{"ok_200", 200, `{"ok":true}`, nil},
		{"client_400", 400, `{"error":"client"}`, &Error{}},
		{"server_500", 500, `{"error":"server"}`, &Error{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockCupidAPI{
				statusCode: tc.statusCode,
				body:       tc.body,
				headers:    map[string]string{"X-Request-Id": "req-123"},
			}
			ts := httptest.NewServer(mock)
			defer ts.Close()

			c, err := New(ts.URL,
				WithUserAgent("cupid-agent"),
				WithTimeout(1*time.Second),
			)
			assert.NoError(t, err)

			respBody, resp, err := c.Do(context.Background(), http.MethodGet, "/path", nil, nil)
			if tc.expectErr == nil {
				defer resp.Body.Close()
				assert.NoError(t, err)
				if assert.NotNil(t, resp) {
					assert.Equal(t, tc.statusCode, resp.StatusCode)
				}
				assert.Equal(t, strings.TrimSpace(tc.body), strings.TrimSpace(string(respBody)))
			} else {
				assert.Error(t, err)
				if tc.statusCode >= 400 && tc.statusCode <= 599 {
					var ce *Error
					assert.True(t, errors.As(err, &ce), "expected client.Error, got %T: %v", err, err)
					if ce != nil {
						assert.Equal(t, tc.statusCode, ce.StatusCode)
						assert.Equal(t, "req-123", ce.RequestID)
					}
				} else {
					assert.Failf(t, "invalid test status range", "status %d not in 4xx/5xx", tc.statusCode)
				}
			}
		})
	}
}
