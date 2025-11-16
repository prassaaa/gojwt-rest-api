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

// Refresh Token Tests

func TestGenerateTokenPair(t *testing.T) {
	secret := "test-secret-key"
	userID := uint(1)
	email := "test@example.com"
	accessExpiry := 15 * time.Minute
	refreshExpiry := 7 * 24 * time.Hour

	t.Run("Generate valid token pair", func(t *testing.T) {
		tokenPair, tokenFamily, err := utils.GenerateTokenPair(userID, email, secret, accessExpiry, refreshExpiry)

		require.NoError(t, err)
		require.NotNil(t, tokenPair)
		assert.NotEmpty(t, tokenPair.AccessToken)
		assert.NotEmpty(t, tokenPair.RefreshToken)
		assert.NotEmpty(t, tokenFamily)
		assert.Equal(t, int64(accessExpiry.Seconds()), tokenPair.ExpiresIn)

		// Validate access token
		claims, err := utils.ValidateToken(tokenPair.AccessToken, secret)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)

		// Verify refresh token is not a JWT (should be random string)
		_, err = utils.ValidateToken(tokenPair.RefreshToken, secret)
		assert.Error(t, err) // Refresh token should not be a JWT
	})

	t.Run("Generate multiple unique token pairs", func(t *testing.T) {
		pair1, family1, err := utils.GenerateTokenPair(userID, email, secret, accessExpiry, refreshExpiry)
		require.NoError(t, err)

		// Sleep for 1 second to ensure different timestamps in JWT
		time.Sleep(1 * time.Second)

		pair2, family2, err := utils.GenerateTokenPair(userID, email, secret, accessExpiry, refreshExpiry)
		require.NoError(t, err)

		// Access tokens should be different due to different timestamps
		assert.NotEqual(t, pair1.AccessToken, pair2.AccessToken)
		// Refresh tokens (random) should always be unique
		assert.NotEqual(t, pair1.RefreshToken, pair2.RefreshToken)
		// Token families should be unique
		assert.NotEqual(t, family1, family2)
	})

	t.Run("Refresh tokens are cryptographically secure", func(t *testing.T) {
		tokens := make(map[string]bool)
		iterations := 100

		for i := 0; i < iterations; i++ {
			pair, _, err := utils.GenerateTokenPair(userID, email, secret, accessExpiry, refreshExpiry)
			require.NoError(t, err)

			// Check for duplicates
			assert.False(t, tokens[pair.RefreshToken], "Duplicate refresh token generated")
			tokens[pair.RefreshToken] = true

			// Verify length (base64 encoded 32 bytes should be ~44 chars)
			assert.Greater(t, len(pair.RefreshToken), 40)
		}
	})
}

func TestExtractTokenExpiry(t *testing.T) {
	secret := "test-secret-key"
	userID := uint(1)
	email := "test@example.com"

	t.Run("Extract expiry from valid token", func(t *testing.T) {
		expiration := 2 * time.Hour
		token, err := utils.GenerateToken(userID, email, secret, expiration)
		require.NoError(t, err)

		expiryTime, err := utils.ExtractTokenExpiry(token, secret)
		require.NoError(t, err)

		expectedExpiry := time.Now().Add(expiration)
		diff := expiryTime.Sub(expectedExpiry)
		assert.Less(t, diff.Abs(), 5*time.Second)
	})

	t.Run("Extract expiry fails with invalid token", func(t *testing.T) {
		_, err := utils.ExtractTokenExpiry("invalid.token", secret)
		assert.Error(t, err)
	})

	t.Run("Extract expiry fails with wrong secret", func(t *testing.T) {
		token, err := utils.GenerateToken(userID, email, secret, time.Hour)
		require.NoError(t, err)

		_, err = utils.ExtractTokenExpiry(token, "wrong-secret")
		assert.Error(t, err)
	})
}

func TestTokenPairIntegration(t *testing.T) {
	secret := "integration-test-secret"
	userID := uint(42)
	email := "integration@test.com"
	accessExpiry := 15 * time.Minute
	refreshExpiry := 7 * 24 * time.Hour

	t.Run("Complete token flow", func(t *testing.T) {
		// 1. Generate initial token pair
		pair1, family1, err := utils.GenerateTokenPair(userID, email, secret, accessExpiry, refreshExpiry)
		require.NoError(t, err)

		// 2. Validate access token works
		claims, err := utils.ValidateToken(pair1.AccessToken, secret)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)

		// Sleep to ensure different timestamps
		time.Sleep(1 * time.Second)

		// 3. Simulate refresh - generate new pair with same user
		pair2, family2, err := utils.GenerateTokenPair(userID, email, secret, accessExpiry, refreshExpiry)
		require.NoError(t, err)

		// 4. Verify new tokens are different
		assert.NotEqual(t, pair1.AccessToken, pair2.AccessToken)
		assert.NotEqual(t, pair1.RefreshToken, pair2.RefreshToken)

		// 5. Token families should be different for different logins
		assert.NotEqual(t, family1, family2)

		// 6. Both access tokens should be valid
		claims1, err := utils.ValidateToken(pair1.AccessToken, secret)
		require.NoError(t, err)
		claims2, err := utils.ValidateToken(pair2.AccessToken, secret)
		require.NoError(t, err)

		assert.Equal(t, claims1.UserID, claims2.UserID)
		assert.Equal(t, claims1.Email, claims2.Email)
	})
}

func BenchmarkGenerateTokenPair(b *testing.B) {
	secret := "benchmark-secret"
	userID := uint(1)
	email := "bench@example.com"
	accessExpiry := 15 * time.Minute
	refreshExpiry := 7 * 24 * time.Hour

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = utils.GenerateTokenPair(userID, email, secret, accessExpiry, refreshExpiry)
	}
}
