// Package postgres предоставляет обертку над репозиторием для работы с картами пользователя.
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

// Storage описывает методы для создания, обновления, удаления и получения карт пользователей.
type Storage struct {
	repo *repository.Repo
	log  *zap.Logger
}

// NewCardsStorage создаёт объект Storage.
func NewCardsStorage(r *repository.Repo, log *zap.Logger) *Storage {
	return &Storage{
		repo: r,
		log:  log.Named("card storage"),
	}
}

// CardCreate создаёт запись карты в репозитории.
func (s *Storage) CardCreate(ctx context.Context, c models.Card, userID int) error {
	q := `INSERT INTO cards (user_id, name, number, cvc, exp_month, exp_year, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	tx, err := s.repo.Pool.Begin(ctx)
	if err != nil {
		s.log.Error("start transaction for card create", zap.Error(err))

		return fmt.Errorf("start transaction for card create: %w", err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			s.log.Error("rollback transaction for card create", zap.Error(err))
		}
	}(tx, ctx)

	_, err = tx.Exec(ctx, q,
		userID, c.Name, c.Number, c.CVC, c.ExpMonth, c.ExpYear, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		s.log.Error("card insert to repo", zap.Error(err))

		return fmt.Errorf("card create in repo: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error("commit transaction for card create", zap.Error(err))

		return fmt.Errorf("commit transaction for card create: %w", err)
	}

	return nil
}

// CardUpdate обновляет запись карты в репозитории.
func (s *Storage) CardUpdate(ctx context.Context, c models.Card, userID int) error {
	q := `UPDATE cards SET(name, number, cvc, exp_month, exp_year, updated_at) 
= ($1, $2, $3, $4, $5, $6) WHERE id = $7 AND user_id = $8`

	tx, err := s.repo.Pool.Begin(ctx)
	if err != nil {
		s.log.Error("start transaction for card update", zap.Error(err))

		return fmt.Errorf("start transaction for card update: %w", err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			s.log.Error("rollback transaction for card update", zap.Error(err))
		}
	}(tx, ctx)

	tag, err := tx.Exec(ctx, q,
		c.Name, c.Number, c.CVC, c.ExpMonth, c.ExpYear, c.UpdatedAt, c.ID, userID)
	if err != nil {
		s.log.Error("card update in repo", zap.Error(err))

		return fmt.Errorf("cart update in repo: %w", err)
	}

	if tag.RowsAffected() == 0 {
		s.log.Error("card update in repo: no rows affected")

		return fmt.Errorf("update card from repo: %w", models.ErrNotFound)
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error("commit transaction for card update", zap.Error(err))

		return fmt.Errorf("commit transaction for card update: %w", err)
	}

	return nil
}

// CardDelete удаляет запись карты по ID.
func (s *Storage) CardDelete(ctx context.Context, id, userID int) error {
	q := `DELETE FROM cards WHERE id = $1 AND user_id = $2`

	tx, err := s.repo.Pool.Begin(ctx)
	if err != nil {
		s.log.Error("start transaction for card delete", zap.Error(err))

		return fmt.Errorf("start transaction for card delete: %w", err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			s.log.Error("rollback transaction for card delete", zap.Error(err))
		}
	}(tx, ctx)

	tag, err := tx.Exec(ctx, q, id, userID)
	if err != nil {
		s.log.Error("card delete from repo", zap.Error(err))

		return fmt.Errorf("delete card from repo: %w", err)
	}

	if tag.RowsAffected() == 0 {
		s.log.Error("card delete from repo: no rows affected")

		return fmt.Errorf("delete card from repo: %w", models.ErrNotFound)
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error("commit transaction for card delete", zap.Error(err))

		return fmt.Errorf("commit transaction for card delete: %w", err)
	}

	return nil
}

// Cards возвращает все карты пользователя по его ID.
func (s *Storage) Cards(ctx context.Context, userID int) ([]models.Card, error) {
	q := `SELECT id, name, number, cvc, exp_month, exp_year, created_at, updated_at
FROM cards WHERE user_id = $1`

	tx, err := s.repo.Pool.Begin(ctx)
	if err != nil {
		s.log.Error("start transaction for get cards", zap.Error(err))

		return nil, fmt.Errorf("start transaction for get cardse: %w", err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			s.log.Error("rollback transaction for get cards", zap.Error(err))
		}
	}(tx, ctx)

	rows, err := tx.Query(ctx, q, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}

		return nil, fmt.Errorf("get cards from repo: %w", err)
	}

	defer rows.Close()

	cards := make([]models.Card, 0)

	for rows.Next() {
		c := models.Card{}

		err = rows.Scan(
			&c.ID,
			&c.Name,
			&c.Number,
			&c.CVC,
			&c.ExpMonth,
			&c.ExpYear,
			&c.CreatedAt,
			&c.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("get cards from repo: %w", err)
		}

		cards = append(cards, c)
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error("commit transaction for get cards", zap.Error(err))

		return nil, fmt.Errorf("commit transaction get cards: %w", err)
	}

	return cards, nil
}
