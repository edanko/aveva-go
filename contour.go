package gen

import (
	"bufio"
	"strconv"
	"strings"
)

type Contour struct {
	ExcessData *ExcessData
	Segments   []*Segment
}

// for profile segment used only amp*, sweep, u and v
type Segment struct {
	AmpU      float64
	AmpV      float64
	Amp       float64
	Radius    float64
	Sweep     float64
	Origin    *Point
	Start     *Point
	End       *Point
	BevelData *BevelData
}

func readContour(s *bufio.Scanner) *Contour {
	con := new(Contour)

	var currentBevel *BevelData

	segmentFinished := false
	seg := new(Segment)
	origin := new(Point)
	start := new(Point)
	end := new(Point)

next:
	for s.Scan() {
		k, v, ok := strings.Cut(s.Text(), "=")

		switch k {
		case "BEVEL_DATA":
			currentBevel = readBevelData(s)
			continue next

		case "EXCESS_DATA":
			con.ExcessData = readExcessData(s)
			continue next

		case "AUXILIARY_DATA":
			for s.Scan() {
				switch s.Text() {
				case "END_OF_AUXILIARY_DATA":
					continue next
				}
			}
		}

		if !ok {
			break
		}

		switch k {
		case "START_U":
			start.X, _ = strconv.ParseFloat(v, 64)

		case "START_V":
			start.Y, _ = strconv.ParseFloat(v, 64)

		case "AMP_U":
			seg.AmpU, _ = strconv.ParseFloat(v, 64)

		case "AMP_V":
			seg.AmpV, _ = strconv.ParseFloat(v, 64)

		case "AMP":
			seg.Amp, _ = strconv.ParseFloat(v, 64)

		case "RADIUS":
			seg.Radius, _ = strconv.ParseFloat(v, 64)

		case "SWEEP":
			seg.Sweep, _ = strconv.ParseFloat(v, 64)

		case "ORIGIN_U":
			origin.X, _ = strconv.ParseFloat(v, 64)

		case "ORIGIN_V":
			origin.Y, _ = strconv.ParseFloat(v, 64)

		case "U":
			end.X, _ = strconv.ParseFloat(v, 64)

		case "V":
			end.Y, _ = strconv.ParseFloat(v, 64)
			segmentFinished = true
		}

		if segmentFinished {
			seg.Origin = origin
			seg.End = end

			if len(con.Segments) > 0 {
				seg.Start = con.Segments[len(con.Segments)-1].End
			} else {
				seg.Start = start
			}

			seg.BevelData = currentBevel
			con.Segments = append(con.Segments, seg)

			seg = new(Segment)
			origin = new(Point)
			end = new(Point)
			segmentFinished = false
		}
	}
	return con
}
