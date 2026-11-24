package gen

import (
	"bufio"
)

type StringData struct {
	Type   string
	Pos    *Point
	Angle  float64
	Height float64
	String string
}

func readStringData(s *bufio.Scanner) (StringData, error) {
	stringData := StringData{
		Pos: &Point{},
	}

	var err error

	for s.Scan() {
		key, value, found := nextKeyValue(s)
		if !found {
			break
		}

		switch key {
		case "STRING_TYPE":
			stringData.Type = string(value)

		case "STRING_POSITION_U":
			stringData.Pos.X, err = atof(value)
			if err != nil {
				return StringData{}, err
			}

		case "STRING_POSITION_V":
			stringData.Pos.Y, err = atof(value)
			if err != nil {
				return StringData{}, err
			}

		case "STRING_ANGLE":
			stringData.Angle, err = atof(value)
			if err != nil {
				return StringData{}, err
			}

		case "STRING_HEIGHT":
			stringData.Height, err = atof(value)
			if err != nil {
				return StringData{}, err
			}

		case "STRING":
			stringData.String = string(value)
		}
	}

	return stringData, nil
}
