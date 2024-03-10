package notes

import (
	"context"
	"errors"
	"fmt"
	"time"

	"GophKeeper/internal/models"
	"GophKeeper/internal/providers/notes/postgres"
	"GophKeeper/internal/repository"
	"GophKeeper/internal/token"

	"go.uber.org/zap"
)

var (
	errNoteProvider = errors.New("notes provider error")
)

type crypto interface {
	EncryptString(value string) (string, error)
	DecryptString(value string) (string, error)
}

type storage interface {
	NoteCreate(ctx context.Context, n models.Note, userID int) error
	NoteUpdate(ctx context.Context, n models.Note, userID int) error
	NoteDelete(ctx context.Context, id, userID int) error
	Notes(ctx context.Context, userID int) ([]models.Note, error)
}

type Provider struct {
	log    *zap.Logger
	store  storage
	crypto crypto
}

func NewProvider(log *zap.Logger, repo *repository.Repo, crypto crypto) *Provider {
	return &Provider{
		log:    log.Named("notes provider"),
		store:  postgres.NewNotesStorage(repo),
		crypto: crypto,
	}
}

func (p *Provider) NoteCreate(ctx context.Context, n models.Note) error {
	paylod, err := token.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", errNoteProvider, err)
	}

	encNoteText, err := p.crypto.EncryptString(n.Text)
	if err != nil {
		return fmt.Errorf("%w: %w", errNoteProvider, err)
	}

	n.Text = encNoteText
	n.CreatedAt = time.Now()
	n.UpdatedAt = time.Now()

	err = p.store.NoteCreate(ctx, n, paylod.ID)
	if err != nil {
		return fmt.Errorf("%w: %w", errNoteProvider, err)
	}

	return nil
}

func (p *Provider) NoteUpdate(ctx context.Context, n models.Note) error {
	paylod, err := token.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", errNoteProvider, err)
	}

	encNoteText, err := p.crypto.EncryptString(n.Text)
	if err != nil {
		return fmt.Errorf("%w: %w", errNoteProvider, err)
	}

	n.Text = encNoteText
	n.UpdatedAt = time.Now()

	err = p.store.NoteUpdate(ctx, n, paylod.ID)
	if err != nil {
		return fmt.Errorf("%w: %w", errNoteProvider, err)
	}

	return nil
}

func (p *Provider) NoteDelete(ctx context.Context, id int) error {
	paylod, err := token.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", errNoteProvider, err)
	}

	err = p.store.NoteDelete(ctx, id, paylod.ID)
	if err != nil {
		return fmt.Errorf("%w: %w", errNoteProvider, err)
	}

	return nil
}

func (p *Provider) Notes(ctx context.Context) ([]models.Note, error) {
	paylod, err := token.FromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errNoteProvider, err)
	}

	res, err := p.store.Notes(ctx, paylod.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errNoteProvider, err)
	}

	if len(res) == 0 {
		return nil, fmt.Errorf("%w: %w", errNoteProvider, models.ErrNotFound)
	}

	for i := 0; i < len(res); i++ {
		decNote, err := p.crypto.DecryptString(res[i].Text)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errNoteProvider, err)
		}

		res[i].Text = decNote
	}

	return res, nil
}
