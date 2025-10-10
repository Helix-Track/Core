package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"helixtrack.ru/core/internal/models"
)

// JWTService handles JWT token generation and validation
type JWTService struct {
	secretKey string
	issuer    string
	expiry    time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey, issuer string, expiryHours int) *JWTService {
	if secretKey == "" {
		secretKey = "helix-track-default-secret-key-change-in-production"
	}
	if issuer == "" {
		issuer = "helixtrack-core"
	}
	if expiryHours == 0 {
		expiryHours = 24 // Default 24 hours
	}

	return &JWTService{
		secretKey: secretKey,
		issuer:    issuer,
		expiry:    time.Duration(expiryHours) * time.Hour,
	}
}

// GenerateToken generates a JWT token for a user
func (s *JWTService) GenerateToken(username, email, name, role string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(s.expiry)

	claims := &models.JWTClaims{
		Username: username,
		Email:    email,
		Name:     name,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *JWTService) ValidateToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}
