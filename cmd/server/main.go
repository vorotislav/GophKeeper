package main

import (
	"GophKeeper/internal/auth"
	"GophKeeper/internal/crypto/asymetry/generator"
	"GophKeeper/internal/crypto/cipher"
	"GophKeeper/internal/providers/cards"
	"GophKeeper/internal/providers/media"
	"GophKeeper/internal/providers/notes"
	"GophKeeper/internal/providers/passwords"
	"GophKeeper/internal/providers/users"
	"GophKeeper/internal/repository"
	"context"
	"errors"
	"fmt"
	stdlog "log"
	"os"
	"sync"
	"time"

	"GophKeeper/internal/http/server"
	"GophKeeper/internal/logger"
	serverSettings "GophKeeper/internal/settings/server"
	"GophKeeper/internal/signals"

	"go.uber.org/zap"
)

const (
	defaultKeysPath        = ".cert/"
	defaultPrivateKey      = "private.pem"
	defaultPublicKey       = "public.pem"
	defaultConfigFile      = "config.yaml"
	serviceShutdownTimeout = 1 * time.Second
)

var (
	buildVersion = "N/A" //nolint:gochecknoglobals
	buildDate    = "N/A" //nolint:gochecknoglobals
	buildCommit  = "N/A" //nolint:gochecknoglobals
)

func main() {
	configFile := parseFlag()
	if configFile == "" {
		configFile = defaultConfigFile
	}

	sets, err := serverSettings.NewSettings(configFile)
	if err != nil {
		stdlog.Fatal(err)
	}

	log, err := logger.New(sets.Log.Level, sets.Log.Format, "stdout", sets.Log.Verbose)
	if err != nil {
		stdlog.Fatal(err)
	}

	nlog := log.Named("main")
	nlog.Debug("Server starting...")
	nlog.Debug(fmt.Sprintf("Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		buildVersion, buildDate, buildCommit))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	oss := signals.NewOSSignals(ctx)

	oss.Subscribe(func(sig os.Signal) {
		nlog.Info("Stopping by OS Signal...",
			zap.String("signal", sig.String()))

		cancel()
	})

	authorizer, err := auth.NewAuthorizer(sets.JWT)
	if err != nil {
		nlog.Fatal("create authorizer", zap.Error(err))
	}

	repo, err := repository.NewRepository(ctx, log, sets.Database.URI)
	if err != nil {
		nlog.Fatal("create repository", zap.Error(err))
	}

	ciph := cipher.NewCipher(sets.Crypto.Key, sets.Crypto.Salt)

	userProvider := users.NewUsersProvider(log, repo, authorizer)
	passProvider := passwords.NewProvider(log, repo, ciph)
	cardProvider := cards.NewProvider(log, repo, ciph)
	notesProvider := notes.NewProvider(log, repo, ciph)
	mediaProvider := media.NewProvider(log, repo, ciph)

	err = checkKeys(log)
	if err != nil {
		nlog.Fatal("read keys", zap.Error(err))
	}

	httpService, err := server.NewService(
		log,
		&sets.API,
		authorizer,
		userProvider,
		passProvider,
		cardProvider,
		notesProvider,
		mediaProvider,
		defaultKeysPath+defaultPrivateKey,
		defaultKeysPath+defaultPublicKey,
	)
	if err != nil {
		nlog.Fatal("create http service", zap.Error(err))
	}

	serviceErrCh := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func(errCh chan<- error, wg *sync.WaitGroup) {
		defer wg.Done()
		defer close(errCh)

		if err := httpService.Run(); err != nil {
			errCh <- err
		}
	}(serviceErrCh, &wg)

	select {
	case err := <-serviceErrCh:
		if err != nil {
			nlog.Error("service error", zap.Error(err))
			cancel()
		}
	case <-ctx.Done():
		nlog.Info("Server stopping...")
		ctxShutdown, ctxCancelShutdown := context.WithTimeout(context.Background(), serviceShutdownTimeout)

		if err := httpService.Stop(ctxShutdown); err != nil {
			nlog.Error("cannot stop server", zap.Error(err))
		}

		ctxCancelShutdown()
	}

	wg.Wait()
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
