package gen

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// These tests pin the payload, not just the record counts. The pre-existing
// tests asserted only len(parts)/Quantity, which is why a parser that silently
// dropped every cut path still passed.

func open(t *testing.T, name string) *os.File {
	t.Helper()

	f, err := os.Open("testdata/" + name)
	require.NoError(t, err)
	t.Cleanup(func() { f.Close() })

	return f
}

// A nest must retain the per-part payload that follows the PART_DATA blocks.
// Regression: an unhandled IDLE_DATA marker used to consume the rest of file.
func TestNestRetainsCutPaths(t *testing.T) {
	t.Parallel()

	parts, err := ReadParts(open(t, "nest.gen"))
	require.NoError(t, err)
	require.Len(t, parts, 394)

	var withBurning, segments int
	for _, p := range parts {
		if len(p.BurningData) == 0 {
			continue
		}
		withBurning++
		for _, b := range p.BurningData {
			require.NotNil(t, b.Contour)
			segments += len(b.Contour.Segments)
		}
	}

	// 919 BURNING_DATA blocks spread over all 394 parts (a bevelled part emits
	// one contour per bevel pass).
	require.Equal(t, 394, withBurning)
	require.Equal(t, 2378, segments)
}

// Every placement of a repeated part must keep its own payload.
// Regression: a map keyed by part name collapsed duplicates onto the last index.
func TestNestRepeatedPartKeepsPayload(t *testing.T) {
	t.Parallel()

	parts, err := ReadParts(open(t, "nest.gen"))
	require.NoError(t, err)

	byName := make(map[string][]*Part)
	for _, p := range parts {
		byName[p.Name] = append(byName[p.Name], p)
	}

	require.Len(t, byName, 392, "nest.gen has 392 distinct part names")

	repeated := 0
	for name, group := range byName {
		if len(group) < 2 {
			continue
		}
		repeated++

		// Every placement of a repeated part must carry its own payload, and
		// placements of one identical part must be indistinguishable in size.
		want := len(group[0].BurningData)
		require.NotZero(t, want, "part %q lost its cut path", name)
		for _, p := range group {
			require.Len(t, p.BurningData, want, "placement of %q has a different payload", name)
		}
	}
	require.Equal(t, 1, repeated, "one part name is placed more than once")

	// No part may end up with an empty payload.
	for _, p := range parts {
		require.NotEmpty(t, p.BurningData, "part %q lost its cut path", p.Name)
	}
}

// LISTED_PROFILE carries its end cuts as siblings of the attribute block, after
// END_OF_PROFILE_DATA. Regression: they were never read (they are emitted after
// the terminator, so a nested reader never saw them).
func TestProfilesRetainEndCuts(t *testing.T) {
	t.Parallel()

	profiles, err := ReadProfiles(open(t, "profiles.gen"))
	require.NoError(t, err)
	require.Len(t, profiles, 440)

	for _, p := range profiles {
		require.NotNil(t, p.LeftEnd, "profile %q has no left end", p.Name)
		require.NotNil(t, p.RightEnd, "profile %q has no right end", p.Name)

		for _, e := range []*End{p.LeftEnd, p.RightEnd} {
			require.NotNil(t, e.Contour)
			require.NotNil(t, e.FlangeContour)
			require.NotEmpty(t, e.Contour.Segments)
			require.NotEmpty(t, e.FlangeContour.Segments)
		}
	}
}

// End-cut numeric fields may be the '-' sentinel, meaning "this ENDCUT_TYPE does
// not define the parameter". Regression: ParseFloat("-") aborted the whole file.
//
// The correlation below is exact in the shipped sample and is what makes the
// sentinel handling meaningful rather than merely tolerant.
func TestEndCutSentinelByType(t *testing.T) {
	t.Parallel()

	profiles, err := ReadProfiles(open(t, "profiles.gen"))
	require.NoError(t, err)

	perType := make(map[int]int)

	for _, p := range profiles {
		for _, e := range []*End{p.LeftEnd, p.RightEnd} {
			require.NotNil(t, e)
			perType[e.EndcutType]++

			switch e.EndcutType {
			case 21:
				// A and R2 are '-' for this end-cut type: no arc / corner radius.
				require.Zero(t, e.A, "ENDCUT_TYPE 21 defines A as '-'")
				require.Zero(t, e.R2, "ENDCUT_TYPE 21 defines R2 as '-'")
			case 27:
				// A and R2 are real radii for this end-cut type.
				require.NotZero(t, e.A)
				require.NotZero(t, e.R2)
			}

			// These are '-' on every end in the sample: reserved fields.
			require.Zero(t, e.B)
			require.Zero(t, e.C)
			require.Zero(t, e.V3)
			require.Zero(t, e.V4)
		}
	}

	require.Equal(t, map[int]int{21: 760, 27: 120}, perType)
}

