package gen

import (
	"bufio"
	"strconv"
	"strings"
)

type ExcessData struct {
	ExcessValue float64
}

func readExcessData(s *bufio.Scanner) ExcessData {
	var e ExcessData

	for s.Scan() {
		k, v, ok := strings.Cut(s.Text(), "=")

		if !ok {
			break
		}

		switch k {
		case "EXCESS_VALUE":
			e.ExcessValue, _ = strconv.ParseFloat(v, 64)
		}
	}
	return e
}
