package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func NormalizeKey(key string) []byte {
	return []byte(strings.TrimSpace(key))
}

func Verify(data []byte, expected string) bool {
	return Digest(data) == strings.ToLower(strings.TrimSpace(expected))
}

func Envelope(key, name string, plaintext []byte) ([]byte, string, int, error) {
	ciphertext, digest, err := Encrypt(NormalizeKey(key), plaintext)
	if err != nil {
		return nil, "", 0, err
	}
	return ciphertext, digest, len(name) + len(ciphertext), nil
}

func DigestParts(parts ...[]byte) string {
	joined := make([]byte, 0)
	for _, part := range parts {
		joined = append(joined, part...)
		joined = append(joined, 0)
	}
	return Digest(joined)
}

func ChecksumLabel(name string, data []byte) string {
	return strings.TrimSpace(name) + ":" + Digest(data)
}

func KeyFingerprint(key []byte) string {
	if len(key) == 0 {
		return ""
	}
	return Digest(key)[:16]
}

func ValidDigest(value string) bool {
	if len(strings.TrimSpace(value)) != 64 {
		return false
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f') || (character >= 'A' && character <= 'F')) {
			return false
		}
	}
	return true
}
