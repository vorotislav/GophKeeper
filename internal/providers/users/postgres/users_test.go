package postgres

import (
	"GophKeeper/internal/models"
	"GophKeeper/internal/repository"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func removeUser(t *testing.T, repo *repository.Repo, userID int) {
	t.Helper()

	q := "DELETE FROM users WHERE id = $1"

	_, err := repo.Pool.Exec(context.Background(), q, userID)
	require.NoError(t, err)
}

func TestUserStore(t *testing.T) {
	repo := repository.TestRepository(t)

	us := NewUserStorage(repo)
	require.NotNil(t, us)

	u := models.User{
		Login:    "login",
		Password: "pass",
	}

	gotUserCreate, err := us.UserCreate(context.Background(), u)
	require.NoError(t, err)
	assert.Equal(t, u.Login, gotUserCreate.Login)
	assert.Equal(t, u.Password, gotUserCreate.Password)
	require.Greater(t, gotUserCreate.ID, 0)

	defer removeUser(t, repo, gotUserCreate.ID)

	_, err = us.UserCreate(context.Background(), u)
	require.Error(t, err)

	gotUserLogin, err := us.UserLogin(context.Background(), u)
	require.NoError(t, err)
	assert.Equal(t, gotUserCreate, gotUserLogin)

	s := models.Session{
		UserID:       gotUserCreate.ID,
		AccessToken:  "access",
		RefreshToken: "4c230485-d287-40bd-a888-7d5c7d46eed4",
		IPAddress:    "127.0.0.1",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	gotSessionCreate, err := us.SessionCreate(context.Background(), s)
	require.NoError(t, err)
	require.Greater(t, gotSessionCreate.ID, int64(0))
	assert.Equal(t, s.UserID, gotSessionCreate.UserID)
	assert.Equal(t, s.AccessToken, gotSessionCreate.AccessToken)
	assert.Equal(t, s.RefreshToken, gotSessionCreate.RefreshToken)
	assert.Equal(t, s.IPAddress, gotSessionCreate.IPAddress)
	assert.Equal(t, s.CreatedAt, gotSessionCreate.CreatedAt)
	assert.Equal(t, s.UpdatedAt, gotSessionCreate.UpdatedAt)

	sessionsIDs, err := us.CheckSessionFromClient(context.Background(), "127.0.0.1")
	require.NoError(t, err)
	assert.Equal(t, gotSessionCreate.ID, sessionsIDs[0])

	sessionsEmpty, err := us.CheckSessionFromClient(context.Background(), "128.9.9.2")
	require.NoError(t, err)
	assert.Empty(t, sessionsEmpty)

	err = us.RemoveSession(context.Background(), gotSessionCreate.ID)
	require.NoError(t, err)
}