// Empty contours must not panic.
func TestEmptyContourHelpers(t *testing.T) {
	t.Parallel()

	c := &Contour{}
	require.Nil(t, c.Vertices())
	require.Nil(t, c.Bulges())
	require.False(t, c.IsClockwise())
}

// Blank lines must not terminate parsing. The shipped samples happen to contain
// none, so this is covered by a synthetic input.
func TestReadPartsToleratesBlankLines(t *testing.T) {
	t.Parallel()

	in := strings.Join([]string{
		"GENERAL_DATA",
		"RAW_THICKNESS=10.00",
		"QUALITY=A40",
		"NO_OF_PARTS=1",
		"END_OF_GENERAL_DATA",
		"",
		"PART_DATA",
		"NAME=P1",
		"POSNO=1",
		"END_OF_PART_DATA",
		"",
		"PART_INFORMATION",
		"PART_NAME=P1",
		"TYPE_OF_WORK=BURNING",
		"END_OF_PART_INFORMATION",
		"",
		"IDLE_DATA",
		"START_OF_CONTOUR",
		"END_OF_CONTOUR",
		"END_OF_IDLE_DATA",
		"",
	}, "\r\n")

	parts, err := ReadParts(strings.NewReader(in))
	require.NoError(t, err)
	require.Len(t, parts, 1)
	require.Equal(t, "P1", parts[0].Name)
	require.Equal(t, float64(10), parts[0].Thickness)
}

// A document may declare its single part only through PART_INFORMATION, with no
// PART_DATA block at all: the attributes live in GENERAL_DATA. Found in 6 files
// of the 45,750-file corpus. The part must be synthesized, not dropped and not
// treated as an error.
func TestReadPartsWithoutPartData(t *testing.T) {
	t.Parallel()

	in := strings.Join([]string{
		"GENERAL_DATA",
		"TYPE_OF_GENERIC_FILE=PLATE_PART",
		"RAW_THICKNESS=8.00",
		"QUALITY=D40_B",
		"NO_OF_PARTS=1",
		"END_OF_GENERAL_DATA",
		"PART_INFORMATION",
		"PART_NAME=P2Z_2P-68BP",
		"TYPE_OF_WORK=BURNING",
		"END_OF_PART_INFORMATION",
		"MARKING_DATA",
		"MARKING_SIDE=TS",
		"END_OF_MARKING_DATA",
	}, "\r\n")

	parts, err := ReadParts(strings.NewReader(in))
	require.NoError(t, err)
	require.Len(t, parts, 1)
	require.Equal(t, "P2Z_2P-68BP", parts[0].Name)
	require.Equal(t, 8.0, parts[0].Thickness) // from GENERAL_DATA
	require.Equal(t, "D40_B", parts[0].Quality)
	require.Len(t, parts[0].MarkingData, 1)
	require.NotNil(t, parts[0].PartCog)

	// Must still support mirroring (PartCog is dereferenced there).
	parts[0].Mirror()
}

// A PART_INFORMATION with an empty PART_NAME is a scope with no owning part
// (e.g. TYPE_OF_WORK=LABELLING). Its payload must be skipped, not attributed to
// the previous part and not fatal. Found in 7 files of the corpus.
func TestReadPartsEmptyScopeName(t *testing.T) {
	t.Parallel()

	in := strings.Join([]string{
		"GENERAL_DATA",
		"RAW_THICKNESS=10.00",
		"QUALITY=A40",
		"NO_OF_PARTS=1",
		"END_OF_GENERAL_DATA",
		"PART_DATA",
		"NAME=P1",
		"END_OF_PART_DATA",
		"PART_INFORMATION",
		"PART_NAME=P1",
		"END_OF_PART_INFORMATION",
		"MARKING_DATA",
		"MARKING_TYPE=PANEL",
		"END_OF_MARKING_DATA",
		"PART_INFORMATION",
		"PART_NAME=",
		"TYPE_OF_WORK=LABELLING",
		"END_OF_PART_INFORMATION",
		"MARKING_DATA",
		"MARKING_TYPE=SHOULD_BE_DROPPED",
		"END_OF_MARKING_DATA",
		"IDLE_DATA",
		"START_OF_CONTOUR",
		"END_OF_CONTOUR",
		"END_OF_IDLE_DATA",
	}, "\r\n")

	parts, err := ReadParts(strings.NewReader(in))
	require.NoError(t, err)
	require.Len(t, parts, 1)
	require.Len(t, parts[0].MarkingData, 1)
	require.Equal(t, "PANEL", parts[0].MarkingData[0].MarkingType,
		"payload from a nameless scope must not attach to the previous part")
}

