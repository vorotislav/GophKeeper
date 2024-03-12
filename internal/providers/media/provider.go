// Package media пакет для описания сущности по управлению медиа пользователя.
package media

import (
	"context"
	"errors"
	"fmt"
	"time"

	"GophKeeper/internal/models"
	"GophKeeper/internal/providers/media/postgres"
	"GophKeeper/internal/repository"
	"GophKeeper/internal/token"

	"go.uber.org/zap"
)

var (
	errMediaProvider = errors.New("media provider error")
)

// crypto описывает доступные методы для шифрования и дешифрования медиа пользователя.
//
//go:generate go run github.com/vektra/mockery/v2@v2.24.0 --name=crypto --exported --with-expecter=true
type crypto interface {
	EncryptString(value string) (string, error)
	DecryptString(value string) (string, error)
}

// storage описывает доступные методы для работы с репозиторием и хранением медиа пользователей.
//
//go:generate go run github.com/vektra/mockery/v2@v2.24.0 --name=storage --exported --with-expecter=true
type storage interface {
	MediaCreate(ctx context.Context, m models.Media, userID int) error
	MediaUpdate(ctx context.Context, m models.Media, userID int) error
	MediaDelete(ctx context.Context, id int, userID int) error
	Medias(ctx context.Context, userID int) ([]models.Media, error)
}

// Provider структура для управления медиа, хранит в себе хранилище и объект для шифрования\дешифрования.
type Provider struct {
	log    *zap.Logger
	store  storage
	crypto crypto
}

// NewProvider конструктор для Provider.
func NewProvider(log *zap.Logger, repo *repository.Repo, crypto crypto) *Provider {
	return &Provider{
		log:    log.Named("media provider"),
		store:  postgres.NewMediaStorage(repo),
		crypto: crypto,
	}
}

// MediaCreate принимает объект модели медиа и сохраняет в репозиторий.
func (p *Provider) MediaCreate(ctx context.Context, m models.Media) error {
	paylod, err := token.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", errMediaProvider, err)
	}

	encBytes, err := p.crypto.EncryptString(string(m.Body))
	if err != nil {
		return fmt.Errorf("%w: %w", errMediaProvider, err)
	}

	m.Body = []byte(encBytes)
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()

	err = p.store.MediaCreate(ctx, m, paylod.ID)
	if err != nil {
		return fmt.Errorf("%w: %w", errMediaProvider, err)
	}

	return nil
}

// MediaUpdate принимает объект модель медиа и обновляет в репозитории.
func (p *Provider) MediaUpdate(ctx context.Context, m models.Media) error {
	paylod, err := token.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", errMediaProvider, err)
	}

	encBytes, err := p.crypto.EncryptString(string(m.Body))
	if err != nil {
		return fmt.Errorf("%w: %w", errMediaProvider, err)
	}

	m.Body = []byte(encBytes)
	m.UpdatedAt = time.Now()

	err = p.store.MediaUpdate(ctx, m, paylod.ID)
	if err != nil {
		return fmt.Errorf("%w: %w", errMediaProvider, err)
	}

	return nil
}

// MediaDelete удаляет медиа из репозитория по ИД.
func (p *Provider) MediaDelete(ctx context.Context, id int) error {
	paylod, err := token.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", errMediaProvider, err)
	}

	err = p.store.MediaDelete(ctx, id, paylod.ID)
	if err != nil {
		return fmt.Errorf("%w: %w", errMediaProvider, err)
	}

	return nil
}

// Medias возвращает список медиа пользователя.
func (p *Provider) Medias(ctx context.Context) ([]models.Media, error) {
	paylod, err := token.FromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errMediaProvider, err)
	}

	res, err := p.store.Medias(ctx, paylod.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errMediaProvider, err)
	}

	if len(res) == 0 {
		return nil, fmt.Errorf("%w: %w", errMediaProvider, models.ErrNotFound)
	}

	for i := 0; i < len(res); i++ {
		decBody, err := p.crypto.DecryptString(string(res[i].Body))
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errMediaProvider, err)
		}

		res[i].Body = []byte(decBody)
	}

	return res, nil
}
