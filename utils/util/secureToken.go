package util

import (
	"crypto/rand"
	"encoding/base64"
)

// SecureToken returns a URL-safe random string built from n bytes of
// crypto/rand entropy. Use it for anything security-sensitive (OAuth state,
// nonces); RandomString in random.go is not cryptographically secure.
func SecureToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
