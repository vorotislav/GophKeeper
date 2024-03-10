package postgres

import (
	"context"
	"errors"
	"fmt"

	"GophKeeper/internal/models"
	"GophKeeper/internal/repository"

	"github.com/jackc/pgx/v5"
)

type Storage struct {
	repo *repository.Repo
}

func NewNotesStorage(r *repository.Repo) *Storage {
	return &Storage{
		repo: r,
	}
}

func (s *Storage) NoteCreate(ctx context.Context, n models.Note, userID int) error {
	q := `INSERT INTO notes(user_id, title, body, expiration_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := s.repo.Pool.Exec(ctx, q,
		userID, n.Title, n.Text, n.ExpiredAt, n.CreatedAt, n.UpdatedAt)
	if err != nil {
		return fmt.Errorf("note create in repo: %w", err)
	}

	return nil
}

func (s *Storage) NoteUpdate(ctx context.Context, n models.Note, userID int) error {
	q := `UPDATE notes SET(title, body, expiration_at, updated_at)
= ($1, $2, $3, $4) WHERE id = $5 AND user_id = $6`

	tag, err := s.repo.Pool.Exec(ctx, q,
		n.Title, n.Text, n.ExpiredAt, n.UpdatedAt, n.ID, userID)
	if err != nil {
		return fmt.Errorf("note update in repo: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("note update in repo: %w", models.ErrNotFound)
	}

	return nil
}

func (s *Storage) NoteDelete(ctx context.Context, id, userID int) error {
	q := `DELETE FROM notes WHERE id = $1 AND user_id = $2`

	tag, err := s.repo.Pool.Exec(ctx, q, id, userID)
	if err != nil {
		return fmt.Errorf("note delete in repo: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("note delete in repo: %w", models.ErrNotFound)
	}

	return nil
}

func (s *Storage) Notes(ctx context.Context, userID int) ([]models.Note, error) {
	q := `SELECT id, title, body, expiration_at, created_at, updated_at FROM notes WHERE user_id = $1`

	rows, err := s.repo.Pool.Query(ctx, q, userID)
	if err != nil {
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
			return nil, fmt.Errorf("get notes form repo: %w", err)
		}

		notes = append(notes, n)
	}

	return notes, nil
}

/*
CREATE TABLE notes (
    id BIGINT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY, -- уникальный идентификатор заметки
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- идентификатор пользователя
    title TEXT NOT NULL, -- заголовок заметки
    body TEXT, -- тело заметки
	expiration_at TIMESTAMP WITH TIME ZONE, -- дата истечения срока действия заметки
    created_at TIMESTAMP WITH TIME ZONE, -- отметка времени создания заметки
    updated_at TIMESTAMP WITH TIME ZONE -- отметка времени обновления заметки
)*/
