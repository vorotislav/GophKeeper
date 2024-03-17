package postgres

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
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

func TestCardsStore(t *testing.T) {
	repo := repository.TestRepository(t)
	require.NotNil(t, repo)

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cs := NewCardsStorage(repo, log)
	require.NotNil(t, cs)

	userID := createUser(t, repo)
	defer removeUser(t, repo, userID)

	c := models.Card{
		Name:      "name",
		Number:    "number",
		CVC:       "cvc",
		ExpMonth:  2,
		ExpYear:   3,
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
	}

	err = cs.CardCreate(context.Background(), c, userID)
	require.NoError(t, err)

	createdCards, err := cs.Cards(context.Background(), userID)
	require.NoError(t, err)
	require.NotNil(t, createdCards)
	require.Len(t, createdCards, 1)

	createdCard := createdCards[0]
	assert.Equal(t, c.Name, createdCard.Name)
	assert.Equal(t, c.Number, createdCard.Number)
	assert.Equal(t, c.CVC, createdCard.CVC)
	assert.Equal(t, c.ExpMonth, createdCard.ExpMonth)
	assert.Equal(t, c.ExpYear, createdCard.ExpYear)
	assert.Equal(t, c.CreatedAt, createdCard.CreatedAt)
	assert.Equal(t, c.UpdatedAt, createdCard.UpdatedAt)

	uc := models.Card{
		ID:        createdCard.ID,
		Name:      "updated name",
		Number:    "updated number",
		CVC:       "updated cvc",
		ExpMonth:  3,
		ExpYear:   4,
		UpdatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.Local),
	}

	err = cs.CardUpdate(context.Background(), uc, userID)
	require.NoError(t, err)

	updatedCards, err := cs.Cards(context.Background(), userID)
	require.NoError(t, err)
	require.NotNil(t, updatedCards)
	require.Len(t, updatedCards, 1)

	updatedCard := updatedCards[0]
	assert.Equal(t, createdCard.ID, updatedCard.ID)
	assert.Equal(t, uc.Name, updatedCard.Name)
	assert.Equal(t, uc.Number, updatedCard.Number)
	assert.Equal(t, uc.CVC, updatedCard.CVC)
	assert.Equal(t, uc.ExpMonth, updatedCard.ExpMonth)
	assert.Equal(t, uc.ExpYear, updatedCard.ExpYear)
	assert.Equal(t, createdCard.CreatedAt, updatedCard.CreatedAt)
	assert.Equal(t, uc.UpdatedAt, updatedCard.UpdatedAt)

	err = cs.CardDelete(context.Background(), updatedCard.ID, userID)
	require.NoError(t, err)

	passwords, err := cs.Cards(context.Background(), userID)
	require.NoError(t, err)
	require.Empty(t, passwords)

	err = cs.CardUpdate(context.Background(), uc, userID)
	require.Error(t, err)
	require.ErrorIs(t, err, models.ErrNotFound)

	err = cs.CardDelete(context.Background(), updatedCard.ID, userID)
	require.Error(t, err)
	require.ErrorIs(t, err, models.ErrNotFound)
}
