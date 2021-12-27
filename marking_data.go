package gen

import (
	"bufio"
	"strconv"
	"strings"
)

type MarkingData struct {
	MarkingSide             string
	MarkingType             string
	MarkingName             string
	MarkingUserType         int
	MarkingAttri            string
	MarkingAttriSb          string
	MarkingShipside         string
	MarkingAssemblyLow      string
	MarkingProfileThickness float64
	MarkingWeld             float64
	MarkingDirection        int
	NumberOfHeads           int
	GeometryValidFor        int
	DistanceY1Y2            float64
	MtrlSideU               float64
	MtrlSideV               float64
	InclinationAngle        float64
	Contour                 *Contour
}

func readMarkingData(s *bufio.Scanner) *MarkingData {
	m := new(MarkingData)
next:
	for s.Scan() {
		k, v, ok := strings.Cut(s.Text(), "=")

		if k == "START_OF_CONTOUR" {
			m.Contour = readContour(s)
			continue next
		}

		if !ok {
			break
		}

		switch k {
		case "MARKING_SIDE":
			m.MarkingSide = v
		case "MARKING_TYPE":
			m.MarkingType = v
		case "MARKING_NAME":
			m.MarkingName = v
		case "MARKING_USER_TYPE":
			m.MarkingUserType, _ = strconv.Atoi(v)
		case "MARKING_ATTRI":
			m.MarkingAttri = v
		case "MARKING_ATTRI_SB":
			m.MarkingAttriSb = v
		case "MARKING_SHIPSIDE":
			m.MarkingShipside = v
		case "MARKING_ASSEMBLY_LOW":
			m.MarkingAssemblyLow = v
		case "MARKING_PROFILE_THICKNESS":
			m.MarkingProfileThickness, _ = strconv.ParseFloat(v, 64)
		case "MARKING_WELD":
			m.MarkingWeld, _ = strconv.ParseFloat(v, 64)
		case "MARKING_DIRECTION":
			m.MarkingDirection, _ = strconv.Atoi(v)
		case "NUMBER_OF_HEADS":
			m.NumberOfHeads, _ = strconv.Atoi(v)
		case "GEOMETRY_VALID_FOR":
			m.GeometryValidFor, _ = strconv.Atoi(v)
		case "DISTANCE_Y1_Y2":
			m.DistanceY1Y2, _ = strconv.ParseFloat(v, 64)
		case "MTRL_SIDE_U":
			m.MtrlSideU, _ = strconv.ParseFloat(v, 64)
		case "MTRL_SIDE_V":
			m.MtrlSideV, _ = strconv.ParseFloat(v, 64)
		case "INCLINATION_ANGLE":
			m.InclinationAngle, _ = strconv.ParseFloat(v, 64)
		}

	}
	return m
}
