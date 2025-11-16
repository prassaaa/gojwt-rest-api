package unit

import (
	"gojwt-rest-api/internal/domain"
	"gojwt-rest-api/internal/utils"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name       string
		userID     uint
		email      string
		secret     string
		expiration time.Duration
		wantErr    bool
	}{
		{
			name:       "Valid token generation",
			userID:     1,
			email:      "test@example.com",
			secret:     "test-secret-key",
			expiration: 24 * time.Hour,
			wantErr:    false,
		},
		{
			name:       "Token with short expiration",
			userID:     2,
			email:      "user@test.com",
			secret:     "another-secret",
			expiration: 1 * time.Minute,
			wantErr:    false,
		},
		{
			name:       "Token with empty email",
			userID:     3,
			email:      "",
			secret:     "secret",
			expiration: 1 * time.Hour,
			wantErr:    false,
		},
		{
			name:       "Token with zero userID",
			userID:     0,
			email:      "zero@test.com",
			secret:     "secret",
			expiration: 1 * time.Hour,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := utils.GenerateToken(tt.userID, tt.email, tt.secret, tt.expiration)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, token)

			// Verify token can be parsed
			parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
				return []byte(tt.secret), nil
			})
			require.NoError(t, err)
			assert.True(t, parsedToken.Valid)
		})
	}
}

func TestValidateToken(t *testing.T) {
	secret := "test-secret-key"
	userID := uint(1)
	email := "test@example.com"

	tests := []struct {
		name        string
		setupToken  func() string
		secret      string
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Valid token",
			setupToken: func() string {
				token, _ := utils.GenerateToken(userID, email, secret, 24*time.Hour)
				return token
			},
			secret:  secret,
			wantErr: false,
		},
		{
			name: "Invalid secret",
			setupToken: func() string {
				token, _ := utils.GenerateToken(userID, email, secret, 24*time.Hour)
				return token
			},
			secret:  "wrong-secret",
			wantErr: true,
		},
		{
			name: "Expired token",
			setupToken: func() string {
				token, _ := utils.GenerateToken(userID, email, secret, -1*time.Hour)
				return token
			},
			secret:  secret,
			wantErr: true,
		},
		{
			name: "Malformed token",
			setupToken: func() string {
				return "invalid.token.here"
			},
			secret:  secret,
			wantErr: true,
		},
		{
			name: "Empty token",
			setupToken: func() string {
				return ""
			},
			secret:  secret,
			wantErr: true,
		},
		{
			name: "Token with only header",
			setupToken: func() string {
				return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
			},
			secret:  secret,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupToken()
			claims, err := utils.ValidateToken(token, tt.secret)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				assert.Nil(t, claims)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, claims)
			assert.Equal(t, userID, claims.UserID)
			assert.Equal(t, email, claims.Email)
		})
	}
}

func TestJWTClaims(t *testing.T) {
	secret := "test-secret"
	userID := uint(123)
	email := "claims@test.com"
	expiration := 2 * time.Hour

	token, err := utils.GenerateToken(userID, email, secret, expiration)
	require.NoError(t, err)

	claims, err := utils.ValidateToken(token, secret)
	require.NoError(t, err)

	t.Run("Claims contain correct user data", func(t *testing.T) {
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
	})

	t.Run("Claims contain issued at time", func(t *testing.T) {
		assert.NotNil(t, claims.IssuedAt)
		assert.True(t, claims.IssuedAt.Before(time.Now()))
	})

	t.Run("Claims contain expiration time", func(t *testing.T) {
		assert.NotNil(t, claims.ExpiresAt)
		assert.True(t, claims.ExpiresAt.After(time.Now()))
	})

	t.Run("Expiration is approximately correct", func(t *testing.T) {
		expectedExpiration := time.Now().Add(expiration)
		actualExpiration := claims.ExpiresAt.Time

		// Allow 5 seconds tolerance
		diff := actualExpiration.Sub(expectedExpiration)
		assert.Less(t, diff.Abs(), 5*time.Second)
	})
}

func TestTokenExpiration(t *testing.T) {
	secret := "test-secret"
	userID := uint(1)
	email := "test@example.com"

	t.Run("Token expires after specified duration", func(t *testing.T) {
		// Create token that expires in 1 second
		token, err := utils.GenerateToken(userID, email, secret, 1*time.Second)
		require.NoError(t, err)

		// Validate immediately - should work
		claims, err := utils.ValidateToken(token, secret)
		require.NoError(t, err)
		assert.NotNil(t, claims)

		// Wait for token to expire
		time.Sleep(2 * time.Second)

		// Validate again - should fail
		claims, err = utils.ValidateToken(token, secret)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestTokenWithDifferentSecrets(t *testing.T) {
	userID := uint(1)
	email := "test@example.com"
	secret1 := "secret-one"
	secret2 := "secret-two"

	token, err := utils.GenerateToken(userID, email, secret1, 24*time.Hour)
	require.NoError(t, err)

	t.Run("Token validates with correct secret", func(t *testing.T) {
		claims, err := utils.ValidateToken(token, secret1)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
	})

	t.Run("Token fails with different secret", func(t *testing.T) {
		claims, err := utils.ValidateToken(token, secret2)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestInvalidSigningMethod(t *testing.T) {
	userID := uint(1)
	email := "test@example.com"
	secret := "test-secret"

	// Create token with different signing method (this will fail to sign properly)
	claims := utils.JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Try to create token with none algorithm (will be rejected)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	// Validation should fail with invalid signing method
	validatedClaims, err := utils.ValidateToken(tokenString, secret)
	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
	assert.ErrorIs(t, err, domain.ErrInvalidSigningMethod)
}

func BenchmarkGenerateToken(b *testing.B) {
	secret := "benchmark-secret"
	userID := uint(1)
	email := "bench@example.com"
	expiration := 24 * time.Hour

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = utils.GenerateToken(userID, email, secret, expiration)
	}
}

func BenchmarkValidateToken(b *testing.B) {
	secret := "benchmark-secret"
	userID := uint(1)
	email := "bench@example.com"
	expiration := 24 * time.Hour

	token, _ := utils.GenerateToken(userID, email, secret, expiration)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = utils.ValidateToken(token, secret)
	}
}

func BenchmarkGenerateTokenParallel(b *testing.B) {
	secret := "benchmark-secret"
	userID := uint(1)
	email := "bench@example.com"
	expiration := 24 * time.Hour

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = utils.GenerateToken(userID, email, secret, expiration)
		}
	})
}
