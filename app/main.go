package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	var s server

	if dsn := os.Getenv("DB_URL"); dsn != "" {
		store, err := NewPostgresStore(dsn)
		if err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}
		defer store.Close()
		s.store = store
		log.Println("connected to PostgreSQL")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /api/items", s.handleListItems)
	mux.HandleFunc("POST /api/items", s.handleCreateItem)
	mux.HandleFunc("GET /api/items/{id}", s.handleGetItem)
	mux.HandleFunc("DELETE /api/items/{id}", s.handleDeleteItem)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server stopped")
}
