package gen

import (
	"bufio"
)

type Excess struct {
	Value float64
}

func readExcess(s *bufio.Scanner) (*Excess, error) {
	excess := &Excess{}

	for s.Scan() {
		key, value, found := nextKeyValue(s)
		if !found {
			break
		}

		switch key {
		case "EXCESS_VALUE":
			value, err := atof(value)
			if err != nil {
				return nil, err
			}
			excess.Value = value
		}
	}
	return excess, nil
}
