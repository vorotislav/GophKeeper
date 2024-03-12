// Package cards пакет для описания сущности по управлению банковскими картами пользователя.
package cards

import (
	"context"
	"errors"
	"fmt"
	"time"

	"GophKeeper/internal/models"
	"GophKeeper/internal/providers/cards/postgres"
	"GophKeeper/internal/repository"
	"GophKeeper/internal/token"

	"go.uber.org/zap"
)

var (
	errCardProvider = errors.New("passwords provider error")
)

// crypto описывает доступные методы для шифрования и дешифрования данных карт пользователя.
//
//go:generate go run github.com/vektra/mockery/v2@v2.24.0 --name=crypto --exported --with-expecter=true
type crypto interface {
	EncryptString(value string) (string, error)
	DecryptString(value string) (string, error)
}

// storage описывает доступные методы для работы с репозиторием и хранением данных карт пользователей.
//
//go:generate go run github.com/vektra/mockery/v2@v2.24.0 --name=storage --exported --with-expecter=true
type storage interface {
	CardCreate(ctx context.Context, c models.Card, userID int) error
	CardUpdate(ctx context.Context, c models.Card, userID int) error
	CardDelete(ctx context.Context, id int, userID int) error
	Cards(ctx context.Context, userID int) ([]models.Card, error)
}

// Provider структура для управления картами, хранит в себе хранилище и объект для шифрования\дешифрования.
type Provider struct {
	log    *zap.Logger
	store  storage
	crypto crypto
}

// NewProvider конструктор для Provider.
func NewProvider(log *zap.Logger, repo *repository.Repo, crypto crypto) *Provider {
	return &Provider{
		log:    log.Named("cards provider"),
		store:  postgres.NewCardsStorage(repo),
		crypto: crypto,
	}
}

// CardCreate принимает объект модели карты и сохраняет в репозиторий.
func (p *Provider) CardCreate(ctx context.Context, c models.Card) error {
	paylod, err := token.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", errCardProvider, err)
	}

	encNumber, err := p.crypto.EncryptString(c.Number)
	if err != nil {
		return fmt.Errorf("%w: %w", errCardProvider, err)
	}

	encCVV, err := p.crypto.EncryptString(c.CVC)
	if err != nil {
		return fmt.Errorf("%w: %w", errCardProvider, err)
	}

	c.Number = encNumber
	c.CVC = encCVV
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()

	err = p.store.CardCreate(ctx, c, paylod.ID)
	if err != nil {
		return fmt.Errorf("%w: %w", errCardProvider, err)
	}

	return nil
}

// CardUpdate принимает объект модель карты и обновляет в репозитории.
func (p *Provider) CardUpdate(ctx context.Context, c models.Card) error {
	paylod, err := token.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", errCardProvider, err)
	}

	encNumber, err := p.crypto.EncryptString(c.Number)
	if err != nil {
		return fmt.Errorf("%w: %w", errCardProvider, err)
	}

	encCVC, err := p.crypto.EncryptString(c.CVC)
	if err != nil {
		return fmt.Errorf("%w: %w", errCardProvider, err)
	}

	c.Number = encNumber
	c.CVC = encCVC
	c.UpdatedAt = time.Now()

	err = p.store.CardUpdate(ctx, c, paylod.ID)
	if err != nil {
		return fmt.Errorf("%w: %w", errCardProvider, err)
	}

	return nil
}

// CardDelete удаляет карту из репозитория по ИД.
func (p *Provider) CardDelete(ctx context.Context, id int) error {
	paylod, err := token.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", errCardProvider, err)
	}

	err = p.store.CardDelete(ctx, id, paylod.ID)
	if err != nil {
		return fmt.Errorf("%w: %w", errCardProvider, err)
	}

	return nil
}

// Cards возвращает список карт пользователя.
func (p *Provider) Cards(ctx context.Context) ([]models.Card, error) {
	paylod, err := token.FromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errCardProvider, err)
	}

	res, err := p.store.Cards(ctx, paylod.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errCardProvider, err)
	}

	if len(res) == 0 {
		return nil, fmt.Errorf("%w: %w", errCardProvider, models.ErrNotFound)
	}

	for i := 0; i < len(res); i++ {
		decNumber, err := p.crypto.DecryptString(res[i].Number)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errCardProvider, err)
		}

		decCVC, err := p.crypto.DecryptString(res[i].CVC)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errCardProvider, err)
		}

		res[i].Number = decNumber
		res[i].CVC = decCVC
	}

	return res, nil
}
