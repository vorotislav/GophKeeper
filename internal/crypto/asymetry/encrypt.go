package asymetry

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

var ErrDecodePublicKey = errors.New("decode pem public key")

func Encrypt(publicKeyPath string, data []byte) ([]byte, error) {
	publicKeyPEM, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}

	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return nil, ErrDecodePublicKey
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, cert.PublicKey.(*rsa.PublicKey), data)
	if err != nil {
		return nil, fmt.Errorf("encrypt data: %w", err)
	}

	return encrypted, nil
}
