package auth

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

func RandomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		log.Printf("[AUTH] Error generating token: %v", err)
		return "", err
	}

    t := base64.RawURLEncoding.EncodeToString(b)

    log.Printf("[AUTH] Generated token (len=%d)", len(t))

	return t, nil
}
