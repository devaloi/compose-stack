package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Item represents a stored item.
type Item struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Store provides data access operations.
type Store interface {
	Ping(ctx context.Context) error
	ListItems(ctx context.Context) ([]Item, error)
	GetItem(ctx context.Context, id int) (*Item, error)
	CreateItem(ctx context.Context, name, description string) (*Item, error)
	DeleteItem(ctx context.Context, id int) error
	DBStats() sql.DBStats
	Close() error
}

type pgStore struct {
	db *sql.DB
}

// NewPostgresStore connects to PostgreSQL and returns a Store.
func NewPostgresStore(dsn string) (Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return &pgStore{db: db}, nil
}

func (s *pgStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *pgStore) ListItems(ctx context.Context) ([]Item, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, description, created_at FROM items ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *pgStore) GetItem(ctx context.Context, id int) (*Item, error) {
	var item Item
	err := s.db.QueryRowContext(ctx, "SELECT id, name, description, created_at FROM items WHERE id = $1", id).
		Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get item: %w", err)
	}
	return &item, nil
}

func (s *pgStore) CreateItem(ctx context.Context, name, description string) (*Item, error) {
	var item Item
	err := s.db.QueryRowContext(ctx,
		"INSERT INTO items (name, description) VALUES ($1, $2) RETURNING id, name, description, created_at",
		name, description).
		Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create item: %w", err)
	}
	return &item, nil
}

func (s *pgStore) DeleteItem(ctx context.Context, id int) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM items WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *pgStore) DBStats() sql.DBStats {
	return s.db.Stats()
}

func (s *pgStore) Close() error {
	return s.db.Close()
}
