package utils

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateShortcode(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func GenerateUniqueShortcode(ctx context.Context, initialShortcode string) (string, error) {
	shortcode := initialShortcode
	if shortcode == "" {
		for i := 0; i < 5; i++ { // Try up to 5 times to generate a unique shortcode
			shortcode = generateShortcode(4)
			// Check if the shortcode already exists
			result, err := GetDynamoDBItem(ctx, shortcode)
			if err != nil {
				return "", err
			}
			if result == nil {
				break // Shortcode is unique
			}
			if i == 4 {
				return "", fmt.Errorf("failed to generate a unique shortcode after 5 attempts")
			}
		}
	}
	return shortcode, nil
}
