package asymetry

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"

	"GophKeeper/internal/settings/common"

	"go.uber.org/zap"
)

const (
	defaultKeysPath   = "./.cert"
	defaultPrivateKey = "private.pem"
	defaultPublicKey  = "public.pem"
)

type Manager struct {
	log            *zap.Logger
	keysDir        string
	privateKeyPath string
	publicKeyPath  string
}

func NewManager(log *zap.Logger, as common.Asymmetry) (*Manager, error) {
	path := as.KeysPath
	if path == "" {
		path = defaultKeysPath
	}

	privKeyName := as.PrivateKey
	if privKeyName == "" {
		privKeyName = defaultPrivateKey
	}

	pubKeyName := as.PublicKey
	if pubKeyName == "" {
		pubKeyName = defaultPublicKey
	}

	if _, err := os.Stat(path + string(os.PathSeparator) + privKeyName); errors.Is(err, os.ErrNotExist) {
		err := generate(log, path, privKeyName, pubKeyName)
		if err != nil {
			return nil, fmt.Errorf("read keys: %w", err)
		}
	}

	return &Manager{
		log:            log.Named("asymetry manager"),
		keysDir:        path,
		privateKeyPath: path + string(os.PathSeparator) + privKeyName,
		publicKeyPath:  path + string(os.PathSeparator) + pubKeyName,
	}, nil
}

func (m *Manager) PublicKeyPath() string {
	return m.publicKeyPath
}

func (m *Manager) PrivateKeyPath() string {
	return m.privateKeyPath
}

func (m *Manager) ReadPublicKey() ([]byte, error) {
	pk, err := os.ReadFile(m.publicKeyPath)
	if err != nil {
		m.log.Error("error reading CA certificate", zap.Error(err))

		return nil, fmt.Errorf("error reading CA certificate: %w", err)
	}

	return pk, nil
}

func generate(log *zap.Logger, path, privateKeyName, publicKeyName string) error {
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

	err = os.Mkdir(path, 0750)
	if err != nil {
		log.Error("create cert dir", zap.Error(err))

		return fmt.Errorf("create cert dir: %w", err)
	}

	err = os.WriteFile(path+string(os.PathSeparator)+privateKeyName, privateKeyPEM.Bytes(), 0o644)
	if err != nil {
		log.Error("write private key into file", zap.Error(err))

		return fmt.Errorf("write private key: %w", err)
	}

	err = os.WriteFile(path+string(os.PathSeparator)+publicKeyName, certPEM.Bytes(), 0o644)
	if err != nil {
		log.Error("write public key into file", zap.Error(err))

		return fmt.Errorf("write public key: %w", err)
	}

	return nil
}
