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

func readStringData(s *bufio.Scanner) *StringData {
	sd := new(StringData)
	pos := new(Point)

	for s.Scan() {
		k, v, ok := strings.Cut(s.Text(), "=")

		if !ok {
			break
		}

		switch k {
		case "STRING_TYPE":
			sd.Type = v

		case "STRING_POSITION_U":
			pos.X, _ = strconv.ParseFloat(v, 64)

		case "STRING_POSITION_V":
			pos.Y, _ = strconv.ParseFloat(v, 64)

		case "STRING_ANGLE":
			sd.Angle, _ = strconv.ParseFloat(v, 64)

		case "STRING_HEIGHT":
			sd.Height, _ = strconv.ParseFloat(v, 64)

		case "STRING":
			sd.String = v
		}
	}

	sd.Pos = pos
	return sd
}
