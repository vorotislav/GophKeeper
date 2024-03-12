package postgres

import (
	"context"
	"testing"
	"time"

	"GophKeeper/internal/models"
	"GophKeeper/internal/repository"

	"github.com/stretchr/testify/assert"
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

func TestNoteStore(t *testing.T) {
	repo := repository.TestRepository(t)
	require.NotNil(t, repo)

	ns := NewNotesStorage(repo)
	require.NotNil(t, ns)

	userID := createUser(t, repo)
	defer removeUser(t, repo, userID)

	n := models.Note{
		Title:     "title",
		Text:      "text",
		ExpiredAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
	}

	err := ns.NoteCreate(context.Background(), n, userID)
	require.NoError(t, err)

	createdNotes, err := ns.Notes(context.Background(), userID)
	require.NoError(t, err)
	require.NotNil(t, createdNotes)
	require.Len(t, createdNotes, 1)

	createdNote := createdNotes[0]
	assert.Equal(t, n.Title, createdNote.Title)
	assert.Equal(t, n.Text, createdNote.Text)
	assert.Equal(t, n.CreatedAt, createdNote.CreatedAt)
	assert.Equal(t, n.UpdatedAt, createdNote.UpdatedAt)

	un := models.Note{
		ID:        createdNote.ID,
		Title:     "updated title",
		Text:      "updated text",
		UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
	}

	err = ns.NoteUpdate(context.Background(), un, userID)
	require.NoError(t, err)

	updatedNotes, err := ns.Notes(context.Background(), userID)
	require.NoError(t, err)
	require.NotNil(t, updatedNotes)
	require.Len(t, updatedNotes, 1)

	updatedNote := updatedNotes[0]
	assert.Equal(t, createdNote.ID, updatedNote.ID)
	assert.Equal(t, un.Title, updatedNote.Title)
	assert.Equal(t, un.Text, updatedNote.Text)
	assert.Equal(t, un.UpdatedAt, updatedNote.UpdatedAt)

	err = ns.NoteDelete(context.Background(), updatedNote.ID, userID)
	require.NoError(t, err)

	notes, err := ns.Notes(context.Background(), userID)
	require.NoError(t, err)
	require.Empty(t, notes)

	err = ns.NoteUpdate(context.Background(), un, userID)
	require.Error(t, err)
	require.ErrorIs(t, err, models.ErrNotFound)

	err = ns.NoteDelete(context.Background(), updatedNote.ID, userID)
	require.Error(t, err)
	require.ErrorIs(t, err, models.ErrNotFound)
}
