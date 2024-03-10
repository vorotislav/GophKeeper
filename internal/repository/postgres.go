package repository

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

//go:embed migrations/*
var migrations embed.FS

const (
	defaultMaxConnSize     = 10
	defaultMaxConnIdleTime = time.Minute * 10
)

var (
	ErrSourceDriver   = errors.New("cannot create source driver")
	ErrSourceInstance = errors.New("cannot create migrate")
	ErrMigrateUp      = errors.New("cannot migrate up")
	ErrCreateStorage  = errors.New("cannot create storage")
)

// Repo defines repository object for interacting with database repository or instances.
type Repo struct {
	log  *zap.Logger
	Pool *pgxpool.Pool
}

// NewRepository returns repository object.
func NewRepository(ctx context.Context, log *zap.Logger, databaseURI string) (*Repo, error) {
	if databaseURI == "" {
		return nil, fmt.Errorf("%w: database uri is empty", ErrCreateStorage)
	}

	c, err := pgxpool.ParseConfig(databaseURI)
	if err != nil {
		return nil, fmt.Errorf("parse connection string failed: %w", err)
	}

	// Set pool settings.
	c.MaxConns = defaultMaxConnSize
	c.MaxConnLifetime = time.Hour
	c.MaxConnIdleTime = defaultMaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("connect to repository failed: %w", err)
	}

	r := &Repo{
		log:  log.Named("repo"),
		Pool: pool,
	}

	// Apply migrations.
	if err = r.migrate(); err != nil {
		r.log.Error("migrations database schema", zap.Error(err))

		return nil, fmt.Errorf("%w: %w", ErrCreateStorage, err)
	}

	r.log.Debug("repo is created", zap.String("uri", databaseURI))

	return r, nil
}

// Migrate established dedicated database connection and applies schema migrations.
func (r *Repo) migrate() error {
	d, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("%w:%w", ErrSourceDriver, err)
	}

	connCfg := r.Pool.Config().ConnConfig
	m, err := migrate.NewWithSourceInstance("iofs", d,
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			connCfg.User, connCfg.Password, connCfg.Host, connCfg.Port, connCfg.Database))
	if err != nil {
		return fmt.Errorf("%w:%w", ErrSourceInstance, err)
	}

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}

		return fmt.Errorf("%w:%w", ErrMigrateUp, err)
	}

	return nil
}

func (r *Repo) Stop() {
	r.Pool.Close()
}
