package gen

import (
	"archive/zip"
	"io/fs"
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

	t.Run("Read multiple files from zip archive", func(t *testing.T) {
		r, err := zip.OpenReader("testdata/archive.zip")
		assert.NoError(t, err)
		defer r.Close()

		var gens []string

		err = fs.WalkDir(r, ".", func(path string, d fs.DirEntry, err error) error {
			assert.NoError(t, err)

			if d.Type().IsRegular() {
				gens = append(gens, path)
			}

			return nil
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, gens)

		for _, path := range gens {
			f, err := r.Open(path)
			assert.NoError(t, err)

			got := ReadPlate(f)
			assert.NotNil(t, got)
		}
	})
}
