package domain

import (
	"crypto/rand"
	"encoding/hex"
)

// NewID generuje losowy identyfikator encji.
func NewID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
