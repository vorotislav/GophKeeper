// Package postgres предоставляет обертку над репозиторием для работы с паролями пользователя.
package postgres

import (
	"GophKeeper/internal/models"
	"GophKeeper/internal/repository"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
)

// Storage описывает методы для создания, обновления, удаления и получения паролей пользователей.
type Storage struct {
	repo *repository.Repo
}

// NewPasswordsStorage создаёт объект Storage.
func NewPasswordsStorage(r *repository.Repo) *Storage {
	return &Storage{
		repo: r,
	}
}

// PasswordCreate создаёт запись пароля в репозитории.
func (s *Storage) PasswordCreate(ctx context.Context, p models.Password, userID int) error {
	q := `INSERT INTO passwords (user_id, title, login, password, site, note, expiration_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := s.repo.Pool.Exec(ctx, q,
		userID, p.Title, p.Login, p.Password, p.URL, p.Note, p.ExpirationDate, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("password create in repo: %w", err)
	}

	return nil
}

// PasswordUpdate обновляет запись пароля в репозитории.
func (s *Storage) PasswordUpdate(ctx context.Context, p models.Password, userID int) error {
	q := `UPDATE passwords SET (title, login, password, site, note, expiration_at, updated_at) 
= ($1, $2, $3, $4, $5, $6, $7) WHERE id = $8 AND user_id = $9`

	tag, err := s.repo.Pool.Exec(ctx, q,
		p.Title,
		p.Login,
		p.Password,
		p.URL,
		p.Note,
		p.ExpirationDate,
		p.UpdatedAt,
		p.ID,
		userID)
	if err != nil {
		return fmt.Errorf("update password in repo: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update password in repo: %w", models.ErrNotFound)
	}

	return nil
}

// PasswordDelete удаляет запись пароля по ID.
func (s *Storage) PasswordDelete(ctx context.Context, id, userID int) error {
	q := `DELETE FROM passwords WHERE id = $1 AND user_id = $2`

	tag, err := s.repo.Pool.Exec(ctx, q, id, userID)
	if err != nil {
		return fmt.Errorf("delete password from repo: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("delete password from repo: %w", models.ErrNotFound)
	}

	return nil
}

// Passwords возвращает все пароли пользователя по его ID.
func (s *Storage) Passwords(ctx context.Context, userID int) ([]models.Password, error) {
	q := `SELECT id, title, login, password, site, note, expiration_at, created_at, updated_at
FROM passwords WHERE user_id = $1`

	rows, err := s.repo.Pool.Query(ctx, q, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}

		return nil, fmt.Errorf("get passwords from repo: %w", err)
	}

	defer rows.Close()

	passwords := make([]models.Password, 0)

	for rows.Next() {
		p := models.Password{}

		err = rows.Scan(
			&p.ID,
			&p.Title,
			&p.Login,
			&p.Password,
			&p.URL,
			&p.Note,
			&p.ExpirationDate,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("get passwords from repo: %w", err)
		}

		passwords = append(passwords, p)
	}

	return passwords, nil
}
