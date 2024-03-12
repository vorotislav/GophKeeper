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

func createUser(t *testing.T, repo *repository.Repo) int {
	t.Helper()

	u := models.User{
		Login:    "login",
		Password: "pass",
	}

	q := "INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id"
	var id int

	err := repo.Pool.QueryRow(context.Background(), q, u.Login, u.Password).Scan(&id)
	require.NoError(t, err)

	return id
}

func TestPasswordsStore(t *testing.T) {
	repo := repository.TestRepository(t)
	require.NotNil(t, repo)

	ps := NewPasswordsStorage(repo)
	require.NotNil(t, ps)

	userID := createUser(t, repo)
	defer removeUser(t, repo, userID)

	p := models.Password{
		Title:          "title",
		Login:          "login",
		Password:       "pass",
		URL:            "url",
		Note:           "note",
		CreatedAt:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		UpdatedAt:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		ExpirationDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local),
	}

	err := ps.PasswordCreate(context.Background(), p, userID)
	require.NoError(t, err)

	createdPasswords, err := ps.Passwords(context.Background(), userID)
	require.NoError(t, err)
	require.NotNil(t, createdPasswords)
	require.Len(t, createdPasswords, 1)

	createdPassword := createdPasswords[0]
	assert.Equal(t, p.Title, createdPassword.Title)
	assert.Equal(t, p.Login, createdPassword.Login)
	assert.Equal(t, p.Password, createdPassword.Password)
	assert.Equal(t, p.URL, createdPassword.URL)
	assert.Equal(t, p.Note, createdPassword.Note)
	assert.Equal(t, p.CreatedAt, createdPassword.CreatedAt)
	assert.Equal(t, p.UpdatedAt, createdPassword.UpdatedAt)

	up := models.Password{
		ID:        createdPassword.ID,
		Title:     "updated title",
		Login:     "updated login",
		Password:  "updated pass",
		URL:       "updated url",
		Note:      "updated note",
		UpdatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.Local),
	}

	err = ps.PasswordUpdate(context.Background(), up, userID)
	require.NoError(t, err)

	updatedPasswords, err := ps.Passwords(context.Background(), userID)
	require.NoError(t, err)
	require.NotNil(t, updatedPasswords)
	require.Len(t, updatedPasswords, 1)

	updatedPassword := updatedPasswords[0]
	assert.Equal(t, createdPassword.ID, updatedPassword.ID)
	assert.Equal(t, up.Title, updatedPassword.Title)
	assert.Equal(t, up.Login, updatedPassword.Login)
	assert.Equal(t, up.Password, updatedPassword.Password)
	assert.Equal(t, up.URL, updatedPassword.URL)
	assert.Equal(t, up.Note, updatedPassword.Note)
	assert.Equal(t, createdPassword.CreatedAt, updatedPassword.CreatedAt)
	assert.Equal(t, up.UpdatedAt, updatedPassword.UpdatedAt)

	err = ps.PasswordDelete(context.Background(), updatedPassword.ID, userID)
	require.NoError(t, err)

	passwords, err := ps.Passwords(context.Background(), userID)
	require.NoError(t, err)
	require.Empty(t, passwords)

	err = ps.PasswordUpdate(context.Background(), up, userID)
	require.Error(t, err)
	require.ErrorIs(t, err, models.ErrNotFound)

	err = ps.PasswordDelete(context.Background(), updatedPassword.ID, userID)
	require.Error(t, err)
	require.ErrorIs(t, err, models.ErrNotFound)
}
