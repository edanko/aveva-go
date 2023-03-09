package gen

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
		got := ReadPlate(f)

		// assert
		assert.Len(t, got, 1)
	})

	t.Run("Read nesting from gen file", func(t *testing.T) {
		t.Parallel()
		// arrange
		f, err := os.Open("testdata/nest.gen")
		require.NoError(t, err)
		defer f.Close()

		// act
		got := ReadPlate(f)

		// assert
		require.Len(t, got, 394)

		var totalParts int
		for _, part := range got {
			totalParts += part.Quantity
		}
		assert.Equal(t, totalParts, 394)
	})

	t.Run("Wrong gen file type", func(t *testing.T) {
		t.Parallel()
		// arrange
		f, err := os.Open("testdata/profiles.gen")
		require.NoError(t, err)
		defer f.Close()

		// act
		got := ReadPlate(f)

		// assert
		assert.Nil(t, got)
	})
}

func BenchmarkReadPlate(b *testing.B) {
	b.Run("Read one part from gen file", func(b *testing.B) {
		b.ReportAllocs()

		f, err := os.Open("testdata/part.gen")
		require.NoError(b, err)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = ReadPlate(f)
			}
		})
	})

	b.Run("Read nesting from gen file", func(b *testing.B) {
		b.ReportAllocs()

		f, err := os.Open("testdata/nest.gen")
		require.NoError(b, err)

		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = ReadPlate(f)
			}
		})
	})
}