// A part may be opened by more scopes than it has PART_DATA placements; the
// surplus must accumulate onto the last placement rather than error.
// Observed widely in the corpus (13 of 25 names in one file alone).
func TestReadPartsOverReferencedScope(t *testing.T) {
	t.Parallel()

	in := strings.Join([]string{
		"GENERAL_DATA",
		"RAW_THICKNESS=10.00",
		"NO_OF_PARTS=1",
		"END_OF_GENERAL_DATA",
		"PART_DATA",
		"NAME=P1",
		"END_OF_PART_DATA",
		"PART_INFORMATION",
		"PART_NAME=P1",
		"END_OF_PART_INFORMATION",
		"MARKING_DATA",
		"MARKING_TYPE=A",
		"END_OF_MARKING_DATA",
		"PART_INFORMATION",
		"PART_NAME=P1",
		"END_OF_PART_INFORMATION",
		"MARKING_DATA",
		"MARKING_TYPE=B",
		"END_OF_MARKING_DATA",
	}, "\r\n")

	parts, err := ReadParts(strings.NewReader(in))
	require.NoError(t, err)
	require.Len(t, parts, 1)
	require.Len(t, parts[0].MarkingData, 2, "both scopes accumulate on the one placement")
}

func numberedLines(parts ...[]string) string {
	var out []string
	for _, p := range parts {
		out = append(out, p...)
	}

	return strings.Join(out, "\r\n")
}

func docHead() []string {
	return []string{"GENERAL_DATA", "TYPE_OF_GENERIC_FILE=PLATE_PART", "RAW_THICKNESS=10.00", "QUALITY=A40", "NO_OF_PARTS=1", "END_OF_GENERAL_DATA"}
}

func onePart() []string {
	return []string{"PART_DATA", "NAME=P1", "END_OF_PART_DATA", "PART_INFORMATION", "PART_NAME=P1", "TYPE_OF_WORK=BURNING", "END_OF_PART_INFORMATION"}
}

// LABELTEXT_DATA / LABELSYMBOL_DATA / EDGE_DATA and contour AUXILIARY_DATA were
// previously skipped; they occur in 4,362 / 4,356 / 124 / 201 corpus files.
func TestReadPartsLabelAndAuxiliaryData(t *testing.T) {
	t.Parallel()

	in := numberedLines(docHead(), onePart(), []string{
		"BURNING_DATA",
		"SHAPE=OUTER_CONTOUR",
		"START_OF_CONTOUR",
		"NO_OF_SEG=1",
		"START_U=0", "START_V=0", "U=10", "V=0",
		"AUXILIARY_DATA", "AUXILIARY_FUNCTION=99", "END_OF_AUXILIARY_DATA",
		"AUXILIARY_DATA", "AUXILIARY_FUNCTION=48", "END_OF_AUXILIARY_DATA",
		"END_OF_CONTOUR",
		"END_OF_BURNING_DATA",
		"LABELTEXT_DATA",
		"TEXT_TYPE=PANEL_NAME                ",
		"TEXT_POSITION_U=2452.92051",
		"TEXT_POSITION_V=2728.90640",
		"TEXT_ANGLE=0.00",
		"TEXT_HEIGHT=10.00",
		"TEXT_DIRECTION=1",
		"TEXT_SIDE=BOTH",
		"TEXT=10123-QRT2100_S/S45S",
		"TEXT_ELEMENTS=1",
		"GEOMETRY_DATA",
		"GEOMETRY_TYPE=",
		"START_OF_CONTOUR",
		"NO_OF_SEG=1",
		"START_U=0", "START_V=0", "U=5", "V=5",
		"END_OF_CONTOUR",
		"END_OF_GEOMETRY_DATA",
		"END_OF_LABELTEXT_DATA",
		"LABELSYMBOL_DATA",
		"SYMBOL_TYPE=POSNO                     ",
		"SYMBOL_POSITION_U=94.1",
		"SYMBOL_POSITION_V=65.15939",
		"SYMBOL_ANGLE=0.00",
		"SYMBOL_HEIGHT=12.00",
		"SYMBOL_FONT=2",
		"SYMBOL_NUMBER=61",
		"END_OF_LABELSYMBOL_DATA",
		"EDGE_DATA",
		"NAME=S8",
		"START_U=21.34827", "START_V=15.0",
		"END_U=1094.64559", "END_V=15.0",
		"ALENGTH=1073.30", "MLENGTH=1072.50",
		"END_OF_EDGE_DATA",
	})

	parts, err := ReadParts(strings.NewReader(in))
	require.NoError(t, err)
	require.Len(t, parts, 1)

	p := parts[0]
	require.Len(t, p.Information, 1)
	require.Equal(t, "BURNING", p.Information[0].TypeOfWork)

	require.Len(t, p.BurningData, 1)
	require.Len(t, p.BurningData[0].Contour.Auxiliaries, 2)
	require.Equal(t, "99", p.BurningData[0].Contour.Auxiliaries[0].Function)
	require.Equal(t, "48", p.BurningData[0].Contour.Auxiliaries[1].Function)

	require.Len(t, p.LabelTexts, 1)
	lt := p.LabelTexts[0]
	require.Equal(t, "PANEL_NAME", lt.Type) // space padding trimmed
	require.Equal(t, "10123-QRT2100_S/S45S", lt.Text)
	require.Equal(t, 1, lt.Elements)
	require.Len(t, lt.Geometry, 1, "glyph stroke must be retained")

	require.Len(t, p.LabelSymbols, 1)
	require.Equal(t, "POSNO", p.LabelSymbols[0].Type)
	require.Equal(t, 61, p.LabelSymbols[0].Number)

	require.Len(t, p.Edges, 1)
	require.Equal(t, "S8", p.Edges[0].Name)
	require.InDelta(t, 1073.30, p.Edges[0].ALength, 1e-6)
}

