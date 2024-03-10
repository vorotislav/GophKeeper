package generator

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"

	"go.uber.org/zap"
)

const (
	defaultKeysPath = "./.cert/"
)

func Generate(log *zap.Logger) error {
	log = log.Named("generator")
	// создаём шаблон сертификата
	cert := &x509.Certificate{
		// указываем уникальный номер сертификата
		SerialNumber: big.NewInt(1659),
		// заполняем базовую информацию о владельце сертификата
		Subject: pkix.Name{
			Organization: []string{"YP"},
			Country:      []string{"RU"},
		},
		// разрешаем использование сертификата для 127.0.0.1 и ::1,
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		// сертификат верен, начиная со времени создания
		NotBefore: time.Now(),
		// время жизни сертификата - 10 лет
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		// устанавливаем использование ключа для цифровой подписи,
		// а также клиентской и серверной авторизации
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	// создаем новый приватный RSA-ключ длиной 4096 бит
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Error("generate private key", zap.Error(err))

		return fmt.Errorf("generate private key: %s", err)
	}

	// создаем сертификат
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Error("create certificate", zap.Error(err))

		return fmt.Errorf("create certificate: %w", err)
	}

	// кодируем сертифик и ключ в формате РЕМ, который используется для хранения и обмена криптографическими ключами
	var certPEM bytes.Buffer
	_ = pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	var privateKeyPEM bytes.Buffer
	_ = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	err = os.Mkdir(defaultKeysPath, 0750)
	if err != nil {
		log.Error("create cert dir", zap.Error(err))

		return fmt.Errorf("create cert dir: %w", err)
	}

	err = os.WriteFile("./.cert/private.pem", privateKeyPEM.Bytes(), 0o644)
	if err != nil {
		log.Error("write private key into file", zap.Error(err))

		return fmt.Errorf("write private key: %w", err)
	}

	err = os.WriteFile("./.cert/public.pem", certPEM.Bytes(), 0o644)
	if err != nil {
		log.Error("write public key into file", zap.Error(err))

		return fmt.Errorf("write public key: %w", err)
	}

	return nil
}
