package gen

import (
	"bufio"
	"strconv"
	"strings"
)

type GeneralData struct {
	Thickness float64
	Quality   string
	NoOfParts int
}

func readGeneralData(s *bufio.Scanner) *GeneralData {
	g := new(GeneralData)
	for s.Scan() {
		k, v, ok := strings.Cut(s.Text(), "=")

		if !ok {
			break
		}

		switch k {
		case "RAW_THICKNESS":
			g.Thickness, _ = strconv.ParseFloat(v, 64)

		case "QUALITY":
			g.Quality = v

		case "NO_OF_PARTS":
			g.NoOfParts, _ = strconv.Atoi(v)
		}

	}
	return g
}
