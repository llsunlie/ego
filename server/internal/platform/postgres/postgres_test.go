package postgres

import (
	"testing"
)

func TestConnect(t *testing.T) {
	url := "postgres://ego:ego@localhost:5432/ego?sslmode=disable"
	pool, err := Connect(url)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	pool.Close()
}

func TestConnect_InvalidURL(t *testing.T) {
	_, err := Connect("postgres://bad:bad@localhost:5432/nonexistent?sslmode=disable")
	if err == nil {
		t.Fatal("expected error with invalid connection URL")
	}
}
