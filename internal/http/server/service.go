package server

import (
	ch "GophKeeper/internal/http"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
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

type asymmetryManager interface {
	PublicKeyPath() string
	PrivateKeyPath() string
	ReadPublicKey() ([]byte, error)
}

type Service struct {
	logger    *zap.Logger
	server    *http.Server
	asManager asymmetryManager
}

var (
	ErrCreateService = errors.New("create service")
)

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
	asManager asymmetryManager,
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
	caCertFile, err := asManager.ReadPublicKey()
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
		logger:    serlog,
		server:    s,
		asManager: asManager,
	}, nil
}

func (s *Service) Run() error {
	s.logger.Debug("Running server on", zap.String("address", s.server.Addr))

	return s.server.ListenAndServeTLS(s.asManager.PublicKeyPath(), s.asManager.PrivateKeyPath())
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
	uh := users.NewHandler(log, provider)

	mux.Post(ch.RegisterPath, uh.Register)
	mux.Post(ch.LoginPath, uh.Login)
}

func createPasswordsHandlerRoutes(
	log *zap.Logger,
	mux *chi.Mux,
	passProvider passwords.PasswordProvider,
) {
	ph := passwords.NewHandler(log, passProvider)

	mux.Post(ch.PasswordsPath, ph.PasswordCreate)
	mux.Put(ch.PasswordsPath, ph.PasswordUpdate)
	mux.Get(ch.PasswordsPath, ph.Passwords)
	mux.Delete(ch.PasswordsPath+"/{passwordID}", ph.PasswordDelete)
}

func createCardsHandlerRoutes(
	log *zap.Logger,
	mux *chi.Mux,
	provider cards.CardProvider,
) {
	crh := cards.NewHandler(log, provider)

	mux.Post(ch.CardsPath, crh.CardCreate)
	mux.Put(ch.CardsPath, crh.CardUpdate)
	mux.Get(ch.CardsPath, crh.Cards)
	mux.Delete(ch.CardsPath+"/{cardID}", crh.CardDelete)
}

func createNoteHandlerRoutes(
	log *zap.Logger,
	mux *chi.Mux,
	provider notes.NoteProvider,
) {
	nh := notes.NewHandler(log, provider)

	mux.Post(ch.NotesPath, nh.NoteCreate)
	mux.Put(ch.NotesPath, nh.NoteUpdate)
	mux.Get(ch.NotesPath, nh.Notes)
	mux.Delete(ch.NotesPath+"/{noteID}", nh.NoteDelete)
}

func createMediasHandlerRoutes(
	log *zap.Logger,
	mux *chi.Mux,
	provider medias.MediaProvider,
) {
	mh := medias.NewHandler(log, provider)

	mux.Post(ch.MediaPath, mh.MediaCreate)
	mux.Put(ch.MediaPath, mh.MediaUpdate)
	mux.Get(ch.MediaPath, mh.Medias)
	mux.Delete(ch.MediaPath+"/{mediaID}", mh.MediaDelete)
}
