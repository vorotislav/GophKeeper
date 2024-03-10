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

type crypto interface {
	EncryptString(value string) (string, error)
	DecryptString(value string) (string, error)
}

type storage interface {
	CardCreate(ctx context.Context, c models.Card, userID int) error
	CardUpdate(ctx context.Context, c models.Card, userID int) error
	CardDelete(ctx context.Context, id int, userID int) error
	Cards(ctx context.Context, userID int) ([]models.Card, error)
}

type Provider struct {
	log    *zap.Logger
	store  storage
	crypto crypto
}

func NewProvider(log *zap.Logger, repo *repository.Repo, crypto crypto) *Provider {
	return &Provider{
		log:    log.Named("cards provider"),
		store:  postgres.NewCardsStorage(repo),
		crypto: crypto,
	}
}

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
