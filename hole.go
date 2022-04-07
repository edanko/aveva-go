package gen

import (
	"bufio"
	"strconv"
	"strings"
)

type Hole struct {
	Open              int
	DistOrigin        float64
	DistOriginToZeroP float64
	DistOriginV       float64
	Type              int
	Rotation          float64
	Mirror            int
	Name              string
	DistLeft          float64
	Contour           Contour
}

func readHolesNotchesCutouts(s *bufio.Scanner) Hole {
	var h Hole

next:
	for s.Scan() {
		k, v, ok := strings.Cut(s.Text(), "=")

		if k == "START_OF_CONTOUR" {
			h.Contour = readContour(s)
			continue next
		}

		if !ok {
			break
		}

		switch k {
		case "OPEN":
			h.Open, _ = strconv.Atoi(v)

		case "DIST_ORIGIN":
			h.DistOrigin, _ = strconv.ParseFloat(v, 64)

		case "DIST_ORIGIN_TO_ZEROP":
			h.DistOriginToZeroP, _ = strconv.ParseFloat(v, 64)

		case "DIST_ORIGIN_V":
			h.DistOriginV, _ = strconv.ParseFloat(v, 64)

		case "TYPE":
			h.Type, _ = strconv.Atoi(v)

		case "ROTATION":
			h.Rotation, _ = strconv.ParseFloat(v, 64)

		case "MIRROR":
			h.Mirror, _ = strconv.Atoi(v)

		case "NAME":
			h.Name = v

		case "DIST_LEFT":
			h.DistLeft, _ = strconv.ParseFloat(v, 64)
		}
	}
	return h
}