// A nest-wide labelling pass opens with a nameless PART_INFORMATION. Its labels
// belong to the document, not to whichever part preceded it.
func TestReadDocumentCollectsUnattributedLabels(t *testing.T) {
	t.Parallel()

	in := numberedLines(docHead(), onePart(), []string{
		"MARKING_DATA", "MARKING_TYPE=PANEL", "END_OF_MARKING_DATA",
		"PART_INFORMATION", "PART_NAME=", "TYPE_OF_WORK=LABELLING", "END_OF_PART_INFORMATION",
		"IDLE_DATA", "START_OF_CONTOUR", "END_OF_CONTOUR", "END_OF_IDLE_DATA",
		"LABELTEXT_DATA", "TEXT_TYPE=PANEL_NAME", "TEXT=OTHER-PART", "END_OF_LABELTEXT_DATA",
		"LABELSYMBOL_DATA", "SYMBOL_TYPE=POSNO", "END_OF_LABELSYMBOL_DATA",
	})

	doc, err := ReadDocument(strings.NewReader(in))
	require.NoError(t, err)
	require.Len(t, doc.Parts, 1)
	require.Empty(t, doc.Parts[0].LabelTexts, "unattributed label must not join the previous part")
	require.Len(t, doc.LabelTexts, 1)
	require.Equal(t, "OTHER-PART", doc.LabelTexts[0].Text)
	require.Len(t, doc.LabelSymbols, 1)

	require.Equal(t, "PLATE_PART", doc.Type)
	require.Equal(t, 10.0, doc.General.Thickness)
}

// Empty numeric values and the '-' sentinel are both real in the corpus, so the
// header must parse without erroring.
func TestReadPartsTolerantNumerics(t *testing.T) {
	t.Parallel()

	in := numberedLines([]string{
		"GENERAL_DATA",
		"TYPE_OF_GENERIC_FILE=",
		"RAW_LENGTH=",
		"REST_LENGTH=",
		"NO_OF_MARKING_STARTS=",
		"TOOL_ID=",
		"RAW_THICKNESS=8.00",
		"DENSITY=7850.00",
		"NO_OF_PARTS=1",
		"END_OF_GENERAL_DATA",
	}, onePart())

	doc, err := ReadDocument(strings.NewReader(in))
	require.NoError(t, err)
	require.Len(t, doc.Parts, 1)
	require.Equal(t, 8.0, doc.General.Thickness)
	require.Equal(t, 7850.0, doc.General.Density)
	require.Zero(t, doc.General.NoOfMarkingStarts)
	require.Empty(t, doc.Type, "empty TYPE_OF_GENERIC_FILE is legal for a single-part export")
}

