package gen

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadListedProfile(t *testing.T) {
	t.Run("Read Listed Profile", func(t *testing.T) {
		f, err := os.Open("testdata/profiles.gen")
		assert.NoError(t, err)

		got := ReadListedProfile(f)
		assert.Len(t, got, 440)

		var totalParts int
		for _, part := range got {
			totalParts += part.Quantity
		}
		assert.Equal(t, totalParts, 440)
	})

	t.Run("Wrong gen file type", func(t *testing.T) {
		f, err := os.Open("testdata/nest.gen")
		assert.NoError(t, err)
		defer f.Close()

		got := ReadListedProfile(f)
		assert.Nil(t, got)
	})
}
