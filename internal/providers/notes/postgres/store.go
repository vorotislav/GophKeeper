// Package postgres предоставляет обертку над репозиторием для работы с заметками пользователя.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"

	"GophKeeper/internal/models"
	"GophKeeper/internal/repository"

	"github.com/jackc/pgx/v5"
)

// Storage описывает методы для создания, обновления, удаления и получения заметок пользователей.
type Storage struct {
	repo *repository.Repo
	log  *zap.Logger
}

// NewNotesStorage создаёт объект Storage.
func NewNotesStorage(r *repository.Repo, log *zap.Logger) *Storage {
	return &Storage{
		repo: r,
		log:  log.Named("note storage"),
	}
}

// NoteCreate создаёт запись замтеки в репозитории.
func (s *Storage) NoteCreate(ctx context.Context, n models.Note, userID int) error {
	q := `INSERT INTO notes(user_id, title, body, expiration_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)`

	tx, err := s.repo.Pool.Begin(ctx)
	if err != nil {
		s.log.Error("start transaction for note create", zap.Error(err))

		return fmt.Errorf("start transaction for note create: %w", err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			s.log.Error("rollback transaction for note create", zap.Error(err))
		}
	}(tx, ctx)

	_, err = tx.Exec(ctx, q,
		userID, n.Title, n.Text, n.ExpiredAt, n.CreatedAt, n.UpdatedAt)
	if err != nil {
		s.log.Error("note create in repo", zap.Error(err))

		return fmt.Errorf("note create in repo: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error("commit transaction for note create", zap.Error(err))

		return fmt.Errorf("commit transaction for note create: %w", err)
	}

	return nil
}

// NoteUpdate обновляет запись заметки в репозитории.
func (s *Storage) NoteUpdate(ctx context.Context, n models.Note, userID int) error {
	q := `UPDATE notes SET(title, body, expiration_at, updated_at)
= ($1, $2, $3, $4) WHERE id = $5 AND user_id = $6`

	tx, err := s.repo.Pool.Begin(ctx)
	if err != nil {
		s.log.Error("start transaction for note update", zap.Error(err))

		return fmt.Errorf("start transaction for note update: %w", err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			s.log.Error("rollback transaction for note update", zap.Error(err))
		}
	}(tx, ctx)

	tag, err := tx.Exec(ctx, q,
		n.Title, n.Text, n.ExpiredAt, n.UpdatedAt, n.ID, userID)
	if err != nil {
		s.log.Error("note update in repo", zap.Error(err))

		return fmt.Errorf("note update in repo: %w", err)
	}

	if tag.RowsAffected() == 0 {
		s.log.Error("note update in repo: no rows affected")

		return fmt.Errorf("note update in repo: %w", models.ErrNotFound)
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error("commit transaction for note update", zap.Error(err))

		return fmt.Errorf("commit transaction for note update: %w", err)
	}

	return nil
}

// NoteDelete удаляет запись заметки по ID.
func (s *Storage) NoteDelete(ctx context.Context, id, userID int) error {
	q := `DELETE FROM notes WHERE id = $1 AND user_id = $2`

	tx, err := s.repo.Pool.Begin(ctx)
	if err != nil {
		s.log.Error("start transaction for note delete", zap.Error(err))

		return fmt.Errorf("start transaction for note delete: %w", err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			s.log.Error("rollback transaction for note delete", zap.Error(err))
		}
	}(tx, ctx)

	tag, err := tx.Exec(ctx, q, id, userID)
	if err != nil {
		s.log.Error("note delete in repo", zap.Error(err))

		return fmt.Errorf("note delete in repo: %w", err)
	}

	if tag.RowsAffected() == 0 {
		s.log.Error("note delete in repo: no rows affected")

		return fmt.Errorf("note delete in repo: %w", models.ErrNotFound)
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error("commit transaction for note delete", zap.Error(err))

		return fmt.Errorf("commit transaction for note delete: %w", err)
	}

	return nil
}

// Notes возвращает все заметки пользователя по его ID.
func (s *Storage) Notes(ctx context.Context, userID int) ([]models.Note, error) {
	q := `SELECT id, title, body, expiration_at, created_at, updated_at FROM notes WHERE user_id = $1`

	tx, err := s.repo.Pool.Begin(ctx)
	if err != nil {
		s.log.Error("start transaction for get notes", zap.Error(err))

		return nil, fmt.Errorf("start transaction for get notes: %w", err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			s.log.Error("rollback transaction for get notes", zap.Error(err))
		}
	}(tx, ctx)

	rows, err := tx.Query(ctx, q, userID)
	if err != nil {
		s.log.Error("get notes form repo", zap.Error(err))

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}

		return nil, fmt.Errorf("get notes form repo: %w", err)
	}

	defer rows.Close()

	notes := make([]models.Note, 0)

	for rows.Next() {
		n := models.Note{}

		err = rows.Scan(
			&n.ID,
			&n.Title,
			&n.Text,
			&n.ExpiredAt,
			&n.CreatedAt,
			&n.UpdatedAt)
		if err != nil {
			s.log.Error("get notes form repo", zap.Error(err))

			return nil, fmt.Errorf("get notes form repo: %w", err)
		}

		notes = append(notes, n)
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error("commit transaction for get notes", zap.Error(err))

		return nil, fmt.Errorf("commit transaction for get notes: %w", err)
	}

	return notes, nil
}