// Document families must be distinguishable, not silently empty.
func TestDocumentTypeAndWrongFamily(t *testing.T) {
	t.Parallel()

	plate := numberedLines(docHead(), onePart())
	typ, err := DocumentType(strings.NewReader(plate))
	require.NoError(t, err)
	require.Equal(t, "PLATE_PART", typ)

	got, err := ReadParts(strings.NewReader(plate))
	require.NoError(t, err)
	require.Len(t, got, 1)

	prof := numberedLines([]string{"TYPE_OF_GENERIC_FILE=LISTED_PROFILE", "VERSION=1.0", "USAGE=LIST", "COMMON_DATA", "END_OF_COMMON_DATA"})
	typ, err = DocumentType(strings.NewReader(prof))
	require.NoError(t, err)
	require.Equal(t, "LISTED_PROFILE", typ)

	got, err = ReadParts(strings.NewReader(prof))
	require.ErrorIs(t, err, ErrNotPlatePart)
	require.Nil(t, got)

	_, err = ReadProfiles(strings.NewReader(plate))
	require.ErrorIs(t, err, ErrNotListedProfile)
}

// WriteDocument must reproduce a document faithfully enough that re-reading it
// yields the same structure. This is the contract the writer is held to; byte
// equality with AVEVA output is not claimed (see docs Appendix E).
func TestWriteRoundTrip(t *testing.T) {
	t.Parallel()

	in := numberedLines(docHead(), onePart(), []string{
		"BURNING_DATA",
		"SHAPE=OUTER_CONTOUR",
		"DIRECTION=-1",
		"NUMBER_OF_HEADS=1",
		"GEOMETRY_VALID_FOR=1",
		"DISTANCE_Y1_Y2=0.00",
		"START_OF_CONTOUR",
		// bevel of segment 0, before NO_OF_SEG
		"BEVEL_DATA", "BEVEL=TS", "BEVEL_CODE=145.00", "BEVEL_NAME=T7-R",
		"BEVEL_TYPE=35", "BEVEL_VARIANT=0", "PLATE_THICKNESS=14.00", "E=14.00",
		"ANGLE_TS=45.00", "CHAMFER_HEIGHT_OS=2.50",
		"END_OF_BEVEL_DATA",
		"NO_OF_SEG=2",
		"START_U=0", "START_V=0",
		"AMP_U=0", "AMP_V=0", "AMP=0", "RADIUS=0", "SWEEP=0",
		"ORIGIN_U=0", "ORIGIN_V=0", "U=10", "V=0",
		// segment 1 changes bevel
		"BEVEL_DATA", "BEVEL=NONE", "BEVEL_CODE=0.00", "BEVEL_NAME=",
		"BEVEL_TYPE=0", "BEVEL_VARIANT=0", "PLATE_THICKNESS=14.00", "E=0.00",
		"ANGLE_TS=0.00", "CHAMFER_HEIGHT_OS=0.00",
		"END_OF_BEVEL_DATA",
		"AMP_U=0", "AMP_V=0", "AMP=0", "RADIUS=100", "SWEEP=90",
		"ORIGIN_U=10", "ORIGIN_V=10", "U=20", "V=10",
		"AUXILIARY_DATA", "AUXILIARY_FUNCTION=99", "END_OF_AUXILIARY_DATA",
		"END_OF_CONTOUR",
		"END_OF_BURNING_DATA",
		"LABELTEXT_DATA",
		"TEXT_TYPE=PANEL_NAME", "TEXT_POSITION_U=2452.92051", "TEXT=ABC",
		"END_OF_LABELTEXT_DATA",
		"EDGE_DATA", "NAME=S8", "START_U=1.5", "START_V=2.5", "END_U=3.5", "END_V=4.5",
		"ALENGTH=1073.30", "MLENGTH=1072.50", "END_OF_EDGE_DATA",
	})

	first, err := ReadDocument(strings.NewReader(in))
	require.NoError(t, err)

	var buf strings.Builder
	require.NoError(t, WriteDocument(&buf, first))

	second, err := ReadDocument(strings.NewReader(buf.String()))
	require.NoError(t, err)

	require.Equal(t, first.General.Thickness, second.General.Thickness)
	require.Equal(t, first.General.Quality, second.General.Quality)
	require.Len(t, second.Parts, 1)

	a, b := first.Parts[0], second.Parts[0]
	require.Equal(t, a.Name, b.Name)
	require.Equal(t, a.PartCog.X, b.PartCog.X)
	require.Equal(t, a.Mirrored, b.Mirrored)

	require.Len(t, b.BurningData, 1)
	ca, cb := a.BurningData[0].Contour, b.BurningData[0].Contour
	require.Equal(t, -1, b.BurningData[0].Direction)
	require.Len(t, cb.Segments, len(ca.Segments))
	require.Len(t, cb.Auxiliaries, 1)
	require.Equal(t, "99", cb.Auxiliaries[0].Function)

	// Sparse per-segment bevels must survive: segment 0 keeps its bevel and
	// segment 1 keeps its own (NONE), not inherited from segment 0.
	require.NotNil(t, cb.Segments[0].Bevel)
	require.Equal(t, "TS", cb.Segments[0].Bevel.Bevel)
	require.Equal(t, 145.0, cb.Segments[0].Bevel.BevelCode)
	require.NotNil(t, cb.Segments[1].Bevel)
	require.Equal(t, "NONE", cb.Segments[1].Bevel.Bevel)
	require.InDelta(t, 100.0, cb.Segments[1].Radius, 1e-9)
	require.InDelta(t, 90.0, cb.Segments[1].Sweep, 1e-9)

	require.Len(t, b.LabelTexts, 1)
	require.Equal(t, "ABC", b.LabelTexts[0].Text)
	require.InDelta(t, 2452.92051, b.LabelTexts[0].PositionU, 1e-9)
	require.Len(t, b.Edges, 1)
	require.InDelta(t, 1073.30, b.Edges[0].ALength, 1e-6)

	// Writing twice must be stable (idempotent).
	var buf2 strings.Builder
	require.NoError(t, WriteDocument(&buf2, second))
	require.Equal(t, buf.String(), buf2.String())
}

