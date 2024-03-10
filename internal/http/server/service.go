package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"GophKeeper/internal/http/server/handlers/cards"
	"GophKeeper/internal/http/server/handlers/medias"
	"GophKeeper/internal/http/server/handlers/notes"
	"GophKeeper/internal/http/server/handlers/passwords"
	"GophKeeper/internal/http/server/handlers/users"
	"GophKeeper/internal/http/server/middlewares/apptype"
	"GophKeeper/internal/http/server/middlewares/auth"
	"GophKeeper/internal/settings/server"
	"GophKeeper/internal/token"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type Service struct {
	logger     *zap.Logger
	server     *http.Server
	privateKey string
	publicKey  string
}

var (
	ErrCreateService = errors.New("create service")
)

type Route struct {
	Pattern string
	Handler http.Handler
}

type authorizer interface {
	ParseToken(string) (token.Payload, error)
}

func NewService(
	log *zap.Logger,
	set *server.APISettings,
	authorizer authorizer,
	provider users.UserProvider,
	passProvider passwords.PasswordProvider,
	cardProvider cards.CardProvider,
	notesProvider notes.NoteProvider,
	mediaProvider medias.MediaProvider,
	privateKey string,
	publicKey string,
) (*Service, error) {
	serlog := log.Named("http-service")

	mux := chi.NewRouter()

	// A good base middleware stack
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Timeout(60 * time.Second))

	mux.Use(apptype.ApplicationType(log))
	mux.Use(auth.CheckAuth(log, authorizer))

	createUserHandlerRoutes(log, mux, provider)
	createPasswordsHandlerRoutes(log, mux, passProvider)
	createCardsHandlerRoutes(log, mux, cardProvider)
	createNoteHandlerRoutes(log, mux, notesProvider)
	createMediasHandlerRoutes(log, mux, mediaProvider)

	var (
		address = set.Address
		port    = strconv.Itoa(set.Port)
	)

	// load CA certificate file and add it to list of client CAs
	caCertFile, err := os.ReadFile(publicKey)
	if err != nil {
		log.Fatal("error reading CA certificate: %v", zap.Error(err))

		return nil, fmt.Errorf("error reading certificate: %w", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertFile)

	// Create the TLS Config with the CA pool and enable Client certificate validation
	tlsConfig := &tls.Config{
		ClientCAs:          caCertPool,
		MinVersion:         tls.VersionTLS12,
		CurvePreferences:   []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		InsecureSkipVerify: true,
	}

	s := &http.Server{
		Addr:      net.JoinHostPort(address, port),
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	return &Service{
		logger:     serlog,
		server:     s,
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

func (s *Service) Run() error {
	s.logger.Debug("Running server on", zap.String("address", s.server.Addr))

	return s.server.ListenAndServeTLS(s.publicKey, s.privateKey)
}

func (s *Service) Stop(ctx context.Context) error {
	s.logger.Debug("stopping http service")

	return s.server.Shutdown(ctx)
}

func createUserHandlerRoutes(
	log *zap.Logger,
	mux *chi.Mux,
	provider users.UserProvider,
) {
	const (
		registerPath = "/v1/users/register"
		loginPath    = "/v1/users/login"
	)

	uh := users.NewHandler(log, provider)

	mux.Post(registerPath, uh.Register)
	mux.Post(loginPath, uh.Login)
}

func createPasswordsHandlerRoutes(
	log *zap.Logger,
	mux *chi.Mux,
	passProvider passwords.PasswordProvider,
) {
	const passwordPath = "/v1/passwords"

	ph := passwords.NewHandler(log, passProvider)

	mux.Post(passwordPath, ph.PasswordCreate)
	mux.Put(passwordPath, ph.PasswordUpdate)
	mux.Get(passwordPath, ph.Passwords)
	mux.Delete(passwordPath+"/{passwordID}", ph.PasswordDelete)
}

func createCardsHandlerRoutes(
	log *zap.Logger,
	mux *chi.Mux,
	provider cards.CardProvider,
) {
	const cardsPath = "/v1/cards"

	ch := cards.NewHandler(log, provider)

	mux.Post(cardsPath, ch.CardCreate)
	mux.Put(cardsPath, ch.CardUpdate)
	mux.Get(cardsPath, ch.Cards)
	mux.Delete(cardsPath+"/{cardID}", ch.CardDelete)
}

func createNoteHandlerRoutes(
	log *zap.Logger,
	mux *chi.Mux,
	provider notes.NoteProvider,
) {
	const notePath = "/v1/notes"

	nh := notes.NewHandler(log, provider)

	mux.Post(notePath, nh.NoteCreate)
	mux.Put(notePath, nh.NoteUpdate)
	mux.Get(notePath, nh.Notes)
	mux.Delete(notePath+"/{noteID}", nh.NoteDelete)
}

func createMediasHandlerRoutes(
	log *zap.Logger,
	mux *chi.Mux,
	provider medias.MediaProvider,
) {
	const mediaPath = "/v1/medias"

	mh := medias.NewHandler(log, provider)

	mux.Post(mediaPath, mh.MediaCreate)
	mux.Put(mediaPath, mh.MediaUpdate)
	mux.Get(mediaPath, mh.Medias)
	mux.Delete(mediaPath+"/{mediaID}", mh.MediaDelete)
}
