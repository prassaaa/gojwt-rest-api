package utils

import (
	"gojwt-rest-api/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var signingMethod = jwt.SigningMethodHS256

// JWTClaims represents JWT claims
type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken generates a new JWT token
func GenerateToken(userID uint, email string, secret string, expiration time.Duration) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrInvalidSigningMethod
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}