// A part with no PART_INFORMATION scope had no payload either, so the writer must
// not invent one.
func TestWriteSkipsScopeForUnopenedPart(t *testing.T) {
	t.Parallel()

	in := numberedLines(docHead(), []string{"PART_DATA", "NAME=ORPHAN", "END_OF_PART_DATA"})

	doc, err := ReadDocument(strings.NewReader(in))
	require.NoError(t, err)
	require.Len(t, doc.Parts, 1)
	require.Empty(t, doc.Parts[0].Information)

	var buf strings.Builder
	require.NoError(t, WriteDocument(&buf, doc))
	require.NotContains(t, buf.String(), "PART_INFORMATION")
}

// Mirror must actually mutate the parsed structure. Its payload slices hold
// values, so a range-loop mutates copies and silently drops the change; and
// segment Start points used to alias the previous segment's End, so inverting
// negated shared points twice and cancelled them out.
func TestMirrorMutatesAndRoundTrips(t *testing.T) {
	t.Parallel()

	in := numberedLines(docHead(), onePart(), []string{
		"MARKING_DATA", "MARKING_TYPE=PANEL", "MARKING_SIDE=TS", "END_OF_MARKING_DATA",
		"BURNING_DATA",
		"SHAPE=OUTER_CONTOUR",
		"START_OF_CONTOUR",
		"NO_OF_SEG=2",
		"START_U=0", "START_V=0",
		"AMP_U=0", "AMP_V=0", "AMP=0", "RADIUS=0", "SWEEP=0",
		"ORIGIN_U=0", "ORIGIN_V=0", "U=10", "V=0",
		"AMP_U=0", "AMP_V=0", "AMP=0", "RADIUS=10", "SWEEP=90",
		"ORIGIN_U=10", "ORIGIN_V=10", "U=20", "V=10",
		"END_OF_CONTOUR",
		"END_OF_BURNING_DATA",
		"LABELTEXT_DATA", "TEXT_TYPE=PANEL_NAME", "TEXT_POSITION_U=100.0", "TEXT=X",
		"END_OF_LABELTEXT_DATA",
		"LABELSYMBOL_DATA", "SYMBOL_TYPE=POSNO", "SYMBOL_POSITION_U=50.0", "END_OF_LABELSYMBOL_DATA",
		"EDGE_DATA", "NAME=S8", "START_U=1.0", "END_U=3.0", "END_OF_EDGE_DATA",
	})

	doc, err := ReadDocument(strings.NewReader(in))
	require.NoError(t, err)
	p := doc.Parts[0]

	// Segment points must not be shared.
	c := p.BurningData[0].Contour
	require.NotSame(t, c.Segments[0].End, c.Segments[1].Start,
		"segment Start must not alias the previous segment's End")

	p.Mirror()
	require.True(t, p.Mirrored)

	// Geometry negated exactly once.
	require.InDelta(t, -10.0, c.Segments[0].End.X, 1e-9)
	require.InDelta(t, -20.0, c.Segments[1].End.X, 1e-9)
	require.InDelta(t, -90.0, c.Segments[1].Sweep, 1e-9)

	// Value-slice scalars must survive the mutation.
	require.Equal(t, "OS", p.MarkingData[0].MarkingSide)
	require.InDelta(t, -100.0, p.LabelTexts[0].PositionU, 1e-9)
	require.InDelta(t, -50.0, p.LabelSymbols[0].PositionU, 1e-9)
	require.InDelta(t, -1.0, p.Edges[0].StartU, 1e-9)
	require.InDelta(t, -3.0, p.Edges[0].EndU, 1e-9)

	// And the mirrored part must survive a write/read cycle.
	var buf strings.Builder
	require.NoError(t, WriteDocument(&buf, doc))
	back, err := ReadDocument(strings.NewReader(buf.String()))
	require.NoError(t, err)
	require.True(t, back.Parts[0].Mirrored)
	require.InDelta(t, -10.0, back.Parts[0].BurningData[0].Contour.Segments[0].End.X, 1e-9)
	require.Equal(t, "OS", back.Parts[0].MarkingData[0].MarkingSide)
}

