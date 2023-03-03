package gen

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadPlate(t *testing.T) {
	t.Run("Read one part from gen file", func(t *testing.T) {
		f, err := os.Open("testdata/part.gen")
		assert.NoError(t, err)
		defer f.Close()

		got := ReadPlate(f)
		assert.Len(t, got, 1)
	})

	t.Run("Read nesting from gen file", func(t *testing.T) {
		f, err := os.Open("testdata/nest.gen")
		assert.NoError(t, err)
		defer f.Close()

		got := ReadPlate(f)
		assert.Len(t, got, 392)

		var totalParts int
		for _, part := range got {
			totalParts += part.Quantity
		}
		assert.Equal(t, totalParts, 394)
	})

	t.Run("Wrong gen file type", func(t *testing.T) {
		f, err := os.Open("testdata/profiles.gen")
		assert.NoError(t, err)
		defer f.Close()

		got := ReadPlate(f)
		assert.Nil(t, got)
	})
}
