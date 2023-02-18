package util

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

func DefaultNonceGenerator(size int) ([]byte, error) {
	nonce := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, nonce)

	return nonce, err
}

func EncryptTwitchID(input string, gcm cipher.AEAD, nonceFn func(int) ([]byte, error)) ([]byte, error) {
	if nonceFn == nil {
		nonceFn = DefaultNonceGenerator
	}

	nonce, err := nonceFn(gcm.NonceSize())
	if err != nil {
		return []byte{}, err
	}

	return gcm.Seal(nonce, nonce, []byte(input), nil), nil
}

func DecryptTwitchID(input string, gcm cipher.AEAD) ([]byte, error) {
	nonceSize := gcm.NonceSize()
	if len(input) < nonceSize {
		return []byte(""), fmt.Errorf("encrypted value smaller than nonce size of %d", nonceSize)
	}

	nonce, ciphertext := input[:nonceSize], input[nonceSize:]
	return gcm.Open(nil, []byte(nonce), []byte(ciphertext), nil)
}
