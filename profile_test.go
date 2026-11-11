package gen

import (
	"os"
	"testing"

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
		got, err := ReadProfiles(f)

		// assert
		require.NoError(t, err)
		require.Len(t, got, 440)

		var totalParts int
		for _, part := range got {
			totalParts += part.Quantity
		}
		require.Equal(t, 440, totalParts)
	})

	t.Run("Wrong gen file type", func(t *testing.T) {
		t.Parallel()
		// arrange
		f, err := os.Open("testdata/nest.gen")
		require.NoError(t, err)
		defer f.Close()

		// act
		got, err := ReadProfiles(f)

		// assert
		require.ErrorIs(t, err, ErrNotListedProfile)
		require.Nil(t, got)
	})
}

func BenchmarkReadListedProfile(b *testing.B) {
	b.Run("Read Profile", func(b *testing.B) {
		b.ReportAllocs()

		f, err := os.Open("testdata/profiles.gen")
		require.NoError(b, err)
		defer f.Close()

		for b.Loop() {
			_, _ = ReadProfiles(f)
		}
	})
}
