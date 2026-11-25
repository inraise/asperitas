package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func NewID() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return hex.EncodeToString(b), nil
}
