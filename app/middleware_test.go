package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricsMiddleware(t *testing.T) {
	handler := metricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/health", "/health"},
		{"/api/items", "/api/items"},
		{"/api/items/123", "/api/items/{id}"},
		{"/api/cache/mykey", "/api/cache/{key}"},
		{"/other", "/other"},
	}
	for _, tt := range tests {
		got := normalizePath(tt.input)
		if got != tt.want {
			t.Errorf("normalizePath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
