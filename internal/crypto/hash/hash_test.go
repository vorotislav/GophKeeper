package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	var (
		s      = "examplestring1234567890"
		h1, h2 string
		err    error
	)

	t.Run("hash strings", func(t *testing.T) {
		h1, err = Password(s)
		assert.NoError(t, err)
		assert.NotEqual(t, "", s)
	})

	t.Run("hashes are different", func(t *testing.T) {
		h2, err = Password(s)
		assert.NoError(t, err)
		assert.NotEqual(t, h1, h2)
	})

	t.Run("check hashes", func(t *testing.T) {
		err = CheckPassword(s, h1)
		assert.NoError(t, err)

		err = CheckPassword(s, h2)
		assert.NoError(t, err)
	})

}
