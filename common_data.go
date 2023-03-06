package gen

import (
	"bufio"
	"strconv"
	"strings"
)

type CommonData struct {
	Shape     string
	Dimension string
	Quality   string
	NoOfProfs uint64
}

func readCommonData(s *bufio.Scanner) *CommonData {
	cd := new(CommonData)

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

		case "NO_OF_PROFS":
			cd.NoOfProfs, _ = strconv.ParseUint(v, 10, 64)
		}
	}
	return cd
}
