package utils

import (
	"crypto/rand"
	"encoding/base64"
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

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // seconds until access token expires
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

// GenerateTokenPair generates both access and refresh tokens
func GenerateTokenPair(userID uint, email string, secret string, accessExpiry, refreshExpiry time.Duration) (*TokenPair, string, error) {
	// Generate access token
	accessToken, err := GenerateToken(userID, email, secret, accessExpiry)
	if err != nil {
		return nil, "", err
	}

	// Generate refresh token (cryptographically secure random string)
	refreshToken, err := generateSecureToken()
	if err != nil {
		return nil, "", err
	}

	// Generate token family for rotation tracking
	tokenFamily, err := generateSecureToken()
	if err != nil {
		return nil, "", err
	}

	pair := &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessExpiry.Seconds()),
	}

	return pair, tokenFamily, nil
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
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

// ExtractTokenExpiry extracts the expiration time from a JWT token
func ExtractTokenExpiry(tokenString string, secret string) (time.Time, error) {
	claims, err := ValidateToken(tokenString, secret)
	if err != nil {
		return time.Time{}, err
	}

	if claims.ExpiresAt == nil {
		return time.Time{}, domain.ErrInvalidToken
	}

	return claims.ExpiresAt.Time, nil
}
