// Package notes пакет для описания сущности по управлению заметками пользователя.
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

// crypto описывает доступные методы для шифрования и дешифрования заметок пользователя.
//
//go:generate go run github.com/vektra/mockery/v2@v2.24.0 --name=crypto --exported --with-expecter=true
type crypto interface {
	EncryptString(value string) (string, error)
	DecryptString(value string) (string, error)
}

// storage описывает доступные методы для работы с репозиторием и хранением заметок пользователей.
//
//go:generate go run github.com/vektra/mockery/v2@v2.24.0 --name=storage --exported --with-expecter=true
type storage interface {
	NoteCreate(ctx context.Context, n models.Note, userID int) error
	NoteUpdate(ctx context.Context, n models.Note, userID int) error
	NoteDelete(ctx context.Context, id, userID int) error
	Notes(ctx context.Context, userID int) ([]models.Note, error)
}

// Provider структура для управления заметками, хранит в себе хранилище и объект для шифрования\дешифрования.
type Provider struct {
	log    *zap.Logger
	store  storage
	crypto crypto
}

// NewProvider конструктор для Provider.
func NewProvider(log *zap.Logger, repo *repository.Repo, crypto crypto) *Provider {
	return &Provider{
		log:    log.Named("notes provider"),
		store:  postgres.NewNotesStorage(repo),
		crypto: crypto,
	}
}

// NoteCreate принимает объект модели заметки и сохраняет в репозиторий.
func (p *Provider) NoteCreate(ctx context.Context, n models.Note) error {
	payload, err := token.FromContext(ctx)
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

	err = p.store.NoteCreate(ctx, n, payload.ID)
	if err != nil {
		return fmt.Errorf("%w: %w", errNoteProvider, err)
	}

	return nil
}

// NoteUpdate принимает объект модель заметки и обновляет в репозитории.
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

// NoteDelete удаляет заметку из репозитория по ИД.
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

// Notes возвращает список заметок пользователя.
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
