// Package postgres предоставляет обертку над репозиторием для работы с пользователями и их сессиями.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"

	"GophKeeper/internal/crypto/hash"
	"GophKeeper/internal/models"
	"GophKeeper/internal/repository"
)

// Storage описывает методы для создания и логин пользователей, создание и удаление сессий и проверку на наличие сессий.
type Storage struct {
	repo *repository.Repo
}

// NewUserStorage создаёт объект Storage.
func NewUserStorage(r *repository.Repo) *Storage {
	return &Storage{repo: r}
}

// UserCreate создаёт пользователя в репозитории.
func (s *Storage) UserCreate(ctx context.Context, user models.User) (models.User, error) {
	qUser := `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id`

	hash, err := hash.Password(user.Password)
	if err != nil {
		return models.User{}, fmt.Errorf("hash password: %w", err)
	}

	var id int

	err = s.repo.Pool.QueryRow(ctx, qUser, user.Login, hash).Scan(&id)
	if err != nil {
		return models.User{}, fmt.Errorf("create user in repo: %w", err)
	}

	user.ID = id

	return user, nil
}

// SessionCreate создаёт сессию пользователя в репозитории.
func (s *Storage) SessionCreate(ctx context.Context, session models.Session) (models.Session, error) {
	q := `INSERT INTO sessions 
    (user_id, refresh_token, ip_address, refresh_token_expired_at, created_at, updated_at)
     VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	var id int64

	err := s.repo.Pool.QueryRow(ctx, q,
		session.UserID,
		session.RefreshToken,
		session.IPAddress,
		session.RefreshTokenExpiredAt,
		session.CreatedAt,
		session.UpdatedAt).Scan(&id)
	if err != nil {
		return models.Session{}, fmt.Errorf("create user session in repo: %w", err)
	}

	session.ID = id

	return session, nil
}

// UserLogin проверяет переданного пользователя в репозитории на наличие, а так же на совпадение хеша пароля.
func (s *Storage) UserLogin(ctx context.Context, user models.User) (models.User, error) {
	q := `SELECT id, password FROM users WHERE login=$1`
	var (
		id           int
		passwordHash string
	)

	err := s.repo.Pool.QueryRow(ctx, q, user.Login).Scan(&id, &passwordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, models.ErrNotFound
		}

		return models.User{}, fmt.Errorf("find user in repo: %w", err)
	}

	if err := hash.CheckPassword(user.Password, passwordHash); err != nil {
		return models.User{}, models.ErrInvalidPassword
	}

	user.ID = id

	return user, nil
}

// CheckSessionFromClient проверяет, есть ли уже сессия для переданного ip-адреса.
func (s *Storage) CheckSessionFromClient(ctx context.Context, ipAddress string) ([]int64, error) {
	q := `SELECT id FROM sessions WHERE ip_address = $1`

	rows, err := s.repo.Pool.Query(ctx, q, ipAddress)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}

		return nil, fmt.Errorf("sessions by ip and mac addresses: %w", err)
	}

	defer rows.Close()

	sessionsIDs := make([]int64, 0)

	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("sessions by ip and mac addresses: %w", err)
		}

		sessionsIDs = append(sessionsIDs, id)
	}

	return sessionsIDs, nil
}

// RemoveSession удаляет запись сессии из репозитория по её id.
func (s *Storage) RemoveSession(ctx context.Context, id int64) error {
	q := `DELETE FROM sessions WHERE id = $1`

	_, err := s.repo.Pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete session from repo: %w", err)
	}

	return nil
}
