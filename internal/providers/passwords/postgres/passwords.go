// Package postgres предоставляет обертку над репозиторием для работы с паролями пользователя.
package postgres

import (
	"GophKeeper/internal/models"
	"GophKeeper/internal/repository"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// Storage описывает методы для создания, обновления, удаления и получения паролей пользователей.
type Storage struct {
	repo *repository.Repo
	log  *zap.Logger
}

// NewPasswordsStorage создаёт объект Storage.
func NewPasswordsStorage(r *repository.Repo, log *zap.Logger) *Storage {
	return &Storage{
		repo: r,
		log:  log.Named("password storage"),
	}
}

// PasswordCreate создаёт запись пароля в репозитории.
func (s *Storage) PasswordCreate(ctx context.Context, p models.Password, userID int) error {
	q := `INSERT INTO passwords (user_id, title, login, password, site, note, expiration_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	tx, err := s.repo.Pool.Begin(ctx)
	if err != nil {
		s.log.Error("start transaction for password create", zap.Error(err))

		return fmt.Errorf("start transaction for password create: %w", err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			s.log.Error("rollback transaction for password create", zap.Error(err))
		}
	}(tx, ctx)

	_, err = tx.Exec(ctx, q,
		userID, p.Title, p.Login, p.Password, p.URL, p.Note, p.ExpiredAt, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		s.log.Error("password create in repo", zap.Error(err))

		return fmt.Errorf("password create in repo: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error("commit transaction for password create", zap.Error(err))

		return fmt.Errorf("commit transaction for password create: %w", err)
	}

	return nil
}

// PasswordUpdate обновляет запись пароля в репозитории.
func (s *Storage) PasswordUpdate(ctx context.Context, p models.Password, userID int) error {
	q := `UPDATE passwords SET (title, login, password, site, note, expiration_at, updated_at) 
= ($1, $2, $3, $4, $5, $6, $7) WHERE id = $8 AND user_id = $9`

	tx, err := s.repo.Pool.Begin(ctx)
	if err != nil {
		s.log.Error("start transaction for password update", zap.Error(err))

		return fmt.Errorf("start transaction for password update: %w", err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			s.log.Error("rollback transaction for password update", zap.Error(err))
		}
	}(tx, ctx)

	tag, err := tx.Exec(ctx, q,
		p.Title,
		p.Login,
		p.Password,
		p.URL,
		p.Note,
		p.ExpiredAt,
		p.UpdatedAt,
		p.ID,
		userID)
	if err != nil {
		s.log.Error("password update in repo", zap.Error(err))

		return fmt.Errorf("update password in repo: %w", err)
	}

	if tag.RowsAffected() == 0 {
		s.log.Error("password update in repo: no rows affected")

		return fmt.Errorf("update password in repo: %w", models.ErrNotFound)
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error("commit transaction for password update", zap.Error(err))

		return fmt.Errorf("commit transaction for password update: %w", err)
	}

	return nil
}

// PasswordDelete удаляет запись пароля по ID.
func (s *Storage) PasswordDelete(ctx context.Context, id, userID int) error {
	q := `DELETE FROM passwords WHERE id = $1 AND user_id = $2`

	tx, err := s.repo.Pool.Begin(ctx)
	if err != nil {
		s.log.Error("start transaction for password delete", zap.Error(err))

		return fmt.Errorf("start transaction for password delete: %w", err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			s.log.Error("rollback transaction for password delete", zap.Error(err))
		}
	}(tx, ctx)

	tag, err := tx.Exec(ctx, q, id, userID)
	if err != nil {
		s.log.Error("password delete in repo", zap.Error(err))

		return fmt.Errorf("delete password from repo: %w", err)
	}

	if tag.RowsAffected() == 0 {
		s.log.Error("password delete in repo: no rows affected")

		return fmt.Errorf("delete password from repo: %w", models.ErrNotFound)
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error("commit transaction for password delete", zap.Error(err))

		return fmt.Errorf("commit transaction for password delete: %w", err)
	}

	return nil
}

// Passwords возвращает все пароли пользователя по его ID.
func (s *Storage) Passwords(ctx context.Context, userID int) ([]models.Password, error) {
	q := `SELECT id, title, login, password, site, note, expiration_at, created_at, updated_at
FROM passwords WHERE user_id = $1`

	tx, err := s.repo.Pool.Begin(ctx)
	if err != nil {
		s.log.Error("start transaction for get passwords", zap.Error(err))

		return nil, fmt.Errorf("start transaction for get passwords: %w", err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			s.log.Error("rollback transaction for get passwords", zap.Error(err))
		}
	}(tx, ctx)

	rows, err := tx.Query(ctx, q, userID)
	if err != nil {
		s.log.Error("get passwords form repo", zap.Error(err))

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
			&p.ExpiredAt,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			s.log.Error("get passwords form repo", zap.Error(err))

			return nil, fmt.Errorf("get passwords from repo: %w", err)
		}

		passwords = append(passwords, p)
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error("commit transaction for get passwords", zap.Error(err))

		return nil, fmt.Errorf("commit transaction for get passwords: %w", err)
	}

	return passwords, nil
}
