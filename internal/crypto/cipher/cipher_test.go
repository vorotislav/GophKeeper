package cipher

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCipher(t *testing.T) {
	t.Parallel()

	c := NewCipher("key", "salt")
	require.NotNil(t, c)

	plainValue := "some string"
	value, err := c.EncryptString(plainValue)
	require.NoError(t, err)
	require.NotEmpty(t, value)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		decryptValue, err := c.DecryptString(value)
		require.NoError(t, err)
		require.NotEmpty(t, decryptValue)
		require.Equal(t, plainValue, decryptValue)
	})

	t.Run("plain text", func(t *testing.T) {
		t.Parallel()

		decryptValue, err := c.DecryptString(plainValue)
		require.Error(t, err)
		require.Empty(t, decryptValue)
		require.ErrorIs(t, err, errInvalidCipherTextLen)
	})

	t.Run("different secretKey", func(t *testing.T) {
		t.Parallel()

		c2 := NewCipher("another key", "salt")
		require.NotNil(t, c2)

		decryptValue, err := c2.DecryptString(value)
		require.Error(t, err)
		require.Empty(t, decryptValue)
		require.Contains(t, err.Error(), "cipher")
	})

	t.Run("different salt", func(t *testing.T) {
		t.Parallel()

		c2 := NewCipher("key", "another salt")
		require.NotNil(t, c2)

		decryptValue, err := c2.DecryptString(value)
		require.Error(t, err)
		require.Empty(t, decryptValue)
		require.Contains(t, err.Error(), "cipher")
	})
}
