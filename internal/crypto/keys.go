package crypto

import "crypto/rand"

const KeySize = 32 // 256-bit key for AES-256-GCM

// GenerateSessionKey returns a cryptographically random symmetric key.
// Both peers receive this key to encrypt/decrypt DataChannel messages.
func GenerateSessionKey() ([]byte, error) {
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}
