package handlers

import (
	"fmt"
	"net/http"
)

func NewServerHandler(path string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverHandler(w, r, path)
	})
}

func serverHandler(w http.ResponseWriter, r *http.Request, path string) {
	userID := r.Header.Get("User-Id")
	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("User-Id header is required"))
		return
	}

	if path != "health" && path != "foo" && path != "bar" && path != "baz" {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(fmt.Sprintf("Not Found: %s", path)))
		return
	}

	w.WriteHeader(http.StatusOK)
}