// IDLE_DATA and BUMP_DATA are contour-bearing payload siblings that were
// previously skipped. IDLE_DATA occurs 93,571 times in 769 corpus files.
func TestReadPartsIdleAndBumpData(t *testing.T) {
	t.Parallel()

	in := numberedLines(docHead(), onePart(), []string{
		"IDLE_DATA",
		"START_OF_CONTOUR",
		"NO_OF_SEG=1",
		"START_U=0", "START_V=0",
		"AMP_U=0", "AMP_V=0", "AMP=0", "RADIUS=0", "SWEEP=0",
		"ORIGIN_U=0", "ORIGIN_V=0", "U=474.68711", "V=293.26502",
		"END_OF_CONTOUR",
		"END_OF_IDLE_DATA",
		"BUMP_DATA",
		"SHAPE=BUMP_ALL",
		"START_OF_CONTOUR",
		"NO_OF_SEG=1",
		"START_U=54.0", "START_V=0",
		"AMP_U=10.40102", "AMP_V=-10.46637", "AMP=-14.75555",
		"RADIUS=50.0", "SWEEP=1.57706",
		"ORIGIN_U=4.0", "ORIGIN_V=0", "U=3.68685", "V=-49.99902",
		"END_OF_CONTOUR",
		"END_OF_BUMP_DATA",
	})

	first, err := ReadDocument(strings.NewReader(in))
	require.NoError(t, err)
	require.Len(t, first.Parts, 1)

	p := first.Parts[0]
	require.Len(t, p.IdleData, 1)
	require.NotNil(t, p.IdleData[0].Contour)
	require.Len(t, p.IdleData[0].Contour.Segments, 1)
	require.InDelta(t, 474.68711, p.IdleData[0].Contour.Segments[0].End.X, 1e-9)

	require.Len(t, p.BumpData, 1)
	require.Equal(t, "BUMP_ALL", p.BumpData[0].Shape)
	require.InDelta(t, 50.0, p.BumpData[0].Contour.Segments[0].Radius, 1e-9)

	// Both must survive a write/read cycle.
	var buf strings.Builder
	require.NoError(t, WriteDocument(&buf, first))
	second, err := ReadDocument(strings.NewReader(buf.String()))
	require.NoError(t, err)
	require.Len(t, second.Parts[0].IdleData, 1)
	require.Len(t, second.Parts[0].BumpData, 1)
	require.Equal(t, "BUMP_ALL", second.Parts[0].BumpData[0].Shape)
	require.InDelta(t, 474.68711, second.Parts[0].IdleData[0].Contour.Segments[0].End.X, 1e-9)
}

// TRANSFORMATION_DATA inside PART_DATA was previously skipped and lost.
func TestReadPartsTransformationData(t *testing.T) {
	t.Parallel()

	in := numberedLines(docHead(), []string{
		"PART_DATA", "NAME=P1",
		"TRANSFORMATION_DATA",
		"TRANSFORMATION_TYPE=REDUCED",
		"TRANSFORMATION_1=1.00000",
		"TRANSFORMATION_2=18100.00000",
		"TRANSFORMATION_3=0.00000",
		"TRANSFORMATION_4=0.00000",
		"TRANSFORMATION_5=0.00000",
		"TRANSFORMATION_6=0.00000",
		"END_OF_TRANSFORMATION_DATA",
		"END_OF_PART_DATA",
	})

	doc, err := ReadDocument(strings.NewReader(in))
	require.NoError(t, err)
	require.Len(t, doc.Parts, 1)
	tr := doc.Parts[0].Transformation
	require.NotNil(t, tr)
	require.Equal(t, "REDUCED", tr.Type)
	require.InDelta(t, 18100.0, tr.Transformation2, 1e-9)

	var buf strings.Builder
	require.NoError(t, WriteDocument(&buf, doc))
	back, err := ReadDocument(strings.NewReader(buf.String()))
	require.NoError(t, err)
	require.NotNil(t, back.Parts[0].Transformation)
	require.Equal(t, "REDUCED", back.Parts[0].Transformation.Type)
}

