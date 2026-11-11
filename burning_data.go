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
	GeometryData     *Geometry
	Contour          *Contour
	HoleShipside     string
	Auxiliaries      []Auxiliary
}

func readBurningData(s *bufio.Scanner) (BurningData, error) {
	b := BurningData{}
next:
	for s.Scan() {
		key, value, found := strings.Cut(s.Text(), "=")

		switch key {
		case "GEOMETRY_DATA":
			geometryData, err := readGeometry(s)
			if err != nil {
				return BurningData{}, err
			}
			b.GeometryData = &geometryData
			continue next

		case "START_OF_CONTOUR":
			contour, err := readContour(s)
			if err != nil {
				return BurningData{}, err
			}
			b.Contour = contour
			continue next

		case "AUXILIARY_DATA":
			a, err := readAuxiliary(s)
			if err != nil {
				return BurningData{}, err
			}
			b.Auxiliaries = append(b.Auxiliaries, a)
			continue next
		}

		if !found {
			break
		}

		switch key {
		case "SHAPE":
			b.Shape = value
		case "START_END_IN_GAP":
			b.StartEndInGap = value
		case "BEVEL_DEFINED":
			b.BevelDefined = value
		case "DIRECTION":
			direction, err := strconv.Atoi(value)
			if err != nil {
				return BurningData{}, err
			}
			b.Direction = direction
		case "DESIGNATION":
			b.Designation = value
		case "PART_ID":
			partID, err := strconv.Atoi(value)
			if err != nil {
				return BurningData{}, err
			}
			b.PartID = partID
		case "NUMBER_OF_HEADS":
			numberOfHeads, err := strconv.Atoi(value)
			if err != nil {
				return BurningData{}, err
			}
			b.NumberOfHeads = numberOfHeads
		case "GEOMETRY_VALID_FOR":
			geometryValidFor, err := strconv.Atoi(value)
			if err != nil {
				return BurningData{}, err
			}
			b.GeometryValidFor = geometryValidFor
		case "DISTANCE_Y1_Y2":
			distanceY1Y2, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return BurningData{}, err
			}
			b.DistanceY1Y2 = distanceY1Y2

		case "HOLE_SHIPSIDE":
			b.HoleShipside = value
		}
	}
	return b, nil
}
