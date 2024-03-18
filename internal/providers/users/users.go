// Package users пакет для описания сущности по управлению пользователями.
package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"GophKeeper/internal/models"
	"GophKeeper/internal/providers/users/postgres"
	"GophKeeper/internal/repository"
	"GophKeeper/internal/token"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	errUserProvider = errors.New("user provider error")
)

// storage описывает доступные методы для работы с репозиторием и хранением пользователей и их сессий.
//
//go:generate go run github.com/vektra/mockery/v2@v2.24.0 --name=storage --exported --with-expecter=true
type storage interface {
	UserCreate(ctx context.Context, user models.User) (models.User, error)
	UserLogin(ctx context.Context, user models.User) (models.User, error)
	SessionCreate(ctx context.Context, session models.Session) (models.Session, error)
	CheckSessionFromClient(ctx context.Context, ipAddress string) ([]int64, error)
	RemoveSession(ctx context.Context, id int64) error
}

// authorizer описывает доступные методы для генерации и парсинга jwt токенов.
//
//go:generate go run github.com/vektra/mockery/v2@v2.24.0 --name=authorizer --exported --with-expecter=true
type authorizer interface {
	GetRefreshTokenDurationLifetime() time.Duration
	GenerateToken(payload token.Payload) (string, error)
	ParseToken(string) (token.Payload, error)
}

// Users структура для управления пользователями, хранит в себе хранилище и объект для работы в токенами.
type Users struct {
	log   *zap.Logger
	store storage
	auth  authorizer
}

// NewUsersProvider конструктор для Users.
func NewUsersProvider(log *zap.Logger, repo *repository.Repo, auth authorizer) *Users {
	u := &Users{
		log:   log,
		store: postgres.NewUserStorage(repo),
		auth:  auth,
	}

	return u
}

// UserCreate принимает на модель с описанием пользователя и его клиента, создаёт в хранилище пользователя и сразу создаёт сессию.
func (u *Users) UserCreate(ctx context.Context, um models.UserMachine) (models.Session, error) {
	user, err := u.store.UserCreate(ctx, um.User)
	if err != nil {
		return models.Session{}, fmt.Errorf("%w: %w", errUserProvider, err)
	}

	tokens, err := u.generateTokens(token.Payload{ID: user.ID})
	if err != nil {
		return models.Session{}, fmt.Errorf("%w: generator tokens: %w", errUserProvider, err)
	}

	session, err := u.store.SessionCreate(ctx, models.Session{
		UserID:                user.ID,
		AccessToken:           tokens.AccessToken,
		RefreshToken:          tokens.RefreshToken,
		IPAddress:             um.Machine.IPAddress,
		RefreshTokenExpiredAt: time.Now().Add(u.auth.GetRefreshTokenDurationLifetime()).Unix(),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	})

	if err != nil {
		return models.Session{}, fmt.Errorf("%w: session create: %w", errUserProvider, err)
	}

	return session, nil
}

// UserLogin проверяет что указанный пользователь существует, а так же его пароль, и после этого создаёт сессию.
func (u *Users) UserLogin(ctx context.Context, um models.UserMachine) (models.Session, error) {
	// проверить наличие пользователя и пароль
	user, err := u.store.UserLogin(ctx, um.User)
	if err != nil {
		return models.Session{}, fmt.Errorf("%w: %w", errUserProvider, err)
	}

	// если сессия с этого устройства уже есть - удалить
	ids, err := u.store.CheckSessionFromClient(ctx, um.Machine.IPAddress)
	if err != nil {
		if !errors.Is(err, models.ErrNotFound) {
			u.log.Warn("find sessions from device",
				zap.String("ip", um.Machine.IPAddress),
				zap.Error(err))
		}
	}

	if len(ids) > 0 {
		for _, id := range ids {
			err := u.store.RemoveSession(ctx, id)
			if err != nil {
				u.log.Warn("delete session from device",
					zap.Int64("session id", id),
					zap.String("ip", um.Machine.IPAddress),
					zap.Error(err))
			}
		}
	}

	// создать новую сессию
	tokens, err := u.generateTokens(token.Payload{ID: user.ID})
	if err != nil {
		return models.Session{}, fmt.Errorf("%w: generator tokens: %w", errUserProvider, err)
	}

	session, err := u.store.SessionCreate(ctx, models.Session{
		UserID:                user.ID,
		AccessToken:           tokens.AccessToken,
		RefreshToken:          tokens.RefreshToken,
		IPAddress:             um.Machine.IPAddress,
		RefreshTokenExpiredAt: time.Now().Add(u.auth.GetRefreshTokenDurationLifetime()).Unix(),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	})

	if err != nil {
		return models.Session{}, fmt.Errorf("%w: session create: %w", errUserProvider, err)
	}

	return session, nil
}

func (u *Users) generateTokens(payload token.Payload) (tokens, error) {
	ss, err := u.auth.GenerateToken(payload)
	if err != nil {
		return tokens{}, err //nolint:wrapcheck
	}

	refreshToken := uuid.New()

	return tokens{
		AccessToken:  ss,
		RefreshToken: refreshToken.String(),
	}, nil
}

type tokens struct {
	AccessToken  string
	RefreshToken string
}
