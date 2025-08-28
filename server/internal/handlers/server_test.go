package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerHandler(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		path       string
		route      string
		userID     string
		wantStatus int
	}{
		{
			name:       "ok health",
			path:       "health",
			route:      "/health",
			userID:     "123",
			wantStatus: http.StatusOK,
		},
		{
			name:       "ok foo",
			path:       "foo",
			route:      "/foo",
			userID:     "123",
			wantStatus: http.StatusOK,
		},
		{
			name:       "ok bar",
			path:       "bar",
			route:      "/bar",
			userID:     "123",
			wantStatus: http.StatusOK,
		},
		{
			name:       "ok baz",
			path:       "baz",
			route:      "/baz",
			userID:     "123",
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing user header",
			path:       "foo",
			route:      "/foo",
			userID:     "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid path returns 404",
			path:       "does-not-exist",
			route:      "/does-not-exist",
			userID:     "123",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			handler := NewServerHandler(tc.path)

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, tc.route, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			if tc.userID != "" {
				req.Header.Set("User-Id", tc.userID)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("unexpected status code: got %d, want %d", rec.Code, tc.wantStatus)
			}
		})
	}
}
