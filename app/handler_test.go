package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockStore struct {
	items []Item
}

func (m *mockStore) Ping(_ context.Context) error { return nil }
func (m *mockStore) Close() error                 { return nil }
func (m *mockStore) DBStats() sql.DBStats          { return sql.DBStats{} }

func (m *mockStore) ListItems(_ context.Context) ([]Item, error) {
	return m.items, nil
}

func (m *mockStore) GetItem(_ context.Context, id int) (*Item, error) {
	for _, item := range m.items {
		if item.ID == id {
			return &item, nil
		}
	}
	return nil, nil
}

func (m *mockStore) CreateItem(_ context.Context, name, description string) (*Item, error) {
	item := Item{ID: len(m.items) + 1, Name: name, Description: description, CreatedAt: time.Now()}
	m.items = append(m.items, item)
	return &item, nil
}

func (m *mockStore) DeleteItem(_ context.Context, id int) error {
	for i, item := range m.items {
		if item.ID == id {
			m.items = append(m.items[:i], m.items[i+1:]...)
			return nil
		}
	}
	return sql.ErrNoRows
}

func TestHandleHealth(t *testing.T) {
	s := &server{store: &mockStore{}}
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	s.handleHealth(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected content-type application/json, got %s", ct)
	}
}

func TestHandleListItems(t *testing.T) {
	s := &server{store: &mockStore{items: []Item{{ID: 1, Name: "test", Description: "desc"}}}}
	req := httptest.NewRequest(http.MethodGet, "/api/items", nil)
	rec := httptest.NewRecorder()
	s.handleListItems(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var items []Item
	_ = json.NewDecoder(rec.Body).Decode(&items)
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestHandleCreateItem(t *testing.T) {
	s := &server{store: &mockStore{}}
	body, _ := json.Marshal(map[string]string{"name": "new", "description": "new item"})
	req := httptest.NewRequest(http.MethodPost, "/api/items", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.handleCreateItem(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandleCreateItemMissingName(t *testing.T) {
	s := &server{store: &mockStore{}}
	body, _ := json.Marshal(map[string]string{"description": "no name"})
	req := httptest.NewRequest(http.MethodPost, "/api/items", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.handleCreateItem(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleDeleteItemNotFound(t *testing.T) {
	s := &server{store: &mockStore{}}
	req := httptest.NewRequest(http.MethodDelete, "/api/items/999", nil)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()
	s.handleDeleteItem(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}
