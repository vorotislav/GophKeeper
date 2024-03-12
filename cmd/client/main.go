package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"go.uber.org/zap"
	stdlog "log"
	"net/http"
	"os"
	"time"

	"GophKeeper/internal/crypto/asymetry/generator"
	"GophKeeper/internal/http/client/cards"
	"GophKeeper/internal/http/client/media"
	"GophKeeper/internal/http/client/notes"
	"GophKeeper/internal/http/client/password"
	"GophKeeper/internal/http/client/session"
	"GophKeeper/internal/logger"
	sesStore "GophKeeper/internal/session"
	"GophKeeper/internal/settings/client"
	"GophKeeper/internal/systems"
	"GophKeeper/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	buildVersion = "N/A" //nolint:gochecknoglobals
	buildDate    = "N/A" //nolint:gochecknoglobals
	buildCommit  = "N/A" //nolint:gochecknoglobals
)

const (
	defaultKeysPath   = ".cert/"
	defaultPrivateKey = "private.pem"
	defaultPublicKey  = "public.pem"
	defaultConfigFile = "config.yaml"
)

func main() {
	configFile := parseFlag()
	if configFile == "" {
		configFile = defaultConfigFile
	}

	sets, err := client.NewSettings(configFile)
	if err != nil {
		stdlog.Fatal(err)
	}

	log, err := logger.New(sets.Log.Level, sets.Log.Format, "client.log", sets.Log.Verbose)
	if err != nil {
		stdlog.Fatal(err)
	}

	nlog := log.Named("main")

	sys := &systems.Systems{}

	sessionStore := &sesStore.Storage{}

	err = checkKeys(log)
	if err != nil {
		log.Fatal("check keys", zap.Error(err))
	}

	transport, err := getTransport()
	if err != nil {
		log.Fatal("cannot create transport for http", zap.Error(err))
	}

	sessionClient := session.NewClient(log, sys, sessionStore, sets.Server.Address, transport)
	passwordClient := password.NewClient(log, sessionStore, sets.Server.Address, transport)
	cardsClient := cards.NewClient(log, sessionStore, sets.Server.Address, transport)
	notesClient := notes.NewClient(log, sessionStore, sets.Server.Address, transport)
	mediaClient := media.NewClient(log, sessionStore, sets.Server.Address, transport)

	t := tea.NewProgram(
		tui.InitModel(sessionClient, passwordClient, cardsClient, notesClient, mediaClient),
		tea.WithAltScreen())

	if _, err := t.Run(); err != nil {
		nlog.Fatal("running program", zap.Error(err))
	}
}

func checkKeys(log *zap.Logger) error {
	if _, err := os.Stat(defaultKeysPath + defaultPrivateKey); errors.Is(err, os.ErrNotExist) {
		err := generator.Generate(log)
		if err != nil {
			return fmt.Errorf("read keys: %w", err)
		}
	}

	return nil
}

func getTransport() (*http.Transport, error) {
	cert, err := tls.LoadX509KeyPair(defaultKeysPath+defaultPublicKey, defaultKeysPath+defaultPrivateKey)
	if err != nil {
		return nil, err
	}

	caCert, err := os.ReadFile(defaultKeysPath + defaultPublicKey)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
	}

	return &http.Transport{
		TLSClientConfig: tlsConfig,
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}, nil
}
