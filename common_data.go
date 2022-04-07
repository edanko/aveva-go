package gen

import (
	"bufio"
	"strings"
)

type CommonData struct {
	Shape     string
	Dimension string
	Quality   string
}

func readCommonData(s *bufio.Scanner) CommonData {
	var cd CommonData

next:
	for s.Scan() {
		k, v, ok := strings.Cut(s.Text(), "=")

		if k == "START_OF_CONTOUR" {
			for s.Scan() {
				switch s.Text() {
				case "END_OF_CONTOUR":
					continue next
				}
			}
		}

		if !ok {
			return cd
		}

		switch k {
		case "SHAPE":
			cd.Shape = v

		case "DIMENSION":
			cd.Dimension = v

		case "QUALITY":
			cd.Quality = v
		}
	}
	return cd
}
