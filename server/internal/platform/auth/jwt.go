package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(userID string, secret []byte, expiration time.Duration, tokenType string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"token_type": tokenType,
		"iat":        time.Now().Unix(),
		"exp":        time.Now().Add(expiration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ParseJWT(tokenStr string, secret []byte) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("user_id not found in token")
	}
	return userID, nil
}

// ParseJWTWithType parses and validates a JWT, additionally checking that the
// token_type claim matches expectedType. This prevents access tokens from
// being used as refresh tokens and vice versa.
func ParseJWTWithType(tokenStr string, secret []byte, expectedType string) (string, error) {
	userID, err := ParseJWT(tokenStr, secret)
	if err != nil {
		return "", err
	}

	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return "", fmt.Errorf("parse unverified: %w", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}
	tokenType, _ := claims["token_type"].(string)
	if tokenType != expectedType {
		return "", fmt.Errorf("unexpected token type: %q, expected %q", tokenType, expectedType)
	}

	return userID, nil
}

// RefreshTokenVerifier implements identity app layer's RefreshTokenVerifier interface.
type RefreshTokenVerifier struct {
	Secret []byte
}

func (v RefreshTokenVerifier) Verify(tokenStr, expectedType string) (string, error) {
	return ParseJWTWithType(tokenStr, v.Secret, expectedType)
}
