package crypto

import (
	"crypto/sha256"
	"errors"
)

var ErrBadKey = errors.New("encryption key is empty")

func derive(key []byte, length int) []byte {
	stream := make([]byte, 0, length)
	var counter byte
	for len(stream) < length {
		h := sha256.New()
		h.Write(key)
		h.Write([]byte{counter})
		stream = append(stream, h.Sum(nil)...)
		counter++
	}
	return stream[:length]
}

func Encrypt(key, plaintext []byte) ([]byte, string, error) {
	if len(key) == 0 {
		return nil, "", ErrBadKey
	}
	mask := derive(key, len(plaintext))
	ciphertext := make([]byte, len(plaintext))
	for i, b := range plaintext {
		ciphertext[i] = b ^ mask[i]
	}
	return ciphertext, Digest(plaintext), nil
}

func Decrypt(key, ciphertext []byte, expectedDigest string) ([]byte, error) {
	if len(key) == 0 {
		return nil, ErrBadKey
	}
	mask := derive(key, len(ciphertext))
	plaintext := make([]byte, len(ciphertext))
	for i, b := range ciphertext {
		plaintext[i] = b ^ mask[i]
	}
	if Digest(plaintext) != expectedDigest {
		return nil, errors.New("digest mismatch")
	}
	return plaintext, nil
}

func EncryptChunks(key, plaintext []byte, chunkSize int) ([][]byte, string, error) {
	if chunkSize <= 0 {
		return nil, "", errors.New("chunk size must be positive")
	}
	chunks := make([][]byte, 0)
	for offset := 0; offset < len(plaintext); offset += chunkSize {
		end := offset + chunkSize
		if end > len(plaintext) {
			end = len(plaintext)
		}
		ciphertext, _, err := Encrypt(key, plaintext[offset:end])
		if err != nil {
			return nil, "", err
		}
		chunks = append(chunks, ciphertext)
	}
	return chunks, Digest(plaintext), nil
}

func DecryptChunks(key []byte, chunks [][]byte, expectedDigest string) ([]byte, error) {
	plain := make([]byte, 0)
	for _, chunk := range chunks {
		decoded, err := Decrypt(key, chunk, Digest(decodedMask(chunk, key)))
		if err != nil {
			return nil, err
		}
		plain = append(plain, decoded...)
	}
	if Digest(plain) != expectedDigest {
		return nil, errors.New("chunk digest mismatch")
	}
	return plain, nil
}

func decodedMask(ciphertext, key []byte) []byte {
	mask := derive(key, len(ciphertext))
	plain := make([]byte, len(ciphertext))
	for index, value := range ciphertext {
		plain[index] = value ^ mask[index]
	}
	return plain
}

func Rotate(key, next []byte, ciphertext []byte, digest string) ([]byte, error) {
	plain, err := Decrypt(key, ciphertext, digest)
	if err != nil {
		return nil, err
	}
	rotated, _, err := Encrypt(next, plain)
	return rotated, err
}
