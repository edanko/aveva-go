package gen

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadPlate(t *testing.T) {
	t.Parallel()
	t.Run("Read one part from gen file", func(t *testing.T) {
		t.Parallel()
		// arrange
		f, err := os.Open("testdata/part.gen")
		require.NoError(t, err)
		defer f.Close()

		// act
		got, err := ReadParts(f)

		// assert
		require.NoError(t, err)
		require.Len(t, got, 1)
	})

	t.Run("Read nesting from gen file", func(t *testing.T) {
		t.Parallel()
		// arrange
		f, err := os.Open("testdata/nest.gen")
		require.NoError(t, err)
		defer f.Close()

		// act
		got, err := ReadParts(f)

		// assert
		require.NoError(t, err)
		require.Len(t, got, 394)

		var totalParts int
		for _, part := range got {
			totalParts += part.Quantity
		}
		require.Equal(t, 394, totalParts)
	})

	t.Run("Wrong gen file type", func(t *testing.T) {
		t.Parallel()
		// arrange
		f, err := os.Open("testdata/profiles.gen")
		require.NoError(t, err)
		defer f.Close()

		// act
		got, err := ReadParts(f)

		// assert
		require.ErrorIs(t, err, ErrNotPlatePart)
		require.Nil(t, got)
	})
}

func BenchmarkReadPlate(b *testing.B) {
	b.Run("Read one part from gen file", func(b *testing.B) {
		b.ReportAllocs()

		f, err := os.Open("testdata/part.gen")
		require.NoError(b, err)
		defer f.Close()

		for b.Loop() {
			_, _ = ReadParts(f)
		}
	})

	b.Run("Read nesting from gen file", func(b *testing.B) {
		b.ReportAllocs()

		f, err := os.Open("testdata/nest.gen")
		require.NoError(b, err)
		defer f.Close()

		for b.Loop() {
			_, _ = ReadParts(f)
		}
	})
}
