package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockCache struct {
	data map[string]string
}

func newMockCache() *mockCache {
	return &mockCache{data: make(map[string]string)}
}

func (m *mockCache) Get(_ context.Context, key string) (string, error) {
	return m.data[key], nil
}

func (m *mockCache) Set(_ context.Context, key, value string, _ time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *mockCache) Ping(_ context.Context) error { return nil }
func (m *mockCache) Close() error                 { return nil }

func TestHandleSetCache(t *testing.T) {
	mc := newMockCache()
	s := &server{cache: mc}

	body := []byte(`{"value":"hello","ttl":60}`)
	req := httptest.NewRequest(http.MethodPut, "/api/cache/mykey", bytes.NewReader(body))
	req.SetPathValue("key", "mykey")
	rec := httptest.NewRecorder()
	s.handleSetCache(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	if mc.data["mykey"] != "hello" {
		t.Errorf("expected cache to contain 'hello', got '%s'", mc.data["mykey"])
	}
}

func TestHandleGetCache(t *testing.T) {
	mc := newMockCache()
	mc.data["mykey"] = "hello"
	s := &server{cache: mc}

	req := httptest.NewRequest(http.MethodGet, "/api/cache/mykey", nil)
	req.SetPathValue("key", "mykey")
	rec := httptest.NewRecorder()
	s.handleGetCache(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandleGetCacheNotFound(t *testing.T) {
	mc := newMockCache()
	s := &server{cache: mc}

	req := httptest.NewRequest(http.MethodGet, "/api/cache/missing", nil)
	req.SetPathValue("key", "missing")
	rec := httptest.NewRecorder()
	s.handleGetCache(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandleSetCacheMissingValue(t *testing.T) {
	mc := newMockCache()
	s := &server{cache: mc}

	body := []byte(`{"ttl":60}`)
	req := httptest.NewRequest(http.MethodPut, "/api/cache/mykey", bytes.NewReader(body))
	req.SetPathValue("key", "mykey")
	rec := httptest.NewRecorder()
	s.handleSetCache(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
