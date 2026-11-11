package gen

import (
	"bufio"
	"strconv"
	"strings"
)

type Hole struct {
	Open               int
	DistOrigin         float64
	DistOriginToZeroP  float64
	DistOriginV        float64
	Type               int
	Rotation           float64
	Mirror             int
	Name               string
	DistLeft           float64
	Contour            *Contour
	Kind               string
	PenetratingProfile *PenetratingProfile
	Drilling           *Drilling
	Slotted            *Slotted
}

func parseHole(s *bufio.Scanner) (*Hole, error) {
	hole := &Hole{}

next:
	for s.Scan() {
		key, value, found := strings.Cut(s.Text(), "=")

		if key == "START_OF_CONTOUR" {
			contour, err := readContour(s)
			if err != nil {
				return nil, err
			}
			hole.Contour = contour
			continue next
		}

		switch key {
		case "PENETRATING_PROFILE":
			pp, err := readPenetratingProfile(s)
			if err != nil {
				return nil, err
			}
			hole.Kind = "PENETRATING_PROFILE"
			hole.PenetratingProfile = pp
			continue next

		case "DRILLING":
			d, err := readDrilling(s)
			if err != nil {
				return nil, err
			}
			hole.Kind = "DRILLING"
			hole.Drilling = d
			continue next

		case "SLOTTED":
			sl, err := readSlotted(s)
			if err != nil {
				return nil, err
			}
			hole.Kind = "SLOTTED"
			hole.Slotted = sl
			continue next
		}

		if !found {
			break
		}

		switch key {
		case "OPEN":
			open, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			hole.Open = open

		case "DIST_ORIGIN":
			distOrigin, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			hole.DistOrigin = distOrigin

		case "DIST_ORIGIN_TO_ZEROP":
			distOriginToZeroP, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			hole.DistOriginToZeroP = distOriginToZeroP

		case "DIST_ORIGIN_V":
			distOriginV, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			hole.DistOriginV = distOriginV

		case "TYPE":
			kind, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			hole.Type = kind

		case "ROTATION":
			rotation, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			hole.Rotation = rotation

		case "MIRROR":
			mirror, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			hole.Mirror = mirror

		case "NAME":
			hole.Name = value

		case "DIST_LEFT":
			distLeft, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			hole.DistLeft = distLeft
		}
	}
	return hole, nil
}

// PenetratingProfile is the PENETRATING_PROFILE sub-block of a hole: a hole cut
// by another profile passing through the plate.
//
// Field set recovered from the mar.dll emission literals. No .gen sample in the
// 45,750-file corpus contains this block, so the parse is unvalidated.
type PenetratingProfile struct {
	Dim               float64
	IntersectingAngle float64
	InclinationAngle  float64
	MtrlPenetrating   string
}

func readPenetratingProfile(s *bufio.Scanner) (*PenetratingProfile, error) {
	p := &PenetratingProfile{}

	for s.Scan() {
		key, value, found := strings.Cut(s.Text(), "=")

		if key == "END_OF_PENETRATING_PROFILE" {
			return p, nil
		}

		if !found {
			return p, nil
		}

		var err error

		switch key {
		case "DIM":
			p.Dim, err = parseOptFloat(value)
		case "INTERSECTING_ANGLE":
			p.IntersectingAngle, err = parseOptFloat(value)
		case "INCLINATION_ANGLE":
			p.InclinationAngle, err = parseOptFloat(value)
		case "MTRL_PENETRATING":
			p.MtrlPenetrating = cleanStr(value)
		}

		if err != nil {
			return nil, err
		}
	}

	return p, nil
}

// Drilling is the DRILLING sub-block of a hole, optionally containing a
// THREADING block. Unvalidated: no corpus sample.
type Drilling struct {
	Threaded    bool
	ThreadStart float64
}

func readDrilling(s *bufio.Scanner) (*Drilling, error) {
	d := &Drilling{}

	for s.Scan() {
		key, _, found := strings.Cut(s.Text(), "=")

		switch key {
		case "THREADING":
			d.Threaded = true
			for s.Scan() {
				k, v, ok := strings.Cut(s.Text(), "=")
				if k == "END_OF_THREADING" {
					break
				}
				if !ok {
					return d, nil
				}
				if k == "THREAD_START" {
					start, err := parseOptFloat(v)
					if err != nil {
						return nil, err
					}
					d.ThreadStart = start
				}
			}
			continue

		case "END_OF_DRILLING":
			return d, nil
		}

		if !found {
			return d, nil
		}
	}

	return d, nil
}

// Slotted is the SLOTTED sub-block of a hole. Unvalidated: no corpus sample.
type Slotted struct {
	Length float64
	Width  float64
}

func readSlotted(s *bufio.Scanner) (*Slotted, error) {
	sl := &Slotted{}

	for s.Scan() {
		key, value, found := strings.Cut(s.Text(), "=")

		if key == "END_OF_SLOTTED" {
			return sl, nil
		}

		if !found {
			return sl, nil
		}

		var err error

		switch key {
		case "LENGTH":
			sl.Length, err = parseOptFloat(value)
		case "WIDTH":
			sl.Width, err = parseOptFloat(value)
		}

		if err != nil {
			return nil, err
		}
	}

	return sl, nil
}
