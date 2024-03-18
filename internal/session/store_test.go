package session

import (
	"GophKeeper/internal/models"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestStorage_GetSession(t *testing.T) {
	session := models.Session{
		ID:                    1,
		UserID:                1,
		AccessToken:           "access",
		RefreshToken:          "refresh",
		IPAddress:             "ip",
		RefreshTokenExpiredAt: 1,
		CreatedAt:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	store := Storage{session: session}

	got := store.GetSession()

	require.Equal(t, session, got)
}

func TestStorage_SaveSession(t *testing.T) {
	session := models.Session{
		ID:                    1,
		UserID:                1,
		AccessToken:           "access",
		RefreshToken:          "refresh",
		IPAddress:             "ip",
		RefreshTokenExpiredAt: 1,
		CreatedAt:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	store := Storage{}

	store.SaveSession(session)

	require.Equal(t, session, store.session)
}
