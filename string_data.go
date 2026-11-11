package gen

import (
	"bufio"
	"strconv"
	"strings"
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
		key, value, found := strings.Cut(s.Text(), "=")
		if !found {
			break
		}

		switch key {
		case "STRING_TYPE":
			stringData.Type = value

		case "STRING_POSITION_U":
			stringData.Pos.X, err = strconv.ParseFloat(value, 64)
			if err != nil {
				return StringData{}, err
			}

		case "STRING_POSITION_V":
			stringData.Pos.Y, err = strconv.ParseFloat(value, 64)
			if err != nil {
				return StringData{}, err
			}

		case "STRING_ANGLE":
			stringData.Angle, err = strconv.ParseFloat(value, 64)
			if err != nil {
				return StringData{}, err
			}

		case "STRING_HEIGHT":
			stringData.Height, err = strconv.ParseFloat(value, 64)
			if err != nil {
				return StringData{}, err
			}

		case "STRING":
			stringData.String = value
		}
	}

	return stringData, nil
}
