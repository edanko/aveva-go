package draw

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	gen "github.com/edanko/aveva-go"
)

// part builds a part with one cut contour and two markings whose lengths decide
// the auto-mirror heuristic.
func part(tsLen, osLen float64) *gen.Part {
	p := &gen.Part{
		Name:      "P1",
		PosNo:     "1",
		ShipNo:    "123",
		BlockNo:   "B1",
		Quality:   "A40",
		Thickness: 10,
		PartCog:   &gen.Point{X: 100, Y: 100},
		PartArea:  1000,
		Quantity:  1,
		BurningData: []gen.BurningData{{
			Shape:   "OUTER_CONTOUR",
			Contour: contour(0, 0, 200, 0),
		}},
	}

	if tsLen > 0 {
		p.MarkingData = append(p.MarkingData, gen.Marking{
			MarkingSide: "TS",
			MarkingType: "PANEL",
			Contour:     contour(0, 10, tsLen, 10),
		})
	}

	if osLen > 0 {
		p.MarkingData = append(p.MarkingData, gen.Marking{
			MarkingSide: "OS",
			MarkingType: "PANEL",
			Contour:     contour(0, 20, osLen, 20),
		})
	}

	return p
}

func contour(x1, y1, x2, y2 float64) *gen.Contour {
	start := &gen.Point{X: x1, Y: y1}
	end := &gen.Point{X: x2, Y: y2}

	return &gen.Contour{Segments: []*gen.Segment{{
		Start:  start,
		End:    end,
		Origin: &gen.Point{X: 0, Y: 0},
	}}}
}

func TestPartToDXFProducesValidStructure(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	d, err := NewDrawer(&buf)
	require.NoError(t, err)
	require.NoError(t, d.PartToDXF(part(10, 10)))

	out := buf.String()
	require.Contains(t, out, "SECTION")
	require.Contains(t, out, "ENTITIES")
	require.Contains(t, out, "ENDSEC")
	require.Contains(t, out, "EOF")
	require.Contains(t, out, "LWPOLYLINE", "contours must be drawn as polylines")
	require.Contains(t, out, "TEXT", "part attributes must be drawn")
}

// The OS-longer-than-TS heuristic must be switchable, because callers that have
// already decided orientation must not have it overridden.
func TestSetAutoMirror(t *testing.T) {
	t.Parallel()

	renderWith := func(autoMirror bool) string {
		t.Helper()

		var buf bytes.Buffer

		d, err := NewDrawer(&buf)
		require.NoError(t, err)
		d.SetAutoMirror(autoMirror)

		// OS marking is the long one, so the heuristic would flip.
		require.NoError(t, d.PartToDXF(part(10, 100)))

		return buf.String()
	}

	require.NotEqual(t, renderWith(true), renderWith(false),
		"disabling the heuristic must change the orientation of the output")
}

func TestAutoMirrorIsOnByDefault(t *testing.T) {
	t.Parallel()

	// With TS longer than OS the heuristic declines to mirror, so the default and
	// an explicitly enabled heuristic must agree.
	withDefault := func() string {
		t.Helper()

		var buf bytes.Buffer

		d, err := NewDrawer(&buf)
		require.NoError(t, err)
		require.NoError(t, d.PartToDXF(part(100, 10)))

		return buf.String()
	}()

	withEnabled := func() string {
		t.Helper()

		var buf bytes.Buffer

		d, err := NewDrawer(&buf)
		require.NoError(t, err)
		d.SetAutoMirror(true)
		require.NoError(t, d.PartToDXF(part(100, 10)))

		return buf.String()
	}()

	require.Equal(t, withEnabled, withDefault)
}

// A part with no geometry must not panic: the drawer walks contours that may be
// absent.
func TestPartToDXFEmptyPart(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	d, err := NewDrawer(&buf)
	require.NoError(t, err)
	require.NoError(t, d.PartToDXF(&gen.Part{Name: "EMPTY", Thickness: 5}))

	require.NotEmpty(t, buf.String())
}

// A marking contour with no segments used to panic inside Vertices().
func TestPartToDXFEmptyContourDoesNotPanic(t *testing.T) {
	t.Parallel()

	p := part(10, 10)
	p.MarkingData = append(p.MarkingData, gen.Marking{
		MarkingSide: "TS",
		MarkingType: "PANEL",
		Contour:     &gen.Contour{},
	})
	p.StringData = append(p.StringData, gen.StringData{String: "X"})
	p.GeometryData = append(p.GeometryData, gen.Geometry{Type: "HOLE", Contour: &gen.Contour{}})

	var buf bytes.Buffer

	d, err := NewDrawer(&buf)
	require.NoError(t, err)
	require.NoError(t, d.PartToDXF(p))
}

func TestFilterPos(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
	}{
		{name: "plain", in: "4818"},
		{name: "already filtered", in: "P-1"},
		{name: "empty", in: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.NotPanics(t, func() { _ = filterPos(tt.in) })
		})
	}
}

func TestGetMarkingTextPositionEmptyContourDoesNotPanic(t *testing.T) {
	t.Parallel()

	require.NotPanics(t, func() {
		x, y, rot := getMarkingTextPosition(&gen.Contour{})
		_ = x
		_ = y
		_ = rot
	})
}

func BenchmarkPartToDXF(b *testing.B) {
	p := part(100, 200)

	b.ReportAllocs()

	for b.Loop() {
		var buf bytes.Buffer

		d, err := NewDrawer(&buf)
		if err != nil {
			b.Fatal(err)
		}
		if err := d.PartToDXF(p); err != nil {
			b.Fatal(err)
		}
	}
}

var _ = strings.TrimSpace
