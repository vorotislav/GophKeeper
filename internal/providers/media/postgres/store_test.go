package postgres

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"GophKeeper/internal/models"
	"GophKeeper/internal/repository"

	"github.com/stretchr/testify/require"
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

func TestMediaStore(t *testing.T) {
	repo := repository.TestRepository(t)
	require.NotNil(t, repo)

	ms := NewMediaStorage(repo)
	require.NotNil(t, ms)

	userID := createUser(t, repo)
	defer removeUser(t, repo, userID)

	m := models.Media{
		Title:     "title",
		Body:      []byte(`some body`),
		MediaType: "type",
		Note:      "note",
		ExpiredAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
	}

	err := ms.MediaCreate(context.Background(), m, userID)
	require.NoError(t, err)

	createdMedias, err := ms.Medias(context.Background(), userID)
	require.NoError(t, err)
	require.NotNil(t, createdMedias)
	require.Len(t, createdMedias, 1)

	createdMedia := createdMedias[0]
	assert.Equal(t, m.Title, createdMedia.Title)
	assert.Equal(t, m.Body, createdMedia.Body)
	assert.Equal(t, m.MediaType, createdMedia.MediaType)
	assert.Equal(t, m.Note, createdMedia.Note)
	assert.Equal(t, m.CreatedAt, createdMedia.CreatedAt)
	assert.Equal(t, m.UpdatedAt, createdMedia.UpdatedAt)

	um := models.Media{
		ID:        createdMedia.ID,
		Title:     "updated title",
		Body:      []byte(`updated some body`),
		MediaType: "updated type",
		Note:      "updated note",
		UpdatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.Local),
	}

	err = ms.MediaUpdate(context.Background(), um, userID)
	require.NoError(t, err)

	updatedMedias, err := ms.Medias(context.Background(), userID)
	require.NoError(t, err)
	require.NotNil(t, updatedMedias)
	require.Len(t, updatedMedias, 1)

	updatedMedia := updatedMedias[0]
	assert.Equal(t, createdMedia.ID, updatedMedia.ID)
	assert.Equal(t, um.Title, updatedMedia.Title)
	assert.Equal(t, um.Body, updatedMedia.Body)
	assert.Equal(t, um.MediaType, updatedMedia.MediaType)
	assert.Equal(t, um.Note, updatedMedia.Note)
	assert.Equal(t, createdMedia.CreatedAt, updatedMedia.CreatedAt)
	assert.Equal(t, um.UpdatedAt, updatedMedia.UpdatedAt)

	err = ms.MediaDelete(context.Background(), updatedMedia.ID, userID)
	require.NoError(t, err)

	passwords, err := ms.Medias(context.Background(), userID)
	require.NoError(t, err)
	require.Empty(t, passwords)

	err = ms.MediaUpdate(context.Background(), um, userID)
	require.Error(t, err)
	require.ErrorIs(t, err, models.ErrNotFound)

	err = ms.MediaDelete(context.Background(), updatedMedia.ID, userID)
	require.Error(t, err)
	require.ErrorIs(t, err, models.ErrNotFound)
}
