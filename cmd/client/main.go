package main

import (
	"crypto/tls"
	"crypto/x509"
	stdlog "log"
	"net/http"
	"time"

	"GophKeeper/cmd/util"
	"GophKeeper/internal/crypto/asymetry"
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
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func main() {
	configFile := util.ParseFlags()

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

	am, err := asymetry.NewManager(log, sets.Asymmetry)
	if err != nil {
		log.Fatal("create asymmetry manager", zap.Error(err))
	}

	transport, err := getTransport(am)
	if err != nil {
		log.Fatal("cannot create transport for http", zap.Error(err))
	}

	sessionClient := session.NewClient(log, sys, sessionStore, sets.Server.Address, transport)
	passwordClient := password.NewClient(log, sessionStore, sets.Server.Address, transport)
	cardsClient := cards.NewClient(log, sessionStore, sets.Server.Address, transport)
	notesClient := notes.NewClient(log, sessionStore, sets.Server.Address, transport)
	mediaClient := media.NewClient(log, sessionStore, sets.Server.Address, transport)

	cr := cron.New()

	t := tea.NewProgram(
		tui.InitModel(sessionClient, passwordClient, cardsClient, notesClient, mediaClient, cr),
		tea.WithAltScreen())

	if _, err := t.Run(); err != nil {
		nlog.Fatal("running program", zap.Error(err))
	}
}

func getTransport(am *asymetry.Manager) (*http.Transport, error) {
	cert, err := tls.LoadX509KeyPair(am.PublicKeyPath(), am.PrivateKeyPath())
	if err != nil {
		return nil, err
	}

	caCert, err := am.ReadPublicKey()
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
