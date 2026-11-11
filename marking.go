package gen

import (
	"bufio"
	"strconv"
	"strings"
)

type Marking struct {
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
	Comment                 string
}

func readMarking(s *bufio.Scanner) (Marking, error) {
	marking := Marking{}

	for s.Scan() {
		key, value, found := strings.Cut(s.Text(), "=")

		if key == "START_OF_CONTOUR" {
			contour, err := readContour(s)
			if err != nil {
				return Marking{}, err
			}
			marking.Contour = contour
			continue
		}

		if !found {
			break
		}

		switch key {
		case "MARKING_SIDE":
			marking.MarkingSide = value

		case "MARKING_TYPE":
			marking.MarkingType = value

		case "MARKING_NAME":
			marking.MarkingName = value

		case "MARKING_USER_TYPE":
			markingType, err := strconv.Atoi(value)
			if err != nil {
				return Marking{}, err
			}
			marking.MarkingUserType = markingType

		case "MARKING_ATTRI":
			marking.MarkingAttri = value

		case "MARKING_ATTRI_SB":
			marking.MarkingAttriSb = value

		case "MARKING_SHIPSIDE":
			marking.MarkingShipside = value

		case "MARKING_ASSEMBLY_LOW":
			marking.MarkingAssemblyLow = value

		case "MARKING_PROFILE_THICKNESS":
			markingProfileThickness, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return Marking{}, err
			}
			marking.MarkingProfileThickness = markingProfileThickness

		case "MARKING_WELD":
			markingWeld, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return Marking{}, err
			}
			marking.MarkingWeld = markingWeld

		case "MARKING_DIRECTION":
			markingDirection, err := strconv.Atoi(value)
			if err != nil {
				return Marking{}, err
			}
			marking.MarkingDirection = markingDirection

		case "NUMBER_OF_HEADS":
			numberOfHeads, err := strconv.Atoi(value)
			if err != nil {
				return Marking{}, err
			}
			marking.NumberOfHeads = numberOfHeads

		case "GEOMETRY_VALID_FOR":
			geometryValidFor, err := strconv.Atoi(value)
			if err != nil {
				return Marking{}, err
			}
			marking.GeometryValidFor = geometryValidFor

		case "DISTANCE_Y1_Y2":
			distanceY1Y2, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return Marking{}, err
			}
			marking.DistanceY1Y2 = distanceY1Y2

		case "MTRL_SIDE_U":
			mtrlSideU, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return Marking{}, err
			}
			marking.MtrlSideU = mtrlSideU

		case "MTRL_SIDE_V":
			mtrlSideV, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return Marking{}, err
			}
			marking.MtrlSideV = mtrlSideV

		case "INCLINATION_ANGLE":
			inclinationAngle, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return Marking{}, err
			}
			marking.InclinationAngle = inclinationAngle

		case "COMMENT":
			marking.Comment = value
		}
	}
	return marking, nil
}
