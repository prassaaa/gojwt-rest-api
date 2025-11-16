package unit

import (
	"gojwt-rest-api/internal/utils"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "Short password",
			password: "pwd",
			wantErr:  false,
		},
		{
			name:     "Long password (within bcrypt 72 byte limit)",
			password: strings.Repeat("a", 70),
			wantErr:  false,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "Password with special characters",
			password: "P@ssw0rd!#$%^&*()",
			wantErr:  false,
		},
		{
			name:     "Unicode password",
			password: "пароль123密码",
			wantErr:  false,
		},
		{
			name:     "Password with spaces",
			password: "pass word 123",
			wantErr:  false,
		},
		{
			name:     "Password with newlines",
			password: "pass\nword",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := utils.HashPassword(tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.NotEqual(t, tt.password, hash, "Hash should not equal plain password")

			// Verify bcrypt format
			assert.True(t, strings.HasPrefix(hash, "$2a$"), "Hash should have bcrypt prefix")
		})
	}
}

func TestHashPasswordDeterminism(t *testing.T) {
	password := "test-password"

	hash1, err1 := utils.HashPassword(password)
	require.NoError(t, err1)

	hash2, err2 := utils.HashPassword(password)
	require.NoError(t, err2)

	// Hashes should be different (bcrypt uses random salt)
	assert.NotEqual(t, hash1, hash2, "Different hashes should be generated for the same password")

	// But both should validate the same password
	assert.NoError(t, utils.CheckPassword(hash1, password))
	assert.NoError(t, utils.CheckPassword(hash2, password))
}

func TestCheckPassword(t *testing.T) {
	password := "correct-password"
	hash, err := utils.HashPassword(password)
	require.NoError(t, err)

	tests := []struct {
		name           string
		hashedPassword string
		plainPassword  string
		wantErr        bool
	}{
		{
			name:           "Correct password",
			hashedPassword: hash,
			plainPassword:  password,
			wantErr:        false,
		},
		{
			name:           "Incorrect password",
			hashedPassword: hash,
			plainPassword:  "wrong-password",
			wantErr:        true,
		},
		{
			name:           "Empty password",
			hashedPassword: hash,
			plainPassword:  "",
			wantErr:        true,
		},
		{
			name:           "Case sensitive - uppercase",
			hashedPassword: hash,
			plainPassword:  "CORRECT-PASSWORD",
			wantErr:        true,
		},
		{
			name:           "Password with extra spaces",
			hashedPassword: hash,
			plainPassword:  " correct-password ",
			wantErr:        true,
		},
		{
			name:           "Password with trailing newline",
			hashedPassword: hash,
			plainPassword:  "correct-password\n",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.CheckPassword(tt.hashedPassword, tt.plainPassword)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, bcrypt.ErrMismatchedHashAndPassword, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckPasswordWithInvalidHash(t *testing.T) {
	tests := []struct {
		name           string
		hashedPassword string
		plainPassword  string
	}{
		{
			name:           "Invalid hash format",
			hashedPassword: "not-a-valid-hash",
			plainPassword:  "password",
		},
		{
			name:           "Empty hash",
			hashedPassword: "",
			plainPassword:  "password",
		},
		{
			name:           "Truncated hash",
			hashedPassword: "$2a$10$",
			plainPassword:  "password",
		},
		{
			name:           "Hash with wrong prefix",
			hashedPassword: "$1$wrongprefix",
			plainPassword:  "password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.CheckPassword(tt.hashedPassword, tt.plainPassword)
			assert.Error(t, err)
		})
	}
}

func TestPasswordHashingRoundTrip(t *testing.T) {
	passwords := []string{
		"simple",
		"P@ssw0rd!",
		"very-long-password-with-many-characters-123456789",
		"短密码",
		"emoji-password-🔐🔑",
		"MixedCasE123",
		"with spaces in it",
	}

	for _, password := range passwords {
		t.Run("Password: "+password, func(t *testing.T) {
			// Hash the password
			hash, err := utils.HashPassword(password)
			require.NoError(t, err)
			require.NotEmpty(t, hash)

			// Verify correct password
			err = utils.CheckPassword(hash, password)
			assert.NoError(t, err, "Should validate correct password")

			// Verify incorrect password fails
			err = utils.CheckPassword(hash, password+"wrong")
			assert.Error(t, err, "Should reject incorrect password")
		})
	}
}

func TestVeryLongPassword(t *testing.T) {
	// bcrypt has a maximum password length of 72 bytes
	// Test passwords around this limit
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "71 bytes",
			password: strings.Repeat("a", 71),
			wantErr:  false,
		},
		{
			name:     "72 bytes",
			password: strings.Repeat("a", 72),
			wantErr:  false,
		},
		{
			name:     "73 bytes (exceeds bcrypt limit)",
			password: strings.Repeat("a", 73),
			wantErr:  true, // bcrypt will reject this
		},
		{
			name:     "100 bytes",
			password: strings.Repeat("a", 100),
			wantErr:  true, // bcrypt will reject this
		},
		{
			name:     "1000 bytes",
			password: strings.Repeat("a", 1000),
			wantErr:  true, // bcrypt will reject this
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := utils.HashPassword(tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			// Should validate the password
			err = utils.CheckPassword(hash, tt.password)
			assert.NoError(t, err)
		})
	}
}

