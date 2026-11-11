package gen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocumentType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "listed profile on first line",
			in:   "TYPE_OF_GENERIC_FILE=LISTED_PROFILE\r\nVERSION=1.0\r\n",
			want: "LISTED_PROFILE",
		},
		{
			name: "plate part inside GENERAL_DATA",
			in:   "GENERAL_DATA\r\nTYPE_OF_GENERIC_FILE=PLATE_PART\r\nEND_OF_GENERAL_DATA\r\n",
			want: "PLATE_PART",
		},
		{
			// testdata/part.gen style: single-part export with an empty type.
			name: "empty type is returned as empty",
			in:   "GENERAL_DATA\r\nTYPE_OF_GENERIC_FILE=\r\nEND_OF_GENERAL_DATA\r\n",
			want: "",
		},
		{name: "absent", in: "GENERAL_DATA\r\nQUALITY=A40\r\n", want: ""},
		{name: "empty input", in: "", want: ""},
		{
			name: "value is trimmed",
			in:   "TYPE_OF_GENERIC_FILE=PLATE_PART   \r\n",
			want: "PLATE_PART",
		},
		{
			// DocumentType only scans a bounded window, so a type far down the
			// file is reported as absent rather than mis-read.
			name: "beyond scan window",
			in:   strings.Repeat("FILLER=x\r\n", 40) + "TYPE_OF_GENERIC_FILE=PLATE_PART\r\n",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := DocumentType(strings.NewReader(tt.in))
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestReadPartsWrongFamily(t *testing.T) {
	t.Parallel()

	profile := "TYPE_OF_GENERIC_FILE=LISTED_PROFILE\r\nVERSION=1.0\r\nUSAGE=LIST\r\n"

	parts, err := ReadParts(strings.NewReader(profile))
	require.ErrorIs(t, err, ErrNotPlatePart)
	require.Nil(t, parts)

	_, err = ReadDocument(strings.NewReader(profile))
	require.ErrorIs(t, err, ErrNotPlatePart)
}

func TestReadProfilesWrongFamily(t *testing.T) {
	t.Parallel()

	plate := numberedLines(docHead(), onePart())

	profiles, err := ReadProfiles(strings.NewReader(plate))
	require.ErrorIs(t, err, ErrNotListedProfile)
	require.Nil(t, profiles)

	_, err = ReadProfileDocument(strings.NewReader(plate))
	require.ErrorIs(t, err, ErrNotListedProfile)
}

// An empty stream is not a plate document and must not be reported as success.
func TestReadPartsEmptyInput(t *testing.T) {
	t.Parallel()

	parts, err := ReadParts(strings.NewReader(""))
	require.ErrorIs(t, err, ErrNotPlatePart)
	require.Nil(t, parts)
}