// The LISTED_PROFILE writer must preserve profiles and their end cuts.
func TestWriteProfilesRoundTrip(t *testing.T) {
	t.Parallel()

	in := numberedLines([]string{
		"TYPE_OF_GENERIC_FILE=LISTED_PROFILE",
		"VERSION=1.0",
		"USAGE=LIST",
		"COMMON_DATA",
		"SHAPE=PP",
		"START_OF_CONTOUR",
		"NO_OF_SEG=1",
		"START_U=0", "START_V=0",
		"AMP_U=0", "AMP_V=0", "AMP=0", "SWEEP=0", "U=180", "V=0",
		"END_OF_CONTOUR",
		"DIMENSION=180*11.0",
		"WEB_HEIGHT=180",
		"WEB_THICKNESS=11",
		"QUALITY=A40",
		"NO_OF_PROFS=1",
		"END_OF_COMMON_DATA",
		"PROFILE_DATA",
		"NAME=FL-TEST",
		"POSNO=1",
		"MLENGTH=8537.26",
		"TLENGTH=9000.00",
		"QUALITY=A40",
		"SHIP_NO=123",
		"SURFACE_TREATMENT=SA2.5",
		"MIRROR=NO",
		"END_OF_PROFILE_DATA",
		"LEFT_END",
		"ENDCUT_TYPE=21",
		// Bit pattern: leading zeros must survive, so this stays a string.
		"ENDCUT_MASK=00010110000",
		"ENDCUT_CODE=2131",
		"A=-",
		"R2=-",
		"V1=90",
		"END_OF_LEFT_END",
		"RIGHT_END",
		"ENDCUT_TYPE=27",
		"ENDCUT_MASK=00010110000",
		"ENDCUT_CODE=2131",
		"A=42",
		"R2=42",
		"V1=90",
		"END_OF_RIGHT_END",
	})

	first, err := ReadProfileDocument(strings.NewReader(in))
	require.NoError(t, err)
	require.Equal(t, "1.0", first.Version)
	require.Equal(t, "LIST", first.Usage)
	require.Len(t, first.Profiles, 1)
	require.NotNil(t, first.Common.Contour)
	require.Equal(t, "PP", first.Common.Shape)

	p := first.Profiles[0]
	require.Equal(t, "FL-TEST", p.Name)
	require.InDelta(t, 8537.26, p.MLength, 1e-6)
	require.Equal(t, "SA2.5", p.SurfaceTreatment)
	require.NotNil(t, p.LeftEnd)
	require.NotNil(t, p.RightEnd)
	require.Equal(t, 21, p.LeftEnd.EndcutType)
	require.Equal(t, "00010110000", p.LeftEnd.EndcutMask, "leading zeros must be preserved")
	// "-" is the sentinel meaning the parameter does not apply to this
	// ENDCUT_TYPE; the reader maps it to 0, so ENDCUT_TYPE stays authoritative.
	require.Zero(t, p.LeftEnd.A)
	require.Zero(t, p.LeftEnd.R2)
	require.Equal(t, 27, p.RightEnd.EndcutType)
	require.InDelta(t, 42.0, p.RightEnd.A, 1e-9)

	var buf strings.Builder
	require.NoError(t, WriteProfiles(&buf, first))
	require.Contains(t, buf.String(), "ENDCUT_MASK=00010110000")

	second, err := ReadProfileDocument(strings.NewReader(buf.String()))
	require.NoError(t, err)
	require.Len(t, second.Profiles, 1)
	require.Equal(t, "FL-TEST", second.Profiles[0].Name)
	require.Equal(t, "00010110000", second.Profiles[0].LeftEnd.EndcutMask)
	require.Equal(t, 21, second.Profiles[0].LeftEnd.EndcutType)
	require.Equal(t, 27, second.Profiles[0].RightEnd.EndcutType)
	require.InDelta(t, 8537.26, second.Profiles[0].MLength, 1e-6)
}