func TestPasswordHashLength(t *testing.T) {
	password := "test-password"
	hash, err := utils.HashPassword(password)
	require.NoError(t, err)

	// Bcrypt hashes are always 60 characters
	assert.Equal(t, 60, len(hash), "Bcrypt hash should be 60 characters")
}

func TestMultiplePasswordHashing(t *testing.T) {
	password := "test-password"
	iterations := 10

	hashes := make(map[string]bool)

	for i := 0; i < iterations; i++ {
		hash, err := utils.HashPassword(password)
		require.NoError(t, err)

		// All hashes should be unique
		assert.False(t, hashes[hash], "Generated duplicate hash")
		hashes[hash] = true

		// But all should validate the same password
		err = utils.CheckPassword(hash, password)
		assert.NoError(t, err)
	}

	assert.Equal(t, iterations, len(hashes), "Should have generated unique hashes")
}

func TestPasswordValidationFailures(t *testing.T) {
	correctPassword := "MySecretPassword123"
	hash, err := utils.HashPassword(correctPassword)
	require.NoError(t, err)

	incorrectPasswords := []string{
		"mysecretpassword123",        // wrong case
		"MySecretPassword124",        // different number
		"MySecretPassword",           // missing number
		"MySecretPassword123 ",       // trailing space
		" MySecretPassword123",       // leading space
		"MySecretPassword123\n",      // trailing newline
		"MySecret Password123",       // extra space
		"",                           // empty
		"CompletelyDifferentPassword", // completely wrong
	}

	for _, wrongPassword := range incorrectPasswords {
		t.Run("Wrong password: "+wrongPassword, func(t *testing.T) {
			err := utils.CheckPassword(hash, wrongPassword)
			assert.Error(t, err, "Should reject incorrect password: %s", wrongPassword)
		})
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "benchmark-password-123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = utils.HashPassword(password)
	}
}

func BenchmarkCheckPassword(b *testing.B) {
	password := "benchmark-password-123"
	hash, _ := utils.HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = utils.CheckPassword(hash, password)
	}
}

func BenchmarkHashPasswordParallel(b *testing.B) {
	password := "benchmark-password-123"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = utils.HashPassword(password)
		}
	})
}

func BenchmarkCheckPasswordParallel(b *testing.B) {
	password := "benchmark-password-123"
	hash, _ := utils.HashPassword(password)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = utils.CheckPassword(hash, password)
		}
	})
}
