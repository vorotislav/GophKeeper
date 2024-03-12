// Package postgres предоставляет обертку над репозиторием для работы с медиа пользователя.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"GophKeeper/internal/models"
	"GophKeeper/internal/repository"

	"github.com/jackc/pgx/v5"
)

// Storage описывает методы для создания, обновления, удаления и получения медиа пользователей.
type Storage struct {
	repo *repository.Repo
}

// NewMediaStorage создаёт объект Storage.
func NewMediaStorage(r *repository.Repo) *Storage {
	return &Storage{repo: r}
}

// MediaCreate создаёт запись медиа в репозитории.
func (s *Storage) MediaCreate(ctx context.Context, m models.Media, userID int) error {
	q := `INSERT INTO media(user_id, title, body, media_type, note, expiration_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := s.repo.Pool.Exec(ctx, q, userID,
		m.Title, m.Body, m.MediaType, m.Note, m.ExpiredAt, m.CreatedAt, m.UpdatedAt)
	if err != nil {
		return fmt.Errorf("media create in repo: %w", err)
	}

	return nil
}

// MediaUpdate обновляет запись медиа в репозитории.
func (s *Storage) MediaUpdate(ctx context.Context, m models.Media, userID int) error {
	q := `UPDATE media SET(title, body, media_type, note, expiration_at, updated_at)
= ($1, $2, $3, $4, $5, $6) WHERE id = $7 AND user_id = $8`

	tag, err := s.repo.Pool.Exec(ctx, q,
		m.Title, m.Body, m.MediaType, m.Note, m.ExpiredAt, m.UpdatedAt, m.ID, userID)
	if err != nil {
		return fmt.Errorf("media update in repo: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("media update in repo: %w", models.ErrNotFound)
	}

	return nil
}

// MediaDelete удаляет запись медиа по ID.
func (s *Storage) MediaDelete(ctx context.Context, id, userID int) error {
	q := `DELETE FROM media WHERE id = $1 AND user_id = $2`

	tag, err := s.repo.Pool.Exec(ctx, q, id, userID)
	if err != nil {
		return fmt.Errorf("delete media from repo: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("delete media from repo: %w", models.ErrNotFound)
	}

	return nil
}

// Medias возвращает все медиа пользователя по его ID.
func (s *Storage) Medias(ctx context.Context, userID int) ([]models.Media, error) {
	q := `SELECT id, title, body, media_type, note, expiration_at, created_at, updated_at 
FROM media WHERE user_id = $1`

	rows, err := s.repo.Pool.Query(ctx, q, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}

		return nil, fmt.Errorf("get media from repo: %w", err)
	}

	defer rows.Close()

	medias := make([]models.Media, 0)

	for rows.Next() {
		m := models.Media{}

		err = rows.Scan(
			&m.ID,
			&m.Title,
			&m.Body,
			&m.MediaType,
			&m.Note,
			&m.ExpiredAt,
			&m.CreatedAt,
			&m.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("get media from repo: %w", err)
		}

		medias = append(medias, m)
	}

	return medias, nil
}
