package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"runtime"

	"golang.org/x/crypto/argon2"
)

const (
	keyHashLength = 32
	keyHashMemory = 64 * 1024
	keyHashTime   = 1
)

var (
	errInvalidCipherTextLen = errors.New("invalid ciphertext len")
)

// Cipher defines object for saving hash of key and salt as key to AES and provide methods for encrypt and decrypt.
type Cipher struct {
	secretHash []byte
}

// NewCipher is constructor for Cipher.
func NewCipher(key, salt string) *Cipher {
	c := &Cipher{
		secretHash: secretKeyHash(key, salt),
	}

	return c
}

// secretKeyHash return hash of secret key of 32 bytes in length by Argon2.
func secretKeyHash(secretKey, salt string) []byte {
	return argon2.IDKey([]byte(secretKey), []byte(salt),
		keyHashTime, keyHashMemory, uint8(runtime.NumCPU()), keyHashLength)
}

// EncryptString accepts a string and encrypts with the AES algorithm.
func (c *Cipher) EncryptString(value string) (string, error) {
	aesCipher, err := aes.NewCipher(c.secretHash)
	if err != nil {
		return "", fmt.Errorf("aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(aesCipher)
	if err != nil {
		return "", fmt.Errorf("gcm wrapped: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())

	_, err = rand.Read(nonce)
	if err != nil {
		return "", fmt.Errorf("rand nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(value), nil)

	return hex.EncodeToString(ciphertext), nil
}

// DecryptString accepts encrypted string and return decrypted it.
func (c *Cipher) DecryptString(value string) (string, error) {
	ciphertext, _ := hex.DecodeString(value)
	aesCipher, err := aes.NewCipher(c.secretHash)

	if err != nil {
		return "", fmt.Errorf("aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(aesCipher)
	if err != nil {
		return "", fmt.Errorf("gcm wrapped: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) <= nonceSize {
		return "", errInvalidCipherTextLen
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("open decrypts and authenticates ciphertext: %w", err)
	}

	return string(plaintext), nil
}
