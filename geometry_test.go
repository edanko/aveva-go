package gen

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

const epsilon = 1e-9

func TestAngleConversion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		deg  float64
	}{
		{name: "zero", deg: 0},
		{name: "45", deg: 45},
		{name: "90", deg: 90},
		{name: "180", deg: 180},
		{name: "360", deg: 360},
		{name: "negative", deg: -135},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.InDelta(t, tt.deg, ToDeg(ToRad(tt.deg)), epsilon)
		})
	}
}

func TestAddAngle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		x, y   float64
		angle  float64
		length float64
		wantX  float64
		wantY  float64
	}{
		{name: "east", x: 0, y: 0, angle: 0, length: 10, wantX: 10, wantY: 0},
		{name: "north", x: 0, y: 0, angle: 90, length: 10, wantX: 0, wantY: 10},
		{name: "west", x: 0, y: 0, angle: 180, length: 10, wantX: -10, wantY: 0},
		{name: "south", x: 0, y: 0, angle: 270, length: 10, wantX: 0, wantY: -10},
		{name: "offset origin", x: 5, y: 5, angle: 0, length: 2, wantX: 7, wantY: 5},
		{name: "zero length", x: 3, y: 4, angle: 45, length: 0, wantX: 3, wantY: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotX, gotY := AddAngle(tt.x, tt.y, tt.angle, tt.length)
			require.InDelta(t, tt.wantX, gotX, epsilon)
			require.InDelta(t, tt.wantY, gotY, epsilon)
		})
	}
}

func TestAngle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		start Point
		end   Point
		want  float64
	}{
		{name: "east", start: Point{X: 0, Y: 0}, end: Point{X: 1, Y: 0}, want: 0},
		{name: "north", start: Point{X: 0, Y: 0}, end: Point{X: 0, Y: 1}, want: math.Pi / 2},
		{name: "diagonal", start: Point{X: 0, Y: 0}, end: Point{X: 1, Y: 1}, want: math.Pi / 4},
		{
			name:  "independent of magnitude",
			start: Point{X: 0, Y: 0},
			end:   Point{X: 100, Y: 100},
			want:  math.Pi / 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.InDelta(t, tt.want, Angle(&tt.start, &tt.end), epsilon)
		})
	}
}

func TestLineMidPoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		start, end Point
		wantX      float64
		wantY      float64
	}{
		{
			name:  "horizontal",
			start: Point{X: 0, Y: 0},
			end:   Point{X: 10, Y: 0},
			wantX: 5,
			wantY: 0,
		},
		{
			name:  "diagonal",
			start: Point{X: -2, Y: -4},
			end:   Point{X: 2, Y: 4},
			wantX: 0,
			wantY: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotX, gotY := LineMidPoint(&tt.start, &tt.end)
			require.InDelta(t, tt.wantX, gotX, epsilon)
			require.InDelta(t, tt.wantY, gotY, epsilon)
		})
	}
}

func TestArcMidPoint(t *testing.T) {
	t.Parallel()

	// Half circle of radius 1 about the origin, from (1,0) to (-1,0): the
	// midpoint lies at angle 90 degrees, i.e. (0,1).
	start := Point{X: 1, Y: 0}
	end := Point{X: -1, Y: 0}
	center := Point{X: 0, Y: 0}

	gotX, gotY := ArcMidPoint(&start, &end, &center, 1)
	require.InDelta(t, 0, gotX, epsilon)
	require.InDelta(t, 1, gotY, epsilon)

	// The midpoint must lie on the circle.
	require.InDelta(t, 1, math.Hypot(gotX, gotY), epsilon)
}

func TestSegmentMidPointDispatchesOnRadius(t *testing.T) {
	t.Parallel()

	start := Point{X: 0, Y: 0}
	end := Point{X: 10, Y: 0}

	// Radius 0 takes the straight-line branch.
	line := Segment{Start: &start, End: &end}
	x, y := SegmentMidPoint(&line)
	require.InDelta(t, 5, x, epsilon)
	require.InDelta(t, 0, y, epsilon)

	// Non-zero radius takes the arc branch, which uses Origin as the centre.
	center := Point{X: 5, Y: 0}
	arc := Segment{
		Start: &start, End: &end, Origin: &center, Radius: 5,
	}
	x, y = SegmentMidPoint(&arc)
	require.InDelta(t, 5, x, epsilon)
	require.InDelta(t, 5, y, epsilon, "arc midpoint must be offset from the chord")
}

func TestSegmentLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		seg  Segment
		want float64
	}{
		{
			name: "straight line",
			seg:  Segment{Start: &Point{X: 0, Y: 0}, End: &Point{X: 3, Y: 4}},
			want: 5,
		},
		{
			name: "quarter arc",
			seg:  Segment{Radius: 2, Sweep: math.Pi / 2},
			want: 2 * math.Pi / 2,
		},
		{
			name: "arc uses absolute sweep",
			seg:  Segment{Radius: 2, Sweep: -math.Pi / 2},
			want: 2 * math.Pi / 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.InDelta(t, tt.want, tt.seg.Length(), epsilon)
		})
	}
}

func TestSegmentBulge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		seg  Segment
		want float64
	}{
		{name: "line has no bulge", seg: Segment{}, want: 0},
		{name: "quarter arc", seg: Segment{Radius: 1, Sweep: math.Pi / 2}, want: math.Tan(math.Pi / 8)},
		{name: "negative amp negates", seg: Segment{Radius: 1, Sweep: math.Pi / 2, Amp: -1}, want: -math.Tan(math.Pi / 8)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.InDelta(t, tt.want, tt.seg.Bulge(), epsilon)
		})
	}
}

func TestContourVerticesEmptyIsNil(t *testing.T) {
	t.Parallel()

	// Regression: these used to panic on a contour with no segments.
	c := Contour{}
	require.Nil(t, c.Vertices())
	require.Nil(t, c.Bulges())
	require.False(t, c.IsClockwise())

	// invert must tolerate an empty and a nil contour.
	c.invert()

	var nilContour *Contour
	nilContour.invert()
}

func TestContourVertices(t *testing.T) {
	t.Parallel()

	start := Point{X: 0, Y: 0}
	mid := Point{X: 10, Y: 0}
	end := Point{X: 10, Y: 10}

	c := Contour{Segments: []*Segment{
		{Start: &start, End: &mid},
		{Start: &mid, End: &end},
	}}

	vertices := c.Vertices()
	require.Len(t, vertices, 3, "one vertex per segment plus the start point")

	// Vertice() returns [x, y, z].
	require.Equal(t, []float64{0, 0, 0}, vertices[0])
	require.Equal(t, []float64{10, 0, 0}, vertices[1])
	require.Equal(t, []float64{10, 10, 0}, vertices[2])
}

func TestContourBulgesLengthMatchesVertices(t *testing.T) {
	t.Parallel()

	start := Point{X: 0, Y: 0}
	mid := Point{X: 10, Y: 0}
	end := Point{X: 10, Y: 10}

	c := Contour{Segments: []*Segment{
		{Start: &start, End: &mid},
		{Start: &mid, End: &end, Radius: 1, Sweep: math.Pi / 2},
	}}

	// A closed LWPOLYLINE needs one bulge per vertex.
	require.Len(t, c.Bulges(), len(c.Vertices()))

	// The trailing bulge closes the loop and must be zero.
	require.Zero(t, c.Bulges()[len(c.Bulges())-1])
}
