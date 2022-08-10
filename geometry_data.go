package gen

import (
	"bufio"
	"strconv"
	"strings"
)

type GeometryData struct {
	Type       string
	RollAxisNo int
	Contour    *Contour
}

func readGeometryData(s *bufio.Scanner) *GeometryData {
	gd := new(GeometryData)

next:
	for s.Scan() {
		k, v, ok := strings.Cut(s.Text(), "=")

		if k == "START_OF_CONTOUR" {
			gd.Contour = readContour(s)
			continue next
		}

		if !ok {
			return gd
		}

		switch k {
		case "GEOMETRY_TYPE", "TYPE":
			gd.Type = v

		case "ROLL_AXIS_NO":
			gd.RollAxisNo, _ = strconv.Atoi(v)
		}
	}
	return gd
}
