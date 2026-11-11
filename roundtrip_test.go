package gen

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

// The shipped fixtures are the regression net for the whole package: they are
// real AVEVA output, so a parse regression or a writer regression shows up here
// rather than only in a full corpus run.
//
// Round-trip means parse -> WriteDocument -> parse again and require the second
// parse to be IDENTICAL. That is the contract the writers are held to; byte
// equality with AVEVA output is not claimed, because the corpus contains several
// writer dialects that disagree on field presence and on empty-vs-zero.
var plateFixtures = []string{
	"testdata/part.gen",
	"testdata/nest.gen",
}

var profileFixtures = []string{
	"testdata/profiles.gen",
}

func readFixture(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoErrorf(t, err, "fixture %s must be present", path)

	return data
}

func TestPlateFixturesParse(t *testing.T) {
	t.Parallel()

	for _, path := range plateFixtures {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			data := readFixture(t, path)

			doc, err := ReadDocument(bytes.NewReader(data))
			require.NoError(t, err)
			require.NotNil(t, doc.General, "GENERAL_DATA must be parsed")
			require.NotEmpty(t, doc.Parts, "fixture must contain parts")

			// Sanity-check the document actually carries usable geometry, so a
			// silent "parsed zero sections" regression cannot pass.
			var segments, withBurning int
			for _, p := range doc.Parts {
				if len(p.BurningData) > 0 {
					withBurning++
				}
				for _, b := range p.BurningData {
					if b.Contour != nil {
						segments += len(b.Contour.Segments)
					}
				}
			}
			require.Positive(t, withBurning, "at least one part must have cut paths")
			require.Positive(t, segments, "at least one contour must have segments")
		})
	}
}

func TestPlateFixturesRoundTrip(t *testing.T) {
	t.Parallel()

	for _, path := range plateFixtures {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			first, err := ReadDocument(bytes.NewReader(readFixture(t, path)))
			require.NoError(t, err)

			var buf bytes.Buffer
			require.NoError(t, WriteDocument(&buf, first))

			second, err := ReadDocument(bytes.NewReader(buf.Bytes()))
			require.NoError(t, err, "written output must be parseable by this package")

			// Whole-structure equality, not a spot check: this is what catches a
			// writer that drops a field or loses numeric precision.
			require.True(t, reflect.DeepEqual(first.General, second.General),
				"GENERAL_DATA must survive a round-trip")
			require.Len(t, second.Parts, len(first.Parts))
			require.Equal(t, first.LabelTexts, second.LabelTexts)
			require.Equal(t, first.LabelSymbols, second.LabelSymbols)

			for i := range first.Parts {
				require.Equalf(t, first.Parts[i], second.Parts[i],
					"part %d (%s) must survive a round-trip", i, first.Parts[i].Name)
			}
		})
	}
}

// Writing twice must be byte-identical: the encoder has to be deterministic.
func TestWriteDocumentIdempotent(t *testing.T) {
	t.Parallel()

	for _, path := range plateFixtures {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			doc, err := ReadDocument(bytes.NewReader(readFixture(t, path)))
			require.NoError(t, err)

			var once, twice bytes.Buffer
			require.NoError(t, WriteDocument(&once, doc))
			require.NoError(t, WriteDocument(&twice, doc))
			require.Equal(t, once.String(), twice.String())
		})
	}
}

// Mirroring then writing must be stable: a mirrored document written and re-read
// is already mirrored and must not flip again.
func TestPlateFixtureMirrorRoundTrip(t *testing.T) {
	t.Parallel()

	doc, err := ReadDocument(bytes.NewReader(readFixture(t, "testdata/part.gen")))
	require.NoError(t, err)
	require.NotEmpty(t, doc.Parts)

	for _, p := range doc.Parts {
		p.Mirror()
	}

	var buf bytes.Buffer
	require.NoError(t, WriteDocument(&buf, doc))

	back, err := ReadDocument(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	require.Len(t, back.Parts, len(doc.Parts))

	for i := range doc.Parts {
		require.True(t, back.Parts[i].Mirrored, "mirror flag must persist")
		require.Equal(t, doc.Parts[i], back.Parts[i],
			"mirrored part %d must survive a round-trip", i)
	}
}

func TestProfileFixturesParse(t *testing.T) {
	t.Parallel()

	for _, path := range profileFixtures {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			doc, err := ReadProfileDocument(bytes.NewReader(readFixture(t, path)))
			require.NoError(t, err)
			require.NotNil(t, doc.Common, "COMMON_DATA must be parsed")
			require.NotEmpty(t, doc.Profiles)

			// The cross-section contour used to be skipped entirely.
			require.NotNil(t, doc.Common.Contour, "COMMON_DATA contour must be parsed")
			require.NotEmpty(t, doc.Common.Contour.Segments)
			require.Equal(t, uint64(len(doc.Profiles)), doc.Common.NoOfProfs,
				"NO_OF_PROFS must agree with the number of profiles")

			withEnds := 0
			for _, p := range doc.Profiles {
				if p.LeftEnd != nil && p.RightEnd != nil {
					withEnds++
				}
			}
			require.Equal(t, len(doc.Profiles), withEnds,
				"every profile must have both end cuts")
		})
	}
}

func TestProfileFixturesRoundTrip(t *testing.T) {
	t.Parallel()

	for _, path := range profileFixtures {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			first, err := ReadProfileDocument(bytes.NewReader(readFixture(t, path)))
			require.NoError(t, err)

			var buf bytes.Buffer
			require.NoError(t, WriteProfiles(&buf, first))

			second, err := ReadProfileDocument(bytes.NewReader(buf.Bytes()))
			require.NoError(t, err)

			require.Equal(t, first.Version, second.Version)
			require.Equal(t, first.Usage, second.Usage)
			require.True(t, reflect.DeepEqual(first.Common, second.Common),
				"COMMON_DATA must survive a round-trip, including untruncated floats")

			require.Len(t, second.Profiles, len(first.Profiles))
			for i := range first.Profiles {
				require.Equalf(t, first.Profiles[i], second.Profiles[i],
					"profile %d (%s) must survive a round-trip, including both end cuts",
					i, first.Profiles[i].Name)
			}
		})
	}
}

// ENDCUT_MASK is a bit pattern; leading zeros must not be lost.
func TestProfileFixtureEndcutMaskPreserved(t *testing.T) {
	t.Parallel()

	doc, err := ReadProfileDocument(bytes.NewReader(readFixture(t, "testdata/profiles.gen")))
	require.NoError(t, err)

	var withMask int
	for _, p := range doc.Profiles {
		if p.LeftEnd != nil && p.LeftEnd.EndcutMask != "" {
			withMask++
			require.Greater(t, len(p.LeftEnd.EndcutMask), 1,
				"ENDCUT_MASK is a bit pattern, got %q", p.LeftEnd.EndcutMask)
		}
	}
	require.Positive(t, withMask, "fixture must exercise ENDCUT_MASK")
}
