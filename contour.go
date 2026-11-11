package gen

import (
	"bufio"
	"math"
	"strconv"
	"strings"
)

type Contour struct {
	ExcessData  *Excess
	Segments    []*Segment
	Auxiliaries []Auxiliary
}

// func (c Contour) IsClosed() bool {
// 	if len(c.Segments) == 0 {
// 		return false
// 	}

// 	return c.Segments[0].Start == c.Segments[len(c.Segments)-1].End
// }

func (c Contour) Vertices() [][]float64 {
	if len(c.Segments) == 0 {
		return nil
	}

	vertices := make([][]float64, len(c.Segments)+1)
	vertices[0] = c.Segments[0].Start.Vertice()

	for i, s := range c.Segments {
		vertices[i+1] = s.End.Vertice()
	}

	return vertices
}

func (c Contour) Bulges() []float64 {
	if len(c.Segments) == 0 {
		return nil
	}

	bulges := make([]float64, 0, len(c.Segments))
	b := c.Segments[0].Bulge()
	bulges = append(bulges, b)

	for i := 0; i < len(c.Segments); i++ {
		b = 0
		if i < len(c.Segments)-1 {
			b = c.Segments[i+1].Bulge()
		}

		bulges = append(bulges, b)
	}
	return bulges
}

func (c Contour) IsClockwise() bool {
	if len(c.Segments) == 0 {
		return false
	}

	sum := 0.0

	for _, s := range c.Segments {
		if s.Origin == nil || s.End == nil {
			continue
		}

		sum += (s.End.X - s.Origin.X) * (s.End.Y + s.Origin.Y)
	}

	return sum > 0
}

// for profile segment used only amp*, sweep, u and v.
type Segment struct {
	AmpU   float64
	AmpV   float64
	Amp    float64
	Radius float64
	Sweep  float64
	Origin *Point
	Start  *Point
	End    *Point
	Bevel  *Bevel
}

func (s Segment) Length() float64 {
	if s.Radius != 0 {
		return s.Radius * math.Abs(s.Sweep)
	}
	return s.Start.DistanceTo(s.End)
}

func (s Segment) Bulge() float64 {
	if s.Radius != 0 {
		bulge := math.Tan(s.Sweep / 4)

		if s.Amp < 0 {
			return -bulge
		}
		return bulge
	}

	return 0
}

func (c *Contour) invert() {
	if c == nil {
		return
	}

	for _, s := range c.Segments {
		if s == nil {
			continue
		}

		s.Sweep = -s.Sweep

		if s.Origin != nil {
			s.Origin.X = -s.Origin.X
		}
		if s.Start != nil {
			s.Start.X = -s.Start.X
		}
		if s.End != nil {
			s.End.X = -s.End.X
		}

		if s.Bevel != nil && s.Bevel.BevelCode != 0 {
			b := s.Bevel

			b.BevelCode = -b.BevelCode
			b.AngleTS, b.AngleOS = b.AngleOS, b.AngleTS
			b.Angle2TS, b.Angle2OS = b.Angle2OS, b.Angle2TS
			b.DepthTS, b.DepthOS = b.DepthOS, b.DepthTS
			b.ChamferWidthTS, b.ChamferWidthOS = b.ChamferWidthOS, b.ChamferWidthTS
			b.Angle2Wts, b.Angle2Wos = b.Angle2Wos, b.Angle2Wts
			b.ChamferHeightTS, b.ChamferHeightOS = b.ChamferHeightOS, b.ChamferHeightTS
			s.Bevel = b
		}
	}
}

func readContour(s *bufio.Scanner) (*Contour, error) {
	contour := &Contour{}

	var currentBevel *Bevel

	segmentFinished := false
	seg := &Segment{}
	origin := &Point{}
	start := &Point{}
	end := &Point{}

next:
	for s.Scan() {
		key, value, found := strings.Cut(s.Text(), "=")

		switch key {
		case "BEVEL_DATA":
			var err error
			currentBevel, err = readBevel(s)
			if err != nil {
				return nil, err
			}
			continue next

		case "EXCESS_DATA":
			excess, err := readExcess(s)
			if err != nil {
				return nil, err
			}
			contour.ExcessData = excess
			continue next

		case "AUXILIARY_DATA":
			aux, err := readAuxiliary(s)
			if err != nil {
				return nil, err
			}
			contour.Auxiliaries = append(contour.Auxiliaries, aux)
			continue next
		}

		if !found {
			break
		}

		switch key {
		case "START_U":
			startX, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			start.X = startX

		case "START_V":
			startY, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			start.Y = startY

		case "AMP_U":
			ampU, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			seg.AmpU = ampU

		case "AMP_V":
			ampV, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			seg.AmpV = ampV

		case "AMP":
			amp, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			seg.Amp = amp

		case "RADIUS":
			radius, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			seg.Radius = radius

		case "SWEEP":
			sweep, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			seg.Sweep = sweep

		case "ORIGIN_U":
			originX, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			origin.X = originX

		case "ORIGIN_V":
			originY, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			origin.Y = originY

		case "U":
			x, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			end.X = x

		case "V":
			y, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			end.Y = y
			segmentFinished = true
		}

		if segmentFinished {
			seg.Origin = origin
			seg.End = end

			// Give the segment its own points instead of aliasing the previous
			// segment's End. Aliasing made mutating one segment silently change
			// its neighbour, and made Contour.invert negate shared points twice.
			src := start
			if len(contour.Segments) > 0 {
				src = contour.Segments[len(contour.Segments)-1].End
			}
			if src != nil {
				copied := *src
				seg.Start = &copied
			}

			seg.Bevel = currentBevel
			contour.Segments = append(contour.Segments, seg)

			seg = &Segment{}
			origin = &Point{}
			end = &Point{}
			segmentFinished = false
		}
	}
	return contour, nil
}
