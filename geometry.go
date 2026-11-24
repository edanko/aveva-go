package gen

import (
	"bufio"
)

type Geometry struct {
	Type       string
	RollAxisNo int
	Contour    *Contour
}

func readGeometry(s *bufio.Scanner) (Geometry, error) {
	geometry := Geometry{}

	for s.Scan() {
		key, value, found := nextKeyValue(s)

		if key == "START_OF_CONTOUR" {
			contour, err := readContour(s)
			if err != nil {
				return Geometry{}, err
			}
			geometry.Contour = contour
			continue
		}

		if !found {
			return geometry, nil
		}

		switch key {
		case "GEOMETRY_TYPE", "TYPE":
			geometry.Type = string(value)

		case "ROLL_AXIS_NO":
			rollAxisNo, err := atoi(value)
			if err != nil {
				return Geometry{}, err
			}
			geometry.RollAxisNo = rollAxisNo
		}
	}
	return geometry, nil
}
