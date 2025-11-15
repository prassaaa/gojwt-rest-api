package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("   JWT Secret Generator")
	fmt.Println("===========================================")
	fmt.Println()

	// Generate 32 bytes random data (256 bits)
	secret, err := generateSecret(32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating secret: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Your JWT Secret (Base64 encoded):")
	fmt.Println("-------------------------------------------")
	fmt.Println(secret)
	fmt.Println("-------------------------------------------")
	fmt.Println()
	fmt.Println("Copy this to your .env file:")
	fmt.Printf("JWT_SECRET=%s\n", secret)
	fmt.Println()
	fmt.Println("⚠️  IMPORTANT:")
	fmt.Println("  - Keep this secret safe!")
	fmt.Println("  - Don't commit to Git")
	fmt.Println("  - Use different secrets for dev/staging/prod")
	fmt.Println()
}

// generateSecret generates a random base64 encoded string
func generateSecret(length int) (string, error) {
	bytes := make([]byte, length)

	// Read random bytes
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Encode to base64
	return base64.StdEncoding.EncodeToString(bytes), nil
}
