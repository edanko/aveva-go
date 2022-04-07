package gen

import (
	"bufio"
	"strconv"
	"strings"
)

type BurningData struct {
	Shape            string
	StartEndInGap    string
	BevelDefined     string
	Direction        int
	Designation      string
	PartID           int
	NumberOfHeads    int
	GeometryValidFor int
	DistanceY1Y2     float64
	GeometryData     *GeometryData
	Contour          Contour
}

func readBurningData(s *bufio.Scanner) BurningData {
	var b BurningData
next:
	for s.Scan() {
		k, v, ok := strings.Cut(s.Text(), "=")

		switch k {
		case "GEOMETRY_DATA":
			gd := readGeometryData(s)
			b.GeometryData = &gd
			continue next

		case "START_OF_CONTOUR":
			b.Contour = readContour(s)
			continue next
		}

		if !ok {
			break
		}

		switch k {
		case "SHAPE":
			b.Shape = v
		case "START_END_IN_GAP":
			b.StartEndInGap = v
		case "BEVEL_DEFINED":
			b.BevelDefined = v
		case "DIRECTION":
			b.Direction, _ = strconv.Atoi(v)
		case "DESIGNATION":
			b.Designation = v
		case "PART_ID":
			b.PartID, _ = strconv.Atoi(v)
		case "NUMBER_OF_HEADS":
			b.NumberOfHeads, _ = strconv.Atoi(v)
		case "GEOMETRY_VALID_FOR":
			b.GeometryValidFor, _ = strconv.Atoi(v)
		case "DISTANCE_Y1_Y2":
			b.DistanceY1Y2, _ = strconv.ParseFloat(v, 64)
		}
	}
	return b
}
