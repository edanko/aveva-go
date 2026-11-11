package gen

import (
	"bufio"
	"strconv"
	"strings"
)

type Excess struct {
	Value float64
}

func readExcess(s *bufio.Scanner) (*Excess, error) {
	excess := &Excess{}

	for s.Scan() {
		key, value, found := strings.Cut(s.Text(), "=")
		if !found {
			break
		}

		switch key {
		case "EXCESS_VALUE":
			value, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			excess.Value = value
		}
	}
	return excess, nil
}
