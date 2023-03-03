package gen

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadListedProfile(t *testing.T) {
	t.Parallel()
	t.Run("Read Profile", func(t *testing.T) {
		t.Parallel()
		// arrange
		f, err := os.Open("testdata/profiles.gen")
		require.NoError(t, err)

		// act
		got := ReadProfile(f)

		// assert
		require.Len(t, got, 440)

		var totalParts int
		for _, part := range got {
			totalParts += part.Quantity
		}
		assert.Equal(t, totalParts, 440)
	})

	t.Run("Wrong gen file type", func(t *testing.T) {
		t.Parallel()
		// arrange
		f, err := os.Open("testdata/nest.gen")
		require.NoError(t, err)
		defer f.Close()

		// act
		got := ReadProfile(f)

		// assert
		assert.Nil(t, got)
	})
}

func BenchmarkReadListedProfile(b *testing.B) {
	b.Run("Read Profile", func(b *testing.B) {
		b.ReportAllocs()

		f, err := os.Open("testdata/profiles.gen")
		require.NoError(b, err)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = ReadProfile(f)
			}
		})

	})
}
