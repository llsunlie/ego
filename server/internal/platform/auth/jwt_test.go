package auth

import (
	"testing"
	"time"
)

func TestGenerateAndParse(t *testing.T) {
	secret := []byte("test-secret")
	exp := time.Hour

	token, err := GenerateJWT("user-123", secret, exp)
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}
	if token == "" {
		t.Fatal("empty token")
	}

	userID, err := ParseJWT(token, secret)
	if err != nil {
		t.Fatalf("ParseJWT: %v", err)
	}
	if userID != "user-123" {
		t.Fatalf("expected user-123, got %s", userID)
	}
}

func TestParseJWT_WrongSecret(t *testing.T) {
	secret := []byte("real-secret")
	token, err := GenerateJWT("user-1", secret, time.Hour)
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	_, err = ParseJWT(token, []byte("wrong-secret"))
	if err == nil {
		t.Fatal("expected error with wrong secret")
	}
}

func TestParseJWT_Expired(t *testing.T) {
	secret := []byte("test-secret")
	token, err := GenerateJWT("user-1", secret, -time.Hour)
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	_, err = ParseJWT(token, secret)
	if err == nil {
		t.Fatal("expected error with expired token")
	}
}

func TestParseJWT_InvalidToken(t *testing.T) {
	_, err := ParseJWT("not-a-jwt", []byte("secret"))
	if err == nil {
		t.Fatal("expected error with invalid token")
	}
}
