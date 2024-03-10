package passwords

import (
	"context"
	"errors"
	"fmt"
	"time"

	"GophKeeper/internal/models"
	"GophKeeper/internal/providers/passwords/postgres"
	"GophKeeper/internal/repository"
	"GophKeeper/internal/token"

	"go.uber.org/zap"
)

var (
	errPasswordProvider = errors.New("passwords provider error")
)

type crypto interface {
	EncryptString(value string) (string, error)
	DecryptString(value string) (string, error)
}

type storage interface {
	PasswordCreate(ctx context.Context, pass models.Password, userID int) error
	PasswordUpdate(ctx context.Context, pass models.Password, userID int) error
	PasswordDelete(ctx context.Context, id int, userID int) error
	Passwords(ctx context.Context, userID int) ([]models.Password, error)
}

type Provider struct {
	log    *zap.Logger
	store  storage
	crypto crypto
}

func NewProvider(log *zap.Logger, repo *repository.Repo, crypto crypto) *Provider {
	return &Provider{
		log:    log.Named("password provider"),
		store:  postgres.NewPasswordsStorage(repo),
		crypto: crypto,
	}
}

func (p *Provider) PasswordCreate(ctx context.Context, pass models.Password) error {
	paylod, err := token.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", errPasswordProvider, err)
	}

	encPass, err := p.crypto.EncryptString(pass.Password)
	if err != nil {
		return fmt.Errorf("%w: %w", errPasswordProvider, err)
	}

	pass.Password = encPass
	pass.CreatedAt = time.Now()
	pass.UpdatedAt = time.Now()

	err = p.store.PasswordCreate(ctx, pass, paylod.ID)
	if err != nil {
		return fmt.Errorf("%w: %w", errPasswordProvider, err)
	}

	return nil
}

func (p *Provider) PasswordUpdate(ctx context.Context, pass models.Password) error {
	paylod, err := token.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", errPasswordProvider, err)
	}

	encPass, err := p.crypto.EncryptString(pass.Password)
	if err != nil {
		return fmt.Errorf("%w: %w", errPasswordProvider, err)
	}

	pass.Password = encPass
	pass.UpdatedAt = time.Now()

	err = p.store.PasswordUpdate(ctx, pass, paylod.ID)
	if err != nil {
		return fmt.Errorf("%w: %w", errPasswordProvider, err)
	}

	return nil
}

func (p *Provider) PasswordDelete(ctx context.Context, id int) error {
	paylod, err := token.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", errPasswordProvider, err)
	}

	err = p.store.PasswordDelete(ctx, id, paylod.ID)
	if err != nil {
		return fmt.Errorf("%w: %w", errPasswordProvider, err)
	}

	return nil
}

func (p *Provider) Passwords(ctx context.Context) ([]models.Password, error) {
	paylod, err := token.FromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errPasswordProvider, err)
	}

	res, err := p.store.Passwords(ctx, paylod.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errPasswordProvider, err)
	}

	if len(res) == 0 {
		return nil, fmt.Errorf("%w: %w", errPasswordProvider, models.ErrNotFound)
	}

	for i := 0; i < len(res); i++ {
		decPass, err := p.crypto.DecryptString(res[i].Password)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errPasswordProvider, err)
		}

		res[i].Password = decPass
	}

	return res, nil
}
