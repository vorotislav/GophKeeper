// Package repository обеспечивает работу с репозиторием на базе PostgreSQL.
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
	// ErrSourceDriver возвращается если не удалось получить объект migrations.
	ErrSourceDriver = errors.New("cannot create source driver")
	// ErrSourceInstance возвращается если не удалось подключиться к БД по указанным параметрам для выполнения миграций.
	ErrSourceInstance = errors.New("cannot create migrate")
	// ErrMigrateUp возвращается если не удалось применить миграции.
	ErrMigrateUp = errors.New("cannot migrate up")
	// ErrCreateStorage возвращается если не удалось создать объект Repo.
	ErrCreateStorage = errors.New("cannot create storage")
)

// Repo описывает структуру репозитория для работы с БД.
type Repo struct {
	log  *zap.Logger
	Pool *pgxpool.Pool
}

// NewRepository принимает строку подключения и пытается создать объект Repo и выполнить подключение к БД.
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

// migrate устанавливает соединение до БД и выполняет миграции схемы.
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

// Stop закрывает пул соединений до БД.
func (r *Repo) Stop() {
	r.Pool.Close()
}
